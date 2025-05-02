package subscription

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"math/rand"
	"strings"
	"time"
	"tinyGW/pkg/plugin/io"
	"tinyGW/pkg/service/conf"
)

type Server struct {
	client mqtt.Client
	stop   chan bool
}

var Mclient *Server

func InitMqttServer(config *conf.Config) *Server {
	subscribeOptions := mqtt.NewClientOptions()
	uri := fmt.Sprintf("%s:%d", config.Subscribe.Host, config.Subscribe.Port)
	subscribeOptions.AddBroker(uri)
	subscribeOptions.SetUsername(config.Subscribe.Username)
	subscribeOptions.SetPassword(config.Subscribe.Password)
	subscribeOptions.SetClientID(config.Cloud.ClientId + generateClientID())
	zap.S().Infoln("订阅采集器服务器配置：", uri)

	subscribeOptions.SetAutoReconnect(false)
	subscribeOptions.SetKeepAlive(60 * time.Second)
	subscribeOptions.SetPingTimeout(10 * time.Second)
	subscribeOptions.SetOnConnectHandler(onConnectHandler)
	subscribeOptions.SetConnectionLostHandler(connectionLostHandler)
	subscribeOptions.SetReconnectingHandler(reconnectingHandler)

	Mclient = &Server{
		client: mqtt.NewClient(subscribeOptions),
		stop:   make(chan bool, 1),
	}
	return Mclient
}

// Connect 连接服务器
func (mc *Server) Connect() {
	zap.S().Infoln("开始连接订阅服务器...")
	go func() {
		for {
			select {
			case <-mc.stop:
				zap.S().Infoln("连接线程停止工作！")
				return
			default:
				if !mc.client.IsConnected() {
					zap.S().Infoln("网络断开，尝试重连...")
					if token := mc.client.Connect(); token.Wait() && token.Error() != nil {
						zap.S().Error("MqttClient：尝试连接失败！")
					}
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()
}

// Disconnect 断开服务器连接
func (mc *Server) Disconnect() {
	zap.S().Error("MServer.Disconnect 关闭服务器连接！")
	mc.stop <- true
	mc.client.Disconnect(0)
}

// Publish 数据上报
func (mc *Server) Publish(clientID string, payload []byte) error {
	if !mc.client.IsConnected() {
		zap.S().Error("无Mqtt.Client连接，无法上报数据。")
		// 存盘返回
		return fmt.Errorf("无Mqtt.Client连接，无法上报数据。")
	}

	if token := mc.client.Publish("/up/gateway/publish/"+clientID, 0, false, payload); token.Wait() && token.Error() != nil {
		zap.S().Error("上报数据失败！")
		return token.Error()
	}
	return nil
}

/** 4G采集器配置
订阅：/up/gateway/subscription
发布：/up/gateway/publish
遗嘱：/up/gateway/close
*/

func onConnectHandler(client mqtt.Client) {
	// 订阅
	if token := client.Subscribe("/up/gateway/publish/+", 0, Mclient.onSubscriptionHandler); token.Wait() && token.Error() != nil {
		zap.S().Error("MqttClient：订阅失败！")
	}
	if token := client.Subscribe("/up/gateway/close/+", 0, Mclient.onSubscriptionHandler); token.Wait() && token.Error() != nil {
		zap.S().Error("MqttClient：订阅失败！")
	}
	//// 广播一个消息看看
	//if token := client.Publish("/up/gateway/subscription", 0, false, "config"); token.Wait() && token.Error() != nil {
	//	zap.S().Error("MqttClient：发布失败！")
	//}
}

// 连接丢失
func connectionLostHandler(client mqtt.Client, err error) {
	zap.S().Info("connectionLostHandler: Mqtt Client 链接断开!")
}

// 掉线重连（开始连接）
func reconnectingHandler(client mqtt.Client, opt *mqtt.ClientOptions) {
	zap.S().Info("reconnectingHandler: Mqtt Client 链接重新链接[永远不会调用]")
}

// 订阅处理
func (mc *Server) onSubscriptionHandler(client mqtt.Client, message mqtt.Message) {
	topicSplit := strings.Split(message.Topic(), "/")
	// 获取设备Id
	length := len(topicSplit)
	clientID := topicSplit[length-1]
	var register registerRequest
	err := json.Unmarshal(message.Payload(), &register)
	if err == nil {
		zap.S().Infoln("收到注册消息：", register.String())
		return
	}
	zap.S().Infoln(fmt.Sprintf("收到[%s]消息：[% X]", clientID, message.Payload()))

	// 处理消息
	io.IConcurrentMap.Set(clientID, message.Payload())
}

// 随机生成clientid
func generateClientID() string {
	str := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := []byte(str)
	clientid := make([]byte, 10)
	for i := 0; i < 10; i++ {
		clientid[i] = bytes[rand.Intn(len(bytes))]
	}
	return "_" + string(clientid)
}

var Module = fx.Provide(
	InitMqttServer,
)
