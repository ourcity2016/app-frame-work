package handler

import (
	"app-frame-work/common"
	"app-frame-work/sync"
	"encoding/json"
	"errors"
)

type RPCClientHandler struct {
	MessageHandlerImpl
}

func (hdl *RPCClientHandler) SendRequestSyncMessage(request *common.Request, c chan []byte) error {
	dataByte, err := json.Marshal(request)
	if err != nil {
		return err
	}
	if request.RequestID == "" {
		return errors.New("request id must be not null")
	}
	errRe := hdl.SendByteMessage(dataByte, c)
	if errRe != nil {
		return errRe
	}
	sync.RPCRequestMessageCache.AddRequest(request)
	return nil
}

func (hdl *RPCClientHandler) HandlerResponseMessage(response *common.Response, c chan []byte) error {
	sync.RPCRequestMessageCache.AddResponse(response)
	myLogger.Debug("return: %s", response)
	return nil
}
