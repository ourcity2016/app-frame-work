package client

import (
	"app-frame-work/client"
	fkcommon "app-frame-work/common"
	"app-frame-work/config"
	fkcontext "app-frame-work/context"
	"app-frame-work/logger"
	"app-frame-work/registry/common"
	"app-frame-work/sync"
	"app-frame-work/util"
	"context"
	"encoding/json"
	"time"
)

var myLogger = logger.BuildMyLogger()

type ConsumerClient interface {
	InitRegistry(*config.AppConfig, *fkcontext.ConnectionManager) error
	CheckConnectManagerReady() bool
	RegisterService(*common.Service) error
	PullService(*common.Service) (*common.ServiceMap, error)
	CheckService() (*common.ServiceMap, error)
	MonitorService() error
}

type ConsumerClientImpl struct {
	ConnectionManager *fkcontext.ConnectionManager
}

func (cc *ConsumerClientImpl) InitRegistry(config *config.AppConfig, connectionManager *fkcontext.ConnectionManager) error {
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
	parentContext := context.Background()
	cc.ConnectionManager = connectionManager
	defer parentContext.Done()
	cc.MonitorService()
	for {
		_ = client.BuildNewConnClient().Conn(network, bindAddr, connectionManager, parentContext)
		myLogger.Error("registry disconnected retry....")
		time.Sleep(10 * time.Second)
	}
}

func (cc *ConsumerClientImpl) RegisterService(service *common.Service) error {
	go func() {
		for {
			if cc.CheckConnectManagerReady() {
				break
			}
			time.Sleep(10 * time.Second)
		}
		jsonData, errJson := json.Marshal(service)
		if errJson != nil {
			return
		}
		fkRequest := fkcommon.Request{RequestID: util.UUID(), Cmd: fkcommon.RPC, Router: "registry.RegistryServiceImpl.RegisterService", Params: string(jsonData)}
		var session fkcontext.Session
		for _, s := range cc.ConnectionManager.Sessions {
			session = *s
			break
		}
		err := session.Handler.SendRequestSyncMessage(&fkRequest, session.SendCh)
		if err != nil {
			return
		}
		_, err = sync.RequestMessageCache.GetResponse(fkRequest.RequestID, session.Context, 5*time.Second)
		if err != nil {
			myLogger.Error("Register Service Error %v", fkRequest)
			return
		}
	}()
	return nil
}

func (cc *ConsumerClientImpl) PullService(service *common.Service) (*common.ServiceMap, error) {
	fkRequest := fkcommon.Request{RequestID: util.UUID(), Cmd: fkcommon.RPC, Router: "registry.RegistryServiceImpl.GetFullService", Params: "{}"}
	var session fkcontext.Session
	for _, s := range cc.ConnectionManager.Sessions {
		session = *s
		break
	}
	err := session.Handler.SendRequestSyncMessage(&fkRequest, session.SendCh)
	if err != nil {
		return nil, err
	}
	response, err := sync.RequestMessageCache.GetResponse(fkRequest.RequestID, session.Context, 5*time.Second)
	if err != nil {
		myLogger.Error("Pull Service Error %v", fkRequest)
		return nil, err
	}
	data := response.Data
	resultData := util.JsonInterfaceToString(data)
	serviceMap := &common.ServiceMap{}
	err = json.Unmarshal([]byte(resultData), serviceMap)
	if err != nil {
		return nil, err
	}
	return serviceMap, nil
}
func (cc *ConsumerClientImpl) CheckService() (*common.ServiceMap, error) {
	fkRequest := fkcommon.Request{RequestID: util.UUID(), Cmd: fkcommon.RPC, Router: "registry.RegistryServiceImpl.ServiceCheck", Params: "{}"}
	var session fkcontext.Session
	for _, s := range cc.ConnectionManager.Sessions {
		session = *s
		break
	}
	err := session.Handler.SendRequestSyncMessage(&fkRequest, session.SendCh)
	if err != nil {
		return nil, err
	}
	response, err := sync.RequestMessageCache.GetResponse(fkRequest.RequestID, session.Context, 5*time.Second)
	if err != nil {
		myLogger.Error("ServiceCheck Service Error %v", fkRequest)
		return nil, err
	}
	data := response.Data
	resultData := util.JsonInterfaceToString(data)
	serviceMap := &common.ServiceMap{}
	err = json.Unmarshal([]byte(resultData), serviceMap)
	if err != nil {
		return nil, err
	}
	return serviceMap, nil
}
func (cc *ConsumerClientImpl) MonitorService() {
	go func() error {
		for {
			for {
				if cc.CheckConnectManagerReady() {
					break
				}
				time.Sleep(3 * time.Second)
			}
			serviceMap, _ := cc.CheckService()

			if serviceMap != nil && serviceMap.ServiceHash != common.ServiceDiscover.ServiceHash {
				serviceMapResult, _ := cc.PullService(nil)
				if serviceMapResult != nil {
					common.ServiceDiscover = *serviceMapResult
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()
}
func (cc *ConsumerClientImpl) CheckConnectManagerReady() bool {
	if cc.ConnectionManager != nil && cc.ConnectionManager.Sessions != nil {
		manager := cc.ConnectionManager.Sessions
		if len(manager) > 0 {
			return true
		}
	}
	return false
}
