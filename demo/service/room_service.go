package service

import (
	fkcommon "app-frame-work/common"
	context2 "app-frame-work/context"
	"app-frame-work/demo/common"
	"app-frame-work/util"
	"context"
	"errors"
)

type RoomService interface {
	CreateRoom(context.Context, *common.Room) (*common.Room, error)
	JoinRoom(context.Context, *common.Room) (*common.Room, error)
	LeaveRoom(context.Context, *common.Room) error
	Talk(context.Context, *common.MessageToRoom) error
}

type RoomServiceImpl struct {
}

func (rom *RoomServiceImpl) CreateRoom(ctx context.Context, room *common.Room) (*common.Room, error) {
	roomName := room.RoomName
	if roomName == "" || len(roomName) <= 5 {
		return nil, errors.New("must be input roomName")
	}
	str := ctx.Value(fkcommon.ContextUserKey).(*context2.Session)
	roomMaps := common.CreatedRoomsList.RoomsMap
	roomHas := common.FindRoomIfUserCreated(str.ConnID)
	user, _ := common.CreatedUsesList.UserMap[str.ConnID]
	if roomHas != nil {
		return roomHas, nil
	}
	roomIdStr := util.UUID()
	userMap := make(map[string]common.User)
	userMap[str.ConnID] = user
	romInfo := common.Room{RoomId: roomIdStr, RoomName: roomName, RoomOwn: user, UserMap: userMap}
	roomMaps[roomIdStr] = romInfo
	myLogger.Info("创建房间成功 %v", roomMaps)
	return &romInfo, nil
}

func (rom *RoomServiceImpl) JoinRoom(ctx context.Context, room *common.Room) (*common.Room, error) {
	roomId := room.RoomId
	if roomId == "" || len(roomId) <= 5 {
		return nil, errors.New("must be input roomId")
	}
	roomMaps := common.CreatedRoomsList.RoomsMap
	roomData, ok := roomMaps[roomId]
	if !ok {
		return nil, errors.New("room not exists")
	}
	str := ctx.Value(fkcommon.ContextUserKey).(*context2.Session)
	user, _ := common.CreatedUsesList.UserMap[str.ConnID]
	roomData.UserMap[str.ConnID] = user
	return &roomData, nil
}

func (rom *RoomServiceImpl) LeaveRoom(ctx context.Context, room *common.Room) error {
	roomId := room.RoomId
	if roomId == "" || len(roomId) <= 5 {
		return errors.New("must be input roomId")
	}
	roomMaps := common.CreatedRoomsList.RoomsMap
	roomFind, ok := roomMaps[roomId]
	if !ok {
		return errors.New("room not exists")
	}
	str := ctx.Value(fkcommon.ContextUserKey).(*context2.Session)
	if roomFind.RoomOwn.UserId == str.ConnID {
		delete(roomMaps, roomId)
	}
	delete(roomFind.UserMap, str.ConnID)
	return nil
}

func (rom *RoomServiceImpl) Talk(ctx context.Context, message *common.MessageToRoom) error {
	msg := message.Message
	if msg == "" {
		return errors.New("must be input msg")
	}
	roomId := message.ToRoom
	if roomId == "" {
		return errors.New("must be input roomId")
	}
	roomMaps := common.CreatedRoomsList.RoomsMap
	roomFind, ok := roomMaps[roomId]
	if !ok {
		return errors.New("room not exists")
	}
	for _, s := range roomFind.UserMap {
		response := fkcommon.Ok(message.Message, "你好啊")
		err := s.Session.Handler.SendResponseMessage(response, s.Session.SendCh)
		if err != nil {
			continue
		}
	}
	return nil
}
