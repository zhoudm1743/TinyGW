package mqtt

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gookit/color"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
	"zsxagw/api/repository"
	"zsxagw/core/collect"
	"zsxagw/core/collect/worker"
	"zsxagw/core/config"
)

type (
	Client struct {
		client           mqtt.Client
		stop             chan bool
		collectorServer  *collect.CollectorServer
		deviceRepository repository.DeviceRepository

		collectorRepository  repository.CollectorRepository
		deviceTypeRepository repository.DeviceTypeRepository
	}
)

var MClient Client

func InitMqttClient(
	repository repository.ReportTaskRepository,
	collectorServer *collect.CollectorServer,
	deviceRepository repository.DeviceRepository,
	collectorRepository repository.CollectorRepository,
	deviceTypeRepository repository.DeviceTypeRepository,
) {
	zap.S().Info("实例化Mqtt Client")

	// 确保上报任务有一条数据
	repository.Migrate()

	reportTask, err := repository.Find("数据上报")
	if err != nil {
		zap.S().Error("没找到上报服务中的【数据上报】那条记录，没有实例化mqtt client.")
		return
	}

	clientOptions := mqtt.NewClientOptions()

	clientOptions.AddBroker(reportTask.Ip + ":" + strconv.Itoa(reportTask.Port))

	// 客户信息
	clientOptions.SetClientID(reportTask.ClientID)
	clientOptions.SetUsername(reportTask.Username)
	clientOptions.SetPassword(reportTask.Password)

	zap.S().Info("mqtt链接设置："+reportTask.Ip, " ", reportTask.Port, " ", reportTask.ClientID, " ", clientOptions)

	// 不使用短线重连机制，自己开线程检测系统在线情况
	clientOptions.SetAutoReconnect(false)

	clientOptions.SetOnConnectHandler(onConnectHandler)
	clientOptions.SetConnectionLostHandler(connectionLostHandler)
	clientOptions.SetReconnectingHandler(reconnectingHandler)

	MClient = Client{
		client:               mqtt.NewClient(clientOptions),
		collectorServer:      collectorServer,
		deviceRepository:     deviceRepository,
		collectorRepository:  collectorRepository,
		deviceTypeRepository: deviceTypeRepository,
		stop:                 make(chan bool, 1),
	}
}

// Connect 连接服务器
func (mc *Client) Connect() {
	go func() {
		for {
			select {
			case <-mc.stop:
				zap.S().Info("连接线程停止工作！")
				return
			default:
				if !mc.client.IsConnected() {
					zap.S().Info("网络断开，尝试重连...")
					if token := mc.client.Connect(); token.Wait() && token.Error() != nil {
						zap.S().Error("MqttClient：尝试连接失败！")
					}
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()
}

// Discount 断开服务器连接
func (mc *Client) Discount() {
	zap.S().Info("MClient.Disconnect 关闭服务器连接！")
	online := fmt.Sprintf(`{"online":false,"ts":%d,"name":"%s"}`, time.Now().Unix(), config.NewConfig().ClientID)
	if token := mc.client.Publish("v1/devices/me/telemetry", 0, false, online); token.Wait() && token.Error() != nil {
		zap.S().Error("上传网关状态！")
	}
	mc.stop <- true
	mc.client.Disconnect(0)
}

// Publish 数据上报
func (mc *Client) Publish(payload []byte) {
	if !mc.client.IsConnected() {
		zap.S().Error("无Mqtt.Client连接，无法上报数据。")
		// 存盘返回
		return
	}

	if token := mc.client.Publish("v1/gateway/telemetry", 0, false, payload); token.Wait() && token.Error() != nil {
		zap.S().Error("上报数据失败！")
		// 存盘返回
	}
}

func onConnectHandler(client mqtt.Client) {
	if token := client.Subscribe("v1/devices/me/rpc/request/+", 0, MClient.receiveMessageHandler); token.Wait() && token.Error() != nil {
		zap.S().Error("onConnectHandler: Mqtt Client 订阅RPC主题失败！")
		return
	}
	zap.S().Info("onConnectHandler: Mqtt Client 链接成功并且订阅RPC主题.")
	// 检测是否有断线缓存数据，如果有就提交，并删除缓存数据
	online := fmt.Sprintf(`{"online":true,"ts":%d,"name":"%s"}`, time.Now().Unix(), config.NewConfig().ClientID)
	color.Blueln("上报数据：" + online)
	if token := client.Publish("v1/devices/me/telemetry", 0, false, online); token.Wait() && token.Error() != nil {
		zap.S().Error("上传网关状态！")
		// 存盘返回
	}
}

// 连接丢失
func connectionLostHandler(client mqtt.Client, err error) {
	zap.S().Info("connectionLostHandler: Mqtt Client 链接断开!")
}

// 掉线重连（开始连接）
func reconnectingHandler(client mqtt.Client, opt *mqtt.ClientOptions) {
	zap.S().Info("reconnectingHandler: Mqtt Client 链接重新链接[永远不会调用]")
}

func (mc *Client) receiveMessageHandler(client mqtt.Client, msg mqtt.Message) {
	clientId := config.NewConfig().ClientID
	zap.S().Infof("收到RPC命令：【%s】,正在解析...", msg.Topic())
	// 0. 不是订阅的设备rpc请求，则返回
	if !strings.Contains(msg.Topic(), "v1/devices/me/rpc/request/") {
		zap.S().Errorf("receiveMessageHandler: 接收到意外主题：%s", msg.Topic())
		return
	}

	// 1. 解析请求
	commandRequest := worker.CommandRequest{}
	if err := commandRequest.FromJson(msg.Payload()); err != nil {
		zap.S().Error("解析设备rpc请求失败: ", err)
		return
	}

	// 判断是否请求本设备
	if commandRequest.Params[0].ClientID != "any" && commandRequest.Params[0].ClientID != clientId {
		zap.S().Errorf("receiveMessageHandler: 不是本设备的RPC命令：%s", msg.Topic())
		return
	}

	// 2. 获取请求ID
	commandRequest.RequestID = ""
	topicToken := strings.Split(msg.Topic(), "/")
	if len(topicToken) == 6 {
		commandRequest.RequestID = topicToken[5]
	}
	if commandRequest.RequestID == "" {
		zap.S().Errorf("主题中不包含请求ID：%s", msg.Topic())
		return
	}

	// 3. 处理命令 ...
	commandResponse := worker.ResponseParam{
		ClientID: clientId,
		CmdName:  commandRequest.Params[0].CmdName,
	}

	switch strings.ToLower(commandRequest.Method) {
	case "gateway":
		command, ok := GatewayCommand[commandRequest.Params[0].CmdName]
		if ok {
			status, devs := command(commandRequest.Params[0].CmdParams, *mc)
			commandResponse.CmdStatus = status
			commandResponse.CmdResult = devs
			mc.SendResponse(commandRequest, commandResponse)
		} else {
			commandResponse.CmdStatus = 1
			commandResponse.CmdResult = "不支持gateway." + commandResponse.CmdName + "命令"
		}
	case "device":
		// 处理设备操作
		// 1. 回调函数
		commandRequest.ResponseParamChan = make(chan worker.ResponseParam, 1)
		if len(commandRequest.Params) < 1 {
			zap.S().Error("对设备的RPC操作缺少请求参数params")
			break
		}
		deviceName := commandRequest.Params[0].DeviceName
		if work, ok := mc.collectorServer.FindByDeviceName(deviceName); ok {
			work.CommandTask(commandRequest)

			responseParam := <-commandRequest.ResponseParamChan
			mc.SendResponse(commandRequest, responseParam)
		}
	default:
		zap.S().Error("无效的方法：", commandRequest.Method)
	}
}

func (mc *Client) SendResponse(request worker.CommandRequest, response worker.ResponseParam) {
	commandResponse := worker.CommandResponse{
		Method: request.Method,
		Params: []worker.ResponseParam{
			response,
		},
	}
	responseTopic := "v1/devices/me/rpc/response/" + request.RequestID
	zap.S().Infof("正确响应RPC调用，主题：%s 响应: %v", responseTopic, commandResponse)
	if token := mc.client.Publish(responseTopic, 0, false, commandResponse.ToJson()); token.Wait() && token.Error() != nil {
		zap.S().Error("应答RPC调用失败！")
	}
}
