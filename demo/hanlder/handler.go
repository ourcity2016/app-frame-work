package hanlder

import (
	"app-frame-work/handler"
	"app-frame-work/logger"
)

var myLogger = logger.BuildMyLogger()

type MyHandler struct {
	handler.MessageHandlerImpl
}

func (hdl *MyHandler) HandlerMessage(request []byte, c chan []byte) error {
	myLogger.Debug("halo: %s", request)
	return nil
}
