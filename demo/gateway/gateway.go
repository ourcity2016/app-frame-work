package main

import (
	"app-frame-work/app"
	"app-frame-work/demo/myfilter"
	"app-frame-work/demo/service"
	"fmt"
)

func main() {
	app := app.BuildFrameAppContext()
	//app.ConnectionManager.Handler = &hanlder.MyHandler{}
	app.Config.Filters.AddFilter(&myfilter.LoginFilter{})
	app.Config.AppName = "demo"
	app.Config.ServiceDiscover.Registry.Enable = true
	app.Config.IgnoreRouters = []string{app.Config.AppName + ".LoginServiceImpl.Login", app.Config.AppName + ".LoginServiceImpl.LoginOut"}
	app.Register(app.Config.AppName, &service.LoginServiceImpl{})
	//app.Register(app.Config.AppName, &service.RoomServiceImpl{})
	//app.RegisterRemote(app.Config.AppName, &service.TestServiceImpl{})
	err := app.Start(&app)
	if err != nil {
		fmt.Println(err)
		return
	}
}
