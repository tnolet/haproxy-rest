package main

import (
	"github.com/gin-gonic/gin"
	"flag"
	"io/ioutil"
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
	lbConfigFile	:= flag.String("lbConfigFile", "resources/haproxy_new.cfg", "Location of the target HAproxy config file")
	lbTemplateFile  := flag.String("lbTemplate", "resources/haproxy_cfg.template", "Template file to build HAproxy load balancer config")
	proxyTemplateFile := flag.String("proxyYemplate", "resources/haproxy_localproxy_cfg.template", "Template file to build HAproxy local proxy config")
	proxyConfigFile  := flag.String("proxyConfigFile", "resources/haproxy_localproxy_new.cfg", "Location of the target HAproxy localproxy config")
	binary        	:= flag.String("binary", "/usr/local/bin/haproxy", "Path to the HAproxy binary")
	kafkaSwitch		:= flag.String("kafkaSwitch","off", "Switch whether to enable Kafka streaming")
	kafkaHost       := flag.String("kafkaHost", "localhost", "The hostname or ip address of the Kafka host")
	kafkaPort       := flag.Int("kafkaPort",9092, "The port of the Kafka host")
	mode			:= flag.String("mode","loadbalancer", "Switch for \"loadbalancer\" or \"localproxy\" mode")
    zooConString    := flag.String("zooConString", "localhost", "A zookeeper ensemble connection string")
	pidFile       	:= flag.String("pidFile", "resources/haproxy-private.pid", "Location of the HAproxy PID file")

	flag.Parse()


	// set intial values based on the mode chosen

	var perstConfFile = ""
	var exampleFile = ""

	if *mode == "loadbalancer" {

		log.Info(" ==> Starting in Load Balancer mode <==")

		perstConfFile = "resources/persistent_lb_config.json"
		exampleFile   = "resources/config_example.json"
		SetTemplateFileName(*lbTemplateFile)
		SetConfigFileName(*lbConfigFile)


	} else if * mode == "localproxy" {

		log.Info(" ==> Starting in Local Proxy mode <==")

		perstConfFile = "resources/persistent_localproxy_config.json"
		exampleFile   = "resources/config_localproxy_example.json"
		SetTemplateFileName(*proxyTemplateFile)
		SetConfigFileName(*proxyConfigFile)

		zkClient := ZookeeperClient{*zooConString}

		log.Info("Connecting to Zookeeper ensemble on " + *zooConString)
		zkConnection := zkClient.connect()
		defer zkConnection.Close()

		snapshots, errors := zkClient.watchServiceTypes(zkConnection,"/magnetic")
		zkClient.watchHandler(zkConnection, snapshots, errors)


	} else {

		log.Error("No correct mode chosen. Please choose either \"loadbalancer\" or \"localproxy\"")
		os.Exit(1)
	}

	// load a persistent config, if any...
	SetFileName(perstConfFile)

	ConfigObj, err := GetConfigFromDisk()
	if err != nil {

		log.Warn("Unable to load persistent config from disk")
		log.Warn("Loading example config")

		// set config temporarily to example config
		SetFileName(exampleFile)
		ConfigObj, err = GetConfigFromDisk()
		SetFileName(perstConfFile)

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

	SetPidFileName(*pidFile)
	SetBinaryFileName(*binary)

	err = RenderConfig(ConfigObj)
	if err != nil {
		log.Error("Error rendering config file")
		return
	} else {
		err = Reload()
		if err != nil {
			log.Error("Error reloading the HAproxy configuration")
			return
		}

	}

	if *kafkaSwitch == "on" {

		// Setup Kafka producer
		setUpProducer(*kafkaHost, *kafkaPort)

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
				err = RenderConfig(ConfigObj)
				if err != nil {
					c.String(500, "Error rendering config file")
					return
				} else {
					err = Reload()
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
