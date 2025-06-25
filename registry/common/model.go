package common

import (
	"sync"
)

var ServiceDiscover = *BuildNewServiceMap()

type ServiceMap struct {
	ServiceMap   map[string]map[string]map[string]*Service `json:"serviceMap"`
	ServiceHash  string                                    `json:"serviceHash"`
	sync.RWMutex `json:"-"`
}

type Service struct {
	MethodName  string            `json:"methodName"`
	ServiceName string            `json:"serviceName"`
	AppName     string            `json:"appName"`
	LoadBalance string            `json:"loadBalance"` //-round-robin -random
	Servers     map[string]Server `json:"servers"`
	Status      int8              `json:"status"` //0 disabled 1 available 2 unavailable
}

type Server struct {
	Ip     string `json:"ip"`
	Port   string `json:"port"`
	Status int8   `json:"status"` //0 disabled 1 available 2 unavailable
}

func BuildNewServiceMap() *ServiceMap {
	return &ServiceMap{ServiceMap: make(map[string]map[string]map[string]*Service)}
}
