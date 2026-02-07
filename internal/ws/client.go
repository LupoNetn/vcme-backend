package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type client struct {
	connection *websocket.Conn
	manager  *Manager
	egress chan []byte
}

type Event struct {
	EventType string `json:"event_type"`
    Payload json.RawMessage `json:"payload"`
}

func NewClient(conn *websocket.Conn, manager *Manager) *client {
	return &client{
		connection: conn,
		manager: manager,
		egress: make(chan []byte),
	}
}


func (c *client) Listen() {
	defer c.connection.Close()

	for {
		_,payload,err := c.connection.ReadMessage()
		if err != nil {
			log.Printf("error reading websocket message: %v", err)
			break
		}

		var event Event
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("Invalid JSON event: %v", err)
			break
		}

		if err := c.manager.RouteEvent(c, event); err != nil {
			log.Printf("error routing event: %v", err)
			break
		}
	}
}

func (c *client) Send(message []byte) {}