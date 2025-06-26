package main

import (
	"app-frame-work/app"
	"app-frame-work/demo/myfilter"
	"app-frame-work/demo/service"
)

func main() {
	appContext := app.BuildFrameAppContext()
	appContext.Config.Filters.AddFilter(&myfilter.LoginFilter{})
	_ = appContext.Register(appContext.Config.AppName, &service.LoginServiceImpl{})
	_ = appContext.Start(&appContext)
}
