package collector

import (
	"go.uber.org/zap"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/io"
	"tinyGW/pkg/service/subscription"
)

// MqttCollector is the collector for MQTT
type MqttCollector struct {
	models.Collector
}

func (t MqttCollector) Open(device *models.Device) bool {
	return true
}

func (t MqttCollector) Close() bool {
	return true
}

func (t MqttCollector) Read(data []byte) int {
	val, exists := io.IConcurrentMap.Get(t.Collector.Mqtt.Name)
	if !exists {
		return 0
	}
	copy(data, val)
	io.IConcurrentMap.Delete(t.Collector.Mqtt.Name)
	return len(val)
}

func (t MqttCollector) Write(data []byte) int {
	err := subscription.Mclient.Publish(t.Mqtt.Name, data)
	if err != nil {
		zap.S().Error(err.Error())
		return 0
	}
	return len(data)
}

func (t MqttCollector) GetName() string {
	return t.Name
}

func (t MqttCollector) GetTimeout() int {
	return t.Timeout
}

func (t MqttCollector) GetInterval() int {
	return t.Interval
}

var _ Collector = (*MqttCollector)(nil)
