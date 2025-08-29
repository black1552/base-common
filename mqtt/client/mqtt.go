package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/os/glog"
)

// Client MQTT客户端结构
type Client struct {
	client     mqtt.Client
	opts       *mqtt.ClientOptions
	ctx        context.Context
	subscribed map[string]byte
	subMutex   sync.RWMutex
	callback   mqtt.MessageHandler
	// 错误处理相关
	onConnectionLost    func(error)
	onReconnect         func()
	onConnect           func()
	onSubscriptionError func(error)
	onPublishError      func(error)
}

// NewClientWithAuth 创建带用户名密码认证的MQTT客户端
func NewClientWithAuth(ctx context.Context, broker, clientId, username, password string) *Client {
	c := &Client{
		ctx:        ctx,
		subscribed: make(map[string]byte),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	opts.SetUsername(username)
	opts.SetPassword(password)

	// 启用自动重连
	opts.SetAutoReconnect(true)

	// 设置连接重试
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(10 * time.Second)

	// 设置连接超时时间
	opts.SetConnectTimeout(60 * time.Second)

	// 设置保活时间(心跳)
	opts.SetKeepAlive(60 * time.Second)

	// 设置ping超时时间
	opts.SetPingTimeout(20 * time.Second)

	// 设置clean session为false，保留会话状态
	opts.SetCleanSession(false)

	// 设置最大重连间隔
	opts.SetMaxReconnectInterval(20 * time.Second)

	// 设置消息排序
	opts.SetOrderMatters(false)

	// 设置协议版本
	opts.SetProtocolVersion(4) // MQTT 3.1.1

	// 设置连接丢失处理函数
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		glog.Error(ctx, "MQTT连接断开:", err)
		if c.onConnectionLost != nil {
			c.onConnectionLost(err)
		}
	})

	// 设置重连处理函数
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		glog.Info(ctx, "MQTT客户端正在尝试重连...")
		if c.onReconnect != nil {
			c.onReconnect()
		}
	})

	// 设置连接成功处理函数，重连成功后重新订阅主题
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		glog.Info(ctx, "MQTT客户端重新连接成功")
		if c.onConnect != nil {
			c.onConnect()
		}
		// 重连成功后重新订阅主题
		go c.resubscribe()
	})

	// 设置其他选项...

	mqttClient := mqtt.NewClient(opts)
	c.client = mqttClient
	c.opts = opts
	return c
}

// Connect 连接到MQTT服务器
func (c *Client) Connect() error {
	glog.Info(c.ctx, "开始连接到MQTT服务器...")
	token := c.client.Connect()
	glog.Info(c.ctx, "等待连接完成...")

	// 使用更长的超时时间，避免网络延迟导致连接失败
	if token.WaitTimeout(30 * time.Second) {
		glog.Info(c.ctx, "连接操作完成")
		if token.Error() != nil {
			err := fmt.Errorf("连接到MQTT代理时发生错误: %w", token.Error())
			glog.Error(c.ctx, "连接MQTT服务器时发生错误:", token.Error())
			return err
		}
	} else {
		// 连接超时
		err := fmt.Errorf("连接到MQTT服务器超时")
		glog.Error(c.ctx, "连接MQTT服务器超时")
		return err
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
	// 保存订阅信息
	c.subMutex.Lock()
	for topic, qos := range topics {
		c.subscribed[topic] = qos
	}
	c.callback = callback
	c.subMutex.Unlock()

	token := c.client.SubscribeMultiple(topics, callback)
	// 增加订阅超时时间
	if token.WaitTimeout(30*time.Second) && token.Error() != nil {
		err := fmt.Errorf("同时订阅多个主题出现错误: %w", token.Error())
		glog.Error(c.ctx, "订阅主题时发生错误:", token.Error())
		if c.onSubscriptionError != nil {
			c.onSubscriptionError(err)
		}
		return err
	}

	// 检查订阅是否成功
	if token.WaitTimeout(30 * time.Second) {
		glog.Info(c.ctx, "成功订阅主题:", topics)
	} else {
		err := fmt.Errorf("订阅主题超时: %v", topics)
		glog.Error(c.ctx, "订阅主题超时:", topics)
		if c.onSubscriptionError != nil {
			c.onSubscriptionError(err)
		}
		return err
	}

	return nil
}

// Publish 发布消息
func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := c.client.Publish(topic, qos, retained, payload)
	// 增加发布超时时间
	if token.WaitTimeout(30*time.Second) && token.Error() != nil {
		err := fmt.Errorf("发送消息到主题%s出现错误: %w", topic, token.Error())
		glog.Error(c.ctx, "发布消息到主题", topic, "时发生错误:", token.Error())
		if c.onPublishError != nil {
			c.onPublishError(err)
		}
		return err
	}

	// 检查发布是否成功
	if token.WaitTimeout(30 * time.Second) {
		glog.Info(c.ctx, "成功发布消息到主题:", topic)
	} else {
		err := fmt.Errorf("发布消息到主题%s超时", topic)
		glog.Error(c.ctx, "发布消息到主题超时:", topic)
		if c.onPublishError != nil {
			c.onPublishError(err)
		}
		return err
	}

	return nil
}

// resubscribe 重新订阅主题
func (c *Client) resubscribe() {
	c.subMutex.RLock()
	defer c.subMutex.RUnlock()

	if len(c.subscribed) == 0 {
		glog.Info(c.ctx, "没有需要重新订阅的主题")
		return
	}

	// 复制订阅信息避免并发问题
	topics := make(map[string]byte)
	for topic, qos := range c.subscribed {
		topics[topic] = qos
	}

	glog.Info(c.ctx, "开始重新订阅主题:", topics)
	// 增加重新订阅的超时时间
	token := c.client.SubscribeMultiple(topics, c.callback)
	if token.WaitTimeout(30 * time.Second) {
		if token.Error() != nil {
			err := fmt.Errorf("重新订阅主题时发生错误: %w", token.Error())
			glog.Error(c.ctx, "重新订阅主题时发生错误:", token.Error())
			if c.onSubscriptionError != nil {
				c.onSubscriptionError(err)
			}
		} else {
			glog.Info(c.ctx, "重新订阅主题成功")
		}
	} else {
		err := fmt.Errorf("重新订阅主题超时")
		glog.Error(c.ctx, "重新订阅主题超时")
		if c.onSubscriptionError != nil {
			c.onSubscriptionError(err)
		}
	}
}

// IsConnected 检查是否连接
func (c *Client) IsConnected() bool {
	isConnected := c.client.IsConnected()
	glog.Debug(c.ctx, "检查连接状态:", isConnected)
	return isConnected
}

// Subscribe 单独订阅一个主题
func (c *Client) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	topics := map[string]byte{topic: qos}
	return c.SubscribeMultiple(topics, callback)
}
