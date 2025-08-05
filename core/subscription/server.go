package subscription

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"TinyGW/core/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

type Server struct {
	client mqtt.Client
	stop   chan struct{}
	log    *logrus.Logger
	cfg    *config.Config
	mu     sync.Mutex
}

type RegisterRequest struct {
	Fver  string `json:"fver"`
	Iccid string `json:"iccid"`
	Imei  string `json:"imei"`
	Csq   int    `json:"csq"`
}

func (r RegisterRequest) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func NewServer(cfg *config.Config, log *logrus.Logger) *Server {
	return &Server{
		stop: make(chan struct{}, 1),
		log:  log,
		cfg:  cfg,
	}
}

func (s *Server) Init() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil && s.client.IsConnected() {
		return
	}
	opts := mqtt.NewClientOptions()
	uri := fmt.Sprintf("tcp://%s:%s", s.cfg.SubMqtt.Host, s.cfg.SubMqtt.Port)
	opts.AddBroker(uri)
	opts.SetUsername(s.cfg.SubMqtt.User)
	opts.SetPassword(s.cfg.SubMqtt.Password)
	opts.SetClientID(s.cfg.Cloud.User + generateClientID())
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetOnConnectHandler(s.onConnectHandler)
	opts.SetConnectionLostHandler(s.connectionLostHandler)
	opts.SetReconnectingHandler(s.reconnectingHandler)
	s.client = mqtt.NewClient(opts)
}

func (s *Server) Connect() {
	s.Init()
	go func() {
		for {
			select {
			case <-s.stop:
				s.log.Info("MQTT 连接线程停止")
				return
			default:
				if !s.client.IsConnected() {
					s.log.Info("MQTT 网络断开，尝试重连...")
					if token := s.client.Connect(); token.Wait() && token.Error() != nil {
						s.log.Error("MQTT 尝试连接失败！", token.Error())
					}
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()
}

func (s *Server) Disconnect() {
	s.stop <- struct{}{}
	if s.client != nil {
		s.client.Disconnect(0)
	}
	s.log.Info("MQTT 服务器连接已关闭")
}

func (s *Server) Publish(clientID string, payload []byte) error {
	if !s.client.IsConnected() {
		s.log.Error("无MQTT连接，无法上报数据。")
		return fmt.Errorf("无MQTT连接，无法上报数据。")
	}
	topic := fmt.Sprintf("/up/gateway/%s/subscription/%s", s.cfg.Cloud.User, clientID)
	s.log.Infof("Publish topic: %s", topic)
	if token := s.client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		s.log.Error("上报数据失败！", token.Error())
		return token.Error()
	}
	return nil
}

func (s *Server) onConnectHandler(client mqtt.Client) {
	cid := s.cfg.Cloud.User
	publishTopic := fmt.Sprintf("/up/gateway/%s/publish/+", cid)
	if token := client.Subscribe(publishTopic, 0, s.onSubscriptionHandler); token.Wait() && token.Error() != nil {
		s.log.Error("MQTT 订阅失败！", token.Error())
	}
	closeTopic := fmt.Sprintf("/up/gateway/%s/close/+", cid)
	if token := client.Subscribe(closeTopic, 0, s.onSubscriptionHandler); token.Wait() && token.Error() != nil {
		s.log.Error("MQTT 订阅失败！", token.Error())
	}
}

func (s *Server) connectionLostHandler(client mqtt.Client, err error) {
	s.log.Warn("MQTT 连接断开: ", err)
}

func (s *Server) reconnectingHandler(client mqtt.Client, opt *mqtt.ClientOptions) {
	s.log.Info("MQTT 正在重连...")
}

func (s *Server) onSubscriptionHandler(client mqtt.Client, message mqtt.Message) {
	topicSplit := strings.Split(message.Topic(), "/")
	clientID := topicSplit[len(topicSplit)-1]
	var register RegisterRequest
	err := json.Unmarshal(message.Payload(), &register)
	if err == nil {
		s.log.Infof("收到注册消息: %s", register.String())
		return
	}
	s.log.Infof("收到[%s]消息: [% X]", clientID, message.Payload())
	// TODO: 处理消息分发，可用 channel/worker pool 优化
}

func generateClientID() string {
	str := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := []byte(str)
	clientid := make([]byte, 10)
	for i := 0; i < 10; i++ {
		clientid[i] = bytes[rand.Intn(len(bytes))]
	}
	return "_" + string(clientid)
}
