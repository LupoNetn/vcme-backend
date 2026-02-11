package ws

import "sync"

type Room struct {
	ID               string
	HostID           string
	HostClient       *Client
	Participants     map[string]*Client
	ParticipantsList []string
	WaitingRoom      map[string]*Client
	mu               sync.Mutex
}

func NewRoom(id string, hostID string, hostClient *Client) *Room {
	return &Room{
		ID:               id,
		HostID:           hostID,
		HostClient:       hostClient,
		Participants:     make(map[string]*Client),
		ParticipantsList: make([]string, 0),
		WaitingRoom:      make(map[string]*Client),
	}
}
