package tcp

import (
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/gtcp"
)

// TcpPoolConfig TCP连接池配置
type TcpPoolConfig struct {
	BufferSize     int           `json:"bufferSize"`     // 缓冲区大小
	MaxConnections int           `json:"maxConnections"` // 缓冲区大小
	ConnectTimeout time.Duration `json:"connectTimeout"`
	ReadTimeout    time.Duration `json:"readTimeout"`  // 读取超时时间
	WriteTimeout   time.Duration `json:"writeTimeout"` // 写入超时时间
	MaxIdleTime    time.Duration `json:"maxIdleTime"`  // 最大空闲时间
}

// TcpConnection TCP连接结构
type TcpConnection struct {
	Id        string       `json:"id"`        // 连接ID
	Address   string       `json:"address"`   // 连接地址
	Server    gtcp.Conn    `json:"conn"`      // 实际连接
	IsActive  bool         `json:"isActive"`  // 是否活跃
	LastUsed  time.Time    `json:"lastUsed"`  // 最后使用时间
	CreatedAt time.Time    `json:"createdAt"` // 创建时间
	Mutex     sync.RWMutex `json:"-"`         // 读写锁
}

// TcpMessage TCP消息结构
type TcpMessage struct {
	Id        string    `json:"id"`        // 消息ID
	ConnId    string    `json:"connId"`    // 连接ID
	Data      []byte    `json:"data"`      // 消息数据
	Timestamp time.Time `json:"timestamp"` // 时间戳
	IsSend    bool      `json:"isSend"`    // 是否是发送的消息
}
