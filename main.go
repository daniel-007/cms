package main

import (
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	router := gin.Default()
	v1 := router.Group("/v1.0")
	{
		v1.POST("/key", apiKeyGenHandler)
		v1.POST("/token", tokenGenHandler)
		v1.POST("/cms", msgSendHandler)
		v1.GET("/cms", msgRecvHandler)
	}
	// router.Run(":8080")
	gracehttp.Serve(
		&http.Server{
			Addr:         ":8080",
			Handler:      router,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		})
}
