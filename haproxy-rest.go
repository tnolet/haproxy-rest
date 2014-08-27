package main

import (
	"github.com/gin-gonic/gin"
	"log/syslog"
	"flag"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"sync"
	"strconv"
)

// override the standard Gin-Gonic middleware to add the CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
}


// set some globally used vars
var (
	ConfigObj *Config
	log, _  = syslog.New(syslog.LOG_DEBUG, "haproxy-rest")
)

func main() {

	port            := flag.Int("port",10001, "Port/IP to use for the REST interface")
	configFile	 	:= flag.String("configFile", "resources/haproxy_new.cfg", "Location of the target HAproxy config file")
	templateFile  	:= flag.String("template", "resources/haproxy_cfg.template", "Template file to build HAproxy config")
	binary        	:= flag.String("binary", "/usr/local/bin/haproxy", "Path to the HAproxy binary")
	pidFile       	:= flag.String("pidFile", "resources/haproxy-private.pid", "Location of the HAproxy PID file")
	flag.Parse()

	s, err := ioutil.ReadFile("resources/config_example.json")
	if err != nil {
		panic("Cannot find config file at location")
	}
	err = json.Unmarshal(s, &ConfigObj)
	if err != nil {
		fmt.Println("Error parsing JSON")
	} else {
		ConfigObj.Mutex = new(sync.RWMutex)
	}


	if ConfigObj.PidFile != *pidFile {
		ConfigObj.PidFile = *pidFile
	}


	err = RenderConfig(*configFile, *templateFile, ConfigObj)
	if err != nil {
		fmt.Println("Error rendering config file")
		return
	} else {
		err = Reload(*binary, *configFile, *pidFile)
		if err != nil {
			fmt.Println("Error reloading the HAproxy configuration")
			return
		}

	}


	log.Info("Starting REST server")
	// initialize the web stack
	r := gin.New()
	// Global middlewares
	r.Use(CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())


	v1 := r.Group("/v1")

	{
		/*

			Backend Actions

		 */

		v1.GET("/backend/:name/:server/weight/:weight", func(c *gin.Context){
				backend := c.Params.ByName("name")
				server :=  c.Params.ByName("server")
				weight  := c.Params.ByName("weight")
				status, err := SetWeight(backend, server, weight)
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.String(200, status)
				}

			})

		/*

			Stats Actions

		 */

		// get standard stats output from haproxy
		v1.GET("/stats", func(c *gin.Context) {
				status, err := GetStats("all")
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.JSON(200, status)
				}

			})
		v1.GET("/stats/backend", func(c *gin.Context) {
				status, err := GetStats("backend")
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.JSON(200, status)
				}

			})


		v1.GET("/stats/frontend", func(c *gin.Context) {
				status, err := GetStats("frontend")
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.JSON(200, status)
				}

			})

		/*

			Full Config Actions

		 */

		// get config file
		v1.GET("/config", func(c *gin.Context){
				c.JSON(200, ConfigObj)
		})

		// set config file

		v1.POST("/config", func(c *gin.Context){

				c.Bind(&ConfigObj)
				err = RenderConfig(*configFile, *templateFile, ConfigObj)
				if err != nil {
					c.String(500, "Error rendering config file")
					return
				} else {
					err = Reload(*binary, *configFile, *pidFile)
					if err != nil {
						c.String(500, "Error reloading the HAproxy configuration")
						return
					} else {
						c.String(200, "Ok")
					}

				}
		})

		/*

			Info

		 */

		// get info on running process
		v1.GET("/info", func(c *gin.Context) {
				status, err := GetInfo()
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.JSON(200, status)
				}

			})
	}

	// Listen and server on port
	r.Run("0.0.0.0:" + strconv.Itoa(*port))
}
