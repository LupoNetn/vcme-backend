package ws

import (
	"sync"

	"github.com/luponetn/vcme/internal/util"
)

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

// Broadcast sends an event to all participants in the room except the excluded client.
func (r *Room) Broadcast(eventType string, payload []byte, excludeClient *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, client := range r.Participants {
		if excludeClient != nil && client.id == excludeClient.id {
			continue
		}
		util.SendEventToClient(client, eventType, payload)
	}
}
