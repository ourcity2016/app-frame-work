package client

import (
	"app-frame-work/adpter"
	"app-frame-work/client"
	fkcommon "app-frame-work/common"
	"app-frame-work/config"
	fkcontext "app-frame-work/context"
	"app-frame-work/logger"
	"app-frame-work/registry/common"
	"app-frame-work/registry/handler"
	"context"
	"encoding/json"
	"time"
)

var myLogger = logger.BuildMyLogger()

type ConsumerClient interface {
	InitRegistry(*config.AppConfig) error
	RegisterService(service *common.Service) error
	RemoveService()
}

type ConsumerClientImpl struct {
	ConnectionManager *fkcontext.ConnectionManager
}

func (cc *ConsumerClientImpl) InitRegistry(config *config.AppConfig) error {
	registry := config.ServiceDiscover.Registry
	bindAddr := registry.BindAddr
	network := registry.Network
	if !registry.Enable {
		myLogger.Info("registry default not enable")
		return nil
	}
	if bindAddr == "" || network == "" {
		myLogger.Info("registry is nil")
		return nil
	}
	registryHandler := hanlder.RegistryHandler{}
	adapterInfo := adpter.AdapterImpl{AdapterCMD: fkcommon.RPC}
	adapterInfo.AddFilter(&adapterInfo)
	registryHandler.MessageHandlerImpl.Adapter = adapterInfo
	newSessionManagerBuilder := fkcontext.NewSessionManagerBuilder(false, &registryHandler)
	parentContext := context.Background()
	cc.ConnectionManager = newSessionManagerBuilder
	defer parentContext.Done()
	for {
		err := client.BuildNewConnClient().Conn(network, bindAddr, newSessionManagerBuilder, parentContext)
		if err != nil {
			continue
		}
		myLogger.Error("registry disconnected retry....")
		time.Sleep(10 * time.Second)
	}
}

func (cc *ConsumerClientImpl) RegisterService(service *common.Service) error {
	go func() {
		for {
			manager := cc.ConnectionManager.Sessions
			if len(manager) > 0 {
				break
			}
		}
		jsonData, errJson := json.Marshal(service)
		if errJson != nil {
			return
		}
		fkRequest := fkcommon.Request{Cmd: fkcommon.RPC, Router: "registry.RegistryServiceImpl.RegisterService", Params: string(jsonData)}
		var session fkcontext.Session
		for _, s := range cc.ConnectionManager.Sessions {
			session = *s
			break
		}
		err := session.Handler.SendRequestMessage(&fkRequest, session.SendCh)
		if err != nil {
			return
		}

	}()
	return nil
}
