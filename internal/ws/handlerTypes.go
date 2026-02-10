package ws

import (
	"encoding/json"
)

type JoinRoomPayload struct {
	CallID   string `json:"call_id"`
	CallLink string `json:"call_link"`
}

type AcceptParticipantPayload struct {
	CallID        string `json:"call_id"`
	ParticipantID string `json:"participant_id"`
}

type LeaveRoomPayload struct {
	CallID        string `json:"call_id"`
	ParticipantID string `json:"participant_id"`
}

type SignalPayload struct {
	CallID   string          `json:"call_id"`
	TargetID string          `json:"target_id"`
	Data     json.RawMessage `json:"data"`
}
