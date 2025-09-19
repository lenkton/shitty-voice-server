package users

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
func (r *Room) Join(u *User) {
	r.Users = append(r.Users, u)
}
