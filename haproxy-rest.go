package main

import (
	"github.com/gin-gonic/gin"
	"flag"
	"io/ioutil"
//	"fmt"
//	"encoding/json"
//	"sync"
	"github.com/jcelliott/lumber"
	"os"
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
	logFile, _ = lumber.NewFileLogger("/tmp/haproxy-rest.log", lumber.INFO, lumber.ROTATE, 1000, 3, 100)
	logConsole = lumber.NewConsoleLogger(lumber.INFO)
	log = lumber.NewMultiLogger()
)

func main() {


	log.Prefix("Haproxy-rest")

	log.AddLoggers(logFile, logConsole)


	// implicit -h prints out help messages
	port            := flag.Int("port",10001, "Port/IP to use for the REST interface. Overrides $PORT0 env variable")
	configFile	 	:= flag.String("configFile", "resources/haproxy_new.cfg", "Location of the target HAproxy config file")
	templateFile  	:= flag.String("template", "resources/haproxy_cfg.template", "Template file to build HAproxy config")
	binary        	:= flag.String("binary", "/usr/local/bin/haproxy", "Path to the HAproxy binary")
	kafkaHost       := flag.String("kafkaHost", "localhost", "The hostname or ip address of the Kafka host")
	kafkaPort       := flag.Int("kafkaPort",9092, "The port of the Kafka host")
	pidFile       	:= flag.String("pidFile", "resources/haproxy-private.pid", "Location of the HAproxy PID file")

	flag.Parse()


	// load a persistent config, if any...
	SetFileName("resources/persistent_config.json")
	ConfigObj, err := GetConfigFromDisk()
	if err != nil {

		log.Warn("Unable to load persistent config from disk")
		log.Warn("Loading example config")

		// set config temporarily to example config
		SetFileName("resources/config_example.json")
		ConfigObj, err = GetConfigFromDisk()
		SetFileName("resources/persistent_config.json")

	} else {

		log.Info("Loading persistent configuration from disk...")

	}

	if ConfigObj.PidFile != *pidFile {
		ConfigObj.PidFile = *pidFile
	}

	//Create and empty pid file on the specified location, if not already there
	if _, err := os.Stat(*pidFile); err == nil {
		log.Info("Pid file exists, proceeding with startup...")
	} else {
		emptyPid := []byte("")
		ioutil.WriteFile(*pidFile, emptyPid, 0644)
	}


	err = RenderConfig(*configFile, *templateFile, ConfigObj)
	if err != nil {
		log.Error("Error rendering config file")
		return
	} else {
		err = Reload(*binary, *configFile, *pidFile)
		if err != nil {
			log.Error("Error reloading the HAproxy configuration")
			return
		}

	}

	// Setup Kafka producer
	setUpProducer(*kafkaHost, *kafkaPort)

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

		v1.POST("/backend/:name/:server/weight/:weight", func(c *gin.Context){
				backend := c.Params.ByName("name")
				server :=  c.Params.ByName("server")
				weight,_  := strconv.Atoi(c.Params.ByName("weight"))
				status, err := SetWeight(backend, server, weight)

				// check on runtime errors
				if err != nil {
					c.String(500, err.Error())
				} else {

					switch status {
					case "No such server.\n\n":
						c.String(404, status)
					case "No such backend.\n\n":
						c.String(404, status)
					default:
						//update the config object with the new weight

						err = UpdateWeightInConfig(backend, server, weight, ConfigObj)

						c.String(200,"Ok")
					}
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
		v1.GET("/stats/server", func(c *gin.Context) {
				status, err := GetStats("server")
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

	// get the Mesos port to listen on
	if *port == 10001 {
		envPort := os.Getenv("PORT0")
		if envPort != "" { *port, err = strconv.Atoi(envPort) }
	}
	// Listen and server on port
	r.Run("0.0.0.0:" + strconv.Itoa(*port))
}
