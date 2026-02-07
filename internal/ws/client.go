package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type client struct {
	id string
	connection *websocket.Conn
	manager  *Manager
	egress chan []byte
}

type Event struct {
	EventType string `json:"event_type"`
    Payload json.RawMessage `json:"payload"`
}

func NewClient(id string, conn *websocket.Conn, manager *Manager) *client {
	return &client{
		id: id,
		connection: conn,
		manager: manager,
		egress: make(chan []byte),
	}
}


func (c *client) Listen() {
	defer func(){
		c.manager.mu.Lock()
		delete(c.manager.clients, c)
		delete(c.manager.clientsByID, c.id)
		c.manager.mu.Unlock()
		c.connection.Close()
	}()

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

func (c *client) Send() {

	defer func() {
		// remove client safely from manager
		c.manager.mu.Lock()
		delete(c.manager.clients, c)
		delete(c.manager.clientsByID, c.id)
		c.manager.mu.Unlock()

		// send close frame before closing
		_ = c.connection.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)

		c.connection.Close()
	}()

	for {
		message, ok := <-c.egress
		if !ok {
			// channel closed -> stop writer
			return
		}

		if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("error writing websocket message: %v", err)
			return
		}
	}
}
