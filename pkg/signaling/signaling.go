package signaling

import (
	"bytes"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var newline = []byte{'\n'}
var space = []byte{' '}

var m = &sync.Mutex{}
var conns = []*websocket.Conn{}

func ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade ws: ", err)
		return
	}

	m.Lock()
	defer m.Unlock()
	conns = append(conns, conn)

	log.Println("connected")
	go read(conn)
}

func read(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("error on read:", err)
			return
		}

		//TODO: jank
		message = bytes.Replace(message, newline, space, -1)
		message = bytes.TrimSpace(message)

		log.Printf("Sending: %v", message)

		m.Lock()
		for _, destConn := range conns {
			write(destConn, message)
		}
		m.Unlock()
	}
}

func write(conn *websocket.Conn, message []byte) {
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Println("Couldn't get next writer:", err)
		return
	}

	w.Write(message)

	if err := w.Close(); err != nil {
		log.Println("Couldn't close writer:", err)
	}
}
