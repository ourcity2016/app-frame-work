package common

import "app-frame-work/context"

type Room struct {
	RoomId   string          `json:"roomId"` //UUID
	RoomName string          `json:"roomName"`
	UserMap  map[string]User `json:"userMap"`
	RoomOwn  User            `json:"roomOwn"`
}

type MessageToRoom struct {
	Message string `json:"message"`
	ToRoom  string `json:"toRoom"`
}

type User struct {
	Name    string          `json:"name"`
	UserId  string          `json:"userId"` //UUID
	Session context.Session `json:"-"`
}

var (
	CreatedRoomsList = &CreatedRooms{RoomsMap: make(map[string]Room)}
	CreatedUsesList  = &CreatedUses{UserMap: make(map[string]User)}
)

type CreatedRooms struct {
	RoomsMap map[string]Room `json:"roomsMap"`
}

type CreatedUses struct {
	UserMap map[string]User `json:"userMap"`
}

func FindRoomIfUserCreated(userId string) *Room {
	for _, s := range CreatedRoomsList.RoomsMap {
		if s.RoomOwn.UserId == userId {
			return &s
		}
	}
	return nil
}
