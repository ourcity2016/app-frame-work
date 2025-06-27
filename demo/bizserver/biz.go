package main

import (
	"app-frame-work/app"
	"app-frame-work/demo/service"
)

func main() {
	appContext := app.BuildFrameAppContext()
	_ = appContext.RegisterRemote(appContext.Config.AppName, &service.TestServiceImpl{})
	_ = appContext.RegisterRemote(appContext.Config.AppName, &service.RoomServiceImpl{})
	_ = appContext.Start(&appContext)
}
