package sync

import (
	"app-frame-work/common"
	"app-frame-work/logger"
	"context"
	"errors"
	"sync"
	"time"
)

var myLogger = logger.BuildMyLogger()

type LocalRequestAndResponse struct {
	Request      *common.Request
	Response     *common.Response
	RequestTime  time.Time
	ResponseTime time.Time
}
type LocalSyncRequestCache struct {
	LocalResponseCache map[string]*LocalRequestAndResponse //requestId
	sync.Mutex
}

type SendChanSync struct {
	SendChan chan *common.Request
}

func (hdl *SendChanSync) SendRequestMessage(message *common.Request) error {
	select {
	case hdl.SendChan <- message:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("响应发送超时")
	}
}

var SendChanSyncF = SendChanSync{SendChan: make(chan *common.Request, 1024)}

var RequestMessageCache = LocalSyncRequestCache{LocalResponseCache: make(map[string]*LocalRequestAndResponse)}

var RPCRequestMessageCache = LocalSyncRequestCache{LocalResponseCache: make(map[string]*LocalRequestAndResponse)}

func (lc *LocalSyncRequestCache) AddRequest(request *common.Request) {
	requestId := request.RequestID
	RequestMessageCache.Lock()
	defer RequestMessageCache.Unlock()
	RequestMessageCache.LocalResponseCache[requestId] = &LocalRequestAndResponse{Request: request, RequestTime: time.Now()}
}

func (lc *LocalSyncRequestCache) AddResponse(response *common.Response) {
	requestId := response.Request.RequestID
	RequestMessageCache.Lock()
	defer RequestMessageCache.Unlock()
	localRequestAndResponse := RequestMessageCache.LocalResponseCache[requestId]
	if localRequestAndResponse == nil {
		return
	}
	localRequestAndResponse.Response = response
	localRequestAndResponse.ResponseTime = time.Now()

}

func (lc *LocalSyncRequestCache) ReadResponse(requestId string) *common.Response {
	RequestMessageCache.Lock()
	defer RequestMessageCache.Unlock()
	localRequestAndResponse := RequestMessageCache.LocalResponseCache[requestId]
	if localRequestAndResponse == nil {
		return nil
	}
	return localRequestAndResponse.Response
}

func (lc *LocalSyncRequestCache) DeleteRequest(requestId string) {
	RequestMessageCache.Lock()
	defer RequestMessageCache.Unlock()
	delete(RequestMessageCache.LocalResponseCache, requestId)
}

func (lc *LocalSyncRequestCache) GetResponse(requestId string, ctx context.Context, duration time.Duration) (*common.Response, error) {
	if duration <= 0 {
		duration = time.Millisecond * 2000
	}
	childCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	defer lc.DeleteRequest(requestId)
	ticker := time.NewTicker(10 * time.Millisecond) // 检查间隔
	defer ticker.Stop()
	for {
		select {
		case <-childCtx.Done():
			myLogger.Error("wait response timeout 5s, requestId: %s", requestId)
			return nil, childCtx.Err()
		case <-ticker.C:
			if response := lc.ReadResponse(requestId); response != nil {
				return response, nil
			}
		}
	}
}
