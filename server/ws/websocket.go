package ws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gorilla/websocket"
)

// 常量定义：默认配置
const (
	// 默认读写缓冲区大小（字节）
	DefaultReadBufferSize  = 1024
	DefaultWriteBufferSize = 1024
	// 默认心跳间隔（秒）：每30秒发送一次心跳
	DefaultHeartbeatInterval = 30 * time.Second
	// 默认心跳超时（秒）：60秒未收到客户端心跳响应则关闭连接
	DefaultHeartbeatTimeout = 60 * time.Second
	// 默认读写超时（秒）
	DefaultReadTimeout  = 60 * time.Second
	DefaultWriteTimeout = 10 * time.Second
	// 消息类型
	MessageTypeText   = websocket.TextMessage
	MessageTypeBinary = websocket.BinaryMessage
)

// Config WebSocket服务端配置
type Config struct {
	// 读写缓冲区大小
	ReadBufferSize  int
	WriteBufferSize int
	// 跨域配置：是否允许所有跨域（生产环境建议指定Origin）
	AllowAllOrigins bool
	// 允许的跨域Origin列表（AllowAllOrigins=false时生效）
	AllowedOrigins []string
	// 心跳配置
	HeartbeatInterval time.Duration // 心跳发送间隔
	HeartbeatTimeout  time.Duration // 心跳超时时间
	// 读写超时
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	MsgType       int
	HeartbeatType string
}

// 默认配置
func DefaultConfig() *Config {
	return &Config{
		ReadBufferSize:    DefaultReadBufferSize,
		WriteBufferSize:   DefaultWriteBufferSize,
		AllowAllOrigins:   true,
		AllowedOrigins:    []string{},
		HeartbeatInterval: DefaultHeartbeatInterval,
		HeartbeatTimeout:  DefaultHeartbeatTimeout,
		ReadTimeout:       DefaultReadTimeout,
		WriteTimeout:      DefaultWriteTimeout,
		MsgType:           MessageTypeText,
		HeartbeatType:     "heartbeat",
	}
}

// Connection WebSocket连接结构体
type Connection struct {
	conn          *websocket.Conn    // 底层连接
	connID        string             // 唯一连接ID
	manager       *Manager           // 所属管理器
	createTime    time.Time          // 连接创建时间
	heartbeatChan chan struct{}      // 心跳通道（用于检测客户端响应）
	ctx           context.Context    // 上下文
	cancel        context.CancelFunc // 上下文取消函数
	writeMutex    sync.Mutex         // 写消息互斥锁（防止并发写）
}

// Manager WebSocket连接管理器
type Manager struct {
	config      *Config                // 配置
	upgrader    *websocket.Upgrader    // HTTP升级器
	connections map[string]*Connection // 所有在线连接（connID -> Connection）
	mutex       sync.RWMutex           // 读写锁（保护connections）
	// 业务回调：收到消息时触发（用户自定义处理逻辑）
	OnMessage func(connID string, msgType int, data []byte)
	// 业务回调：连接建立时触发
	OnConnect func(connID string)
	// 业务回调：连接关闭时触发
	OnDisconnect func(connID string, err error)
}

// NewManager 创建连接管理器
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	// 初始化升级器
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			// 跨域检查
			if config.AllowAllOrigins {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, allowed := range config.AllowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}

	return &Manager{
		config:      config,
		upgrader:    upgrader,
		connections: make(map[string]*Connection),
		mutex:       sync.RWMutex{},
		// 默认回调（用户可覆盖）
		OnMessage: func(connID string, msgType int, data []byte) {
			log.Printf("[默认回调] 收到连接[%s]消息：%s", connID, string(data))
		},
		OnConnect: func(connID string) {
			log.Printf("[默认回调] 连接[%s]已建立", connID)
		},
		OnDisconnect: func(connID string, err error) {
			log.Printf("[默认回调] 连接[%s]已关闭：%v", connID, err)
		},
	}
}

// Upgrade HTTP升级为WebSocket连接
// connID：自定义连接唯一ID（如用户ID、设备ID）
func (m *Manager) Upgrade(w http.ResponseWriter, r *http.Request, connID string) (*Connection, error) {
	if connID == "" {
		return nil, errors.New("连接ID不能为空")
	}

	// 检查连接ID是否已存在
	m.mutex.RLock()
	_, exists := m.connections[connID]
	m.mutex.RUnlock()
	if exists {
		return nil, fmt.Errorf("连接ID[%s]已存在", connID)
	}

	// 升级HTTP连接
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("升级WebSocket失败：%w", err)
	}

	// 创建上下文（用于优雅关闭）
	ctx, cancel := context.WithCancel(context.Background())

	// 创建连接实例
	wsConn := &Connection{
		conn:          conn,
		connID:        connID,
		manager:       m,
		createTime:    time.Now(),
		heartbeatChan: make(chan struct{}, 1),
		ctx:           ctx,
		cancel:        cancel,
		writeMutex:    sync.Mutex{},
	}

	// 添加到管理器
	m.mutex.Lock()
	m.connections[connID] = wsConn
	m.mutex.Unlock()

	// 触发连接建立回调
	m.OnConnect(connID)

	// 启动读消息协程
	go wsConn.ReadPump()
	// 启动写消息协程（处理异步发送）
	go wsConn.WritePump()
	// 启动心跳检测协程
	go wsConn.Heartbeat()

	return wsConn, nil
}

// ReadPump 读取客户端消息（持续运行）
func (c *Connection) ReadPump() {
	defer func() {
		// 发生panic时关闭连接
		if err := recover(); err != nil {
			log.Printf("连接[%s]读消息协程panic：%v", c.connID, err)
		}
		// 关闭连接并清理
		c.Close(fmt.Errorf("读消息协程退出"))
	}()

	// 设置读超时
	c.conn.SetReadDeadline(time.Now().Add(c.manager.config.ReadTimeout))

	// 设置消息分片大小（默认不限制）
	c.conn.SetReadLimit(0)

	for {
		select {
		case <-c.ctx.Done():
			return // 上下文已取消，退出
		default:
			// 读取客户端消息
			msgType, data, err := c.conn.ReadMessage()
			if err != nil {
				// 区分正常关闭和异常错误
				var closeErr *websocket.CloseError
				if errors.As(err, &closeErr) {
					c.Close(fmt.Errorf("客户端主动关闭：%s", closeErr.Text))
				} else {
					c.Close(fmt.Errorf("读取消息失败：%w", err))
				}
				return
			}

			// 更新读超时（收到消息则重置超时）
			c.conn.SetReadDeadline(time.Now().Add(c.manager.config.ReadTimeout))

			// 心跳响应：如果是客户端的pong消息，触发心跳通道
			var msg *Msg
			_ = gconv.Struct(data, &msg)
			if msgType == c.manager.config.MsgType && msg.Type == c.manager.config.HeartbeatType {
				log.Printf("[心跳] 收到连接[%s]心跳：%s", c.connID, string(data))
				select {
				case c.heartbeatChan <- struct{}{}:
				default:
				}
				continue
			}

			// 触发业务消息回调
			c.manager.OnMessage(c.connID, msgType, data)
		}
	}
}

type Msg struct {
	Type      string `json:"type"`
	Data      any    `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

// WritePump 处理异步写消息（持续运行）
// 注：实际项目可扩展为消息队列，此处简化为直接写
func (c *Connection) WritePump() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("连接[%s]写消息协程panic：%v", c.connID, err)
		}
	}()

	// 暂时无需循环，实际可扩展为监听写队列
	<-c.ctx.Done()
}

// Heartbeat 心跳检测（持续运行）
func (c *Connection) Heartbeat() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("连接[%s]心跳协程panic：%v", c.connID, err)
		}
	}()

	// 心跳定时器
	ticker := time.NewTicker(c.manager.config.HeartbeatInterval)
	defer ticker.Stop()

	// 超时定时器
	timeoutTimer := time.NewTimer(c.manager.config.HeartbeatTimeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// 发送心跳ping消息
			err := c.Send(Msg{Type: c.manager.config.HeartbeatType, Timestamp: gtime.Timestamp()})
			if err != nil {
				c.Close(fmt.Errorf("发送心跳失败：%w", err))
				return
			}
			// 重置超时定时器
			if !timeoutTimer.Stop() {
				<-timeoutTimer.C
			}
			timeoutTimer.Reset(c.manager.config.HeartbeatTimeout)
		case <-timeoutTimer.C:
			// 心跳超时，关闭连接
			c.Close(errors.New("心跳超时，客户端无响应"))
			return
		case <-c.heartbeatChan:
			// 收到客户端pong响应，重置超时定时器
			if !timeoutTimer.Stop() {
				<-timeoutTimer.C
			}
			timeoutTimer.Reset(c.manager.config.HeartbeatTimeout)
		}
	}
}

// Send 发送消息到客户端（线程安全）
func (c *Connection) Send(data any) error {
	select {
	case <-c.ctx.Done():
		return errors.New("连接已关闭，无法发送消息")
	default:
		// 加锁防止并发写
		c.writeMutex.Lock()
		defer c.writeMutex.Unlock()

		// 设置写超时
		c.conn.SetWriteDeadline(time.Now().Add(c.manager.config.WriteTimeout))

		// 发送消息
		err := c.conn.WriteMessage(c.manager.config.MsgType, gconv.Bytes(data))
		if err != nil {
			return fmt.Errorf("发送消息失败：%w", err)
		}
		return nil
	}
}

// Close 关闭连接（优雅清理）
func (c *Connection) Close(err error) {
	// 取消上下文（终止所有协程）
	c.cancel()

	// 关闭底层连接
	_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
	_ = c.conn.Close()

	// 从管理器移除
	c.manager.mutex.Lock()
	delete(c.manager.connections, c.connID)
	c.manager.mutex.Unlock()

	// 触发断开回调
	c.manager.OnDisconnect(c.connID, err)

	log.Printf("连接[%s]已关闭，当前在线数：%d", c.connID, c.manager.GetOnlineCount())
}

// GetOnlineCount 获取在线连接数
func (m *Manager) GetOnlineCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.connections)
}

// Broadcast 广播消息到所有在线连接
func (m *Manager) Broadcast(data any) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.connections) == 0 {
		return errors.New("无在线连接")
	}

	// 并发发送（非阻塞）
	var wg sync.WaitGroup
	var errMsg string

	for _, conn := range m.connections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if err := c.Send(data); err != nil {
				errMsg += fmt.Sprintf("连接[%s]广播失败：%v；", c.connID, err)
			}
		}(conn)
	}

	wg.Wait()

	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

// SendToConn 定向发送消息到指定连接
func (m *Manager) SendToConn(connID string, data any) error {
	m.mutex.RLock()
	conn, exists := m.connections[connID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("连接[%s]不存在", connID)
	}

	return conn.Send(data)
}

func (m *Manager) GetAllConn() map[string]*Connection {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.connections
}

// CloseAll 关闭所有连接
func (m *Manager) CloseAll() {
	m.mutex.RLock()
	connIDs := make([]string, 0, len(m.connections))
	for connID := range m.connections {
		connIDs = append(connIDs, connID)
	}
	m.mutex.RUnlock()

	for _, connID := range connIDs {
		m.mutex.RLock()
		conn := m.connections[connID]
		m.mutex.RUnlock()
		if conn != nil {
			conn.Close(errors.New("服务端主动关闭所有连接"))
		}
	}
}
