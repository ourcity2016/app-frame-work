package service

import (
	"app-frame-work/demo/common"
	"context"
)

type TestService interface {
	Test(context.Context, *common.Room) (*common.Room, error)
}

type TestServiceImpl struct {
}

func (test *TestServiceImpl) Test(ctx context.Context, room *common.Room) (*common.Room, error) {
	myLogger.Info("Hello World")
	return nil, nil
}
