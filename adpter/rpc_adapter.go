package adpter

import (
	"app-frame-work/common"
	"app-frame-work/logger"
	"app-frame-work/rpc"
	"app-frame-work/sync"
	"errors"
	"strings"
	"time"
)

var myLogger = logger.BuildMyLogger()

type Adapter interface {
	DoAdapter(request *common.Request) (interface{}, bool, error)

	AdapterMatch(request *common.Request) bool
}

type AdapterImpl struct {
	AdapterCMD string
	Adapters   []Adapter
}

func (adapter *AdapterImpl) AddAdapter(ad Adapter) {
	adapter.Adapters = append(adapter.Adapters, ad)
}
func (adapter *AdapterImpl) DoAdapter(request *common.Request) (interface{}, bool, error) {
	if !adapter.AdapterMatch(request) {
		return nil, false, nil
	}
	routerString := request.Router
	parts := strings.Split(routerString, ".")
	result, find, err := rpc.NewServiceContext().JSONRPC(parts[0], parts[1], parts[2], request.Ctx, request.Params)
	////远程调用RPC 需要同步等待
	if !find {
		err := sync.SendChanSyncF.SendRequestMessage(request)
		if err != nil {
			return nil, false, err
		}
		response, rpcError := sync.RPCRequestMessageCache.GetResponse(request.RequestID, request.Ctx, time.Second*5)
		if rpcError != nil {
			return response, true, rpcError
		}
		return response, true, nil
	}
	return result, true, err
}

func (adapter *AdapterImpl) AdapterMatch(request *common.Request) bool {
	return adapter.AdapterCMD == request.Cmd
}

func (adapter *AdapterImpl) Execute(input *common.Request) (interface{}, bool, error) {
	for _, f := range adapter.Adapters {
		output, findAdapter, err := f.DoAdapter(input)
		if findAdapter {
			return output, findAdapter, err
		}
	}
	myLogger.Warn("adapter not found for %v", input)
	return nil, false, errors.New("service can not exec your command please check")
}
