package rpc

import (
	"app-frame-work/util"
	"context"
	"errors"
	"reflect"
	"sync"
)

type RPC interface {
	Register(moduleName string, obj interface{}) error

	RPC(moduleName string, obj string, methodName string, params interface{}) (interface{}, error)

	JSONRPC(moduleName string, obj string, methodName string, ctx context.Context, param string) (interface{}, bool, error)
}

var (
	ServiceList = NewServiceContext()
)

type ServiceContext struct {
	mu           sync.RWMutex
	ServiceTable map[string]map[string]map[string]interface{}
}

func NewServiceContext() *ServiceContext {
	return &ServiceContext{
		ServiceTable: make(map[string]map[string]map[string]interface{}),
	}
}
func (rc *ServiceContext) Register(moduleName string, obj interface{}) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	t := reflect.TypeOf(obj)
	objName := t.Elem().Name()
	// 初始化第一级
	if _, ok := ServiceList.ServiceTable[moduleName]; !ok {
		ServiceList.ServiceTable[moduleName] = make(map[string]map[string]interface{})
	}

	// 初始化第二级
	if _, ok := ServiceList.ServiceTable[moduleName][objName]; !ok {
		ServiceList.ServiceTable[moduleName][objName] = make(map[string]interface{})
	}
	// 遍历所有方法
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		ServiceList.ServiceTable[moduleName][objName][method.Name] = obj
	}
	return nil
}

func (rc *ServiceContext) RPC(moduleName string, obj string, methodName string, params interface{}) (interface{}, error) {
	t := reflect.TypeOf(obj)
	rpcService, ok := GetFrom3LevelMap(moduleName, t.Name(), methodName)
	if !ok {
		return nil, errors.New("not found service")
	}
	return util.DynamicInvoke(rpcService, methodName, params)
}

func (rc *ServiceContext) JSONRPC(moduleName string, obj string, methodName string, ctx context.Context, param string) (interface{}, bool, error) {
	rpcService, ok := GetFrom3LevelMap(moduleName, obj, methodName)
	if !ok {
		return nil, false, errors.New("not found service")
	}
	result, err := util.JSONRPCWithCtx(rpcService, methodName, ctx, param)
	return result, true, err
}

func GetFrom3LevelMap(moduleName string, obj string, methodName string) (interface{}, bool) {
	m := ServiceList.ServiceTable
	// 检查第一级
	level1, ok := m[moduleName]
	if !ok {
		return level1, false
	}

	// 检查第二级
	level2, ok := level1[obj]
	if !ok {
		return level2, false
	}
	// 检查第3级
	level3, ok := level2[methodName]
	if !ok {
		return level3, false
	}
	// 检查第三级
	return level3, ok
}
