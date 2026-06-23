package users

import (
	"log"
	"slices"
)

// TODO: hide the fields and introduce some DTO
type Room struct {
	ID    string  `json:"id"`
	Users []*User `json:"users"`
}

func NewRoom(id string) *Room {
	return &Room{ID: id, Users: make([]*User, 0)}
}

// TODO: check if the user is already in the room
// TODO: add leaving a room
func (r *Room) Join(user *User) {
	for _, u := range r.Users {
		sender, err := u.pc.AddTrack(user.localTrack)
		if err != nil {
			log.Printf("ERROR: adding track from %v to %v: %v\n", user.ID, u.ID, err)
		} else {
			u.trackSenders[user.ID] = sender
			log.Printf("DEBUG: added a track from %v to %v\n", user.ID, u.ID)
		}
		sender, err = user.pc.AddTrack(u.localTrack)
		if err != nil {
			log.Printf("ERROR: adding track from %v to %v: %v\n", u.ID, user.ID, err)
		} else {
			user.trackSenders[u.ID] = sender
			log.Printf("DEBUG: added a track from %v to %v\n", u.ID, user.ID)
		}
	}
	user.room = r
	r.Users = append(r.Users, user)
}

func (r *Room) Leave(user *User) {
	for i, u := range r.Users {
		if u != user {
			continue
		}

		// not cool for slices which we iterate upon,
		// but ok here, because we break
		r.Users = slices.Delete(r.Users, i, i+1)
		break
	}

	for _, u := range r.Users {
		user.removePeerTrack(u.ID)
	}
	for _, u := range r.Users {
		u.removePeerTrack(user.ID)
	}
	user.room = nil
}
