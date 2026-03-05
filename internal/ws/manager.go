package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/luponetn/vcme/internal/config"
	"github.com/luponetn/vcme/internal/db"
	"github.com/luponetn/vcme/internal/util"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type clientList map[*Client]bool
type clientsByID map[string]*Client
type eventHandler func(client *Client, event Event) error

type Manager struct {
	r           *gin.Engine
	clients     clientList
	clientsByID clientsByID
	handlers    map[string]eventHandler
	rooms       map[string]*Room
	mu          sync.Mutex
	queries     db.Queries
	cfg         *config.Config
}

func NewManager(r *gin.Engine, querier *db.Queries, cfg *config.Config) *Manager {
	m := &Manager{
		r:           r,
		clients:     make(clientList),
		clientsByID: make(clientsByID),
		handlers:    make(map[string]eventHandler),
		rooms:       make(map[string]*Room),
		queries:     *querier,
		cfg:         cfg,
	}
	m.RegisterEventHandler()
	return m
}

func (m *Manager) ServeWS(c *gin.Context) {
	log.Println("new websocket connection")

	//get connected client's id from token in query param - Validate BEFORE upgrading
	token := c.Query("token")
	if token == "" {
		log.Printf("no token provided in request")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is required"})
		return
	}

	claims, err := util.VerifyToken(token, m.cfg.JWTAccessSecret)
	if err != nil {
		log.Printf("invalid token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	clientIDStr := claims.UserID
	/*
	 TODO: Convert clientIDStr from uuid.UUID -> string
	*/

	//upgrade http connection to websocket
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("error upgrading connection: %v", err)
		return
	}

	client := NewClient(clientIDStr.String(), conn, m)
	m.AddClient(client, clientIDStr.String())

	//start listening for processes from other clients
	go client.Listen()
	go client.Send()
}

// register event handlers for the different event types
func (m *Manager) RegisterEventHandler() {
	m.handlers[EventTypeAnswer] = m.handleAnswer
	m.handlers[EventTypeOffer] = m.handleOffer
	m.handlers[EventTypeJoinRoom] = m.handleJoinRoom
	m.handlers[EventTypeLeaveRoom] = m.handleLeaveRoom
	m.handlers[EventTypeICECandidate] = m.handleICECandidate
	m.handlers[EventTypeAcceptParticipant] = m.handleAcceptParticipant
	m.handlers[EventTypeDeclineParticipant] = m.handleDeclineParticipant
	m.handlers[EventTypeGetInitiator] = m.handleGetInitiator
	m.handlers[EventTypeSendEmoji] = m.handleSendEmoji
}

// send event to the manager for routing
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
	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}

	if clientID != "" {
		delete(m.clientsByID, clientID)
	}
	m.mu.Unlock()

	// Cleanup from any room they were in
	if client.RoomID != "" {
		m.RemoveClientFromRoom(client)
	}
}

func (m *Manager) RemoveClientFromRoom(c *Client) {
	m.mu.Lock()
	room, ok := m.rooms[c.RoomID]
	m.mu.Unlock()

	if !ok {
		return
	}

	room.mu.Lock()
	delete(room.Participants, c.id)
	delete(room.WaitingRoom, c.id)

	// Remove from ParticipantsList
	for i, id := range room.ParticipantsList {
		if id == c.id {
			room.ParticipantsList = append(room.ParticipantsList[:i], room.ParticipantsList[i+1:]...)
			break
		}
	}

	isEmpty := len(room.Participants) == 0 && len(room.WaitingRoom) == 0
	room.mu.Unlock()

	if isEmpty {
		m.mu.Lock()
		// Re-check under manager lock to avoid race conditions
		if r, ok := m.rooms[c.RoomID]; ok {
			r.mu.Lock()
			if len(r.Participants) == 0 && len(r.WaitingRoom) == 0 {
				delete(m.rooms, c.RoomID)
				log.Printf("room %s deleted as it is empty", c.RoomID)
			}
			r.mu.Unlock()
		}
		m.mu.Unlock()
	}
}
