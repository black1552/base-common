package server

import (
	"context"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/os/glog"
)

// Client MQTT客户端结构
type Client struct {
	client mqtt.Client
	opts   *mqtt.ClientOptions
	ctx    context.Context
}

// NewClientWithAuth 创建带用户名密码认证的MQTT客户端
func NewClientWithAuth(ctx context.Context, broker, clientId, username, password string) *Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(15 * time.Second)
	opts.SetConnectTimeout(10 * time.Second)

	// 设置其他选项...

	mqttClient := mqtt.NewClient(opts)
	return &Client{
		client: mqttClient,
		opts:   opts,
		ctx:    ctx,
	}
}

// Connect 连接到MQTT服务器
func (c *Client) Connect() error {
	glog.Info(c.ctx, "开始连接到MQTT服务器...")
	token := c.client.Connect()
	glog.Info(c.ctx, "等待连接完成...")

	// 使用带超时的等待，避免无限期阻塞
	if token.WaitTimeout(10 * time.Second) {
		glog.Info(c.ctx, "连接操作完成")
		if token.Error() != nil {
			glog.Error(c.ctx, "连接MQTT服务器时发生错误:", token.Error())
			return fmt.Errorf("连接到MQTT代理时发生错误: %w", token.Error())
		}
	} else {
		// 连接超时
		glog.Error(c.ctx, "连接MQTT服务器超时")
		return fmt.Errorf("连接到MQTT服务器超时")
	}
	glog.Info(c.ctx, "成功连接到MQTT服务器")
	return nil
}

// Disconnect 断开MQTT连接
func (c *Client) Disconnect() {
	if c.client.IsConnected() {
		c.client.Disconnect(250)
		glog.Info(c.ctx, "已断开MQTT连接")
	}
}

// SubscribeMultiple 同时订阅多个主题
func (c *Client) SubscribeMultiple(topics map[string]byte, callback mqtt.MessageHandler) error {
	if token := c.client.SubscribeMultiple(topics, callback); token.Wait() && token.Error() != nil {
		return fmt.Errorf("同时订阅多个主题出现错误: %w", token.Error())
	}
	return nil
}

// Publish 发布消息
func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	if token := c.client.Publish(topic, qos, retained, payload); token.Wait() && token.Error() != nil {
		return fmt.Errorf("发送消息到主题%s出现错误: %w", topic, token.Error())
	}
	return nil
}

// IsConnected 检查是否连接
func (c *Client) IsConnected() bool {
	isConnected := c.client.IsConnected()
	glog.Debug(c.ctx, "检查连接状态:", isConnected)
	return isConnected
}
