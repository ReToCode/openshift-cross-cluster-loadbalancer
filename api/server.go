package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func RunAPI() {
	glog.Infof("Starting api server on 8080")

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT"},
		AllowHeaders:     []string{"Origin"},
	}))

	// Todo: Add authentication

	router.Run()
}
