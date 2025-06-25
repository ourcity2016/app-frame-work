package service

import (
	"app-frame-work/registry/common"
	"app-frame-work/util"
	"context"
	"errors"
)

type RegistryService interface {
	RegisterService(context.Context, *common.Service) error

	GetFullService(context.Context) *common.ServiceMap

	RemoveService(context.Context, *common.Service) error

	ServiceCheck(context.Context, *common.Service) string
}

type RegistryServiceImpl struct {
}

func (rst *RegistryServiceImpl) RegisterService(ctx context.Context, service *common.Service) error {
	serviceMap := common.ServiceDiscover
	serviceMap.RLock()
	defer serviceMap.RUnlock()
	moduleName := service.AppName
	// 初始化第一级
	if _, ok := serviceMap.ServiceMap[moduleName]; !ok {
		serviceMap.ServiceMap[moduleName] = make(map[string]map[string]*common.Service)
	}
	serviceName := service.ServiceName
	// 初始化第二级
	if _, ok := serviceMap.ServiceMap[moduleName][serviceName]; !ok {
		serviceMap.ServiceMap[moduleName][serviceName] = make(map[string]*common.Service)
	}
	methodName := service.MethodName
	serviceInfo, ok := serviceMap.ServiceMap[moduleName][serviceName][methodName]
	if ok {
		for _, serversData := range service.Servers {
			serviceInfo.Servers[serversData.Ip+serversData.Port] = serversData
		}
	} else {
		serviceMap.ServiceMap[moduleName][serviceName][methodName] = service
	}
	serviceMap.ServiceHash = util.UUID()
	return nil
}

func (rst *RegistryServiceImpl) GetFullService(ctx context.Context, service *common.Service) *common.ServiceMap {
	return &common.ServiceDiscover
}

func (rst *RegistryServiceImpl) RemoveService(ctx context.Context, service *common.Service) error {
	serviceMap := common.ServiceDiscover
	serviceMap.RLock()
	defer serviceMap.RUnlock()
	moduleName := service.AppName
	if _, ok := serviceMap.ServiceMap[moduleName]; !ok {
		return errors.New("module not exist")
	}
	serviceName := service.ServiceName
	if _, ok := serviceMap.ServiceMap[moduleName][serviceName]; !ok {
		return errors.New("service  not exist")
	}
	methodName := service.MethodName
	serviceInfo := serviceMap.ServiceMap[moduleName][serviceName][methodName]
	if serviceInfo == nil {
		return errors.New("service impl not exist")
	}
	servers := serviceInfo.Servers
	serverNeedRemove := service.Servers
	if len(servers) > 0 && len(serverNeedRemove) > 0 {
		for _, serversData := range serverNeedRemove {
			delete(servers, serversData.Ip+serversData.Port)
		}
	} else {
		delete(serviceMap.ServiceMap[moduleName][serviceName], methodName)
	}
	serviceMap.ServiceHash = util.UUID()
	return nil
}

func (rst *RegistryServiceImpl) ServiceCheck(ctx context.Context, serviceHash *common.Service) string {
	return common.ServiceDiscover.ServiceHash
}
