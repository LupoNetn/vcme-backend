package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/luponetn/vcme/internal/db"
	"github.com/luponetn/vcme/internal/util"
)

var (
	//event types
	EventTypeJoinRoom          = "join_room"
	EventTypeLeaveRoom         = "leave_room"
	EventTypeOffer             = "offer"
	EventTypeAnswer            = "answer"
	EventTypeICECandidate      = "ice_candidate"
	EventTypeAcceptParticipant = "accept_participant"
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

func (m *Manager) handleAcceptParticipant(c *Client, event Event) error {
	var payload AcceptParticipantPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Printf("invalid accept participant payload: %v", err)
		return err
	}

	room, ok := m.rooms[payload.CallID]
	if !ok {
		log.Printf("room not found for call id: %v", payload.CallID)
		util.SendEventToClient(c, "error", []byte(`{"message":"room not found"}`))
		return nil
	}

	if c.id != room.HostID {
		util.SendEventToClient(c, "error", []byte(`{"message":"only host can accept participants"}`))
		return nil
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	participantClient, ok := room.WaitingRoom[payload.ParticipantID]
	if !ok {
		log.Printf("participant client not found in waiting room for participant id: %v", payload.ParticipantID)
		util.SendEventToClient(c, "error", []byte(`{"message":"participant client not found in waiting room"}`))
		return nil
	}
	delete(room.WaitingRoom, payload.ParticipantID)
	room.Participants[payload.ParticipantID] = participantClient
	log.Printf("participant %v accepted into room %v", payload.ParticipantID, payload.CallID)

	//participant's list
	participantList := make([]string, 0, len(room.Participants))
	for participantID := range room.Participants {
		participantList = append(participantList, participantID)
	}

	//add to participant's table
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	callUUID, err := uuid.Parse(payload.CallID)
	if err != nil {
		return err
	}

	userUUID, err := uuid.Parse(payload.ParticipantID)
	if err != nil {
		return err
	}

	params := db.AddUserToCallParticipantsParams{
		CallID: callUUID,
		UserID: userUUID,
	}
	_, err = m.queries.AddUserToCallParticipants(ctx, params)
	if err != nil {
		log.Printf("error adding user to call participants: %v", err)
		util.SendEventToClient(c, "error", []byte(`{"message":"error adding participant to call in database"}`))
		return nil
	}

	//send acceptance event to participant with list of current participants in the room
	acceptedPayload := struct {
		Message       string   `json:"message"`
		CallID        string   `json:"call_id"`
		Participants  []string `json:"participants"`
		ParticipantID string   `json:"participant_id"`
	}{
		Message:       "accepted into room",
		CallID:        payload.CallID,
		Participants:  participantList,
		ParticipantID: payload.ParticipantID,
	}
	b, err := json.Marshal(acceptedPayload)
	if err != nil {
		log.Printf("error marshaling accepted payload: %v", err)
		util.SendEventToClient(c, "error", []byte(`{"message":"error preparing acceptance payload"}`))
		return nil
	}

	util.SendEventToClient(participantClient, "accepted_into_room", b)
	return nil
}

func (m *Manager) handleLeaveRoom(c *Client, event Event) error {
	//handle leave room event
	return nil
}

func (m *Manager) handleOffer(c *Client, event Event) error {
	var offerPayload SignalPayload
	if err := json.Unmarshal(event.Payload, &offerPayload); err != nil {
		log.Printf("invalid offer payload: %v", err)
		return err
	}

	room, ok := m.rooms[offerPayload.CallID]
	if !ok {
		log.Printf("room not found for call id: %v", offerPayload.CallID)
		util.SendEventToClient(c, "error", []byte(`{"message":"room not found"}`))
		return nil
	}

	// Check membership while holding the lock, then release before sending
	room.mu.Lock()
	_, senderPresent := room.Participants[c.id]
	targetClient, targetPresent := room.Participants[offerPayload.TargetID]
	room.mu.Unlock()

	if !senderPresent {
		log.Printf("sender not in room: %v", c.id)
		util.SendEventToClient(c, "error", []byte(`{"message":"sender not in room"}`))
		return nil
	}

	if !targetPresent {
		log.Printf("target client not found in room for target id: %v", offerPayload.TargetID)
		util.SendEventToClient(c, "error", []byte(`{"message":"target client not found in room"}`))
		return nil
	}

	// Forward the offer and include the sender id so recipient can identify them
	forwardPayload := struct {
		CallID   string          `json:"call_id"`
		TargetID string          `json:"target_id"`
		SenderID string          `json:"sender_id"`
		Data     json.RawMessage `json:"data"`
	}{
		CallID:   offerPayload.CallID,
		TargetID: offerPayload.TargetID,
		SenderID: c.id,
		Data:     offerPayload.Data,
	}

	b, err := json.Marshal(forwardPayload)
	if err != nil {
		log.Printf("error marshaling forward offer payload: %v", err)
		util.SendEventToClient(c, "error", []byte(`{"message":"error preparing offer payload"}`))
		return nil
	}

	util.SendEventToClient(targetClient, "offer", b)
	return nil
}

func (m *Manager) handleAnswer(c *Client, event Event) error {
	var answerPayload SignalPayload
	if err := json.Unmarshal(event.Payload, &answerPayload); err != nil {
		log.Printf("invalid answer payload: %v", err)
		return err
	}

	room, ok := m.rooms[answerPayload.CallID]
	if !ok {
		log.Printf("room not found for call id: %v", answerPayload.CallID)
		util.SendEventToClient(c, "error", []byte(`{"message":"room not found"}`))
		return nil
	}

	// Verify membership and target while holding the lock
	room.mu.Lock()
	_, senderPresent := room.Participants[c.id]
	targetClient, targetPresent := room.Participants[answerPayload.TargetID]
	room.mu.Unlock()

	if !senderPresent {
		log.Printf("sender not in room: %v", c.id)
		util.SendEventToClient(c, "error", []byte(`{"message":"sender not in room"}`))
		return nil
	}

	if !targetPresent {
		log.Printf("target client not found in room for target id: %v", answerPayload.TargetID)
		util.SendEventToClient(c, "error", []byte(`{"message":"target client not found in room"}`))
		return nil
	}

	// Forward the answer and include the sender id so recipient can identify them
	forwardPayload := struct {
		CallID   string          `json:"call_id"`
		TargetID string          `json:"target_id"`
		SenderID string          `json:"sender_id"`
		Data     json.RawMessage `json:"data"`
	}{
		CallID:   answerPayload.CallID,
		TargetID: answerPayload.TargetID,
		SenderID: c.id,
		Data:     answerPayload.Data,
	}

	b, err := json.Marshal(forwardPayload)
	if err != nil {
		log.Printf("error marshaling forward answer payload: %v", err)
		util.SendEventToClient(c, "error", []byte(`{"message":"error preparing answer payload"}`))
		return nil
	}

	util.SendEventToClient(targetClient, "answer", b)
	return nil
}

func (m *Manager) handleICECandidate(c *Client, event Event) error {
	var icePayload SignalPayload
	if err := json.Unmarshal(event.Payload, &icePayload); err != nil {
		log.Printf("invalid ice candidate payload: %v", err)
		return err
	}

	room, ok := m.rooms[icePayload.CallID]
	if !ok {
		log.Printf("room not found for call id: %v", icePayload.CallID)
		util.SendEventToClient(c, "error", []byte(`{"message":"room not found"}`))
		return nil
	}

	// Verify membership and target while holding the lock
	room.mu.Lock()
	_, senderPresent := room.Participants[c.id]
	targetClient, targetPresent := room.Participants[icePayload.TargetID]
	room.mu.Unlock()

	if !senderPresent {
		log.Printf("sender not in room: %v", c.id)
		util.SendEventToClient(c, "error", []byte(`{"message":"sender not in room"}`))
		return nil
	}

	if !targetPresent {
		log.Printf("target client not found in room for target id: %v", icePayload.TargetID)
		util.SendEventToClient(c, "error", []byte(`{"message":"target client not found in room"}`))
		return nil
	}

	// Forward the ICE candidate and include the sender id so recipient can identify them
	forwardPayload := struct {
		CallID   string          `json:"call_id"`
		TargetID string          `json:"target_id"`
		SenderID string          `json:"sender_id"`
		Data     json.RawMessage `json:"data"`
	}{
		CallID:   icePayload.CallID,
		TargetID: icePayload.TargetID,
		SenderID: c.id,
		Data:     icePayload.Data,
	}

	b, err := json.Marshal(forwardPayload)
	if err != nil {
		log.Printf("error marshaling forward ice candidate payload: %v", err)
		util.SendEventToClient(c, "error", []byte(`{"message":"error preparing ice candidate payload"}`))
		return nil
	}

	util.SendEventToClient(targetClient, "ice_candidate", b)
	return nil
}
