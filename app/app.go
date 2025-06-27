package app

import (
	"app-frame-work/adpter"
	client2 "app-frame-work/client"
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
	"sync"
)

var myLogger = logger.BuildMyLogger()

type App interface {
	Start(appContext *FrameAppContext) error
	Shutdown(appContext *FrameAppContext) error
	Register(moduleName string, obj interface{}) error
	RegisterRemote(moduleName string, obj interface{}) error
}

type FrameAppContext struct {
	Config                 *config.AppConfig
	Context                context.Context
	ConnectionManager      *appcontext.ConnectionManager
	ClientSessionManager   *appcontext.ConnectionManager
	RegistrySessionManager *appcontext.ConnectionManager
	ConsumerClient         *client.ConsumerClientImpl
	RemoteRPCClient        *client2.RemoteRPCImpl
}

func BuildFrameAppContext() FrameAppContext {
	myLogger.Info("初始化Context")
	defaultAdapter := adpter.AdapterImpl{AdapterCMD: fkcommon.RPC}
	defaultAdapter.AddAdapter(&defaultAdapter)
	messageHandler := handler.MessageHandlerImpl{Adapter: defaultAdapter}
	//------------------------------------------------------------------------------------
	defaultAdapterForRegistry := adpter.AdapterImpl{AdapterCMD: fkcommon.RPC}
	defaultAdapterForRegistry.AddAdapter(&defaultAdapterForRegistry)
	registryClientHandler := handler.RPCClientHandler{}
	registryClientHandler.MessageHandlerImpl.Adapter = defaultAdapterForRegistry
	//------------------------------------------------------------------------------------
	defaultAdapterForRPC := adpter.AdapterImpl{AdapterCMD: fkcommon.RPC}
	defaultAdapterForRPC.AddAdapter(&defaultAdapterForRPC)
	rpcClientHandler := handler.RPCClientHandler{}
	rpcClientHandler.MessageHandlerImpl.Adapter = defaultAdapterForRPC
	//------------------------------------------------------------------------------------
	rpcSessionManager := appcontext.NewSessionManagerBuilder(false, &rpcClientHandler)
	return FrameAppContext{
		Config:                 config.LoadAppConfig(),
		ConnectionManager:      appcontext.NewSessionManagerBuilder(true, &messageHandler),
		ClientSessionManager:   appcontext.NewSessionManagerBuilder(false, &rpcClientHandler),
		RegistrySessionManager: appcontext.NewSessionManagerBuilder(false, &registryClientHandler),
		ConsumerClient:         &client.ConsumerClientImpl{},
		RemoteRPCClient:        &client2.RemoteRPCImpl{Ctx: context.Background(), ConnectionManager: rpcSessionManager, LocalServers: make(map[string]*common.Server)},
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
	wg.Add(3)
	go func() {
		defer wg.Done()
		errorData := server.BuildNewTCPServer().Listen(appContext.Config, appContext.ConnectionManager, ctx)
		if errorData != nil {
			return
		}
	}()
	go func() {
		defer wg.Done()
		err := appContext.ConsumerClient.InitRegistry(appContext.Config, appContext.RegistrySessionManager)
		if err != nil {
			return
		}
	}()
	go func() {
		defer wg.Done()
		appContext.RemoteRPCClient.StartRPCClient()
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
	err := app.Register(moduleName, obj)
	if err != nil {
		return err
	}
	err = app.ConsumerClient.RegisterRemoteService(moduleName, app.Config.ServerConfig.BindAddr, obj)
	if err != nil {
		return err
	}
	return nil
}
