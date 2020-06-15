package signaling

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var newline = []byte{'\n'}
var space = []byte{' '}

var m = &sync.Mutex{}
var conns = []*websocket.Conn{}

type Hub struct {
	registerChan chan *WsClient
	messageChan  chan *Message
}

type Message struct {
	sourceClient *WsClient
	content      []byte
}

type WsClient struct {
	namespace string
	conn      *websocket.Conn
}

func NewHub() *Hub {
	hub := &Hub{
		registerChan: make(chan *WsClient),
		messageChan:  make(chan *Message),
	}

	go func() {
		clients := make(map[string][]*WsClient)

		for {
			select {
			case client := <-hub.registerChan:
				log.Printf("add client to %v", client.namespace)
				nsList, ok := clients[client.namespace]
				if !ok {
					log.Printf("new list")
					nsList = []*WsClient{}
				}
				nsList = append(nsList, client)
				clients[client.namespace] = nsList

				if len(nsList) == 2 {
					// we have pair, tell the first one to iniate the call
					nsList[0].write([]byte("initiateCall"))
				}

			case message := <-hub.messageChan:
				log.Printf("send message %v to %v", string(message.content), message.sourceClient.namespace)
				nsList, ok := clients[message.sourceClient.namespace]
				if ok {
					for _, destClient := range nsList {
						if destClient != message.sourceClient {
							destClient.write(message.content)
						}
					}
				}
			}
		}
	}()

	return hub
}

func (h *Hub) addClient(namespace string, conn *websocket.Conn) {
	client := &WsClient{
		namespace: namespace,
		conn:      conn,
	}

	go read(h, client)

	h.registerChan <- client
}

// ServeWebSocket needs documentation <- TODO
func ServeWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	//TODO: use URL to create namespace / room

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade ws: ", err)
		return
	}

	hub.addClient(r.URL.String(), conn)
}

func read(hub *Hub, client *WsClient) {
	for {
		//TODO: close on error
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Println("error on read:", err)
			return
		}

		//TODO: jank
		// message = bytes.Replace(message, newline, space, -1)
		// message = bytes.TrimSpace(message)

		m := &Message{
			sourceClient: client,
			content:      message,
		}
		hub.messageChan <- m
	}
}

func (c *WsClient) write(message []byte) {
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Println("Couldn't get next writer:", err)
		return
	}

	w.Write(message)

	if err := w.Close(); err != nil {
		log.Println("Couldn't close writer:", err)
	}
}
