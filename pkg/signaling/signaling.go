package signaling

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var newline = []byte{'\n'}
var space = []byte{' '}

type Hub struct {
	registerChan   chan *WsClient
	deregisterChan chan *WsClient
	messageChan    chan *Message
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
		registerChan:   make(chan *WsClient),
		deregisterChan: make(chan *WsClient),
		messageChan:    make(chan *Message),
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
			case client := <-hub.deregisterChan:
				log.Printf("remove client from %v", client.namespace)
				list, ok := clients[client.namespace]
				if !ok {
					log.Printf("Tried to remove from room that doesn't exist!")
					continue
				}

				//TODO: all of this needs to be a Room Struct with add / remove

				//find idx of client
				idx := -1
				for i, listClient := range list {
					if client == listClient {
						idx = i
						break
					}
				}

				if idx == -1 {
					log.Printf("Tried to remove client that is not in the room!")
					continue
				}

				//overwrite client idx with last idx and shorten the slice
				list[idx] = list[len(list)-1]
				list = list[:len(list)-1]

				if len(list) > 0 {
					log.Printf("removed client, size is now: %v", len(list))
					clients[client.namespace] = list
				} else {
					delete(clients, client.namespace)
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
	defer func() {
		hub.deregisterChan <- client
		client.conn.Close()
	}()

	for {
		//TODO: close on error
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error closing: %v", err)
			} else {
				log.Printf("normal close: %v", err)
			}
			return
		}

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
