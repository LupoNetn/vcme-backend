package ws

import "sync"

type Room struct {
	ID           string
	HostID       string
	HostClient   *Client
	Participants map[string]*Client
	WaitingRoom  map[string]*Client
	mu           sync.Mutex
}

func NewRoom(id string, hostID string, hostClient *Client) *Room {
	return &Room{
		ID: id,
		HostID: hostID,
		HostClient: hostClient,
		Participants: make(map[string]*Client),
		WaitingRoom: make(map[string]*Client),
	}
}