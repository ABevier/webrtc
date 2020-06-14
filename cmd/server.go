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
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", signaling.ServeWebSocket)

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("Failed to start http server: ", err)
	}
}
