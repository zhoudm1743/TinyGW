package collector

import "zsxagw/api/domain"

// Collector 【采集器接口】
type Collector interface {
	Open(device *domain.Device) bool
	Close() bool
	Read(data []byte) int
	Write(data []byte) int
	GetName() string
	GetTimeout() int
	GetInterval() int
}

// ConnectorFactory 【采集器接口】工厂，根据【采集接口】创建不同的【采集器】
func ConnectorFactory(collector domain.Collector) Collector {
	switch collector.Type {
	case "Serial":
		return &SerialCollector{
			Collector: collector,
		}
	case "TcpClient":
		return &TcpClientCollector{
			Collector: collector,
		}
	case "TcpServer":
		return &TcpServerCollector{
			Collector: collector,
		}
	case "Mqtt":
		return &MqttCollector{
			Collector: collector,
		}
	}

	return nil
}
