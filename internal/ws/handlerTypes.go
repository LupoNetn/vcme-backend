package ws

type JoinRoomPayload struct {
	CallID string `json:"call_id"`
	CallLink string `json:"call_link"`
}


type AcceptParticipantPayload struct {
	CallID string `json:"call_id"`
    ParticipantID string `json:"participant_id"`
}