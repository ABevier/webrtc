package main

import (
	"log"

	"github.com/abevier/webrtc/pkg/signaling"
	"github.com/gin-gonic/gin"
)

func main() {
	hub := signaling.NewHub()

	router := gin.Default()

	router.StaticFile("/", "web/static/index.html")

	router.GET("/ws/:room", func(c *gin.Context) {
		room := c.Param("room")
		signaling.ServeWebSocket(hub, room, c.Writer, c.Request)
	})

	err := router.Run()
	if err != nil {
		log.Fatal("Failed to start http server: ", err)
	}
}
