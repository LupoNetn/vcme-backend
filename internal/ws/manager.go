package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/luponetn/vcme/internal/util"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true},
		}
)

type clientList map[*client]bool
type clientsByID map[string]*client
type eventHandler func(client *client, event Event) error

type Manager struct {
	r *gin.Engine
	clients  clientList
	clientsByID clientsByID
	handlers map[string]eventHandler
	mu sync.Mutex
}

func NewManager(r *gin.Engine) *Manager {
	return &Manager{
		r: r,
		clients: make(clientList),
		clientsByID: make(clientsByID),
		handlers: make(map[string]eventHandler),
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

	client := NewClient(conn,m)
	

	//get connected client's id and add to clientsByID map
	user,exists := c.Get("user")
	if !exists {
		log.Printf("error getting client id from context")
		return
	}
	clientID := user.(*util.Claims).UserID
	m.clientsByID[clientID.String()] = client

	m.AddClient(client, clientID.String())


	//start listening for processes from the clients
	go client.Listen()
	go client.Send([]byte("hello from server"))
}

//register event handlers for the different event types
func (m *Manager) RegisterEventHandler() {
	m.handlers[EventTypeAnswer] = m.handleAnswer
	m.handlers[EventTypeOffer] = m.handleOffer
	m.handlers[EventTypeJoinRoom] = m.handleJoinRoom
	m.handlers[EventTypeLeaveRoom] = m.handleLeaveRoom
	m.handlers[EventTypeICECandidate] = m.handleICECandidate
}

//send event to the manager for routing
func (m *Manager) RouteEvent(c *client, event Event) error {
  if _, ok := m.handlers[event.EventType]; !ok {
	log.Printf("no handler for specified event: %v", event.EventType)
	return fmt.Errorf("no handler for specified event: %v", event.EventType)
  }

  return m.handlers[event.EventType](c, event)
}


func (m *Manager) AddClient(client *client, clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client] = true
	m.clientsByID[clientID] = client
}

func (m *Manager) RemoveClient(client *client, clientID string) {
	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}

	if clientID != "" {
		delete(m.clientsByID, clientID)
	}
}