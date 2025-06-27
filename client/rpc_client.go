package client

import (
	"app-frame-work/common"
	fkcontext "app-frame-work/context"
	registrycommon "app-frame-work/registry/common"
	"app-frame-work/sync"
	"context"
	"errors"
	"strings"
	"time"
)

type RegistryCommon struct {
	Ip   string
	Port string
}
type RemoteRPC interface {
	RemoteRPC(*common.Request) (interface{}, bool, error)
	CheckConnectManagerReady() bool
	StartRPCClient()
	CheckNeedInitConn() bool
}
type RemoteRPCImpl struct {
	Ctx               context.Context
	ConnectionManager *fkcontext.ConnectionManager
	LocalServers      map[string]*registrycommon.Server
	ServerHasChange   bool
}

func (rr *RemoteRPCImpl) CheckNeedInitConn(registryServer *registrycommon.Server) bool {
	session := rr.ConnectionManager.FindAnySession(registryServer.Ip + ":" + registryServer.Port)
	return session != nil
}

func (rr *RemoteRPCImpl) StartRPCClient() {
	defer rr.Ctx.Done()
	connClient := BuildNewConnClient()
	myLogger.Debug("start rpc client service ....")
	go func() {
		err := rr.GetMessageChan()
		if err != nil {
			myLogger.Debug("get message chan error %s", err.Error())
			return
		}
	}()
	for {
		if !rr.ServerHasChange {
			continue
		}
		localServer := rr.LocalServers
		if len(localServer) <= 0 {
			continue
		}
		myLogger.Debug("init rpc connect ....")
		for _, s := range localServer {
			ifFind := rr.CheckNeedInitConn(s)
			if ifFind {
				continue
			}
			bindAddr := s.Ip + ":" + s.Port
			network := "tcp"
			go func() {
				cdCtx := context.WithoutCancel(rr.Ctx)
				defer cdCtx.Done()
				_ = connClient.Conn(network, bindAddr, rr.ConnectionManager, cdCtx)
				myLogger.Error("rpc client disconnected retry....")
			}()
		}
		rr.ServerHasChange = false
	}
}

func (rr *RemoteRPCImpl) RemoteRPC(req *common.Request) (interface{}, bool, error) {
	routerString := req.Router
	parts := strings.Split(routerString, ".")
	//查找服务
	remoteServiceMap := registrycommon.ServiceDiscover.ServiceMap
	_, ok := remoteServiceMap[parts[0]]
	if !ok {
		return nil, false, nil
	}
	_, ok1 := remoteServiceMap[parts[0]][parts[1]]
	if !ok1 {
		return nil, false, nil
	}
	_, ok2 := remoteServiceMap[parts[0]][parts[1]][parts[2]]
	if !ok2 {
		return nil, false, nil
	}
	service := remoteServiceMap[parts[0]][parts[1]][parts[2]]
	if service == nil {
		return nil, false, nil
	}
	servers := service.Servers
	for _, sv := range servers {
		session := rr.ConnectionManager.FindAnySession(sv.Ip + ":" + sv.Port)
		if session == nil {
			rr.ServerHasChange = true
			rr.LocalServers[sv.Ip+":"+sv.Port] = &sv
			getSession, err := rr.GetSession(context.WithoutCancel(rr.Ctx), &sv, time.Second*5)
			if err != nil {
				return nil, false, err
			}
			session = getSession
		}
		if session != nil {
			err := session.Handler.SendRequestSyncMessage(req, session.SendCh)
			if err != nil {
				return nil, true, err
			}
		}
	}
	return nil, false, nil
}
func (rr *RemoteRPCImpl) CheckConnectManagerReady() bool {
	if rr.ConnectionManager != nil && rr.ConnectionManager.Sessions != nil {
		manager := rr.ConnectionManager.Sessions
		if len(manager) > 0 {
			return true
		}
	}
	return false
}

func (rr *RemoteRPCImpl) GetSession(ctx context.Context, server *registrycommon.Server, duration time.Duration) (*fkcontext.Session, error) {
	if duration <= 0 {
		duration = time.Millisecond * 2000
	}
	childCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond) // 检查间隔
	defer ticker.Stop()
	for {
		select {
		case <-childCtx.Done():
			myLogger.Error("wait Session timeout 5s, requestId: %s")
			return nil, childCtx.Err()
		case <-ticker.C:
			session := rr.ConnectionManager.FindAnySession(server.Ip + ":" + server.Port)
			if session != nil {
				return session, nil
			}
		}
	}
}

func (rr *RemoteRPCImpl) GetMessageChan() error {
	for {
		select {
		case msg, ok := <-sync.SendChanSyncF.SendChan:
			if !ok {
				return errors.New("读循环已关闭通道") // 读循环已关闭通道
			}
			sync.RPCRequestMessageCache.AddRequest(msg)
			_, _, err := rr.RemoteRPC(msg)
			if err != nil {
				return err
			}
			//case <-time.After(100 * time.Millisecond):
		}
	}
}
