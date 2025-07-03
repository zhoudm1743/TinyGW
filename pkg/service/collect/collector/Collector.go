package collector

import "tinyGW/app/models"

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
func ConnectorFactory(collector models.Collector) Collector {
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
	case "fouGPRS", "FourGPRS":
		return &FourGDirectCollector{
			Collector: collector,
		}
	}

	return nil
}
