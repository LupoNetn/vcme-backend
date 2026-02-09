package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/luponetn/vcme/internal/db"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true},
		}
)

type clientList map[*Client]bool
type clientsByID map[string]*Client
type eventHandler func(client *Client, event Event) error

type Manager struct {
	r *gin.Engine
	clients  clientList
	clientsByID clientsByID
	handlers map[string]eventHandler
	rooms map[string]*Room
	mu sync.Mutex
	queries db.Queries
}

func NewManager(r *gin.Engine, querier *db.Queries) *Manager {
	return &Manager{
		r: r,
		clients: make(clientList),
		clientsByID: make(clientsByID),
		handlers: make(map[string]eventHandler),
		rooms: make(map[string]*Room),
		queries: *querier,
	}
}


func (m *Manager) ServeWS(c *gin.Context) {
	log.Println("new websocket connection")
	//upgrade http connection to websocket
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("error upgrading connection: %v", err)
		return
	}

	//get connected client's id and add to clientsByID map
	user, exists := c.Get("user")
	if !exists {
		log.Printf("error getting client id from context")
		conn.Close()
		return
	}

	var clientIDStr string
	switch v := user.(type) {
	case string:
		clientIDStr = v
	default:
		if s, ok := user.(fmt.Stringer); ok {
			clientIDStr = s.String()
		} else {
			log.Printf("unexpected user type in context: %T", user)
			conn.Close()
			return
		}
	}

	client := NewClient(clientIDStr, conn, m)
	m.AddClient(client, clientIDStr)

	//start listening for processes from the clients
	go client.Listen()
	go client.Send()
}

//register event handlers for the different event types
func (m *Manager) RegisterEventHandler() {
	m.handlers[EventTypeAnswer] = m.handleAnswer
	m.handlers[EventTypeOffer] = m.handleOffer
	m.handlers[EventTypeJoinRoom] = m.handleJoinRoom
	m.handlers[EventTypeLeaveRoom] = m.handleLeaveRoom
	m.handlers[EventTypeICECandidate] = m.handleICECandidate
	m.handlers[EventTypeAcceptParticipant] = m.handleAcceptParticipant
}

//send event to the manager for routing
func (m *Manager) RouteEvent(c *Client, event Event) error {
  if _, ok := m.handlers[event.EventType]; !ok {
	log.Printf("no handler for specified event: %v", event.EventType)
	return fmt.Errorf("no handler for specified event: %v", event.EventType)
  }

  return m.handlers[event.EventType](c, event)
}


func (m *Manager) AddClient(client *Client, clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client] = true
	m.clientsByID[clientID] = client
}

func (m *Manager) RemoveClient(client *Client, clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}

	if clientID != "" {
		delete(m.clientsByID, clientID)
	}
}