package main

import (
	"app-frame-work/adpter"
	"app-frame-work/app"
	fkcommon "app-frame-work/common"
	"app-frame-work/demo/service"
	hanlder "app-frame-work/registry/handler"
	"fmt"
)

func main() {
	app := app.BuildFrameAppContext()
	//app.ConnectionManager.Handler = &hanlder.MyHandler{}
	//app.Config.Filters.AddFilter(&myfilter.LoginFilter{})
	app.Config.AppName = "demo"
	app.Config.ServiceDiscover.Registry.Enable = true
	adapterRegistry := adpter.AdapterImpl{AdapterCMD: fkcommon.RPC}
	adapterRegistry.AddAdapter(&adapterRegistry)
	registryHandlerInfo := hanlder.RegistryHandler{}
	registryHandlerInfo.Adapter = adapterRegistry
	app.RegistrySessionManager.Handler = &registryHandlerInfo
	app.Config.ServerConfig.BindAddr = "127.0.0.1:8888"
	//app.Config.IgnoreRouters = []string{app.Config.AppName + ".LoginServiceImpl.Login", app.Config.AppName + ".LoginServiceImpl.LoginOut"}
	//app.Register(app.Config.AppName, &service.LoginServiceImpl{})
	//app.Register(app.Config.AppName, &service.RoomServiceImpl{})
	app.RegisterRemote(app.Config.AppName, &service.TestServiceImpl{})
	app.RegisterRemote(app.Config.AppName, &service.RoomServiceImpl{})
	err := app.Start(&app)
	if err != nil {
		fmt.Println(err)
		return
	}
}
