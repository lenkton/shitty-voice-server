package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type ConnectionKeyType struct{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR: upgrading connection: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, ConnectionKeyType{}, conn)
	for {
		messageType, r, err := conn.NextReader()
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("INFO: ws message type: %v\n", messageType)
		w, err := conn.NextWriter(messageType)
		if err != nil {
			log.Println(err)
			return
		}
		if err := handleMessage(w, r, ctx); err != nil {
			log.Println(err)
			// I think it is better to try to close the w
			// return
		}
		if err := w.Close(); err != nil {
			log.Println(err)
			return
		}
	}
}

func handleMessage(w io.Writer, r io.Reader, ctx context.Context) error {
	message := make(map[string]any)
	err := json.NewDecoder(r).Decode(&message)
	if err != nil {
		log.Printf("ERROR: parsing ws message: %v\n", err)
		_, innerErr := w.Write([]byte(`{"error":"malformed message"}`))
		if innerErr != nil {
			log.Printf("ERROR: writing ws message: %v\n", err)
		}
		return err
	}
	answer, err := handleParsedMessage(message, ctx)
	if err != nil {
		log.Printf("ERROR: processing ws message: %v\n", err)
		return err
	}

	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		log.Printf("ERROR: writing ws message: %v\n", err)
		return err
	}

	return nil
}

// TODO: clear the storage on connection close
var userConnsStorage = make(map[string]*websocket.Conn)

// TODO: rewrite into a type: MessageWithContext
func handleParsedMessage(message map[string]any, ctx context.Context) (any, error) {
	switch message["type"] {
	case "ping":
		return map[string]any{"type": "pong"}, nil
	case "login":
		// WARN: it could be an integer...
		userID, ok := message["userId"].(string)
		if !ok || userID == "" {
			// WARN: it is kind of an error...
			return map[string]any{"error": "invalid user id"}, nil
		}
		connAny := ctx.Value(ConnectionKeyType{})
		if connAny == nil {
			return nil, fmt.Errorf("no connection in the context")
		}
		// WARN: could panic
		conn := connAny.(*websocket.Conn)
		userConnsStorage[userID] = conn
		log.Printf("INFO: ws user %v logged in\n", userID)
		return map[string]any{
				"type":    "response",
				"message": "logged in as " + userID,
			},
			nil
	}
	return make(map[string]any), nil
}
