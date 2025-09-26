package users

import "log"

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
		// TODO: it could return an error
		_, err := u.pc.AddTrack(user.localTrack)
		if err != nil {
			log.Printf("ERROR: adding track: %v\n", err)
		}
		log.Printf("DEBUG: added a track from %v to %v\n", user.ID, u.ID)
		// TODO: it could return an error
		_, err = user.pc.AddTrack(u.localTrack)
		if err != nil {
			log.Printf("ERROR: adding track: %v\n", err)
		}
		log.Printf("DEBUG: added a track from %v to %v\n", u.ID, user.ID)
	}
	r.Users = append(r.Users, user)
}
