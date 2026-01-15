package tcp

import (
	"fmt"
	"time"
)

// Example 展示如何使用TCP服务
func Example() {
	// 创建配置
	config := &TcpPoolConfig{
		BufferSize:     2048,
		MaxConnections: 100000,
		ConnectTimeout: time.Second * 5,
		ReadTimeout:    time.Second * 30,
		WriteTimeout:   time.Second * 10,
		MaxIdleTime:    time.Minute * 5,
	}

	// 创建TCP服务器
	server := NewTCPServer("0.0.0.0:8888", config)

	// 设置消息处理函数
	server.SetMessageHandler(func(conn *TcpConnection, msg *TcpMessage) error {
		fmt.Printf("Received message from %s: %s\n", conn.Id, string(msg.Data))

		// 回显消息
		return server.SendTo(conn.Id, []byte(fmt.Sprintf("Echo: %s", msg.Data)))
	})

	// 启动服务器
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}

	// 运行10秒后停止
	fmt.Println("TCP server started. Running for 10 seconds...")
	time.Sleep(time.Second * 10)

	// 停止服务器
	if err := server.Stop(); err != nil {
		fmt.Printf("Failed to stop server: %v\n", err)
	}

	fmt.Println("TCP server stopped.")
}
