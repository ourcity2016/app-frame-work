package service

import (
	fkcommon "app-frame-work/common"
	context2 "app-frame-work/context"
	"app-frame-work/demo/common"
	"app-frame-work/logger"
	"context"
	"errors"
)

var myLogger = logger.BuildMyLogger()

type LoginService interface {
	Login(context.Context, *common.User) error
	LoginOut(context.Context, *common.User) error
}

type LoginServiceImpl struct {
}

func (service *LoginServiceImpl) Login(ctx context.Context, user *common.User) error {
	myLogger.Info("user login => %v", user)
	userNameStr := user.Name
	if userNameStr == "" || len(userNameStr) <= 5 {
		return errors.New("must be input name")
	}
	str := ctx.Value(fkcommon.ContextUserKey).(*context2.Session)
	userList := common.CreatedUsesList.UserMap
	userData, ok := userList[str.ConnID]
	if ok {
		return errors.New("用户已经登录")
	}
	userList[str.ConnID] = common.User{Name: userNameStr, UserId: str.ConnID, Session: *str}
	myLogger.Info("user login => %v", userData)
	return nil
}
func (service *LoginServiceImpl) LoginOut(ctx context.Context, user *common.User) error {
	str := ctx.Value(fkcommon.ContextUserKey).(*context2.Session)
	userList := common.CreatedUsesList.UserMap
	_, ok := userList[str.ConnID]
	if !ok {
		return errors.New("未发现登录数据")
	}
	delete(userList, str.ConnID)
	myLogger.Info("user loginOut => %s", str.ConnID)
	return nil
}
