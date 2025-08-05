package event

import (
	"tinyGW/app/api/repository"
	"tinyGW/pkg/service/event"

	"go.uber.org/zap"
)

// Handle3761EventWithPublish 处理3761-BY协议事件并触发设备采集完成事件
func Handle3761EventWithPublish(deviceAddr string, data []byte, deviceRepository repository.DeviceRepository, eventBus *event.EventService) {
	if len(data) < 18 {
		zap.S().Warnln("3761-BY事件数据长度不足")
		return
	}

	// 检查是否是F254主动上报数据
	if !IsF254ActiveReport(data) {
		return
	}

	// 解析事件数据
	event, err := parse3761Event(deviceAddr, data)
	if err != nil {
		zap.S().Errorf("解析3761-BY事件失败: %v", err)
		return
	}

	// 调用对应的事件处理函数
	handle3761Event(event)

	// 转换为设备对象并触发DeviceCollectFinish事件
	ConvertToDeviceAndPublish3761(event, deviceRepository, eventBus)
}
