package util

import (
	"encoding/json"
	"log"
)

func SendEventToClient(c any, eventType string, payload json.RawMessage) {
	event := struct {
		EventType string
		Payload   json.RawMessage
	}{
		EventType: eventType,
		Payload:   payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("error marshaling event: %v", err)
		return
	}

	// Send the event to the client
	if sender, ok := c.(interface{ SendToClientEgress([]byte) }); ok {
		sender.SendToClientEgress(data)
	} else {
		log.Printf("error sending event: client does not implement SendToClientEgress([]byte)")
	}
}
