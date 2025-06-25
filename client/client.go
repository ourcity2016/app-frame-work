package client

import (
	"app-frame-work/common"
	"app-frame-work/config"
	"app-frame-work/context"
	"app-frame-work/logger"
	"app-frame-work/util"
	"bufio"
	oscontext "context"
	"net"
)

var myLogger = logger.BuildMyLogger()

type Client interface {
	Conn(*config.AppConfig, *context.ConnectionManager, oscontext.Context) error
}

type ConnClient struct {
}

func BuildNewConnClient() *ConnClient {
	return &ConnClient{}
}

func (c *ConnClient) Conn(network string, connAddr string, sessions *context.ConnectionManager, parentContext oscontext.Context) error {
	conn, err := net.Dial(network, connAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	myLogger.Info("与服务器建立连接: %s %s \n", network, connAddr)
	session := &context.Session{
		Conn:    conn,
		Reader:  bufio.NewReaderSize(conn, 8192),
		Writer:  bufio.NewWriterSize(conn, 8192),
		SendCh:  make(chan []byte, 100),
		ConnID:  util.UUID(),
		Handler: sessions.Handler,
	}
	session.Context = oscontext.WithValue(parentContext, common.ContextUserKey, session)
	defer session.Context.Done()
	sessions.AddSession(session)
	context.HandlerClientSession(session)
	myLogger.Warn("shutdown client session %s", session.ConnID)
	sessions.RemoveSession(session.Conn.RemoteAddr().String())
	return nil
}
