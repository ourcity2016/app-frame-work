package main

import (
	"app-frame-work/app"
	"app-frame-work/registry/service"
	"fmt"
)

func main() {
	app := app.BuildFrameAppContext()
	app.Config.AppName = "registry"
	app.Config.ServerConfig.BindAddr = "127.0.0.1:8848"
	app.Register(app.Config.AppName, &service.RegistryServiceImpl{})
	err := app.Start(&app)
	if err != nil {
		fmt.Println(err)
		return
	}
}
