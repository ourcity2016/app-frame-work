package context

import (
	"app-frame-work/common"
	"app-frame-work/filters"
	"app-frame-work/handler"
	"app-frame-work/logger"
	"app-frame-work/util"
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var myLogger = logger.BuildMyLogger()

type ConnectionManager struct {
	sync.RWMutex
	Sessions map[string]*Session
	IsServer bool
	Handler  handler.MessageHandler
}

type Session struct {
	Conn          net.Conn
	Reader        *bufio.Reader
	Writer        *bufio.Writer
	userName      string
	SendCh        chan []byte
	Done          chan struct{}
	Filters       filters.Filters
	IgnoreRouters []string
	ConnID        string //UUID
	Context       context.Context
	IsServer      bool
	Handler       handler.MessageHandler
}

func NewSessionManagerBuilder(isServer bool, handler handler.MessageHandler) *ConnectionManager {
	return &ConnectionManager{
		Sessions: map[string]*Session{},
		IsServer: isServer,
		Handler:  handler,
	}
}
func (s *Session) IgnoreFilters(requestRouter string) bool {
	ignoreList := s.IgnoreRouters
	for _, ignore := range ignoreList {
		if ignore == requestRouter {
			return true
		}
	}
	return false
}
func (s *Session) readLoop() {
	defer close(s.SendCh) // 关闭发送通道
	for {
		msg, err := s.decodeMessage()
		if err != nil {
			if err == io.EOF {
				myLogger.Error("客户端 %s 断开连接", s.Conn.RemoteAddr())
			} else {
				myLogger.Error("读取错误: %v", err)
			}
			return
		}
		go func() {
			if s.IsServer {
				request := common.Request{SessionID: s.ConnID, Ctx: s.Context}
				request.OriginalMsg = string(msg)
				errJson := json.Unmarshal([]byte(msg), &request)
				if request.RequestID == "" {
					request.RequestID = util.UUID()
				}
				if errJson != nil || request.Cmd == "" {
					errRes := common.ERROR(nil, fmt.Sprintf("unkown command: %s", string(msg)))
					errRes.Request = request
					err := s.Handler.SendResponseMessage(errRes, s.SendCh)
					if err != nil {
						return
					}
					return
				}
				ignoreFilter := s.IgnoreFilters(request.Router)
				if ignoreFilter {
					errInfo := s.Handler.HandlerRequestMessage(&request, s.SendCh)
					if errInfo != nil {
						return
					}
				} else {
					result, continued := s.Filters.Execute(&request)
					if continued {
						errInfo := s.Handler.HandlerRequestMessage(&request, s.SendCh)
						if errInfo != nil {
							return
						}
					} else {
						errInfo := s.Handler.SendResponseMessage(result, s.SendCh)
						if errInfo != nil {
							return
						}
					}
				}
			} else {
				response := common.Response{}
				errJson := json.Unmarshal([]byte(msg), &response)
				if errJson != nil {
					myLogger.Error("response msg error %s", string(msg))
					return
				}
				myLogger.Debug("response from server: %s", string(msg))
				errInfo := s.Handler.HandlerResponseMessage(&response, s.SendCh)
				if errInfo != nil {
					return
				}
			}
		}()
	}
}

func (s *Session) decodeMessage() ([]byte, error) {
	// 读取并消费长度前缀
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(s.Reader, lengthBytes); err != nil {
		return nil, fmt.Errorf("读取长度前缀失败: %w", err)
	}
	length := binary.LittleEndian.Uint32(lengthBytes)
	// 更严格的大小限制检查
	const maxMessageSize = 10 * 1024 * 1024
	if length == 0 {
		myLogger.Debug("丢弃无效包数据")
		return nil, nil
	}
	if length > maxMessageSize {
		myLogger.Debug("消息长度 %d 超过限制 %d", length, maxMessageSize)
		return nil, nil
	}
	// 读取消息体
	message := make([]byte, length)
	if _, err := io.ReadFull(s.Reader, message); err != nil {
		myLogger.Debug("读取消息体失败: %w", err)
		return nil, nil
	}
	if len(message) <= 4 {
		myLogger.Debug("丢弃无效包数据: %s", string(message))
		return message, nil
	}
	return message[4:len(message)], nil
}
func (s *Session) writeLoop() {
	defer func() {
		// 清空发送队列
		for range s.SendCh {
			// 丢弃未发送的消息
		}
	}()

	for {
		select {
		case msg, ok := <-s.SendCh:
			if !ok {
				return // 读循环已关闭通道
			}

			if _, err := s.Writer.Write(util.EncodeMessage(msg)); err != nil {
				myLogger.Error("写入错误: %v", err)
				return
			}

			// 批量刷新写入
			if len(s.SendCh) == 0 {
				if err := s.Writer.Flush(); err != nil {
					myLogger.Error("刷新错误: %v", err)
					return
				}
			}

		case <-time.After(100 * time.Millisecond):
			// 定期刷新缓冲区
			if err := s.Writer.Flush(); err != nil {
				myLogger.Error("定时刷新错误: %v", err)
				return
			}
		}
	}
}

func (s *Session) heartBeat() {
	const (
		defaultHeartbeatInterval = 10 * time.Second
	)

	heartbeatTicker := time.NewTicker(defaultHeartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-heartbeatTicker.C:
			// 尝试发送心跳
			err := s.sendHeartbeat()
			if err != nil {
				// 延迟重试
				return
			}

		case <-s.SendCh:
			myLogger.Info("收到关闭信号，停止心跳检测,session:%s", s.ConnID)
			return
		}
	}
}

func (s *Session) sendHeartbeat() error {
	select {
	case <-s.SendCh:
		myLogger.Warn("通道已关闭 session:%s", s.ConnID)
		return errors.New("通道已关闭")
	case s.SendCh <- util.EncodeMessage([]byte("{\"cmd\":\"PING\"}")):
		return nil
	default:
		return fmt.Errorf("发送队列已满")
	}
}

func (m *ConnectionManager) AddSession(session *Session) {
	m.Lock()
	defer m.Unlock()
	addr := session.Conn.RemoteAddr().String()
	myLogger.Info("新客户端 %s 连接加入\n", addr)
	m.Sessions[session.ConnID] = session
}

func (m *ConnectionManager) FindAnySession(serverAndPort string) *Session {
	m.Lock()
	defer m.Unlock()
	for _, s := range m.Sessions {
		if s.Conn.RemoteAddr().String() == serverAndPort {
			return s
		}
	}
	return nil
}

func (m *ConnectionManager) RemoveSession(connId string) {
	m.Lock()
	defer m.Unlock()
	myLogger.Info("客户端 %s 被移除\n", connId)
	delete(m.Sessions, connId)
}

func (m *ConnectionManager) Broadcast(msg []byte) {
	m.RLock()
	defer m.RUnlock()
	for _, session := range m.Sessions {
		select {
		case session.SendCh <- msg:
		default:
			myLogger.Error("广播到 %s 失败，发送队列满", session.Conn.RemoteAddr())
		}
	}
}

func HandlerClientSession(session *Session) {
	// 启动读写协程
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		session.readLoop()
		myLogger.Warn("session.readLoop end session:%s", session.ConnID)
	}()

	go func() {
		defer wg.Done()
		session.writeLoop()
		myLogger.Warn("session.writeLoop end session:%s", session.ConnID)
	}()

	go func() {
		defer wg.Done()
		//session.heartBeat()
		myLogger.Warn("session.heartBeat end session:%s", session.ConnID)
	}()
	wg.Wait()
}
