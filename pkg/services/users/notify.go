package users

import (
	"echo-webrtc-test/pkg/amqp"
	"echo-webrtc-test/pkg/socket"
	"encoding/json"
	"fmt"
)

// TODO: introduce some kind of a flag to decide which path to take
func Notify(id string, message map[string]any) error {
	err := socket.SendMessage(id, message)
	if err != socket.ErrConnectionNotFound {
		return err
	}
	if err == nil {
		return nil
	}
	// so we are using this server as a media node (or not...)
	return sendAMQPUpdate(id, message)
}

type amqpMessage struct {
	UserID  string         `json:"user_id"`
	Message map[string]any `json:"message"`
}

const UPDATES_QUEUE_NAME = "voice-server-updates"

func sendAMQPUpdate(id string, message map[string]any) error {
	msg := &amqpMessage{UserID: id, Message: message}
	encoded, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("encoding message: %d", err)
	}
	err = amqp.MainAMQPClient.SendMessageToQueue(string(encoded), UPDATES_QUEUE_NAME)
	if err != nil {
		return fmt.Errorf("sending message: %d", err)
	}

	return nil
}
