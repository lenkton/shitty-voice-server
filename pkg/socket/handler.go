package socket

import (
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
		if _, err := io.Copy(w, r); err != nil {
			log.Println(err)
			return
		}
		if err := w.Close(); err != nil {
			log.Println(err)
			return
		}
	}

}
