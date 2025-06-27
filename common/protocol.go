package common

import (
	"context"
)

// 登录
// {"cmd":"rpc","router":"gateway.LoginServiceImpl.Login","params":"{\"name\":\"xxxxxx1\"}"}
// 注销登录
// {"cmd":"rpc","router":"gateway.LoginServiceImpl.LoginOut","params":"{\"name\":\"xxxxxx\"}"}
// 加入房间
// {"cmd":"rpc","router":"chat-bizserver.RoomServiceImpl.JoinRoom","params":"{\"roomId\":\"aa414d89-d2e9-eef7-18d9-4351bdd68516\"}"}
// 创建房间
// {"cmd":"rpc","router":"chat-bizserver.RoomServiceImpl.CreateRoom","params":"{\"roomName\":\"xxxxxx\"}"}
// Talk
// {"cmd":"rpc","router":"chat-bizserver.RoomServiceImpl.Talk","params":"{\"toRoom\":\"xxxxxx\",\"message\":\"你好\"}"}
// {"cmd":"rpc","router":"registry.RegistryServiceImpl.RegisterService","params":"{\"toRoom\":\"xxxxxx\",\"message\":\"你好\"}"}
// {"cmd":"rpc","router":"registry.RegistryServiceImpl.GetFullService","params":"{\"toRoom\":\"xxxxxx\",\"message\":\"你好\"}"}
// {"cmd":"rpc","router":"registry.RegistryServiceImpl.ServiceCheck","params":"{\"toRoom\":\"xxxxxx\",\"message\":\"你好\"}"}

// {"cmd":"rpc","session_id":"","router":"registry.RegistryServiceImpl.RegisterService","params":"{\"methodName\":\"Test\",\"serviceName\":\"TestServiceImpl\",\"appName\":\"demo\",\"loadBalance\":\"round-robin\",\"servers\":{},\"status\":1}}
const (
	RPC            = "rpc"
	OK             = "ok"
	ContextUserKey = "USER_KEY"
)

type Request struct {
	Cmd         string            `json:"cmd"` //PING RPC
	SessionID   string            `json:"sessionId"`
	RequestID   string            `json:"requestId"`
	Router      string            `json:"router"`
	Params      string            `json:"params"`
	Ctx         context.Context   `json:"-"`
	OriginalMsg string            `json:"originalMsg"`
	Headers     map[string]string `json:"Headers"`
}

type Response struct {
	Request Request
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
	Code    int16       `json:"code"`
	Message string      `json:"message"`
}

func OkWithNil() *Response {
	return &Response{Status: "success", Data: nil, Code: 200, Message: ""}
}

func Ok(data interface{}, message string) *Response {
	return &Response{Status: "success", Data: data, Code: 200, Message: message}
}
func ERRORWithNil() *Response {
	return &Response{Status: "fault", Data: nil, Code: 500, Message: ""}
}
func ERROR(data interface{}, message string) *Response {
	return &Response{Status: "fault", Data: data, Code: 500, Message: message}
}
func ERRORWithoutLogin(message string) *Response {
	return &Response{Status: "fault", Data: nil, Code: 401, Message: message}
}
func ERRORWithoutPermission(message string) *Response {
	return &Response{Status: "fault", Data: nil, Code: 403, Message: message}
}
