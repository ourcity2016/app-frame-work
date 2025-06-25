package handler

import (
	"app-frame-work/adpter"
	"app-frame-work/common"
	"app-frame-work/logger"
	"app-frame-work/util"
	"encoding/json"
	"errors"
	"time"
)

var myLogger = logger.BuildMyLogger()

type MessageHandler interface {
	HandlerRequestMessage(request *common.Request, c chan []byte) error
	HandlerResponseMessage(response *common.Response, c chan []byte) error
	SendByteMessage(message []byte, c chan []byte) error
	SendMessage(message string, c chan []byte) error
	SendResponseMessage(response *common.Response, c chan []byte) error
	SendRequestMessage(request *common.Request, c chan []byte) error
}

type MessageHandlerImpl struct {
	Adapter adpter.AdapterImpl
}

func BuildMessageHandler() MessageHandler {
	return &MessageHandlerImpl{}
}
func (hdl *MessageHandlerImpl) SendByteMessage(message []byte, c chan []byte) error {
	select {
	case c <- util.EncodeMessage(message):
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("响应发送超时")
	}
}
func (hdl *MessageHandlerImpl) SendMessage(message string, c chan []byte) error {
	return hdl.SendByteMessage([]byte(message), c)
}
func (hdl *MessageHandlerImpl) SendResponseMessage(response *common.Response, c chan []byte) error {
	dataByte, err := json.Marshal(response)
	if err != nil {
		return err
	}
	return hdl.SendByteMessage(dataByte, c)
}
func (hdl *MessageHandlerImpl) SendRequestMessage(request *common.Request, c chan []byte) error {
	dataByte, err := json.Marshal(request)
	if err != nil {
		return err
	}
	return hdl.SendByteMessage(dataByte, c)
}

// {cmd:"request",router:"",params:{}}
func (hdl *MessageHandlerImpl) HandlerRequestMessage(request *common.Request, c chan []byte) error {
	myLogger.Debug("正在处理请求业务内容: %v", request)
	result, _, err := hdl.Adapter.Execute(request)
	if err != nil {
		response := common.ERROR(nil, err.Error())
		response.Request = *request
		errData := hdl.SendResponseMessage(response, c)
		if errData != nil {
			return errData
		}
		return nil
	}
	response := common.Ok(result, "")
	response.Request = *request
	errData := hdl.SendResponseMessage(response, c)
	if errData != nil {
		return errData
	}
	return nil
}
func (hdl *MessageHandlerImpl) HandlerResponseMessage(response *common.Response, c chan []byte) error {
	myLogger.Debug("服务器返回数据处理中: %v", response)
	return nil
}
