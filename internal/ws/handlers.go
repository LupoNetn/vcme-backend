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
	c.RoomID = payload.CallID // Store the room the client is currently in
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

		hostJoinedPayload := struct {
			Message string `json:"message"`
			CallID  string `json:"call_id"`
		}{
			Message: "host has joined call",
			CallID:  payload.CallID,
		}

		data, err := json.Marshal(hostJoinedPayload)
		if err != nil {
			log.Printf("error marshaling waiting room payload: %v", err)
		} else {
			util.SendEventToClient(c, "host_joined_room", data)
		}
	} else {
		room, ok := m.rooms[payload.CallID]
		if !ok {

			util.SendEventToClient(c, "error", []byte(`{"message":"host not yet connected"}`))
			return nil
		}
		room.mu.Lock()
		room.WaitingRoom[c.id] = c
		room.mu.Unlock()
		log.Printf("client %v added to waiting room for room %v", c.id, payload.CallID)

		//send waiting room event to client
		waitingRoomPayload := struct {
			Message string `json:"message"`
			CallID  string `json:"call_id"`
		}{
			Message: "added to waiting room",
			CallID:  payload.CallID,
		}
		data, err := json.Marshal(waitingRoomPayload)
		if err != nil {
			log.Printf("error marshaling waiting room payload: %v", err)
		} else {
			util.SendEventToClient(c, "waiting_room", data)
		}
		// get requester name
		reqUUID, _ := uuid.Parse(c.id)
		reqUser, err := m.queries.GetUserById(ctx, reqUUID)
		userName := "Anonymous"
		if err == nil {
			userName = reqUser.Name
		}

		// include the new participant's client id so the host can identify them
		newPartPayload := struct {
			Message    string `json:"message"`
			ClientID   string `json:"client_id"`
			ClientName string `json:"client_name"`
		}{
			Message:    userName + " wants to join the call",
			ClientID:   c.id,
			ClientName: userName,
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
		util.SendEventToClient(c, "error", []byte(`{"message":"room not found"}`))
		return nil
	}

	// only host can accept
	if c.id != room.HostID {
		util.SendEventToClient(c, "error", []byte(`{"message":"only host can accept participants"}`))
		return nil
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// ---------------------------
	// get client from waiting room
	// ---------------------------
	participantClient, ok := room.WaitingRoom[payload.ParticipantID]
	if !ok {
		util.SendEventToClient(c, "error", []byte(`{"message":"participant not in waiting room"}`))
		return nil
	}

	// if already in memory just ignore
	if _, exists := room.Participants[payload.ParticipantID]; exists {
		log.Printf("participant already in memory room: %s", payload.ParticipantID)
		return nil
	}

	// ---------------------------
	// DB logic (UPSERT-LIKE)
	// ---------------------------
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	callUUID, err := uuid.Parse(payload.CallID)
	if err != nil {
		return err
	}

	userUUID, err := uuid.Parse(payload.ParticipantID)
	if err != nil {
		return err
	}

	// check if already exists in DB
	_, err = m.queries.GetCallParticipant(ctx, db.GetCallParticipantParams{
		CallID: callUUID,
		UserID: userUUID,
	})

	if err != nil {
		// not found → insert
		addParams := db.AddUserToCallParticipantsParams{
			CallID: callUUID,
			UserID: userUUID,
		}

		_, err = m.queries.AddUserToCallParticipants(ctx, addParams)
		if err != nil {
			log.Printf("error adding participant to db: %v", err)
			util.SendEventToClient(c, "error", []byte(`{"message":"db insert failed"}`))
			return nil
		}
	}

	// ---------------------------
	// MOVE USER INTO ROOM MEMORY
	// ---------------------------
	delete(room.WaitingRoom, payload.ParticipantID)
	room.Participants[payload.ParticipantID] = participantClient

	log.Printf("participant %s accepted into room %s", payload.ParticipantID, payload.CallID)

	// build participant list
	participantList := make([]string, 0, len(room.Participants))
	for id := range room.Participants {
		participantList = append(participantList, id)
	}

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
		util.SendEventToClient(c, "error", []byte(`{"message":"marshal error"}`))
		return nil
	}

	util.SendEventToClient(participantClient, "accepted_into_room", b)
	return nil
}

func (m *Manager) handleLeaveRoom(c *Client, event Event) error {
	var payload LeaveRoomPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Printf("invalid leave room payload: %v", err)
		return err
	}

	_, ok := m.rooms[payload.CallID]
	if !ok {
		util.SendEventToClient(c, "error", []byte(`{"message":"room not found"}`))
		return nil
	}

	// Cleanup from the room
	m.RemoveClientFromRoom(c)
	c.RoomID = "" // Clear the room assignment

	log.Printf("participant %s left room %s", payload.ParticipantID, payload.CallID)
	util.SendEventToClient(c, "left_room", []byte(`{"message":"left room"}`))
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
