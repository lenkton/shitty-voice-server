package socket

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrConnectionNotFound = errors.New("connection for the user not found")

func SendMessage(userID string, message map[string]any) error {
	conn, found := userConnsStorage[userID]
	if !found {
		return ErrConnectionNotFound
	}

	encodedMessage, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshalling json: %v", err)
	}

	err = conn.WriteMessage(1, encodedMessage)
	if err != nil {
		return fmt.Errorf("writing message to socket: %v", err)
	}

	return nil
}
