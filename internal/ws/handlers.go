package ws

var (
	//event types
	EventTypeJoinRoom = "join_room"
	EventTypeLeaveRoom = "leave_room"
	EventTypeOffer = "offer"
	EventTypeAnswer = "answer"
	EventTypeICECandidate = "ice_candidate"
)

//handlers for the different event types
func (m *Manager) handleJoinRoom(c *client, event Event) error {
	//handle join room event
	return nil
}

func (m *Manager) handleLeaveRoom(c *client, event Event) error {
	//handle leave room event
	return nil
}

func (m *Manager) handleOffer(c *client, event Event) error {
	//handle offer event
	return nil
}

func (m *Manager) handleAnswer(c *client, event Event) error {
	//handle answer event
	return nil
}

func (m *Manager) handleICECandidate(c *client, event Event) error {
	//handle ice candidate event
	return nil
}

