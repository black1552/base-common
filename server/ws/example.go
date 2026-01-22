package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/util/gconv"
)

var manager = NewWs()

func NewWs() *Manager {
	// 1. 自定义配置（可选，也可使用默认配置）
	customConfig := &Config{
		AllowAllOrigins:   true,
		HeartbeatInterval: 20 * time.Second, // 20秒发一次心跳
		HeartbeatTimeout:  40 * time.Second, // 40秒超时
	}

	// 2. 创建管理器
	m := NewManager(customConfig)

	// 3. 覆盖业务回调（核心：自定义消息处理逻辑）
	// 连接建立回调
	m.OnConnect = func(connID string) {
		log.Printf("业务回调：连接[%s]上线，当前在线数：%d", connID, m.GetOnlineCount())
		// 欢迎消息
		_ = m.SendToConn(connID, []byte("欢迎连接WebSocket服务！"))
	}

	// 收到消息回调
	m.OnMessage = func(connID string, msgType int, data any) {
		log.Printf("业务回调：收到连接[%s]消息：%s", connID, gconv.String(data))
		// 示例：echo回复
		reply := []byte("服务端回复：" + gconv.String(data))
		_ = m.SendToConn(connID, reply)

		// 示例：广播消息给所有连接
		_ = m.Broadcast([]byte("广播：" + connID + "说：" + gconv.String(data)))
	}

	// 连接断开回调
	m.OnDisconnect = func(connID string, err error) {
		log.Printf("业务回调：连接[%s]下线，原因：%v，当前在线数：%d", connID, err, m.GetOnlineCount())
	}
	return m
}
func Upgrade(w http.ResponseWriter, r *http.Request, connID string) {
	_, err := manager.Upgrade(w, r, connID)
	if err != nil {
		log.Printf("升级连接失败：%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func main() {
	// 4. 注册WebSocket路由
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 自定义连接ID（示例：使用请求参数中的user_id）
		connID := r.URL.Query().Get("user_id")
		if connID == "" {
			http.Error(w, "user_id不能为空", http.StatusBadRequest)
			return
		}
		// 升级连接
		Upgrade(w, r, connID)
	})

	// 5. 启动服务
	log.Println("WebSocket服务启动：http://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
