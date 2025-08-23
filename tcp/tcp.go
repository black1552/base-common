package tcp

import (
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtcp"
	"github.com/gogf/gf/v2/os/glog"
)

// TCPServer TCP服务器结构
type TCPServer struct {
	Address    string
	Config     *TcpPoolConfig
	Listener   *gtcp.Server
	Connection *ConnectionPool
	Logger     *glog.Logger
}

// ConnectionPool 连接池结构
type ConnectionPool struct {
	connections map[string]*TcpConnection
	mutex       sync.RWMutex
	config      *TcpPoolConfig
	logger      *glog.Logger
}

// NewTCPServer 创建一个新的TCP服务器
func NewTCPServer(address string, config *TcpPoolConfig) *TCPServer {
	logger := g.Log("tcp-server")

	pool := &ConnectionPool{
		connections: make(map[string]*TcpConnection),
		config:      config,
		logger:      logger,
	}
	server := gtcp.NewServer(address, connFunc)
	return &TCPServer{
		Address:    address,
		Config:     config,
		Listener:   server,
		Connection: pool,
		Logger:     logger,
	}
}

func connFunc(conn *gtcp.Conn) {
	// TODO 尚未开发
}
