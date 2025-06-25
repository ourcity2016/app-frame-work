package server

import (
	"app-frame-work/common"
	"app-frame-work/config"
	"app-frame-work/context"
	"app-frame-work/logger"
	"app-frame-work/util"
	"bufio"
	oscontext "context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var myLogger = logger.BuildMyLogger()

type Server interface {
	Listen(config *config.AppConfig, sessions *context.ConnectionManager, ctx oscontext.Context) error
}

type TCPServer struct {
}

func BuildNewTCPServer() *TCPServer {
	return &TCPServer{}
}

func (tcpServer *TCPServer) Listen(appConfig *config.AppConfig, sessions *context.ConnectionManager, ctx oscontext.Context) error {
	serverConfig := appConfig.ServerConfig
	myLogger.Info("正在配置服务器...")
	listen, err := net.Listen(serverConfig.Network, serverConfig.BindAddr)
	if err != nil {
		return err
	}
	defer func(listen net.Listener) {
		errClose := listen.Close()
		if errClose != nil {
			myLogger.Error(errClose.Error())
			return
		}
	}(listen)
	defer gracefulShutdown(listen, sessions)
	parentContext := ctx
	myLogger.Info("服务器: %s : %s 等待客户端加入 \n", serverConfig.Network, serverConfig.BindAddr)
	for {
		conn, errConn := listen.Accept()
		if errConn != nil {
			if ne, ok := errConn.(net.Error); ok && ne.Temporary() {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return fmt.Errorf("接受连接错误: %w", err)
		}
		session := &context.Session{
			Conn:          conn,
			Reader:        bufio.NewReaderSize(conn, 8192),
			Writer:        bufio.NewWriterSize(conn, 8192),
			SendCh:        make(chan []byte, 100),
			Filters:       appConfig.Filters,
			ConnID:        util.UUID(),
			IgnoreRouters: appConfig.IgnoreRouters,
			IsServer:      sessions.IsServer,
			Handler:       sessions.Handler,
		}
		session.Context = oscontext.WithValue(parentContext, common.ContextUserKey, session)
		go func() {
			defer session.Context.Done()
			sessions.AddSession(session)
			context.HandlerClientSession(session)
			myLogger.Warn("Shutdown session %s", session.ConnID)
			sessions.RemoveSession(session.Conn.RemoteAddr().String())
		}()

	}
}

func gracefulShutdown(listener net.Listener, manager *context.ConnectionManager) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	myLogger.Info("接收到关闭信号，开始优雅关闭...")

	// 停止接受新连接
	listener.Close()

	// 关闭所有现有连接
	manager.RLock()
	defer manager.RUnlock()

	var wg sync.WaitGroup
	for _, session := range manager.Sessions {
		wg.Add(1)
		go func(s *context.Session) {
			defer wg.Done()
			session.Conn.Close()
			<-session.Done // 等待会话完全结束
		}(session)
	}

	wg.Wait()
	myLogger.Info("所有连接已关闭，服务退出")
}
