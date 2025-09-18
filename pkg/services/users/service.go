package users

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pion/webrtc/v4"
)

type UsersService struct {
	api  *webrtc.API
	user *User
}

func NewService(api *webrtc.API) *UsersService {
	return &UsersService{api: api, user: &User{}}
}

func (us *UsersService) HTTPHandleOffer(w http.ResponseWriter, r *http.Request) {
	err := us.user.CreatePeerConnection(us.api)
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
	answer, err := us.user.HandleOffer(offer)
	if err != nil {
		log.Printf("ERROR: HandleOffer: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		log.Println(err)
		return
	}
}
