package hanlder

import (
	"app-frame-work/handler"
	"app-frame-work/logger"
)

var myLogger = logger.BuildMyLogger()

type RegistryHandler struct {
	handler.MessageHandlerImpl
}

func (hdl *RegistryHandler) HandlerMessage(request []byte, c chan []byte) error {
	myLogger.Debug("halo: %s", request)
	return nil
}
