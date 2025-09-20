package socket

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

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
		if err := handleMessage(w, r); err != nil {
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

func handleMessage(w io.Writer, r io.Reader) error {
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
	answer, err := handleParsedMessage(message)
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

func handleParsedMessage(message map[string]any) (any, error) {
	switch message["type"] {
	case "ping":
		return map[string]any{"type": "pong"}, nil
	}
	return make(map[string]any), nil
}
