package collector

import (
	"net"
	"sync"
	"tinyGW/app/models"
	"tinyGW/pkg/service/command"
)

var (
	// CollectorInstancesMutex 采集器实例互斥锁
	CollectorInstancesMutex sync.RWMutex
	// CollectorInstances 采集器实例映射
	CollectorInstances = make(map[string]Collector)
)

// Collector 【采集器接口】
type Collector interface {
	Open(device *models.Device) bool
	Close() bool
	Read(data []byte) int
	Write(data []byte) int
	GetName() string
	GetTimeout() int
	GetInterval() int
}

// ConnectorFactory 【采集器接口】工厂，根据【采集接口】创建不同的【采集器】
func ConnectorFactory(collector models.Collector, commandManager *command.Manager) Collector {
	// 检查是否已经存在该采集器的实例
	CollectorInstancesMutex.RLock()
	if instance, exists := CollectorInstances[collector.Name]; exists {
		CollectorInstancesMutex.RUnlock()
		return instance
	}
	CollectorInstancesMutex.RUnlock()

	// 创建新的采集器实例
	var instance Collector

	switch collector.Type {
	case "Serial":
		instance = &SerialCollector{
			Collector: collector,
		}
	case "TcpClient":
		instance = &TcpClientCollector{
			Collector: collector,
		}
	case "TcpServer":
		// 使用增强版的TcpServerCollector，支持4G水表等协议
		tcpServer := &TcpServerCollector{
			Collector:         collector,
			commandManager:    commandManager,
			deviceConnections: make(map[string]net.Conn),
			commandResponses:  make(map[string]*CommandResponse),
		}
		// 初始化条件变量
		tcpServer.responseCond = sync.NewCond(&tcpServer.responseMutex)
		// 启动定期清理任务
		go tcpServer.startPeriodicCleanup()
		instance = tcpServer
	case "Mqtt":
		instance = &MqttCollector{
			Collector: collector,
		}
	case "fouGPRS", "FourGPRS":
		instance = &FourGDirectCollector{
			Collector: collector,
		}
	}

	if instance != nil {
		// 存储实例
		CollectorInstancesMutex.Lock()
		CollectorInstances[collector.Name] = instance
		CollectorInstancesMutex.Unlock()
	}

	return instance
}
