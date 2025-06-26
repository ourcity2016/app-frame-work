package main

import (
	"app-frame-work/app"
	"app-frame-work/registry/service"
)

func main() {
	appContext := app.BuildFrameAppContext()
	_ = appContext.Register(appContext.Config.AppName, &service.RegistryServiceImpl{})
	_ = appContext.Start(&appContext)
}
