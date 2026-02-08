package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/luponetn/vcme/internal/util"
)

var (
	//event types
	EventTypeJoinRoom     = "join_room"
	EventTypeLeaveRoom    = "leave_room"
	EventTypeOffer        = "offer"
	EventTypeAnswer       = "answer"
	EventTypeICECandidate = "ice_candidate"
)

// handlers for the different event types
func (m *Manager) handleJoinRoom(c *Client, event Event) error {
	var payload JoinRoomPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Printf("invalid join room payload: %v", err)
		return err
	}
	//get host id
	callID, err := uuid.Parse(payload.CallID)
	if err != nil {
		log.Printf("invalid call id: %v", err)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	host_id, err := m.queries.GetHostIDByCallID(ctx, callID)
	if err != nil {
		log.Printf("error getting host id: %v", err)
		return err
	}

	//check if host is in the room
	if c.id == host_id.String() {
		room := NewRoom(payload.CallID, host_id.String(), c)
		m.mu.Lock()
		m.rooms[payload.CallID] = room
		m.mu.Unlock()
		log.Printf("host %v created room %v", host_id.String(), payload.CallID)

		room.mu.Lock()
		room.Participants[host_id.String()] = c
		room.mu.Unlock()
		log.Printf("host %v joined room %v", host_id.String(), payload.CallID)
	} else {
		room, ok := m.rooms[payload.CallID]
		if !ok {

			util.SendEventToClient(c, "error", []byte(`{"message":"host not yet connected"}`))
		}
		room.mu.Lock()
		room.WaitingRoom[c.id] = c
		room.mu.Unlock()
		log.Printf("client %v added to waiting room for room %v", c.id, payload.CallID)
		util.SendEventToClient(c, "waiting_room", []byte(`{"message":"added to waiting room"}`))
		// include the new participant's client id so the host can identify them
		newPartPayload := struct {
			Message  string `json:"message"`
			ClientID string `json:"client_id"`
		}{
			Message:  "a new user wants to join call",
			ClientID: c.id,
		}
		b, err := json.Marshal(newPartPayload)
		if err != nil {
			log.Printf("error marshaling new_participant payload: %v", err)
		} else {
			util.SendEventToClient(room.HostClient, "new_participant", b)
		}
	}

	return nil
}

func (m *Manager) handleLeaveRoom(c *Client, event Event) error {
	//handle leave room event
	return nil
}

func (m *Manager) handleOffer(c *Client, event Event) error {
	//handle offer event
	return nil
}

func (m *Manager) handleAnswer(c *Client, event Event) error {
	//handle answer event
	return nil
}

func (m *Manager) handleICECandidate(c *Client, event Event) error {
	//handle ice candidate event
	return nil
}
