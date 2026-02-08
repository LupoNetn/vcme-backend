package util

import (
	"encoding/json"
	"log"
)

func SendEventToClient(c interface{}, eventType string, payload []byte) {
	event := struct {
		EventType string
		Payload   []byte
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
	if sender, ok := c.(interface{ Send([]byte) }); ok {
		sender.Send(data)
	} else {
		log.Printf("error sending event: client does not implement Send([]byte)")
	}
}
