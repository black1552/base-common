package handler

import (
	"context"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Handler 消息处理接口
type Handler interface {
	context.Context
	handleMessage(client mqtt.Client, msg mqtt.Message)
}
