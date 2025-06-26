package hanlder

import (
	"app-frame-work/common"
	"app-frame-work/handler"
	"app-frame-work/logger"
	"app-frame-work/sync"
)

var myLogger = logger.BuildMyLogger()

type RegistryHandler struct {
	handler.MessageHandlerImpl
}

func (hdl *RegistryHandler) HandlerResponseMessage(response *common.Response, c chan []byte) error {
	sync.RequestMessageCache.AddResponse(response)
	myLogger.Debug("return: %s", response)
	return nil
}
