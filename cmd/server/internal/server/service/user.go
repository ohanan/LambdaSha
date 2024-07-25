package service

func NewUser(id string) *User {
	return &User{
		id:       id,
		roomInfo: &UserRoomInfo{},
	}
}

type User struct {
	id       string
	roomInfo *UserRoomInfo
}

func (a *User) ID() string {
	return a.id
}
