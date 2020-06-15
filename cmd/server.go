package main

import (
	"log"
	"net/http"

	"github.com/abevier/webrtc/pkg/signaling"
)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	//TODO: validate

	http.ServeFile(w, r, "web/static/index.html")
}

func main() {
	hub := signaling.NewHub()

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		signaling.ServeWebSocket(hub, w, r)
	})

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("Failed to start http server: ", err)
	}
}
