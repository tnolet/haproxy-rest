package main

import (
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	}
}

func main() {

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
				status, err := GetStats()
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.JSON(200, status)
				}

			})
		v1.GET("/stats/backend", func(c *gin.Context) {
				status, err := GetStatsBackend()
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.JSON(200, status)
				}

			})


		v1.GET("/stats/frontend", func(c *gin.Context) {
				status, err := GetStatsFrontend()
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.JSON(200, status)
				}

			})


		// get info on running process
		v1.GET("/info", func(c *gin.Context) {
				status, err := GetInfo()
				if err != nil {
					c.String(500, err.Error())
				} else {
					c.String(200, status)
				}

			})
	}


	// Listen and server on 0.0.0.0:10001
	r.Run(":10001")
}


