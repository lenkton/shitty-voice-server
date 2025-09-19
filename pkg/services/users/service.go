package users

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pion/webrtc/v4"
)

type UsersService struct {
	api   *webrtc.API
	users map[string]*User
	rooms map[string]*Room
}

func NewService(api *webrtc.API) *UsersService {
	return &UsersService{api: api, users: make(map[string]*User), rooms: make(map[string]*Room)}
}

func (us *UsersService) HTTPHandleOffer(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	user, found := us.users[userID]
	if !found {
		user = NewUser(userID)
		us.users[userID] = user
	}

	err := user.CreatePeerConnection(us.api)
	// TODO: somehow kill the pc after the call is over
	if err == ErrPCAlreadyCreated {
		log.Println("ERROR: there is a PeerConnection already")
		http.Error(w, "the peer has already been connected", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("ERROR: CreatePeerConnection: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	var offer webrtc.SessionDescription
	err = json.NewDecoder(r.Body).Decode(&offer)
	if err != nil {
		log.Printf("ERROR: parsing offer: %v\n", err)
		http.Error(w, "malformed offer", http.StatusUnprocessableEntity)
		return
	}
	answer, err := user.HandleOffer(offer)
	if err != nil {
		log.Printf("ERROR: HandleOffer: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		log.Printf("ERROR: encoding answer: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// TODO: check if the user has already joined some room
func (us *UsersService) HTTPHandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	dto := &struct {
		UserID string `json:"userId"`
	}{}
	err := json.NewDecoder(r.Body).Decode(dto)
	if err != nil {
		log.Printf("ERROR: decoding body: %v\n", err)
		http.Error(w, "malformed request body", http.StatusUnprocessableEntity)
		return
	}
	user, found := us.users[dto.UserID]
	if !found {
		user = NewUser(dto.UserID)
		us.users[dto.UserID] = user
	}
	roomID := r.PathValue("room_id")
	room, found := us.rooms[roomID]
	if !found {
		room = NewRoom(roomID)
		us.rooms[roomID] = room
	}
	room.Join(user)
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(room)
	if err != nil {
		log.Printf("ERROR: encoding json: %v\n", err)
	}
}
