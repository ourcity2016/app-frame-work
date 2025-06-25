package app

import (
	"app-frame-work/adpter"
	fkcommon "app-frame-work/common"
	"app-frame-work/config"
	appcontext "app-frame-work/context"
	"app-frame-work/handler"
	"app-frame-work/logger"
	"app-frame-work/registry/client"
	"app-frame-work/registry/common"
	"app-frame-work/rpc"
	"app-frame-work/server"
	"context"
	"reflect"
	"strings"
	"sync"
)

var myLogger = logger.BuildMyLogger()
var consumerClientImpl = client.ConsumerClientImpl{}

type App interface {
	Start(appContext *FrameAppContext) error
	Shutdown(appContext *FrameAppContext) error
	Register(moduleName string, obj interface{}) error
	RegisterRemote(moduleName string, obj interface{}) error
}

type FrameAppContext struct {
	Config               *config.AppConfig
	Context              context.Context
	ConnectionManager    *appcontext.ConnectionManager
	ClientSessionManager *appcontext.ConnectionManager
}

func BuildFrameAppContext() FrameAppContext {
	myLogger.Info("初始化Context")
	defaultAdapter := adpter.AdapterImpl{AdapterCMD: fkcommon.RPC}
	defaultAdapter.AddFilter(&defaultAdapter)
	messageHandler := handler.MessageHandlerImpl{Adapter: defaultAdapter}
	connectionManager := appcontext.NewSessionManagerBuilder(true, &messageHandler)
	return FrameAppContext{
		Config:               config.NewDefaultAppConfig(),
		ConnectionManager:    connectionManager,
		ClientSessionManager: appcontext.NewSessionManagerBuilder(true, &messageHandler),
	}
}

func BuildFrameAppContextWithConfig(config *config.AppConfig) FrameAppContext {
	frameAppContext := BuildFrameAppContext()
	frameAppContext.Config = config
	return frameAppContext
}

func (app *FrameAppContext) Start(appContext *FrameAppContext) error {
	ctx := context.Background()
	appContext.Context = ctx
	defer appContext.Context.Done()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		errorData := server.BuildNewTCPServer().Listen(appContext.Config, appContext.ConnectionManager, ctx)
		if errorData != nil {
			return
		}
	}()
	go func() {
		defer wg.Done()
		err := consumerClientImpl.InitRegistry(appContext.Config)
		if err != nil {
			return
		}
	}()
	wg.Wait()
	return nil
}

func (app *FrameAppContext) Shutdown() error {
	return nil
}

func (app *FrameAppContext) Register(moduleName string, obj interface{}) error {
	service := rpc.ServiceContext{}
	return service.Register(moduleName, obj)
}
func (app *FrameAppContext) RegisterRemote(moduleName string, obj interface{}) error {
	t := reflect.TypeOf(obj)
	objName := t.Elem().Name()
	bindAddr := app.Config.ServerConfig.BindAddr
	parts := strings.Split(bindAddr, ":")
	serverInfo := common.Server{Status: 1, Ip: parts[0], Port: parts[1]}
	serverMap := make(map[string]common.Server, 1)
	serverMap[serverInfo.Ip+serverInfo.Port] = serverInfo
	// 遍历所有方法
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		service := common.Service{AppName: moduleName, ServiceName: objName, MethodName: method.Name, LoadBalance: "round-robin", Status: 1, Servers: serverMap}
		myLogger.Info("register remote service %v", service)
		err := consumerClientImpl.RegisterService(&service)
		if err != nil {
			myLogger.Error(err.Error())
			return err
		}
	}
	return nil
}
