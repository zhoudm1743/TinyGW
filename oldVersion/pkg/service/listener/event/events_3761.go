package event

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"tinyGW/app/api/repository"
	"tinyGW/pkg/service/event"
)

// Event3761 3761-BY协议事件结构
type Event3761 struct {
	DeviceAddr  string                 // 设备地址
	EventType   string                 // 事件类型
	EventTime   time.Time              // 事件时间
	TableNumber string                 // 表号
	EventData   map[string]interface{} // 事件数据
	RawData     []byte                 // 原始数据
}

// EventHandler3761 3761-BY事件处理函数类型
type EventHandler3761 func(event *Event3761)

var (
	// 事件处理函数映射
	eventHandlers3761 = make(map[string]EventHandler3761)
)

// RegisterEventHandler3761 注册3761-BY事件处理函数
func RegisterEventHandler3761(eventType string, handler EventHandler3761) {
	eventHandlers3761[eventType] = handler
	zap.S().Infof("注册3761-BY事件处理函数: %s", eventType)
}

// UnregisterEventHandler3761 注销3761-BY事件处理函数
func UnregisterEventHandler3761(eventType string) {
	delete(eventHandlers3761, eventType)
	zap.S().Infof("注销3761-BY事件处理函数: %s", eventType)
}

// IsF254ActiveReport 检查是否是F254主动上报数据
func IsF254ActiveReport(data []byte) bool {
	if len(data) < 18 {
		return false
	}

	// 检查AFN和DT来判断是否是F254主动上报
	afn := data[12] // AFN在第13字节（索引12）
	dt1 := data[16] // DT1在第17字节（索引16）
	dt2 := data[17] // DT2在第18字节（索引17）

	// AFN=10H(数据转发), DT1=7FH, DT2=04H 对应F254主动上报
	return afn == 0x10 && dt1 == 0x7F && dt2 == 0x04
}

// parse3761Event 解析3761-BY事件数据
func parse3761Event(deviceAddr string, data []byte) (*Event3761, error) {
	event := &Event3761{
		DeviceAddr: deviceAddr,
		EventData:  make(map[string]interface{}, 8), // 预分配合理容量，减少扩容
		EventTime:  time.Now(),                      // 默认使用当前时间
	}

	// 只在Debug模式下复制原始数据
	if zap.S().Level() == zap.DebugLevel {
		event.RawData = make([]byte, len(data))
		copy(event.RawData, data)
	}

	// 确保数据长度足够
	if len(data) < 18 {
		return nil, fmt.Errorf("数据长度不足")
	}

	// 检查是否是F254主动上报
	if !IsF254ActiveReport(data) {
		return nil, fmt.Errorf("非F254主动上报数据")
	}

	// 提取表地址（通常在用户数据区域）
	if len(data) >= 30 {
		// 表号通常在用户数据区域的特定位置，这里假设在位置20-25
		var tableNumber string
		for i := 20; i < 26 && i < len(data); i++ {
			tableNumber += fmt.Sprintf("%02X", data[i])
		}
		event.TableNumber = tableNumber
	}

	// 设置事件类型
	event.EventType = "f254_report" // F254主动上报

	// 解析用户数据区域
	if len(data) >= 40 {
		// 这里根据3761-BY协议的具体格式解析用户数据
		// 示例：解析电能数据
		parseEnergyData(data, event)
	}

	return event, nil
}

// parseEnergyData 解析电能数据
func parseEnergyData(data []byte, event *Event3761) {
	if len(data) < 40 {
		zap.S().Warnln("电能数据长度不足")
		return
	}

	// 解析电能值（示例位置30-33）
	if len(data) >= 34 {
		energyValue := uint32(data[30]) +
			(uint32(data[31]) << 8) +
			(uint32(data[32]) << 16) +
			(uint32(data[33]) << 24)
		event.EventData["energy"] = float64(energyValue) / 100.0 // 单位kWh
	}

	// 解析功率（示例位置34-35）
	if len(data) >= 36 {
		power := (uint16(data[34]) << 8) | uint16(data[35])
		event.EventData["power"] = float64(power) / 10.0 // 单位kW
	}

	// 解析电压（示例位置36-37）
	if len(data) >= 38 {
		voltage := (uint16(data[36]) << 8) | uint16(data[37])
		event.EventData["voltage"] = float64(voltage) / 10.0 // 单位V
	}

	// 解析电流（示例位置38-39）
	if len(data) >= 40 {
		current := (uint16(data[38]) << 8) | uint16(data[39])
		event.EventData["current"] = float64(current) / 100.0 // 单位A
	}

	// 解析状态字（示例位置40-41）
	if len(data) >= 42 {
		statusWord := (uint16(data[40]) << 8) | uint16(data[41])
		event.EventData["status_word"] = statusWord

		// 解析状态位
		event.EventData["cover_open"] = (statusWord & 0x0200) != 0   // 开表盖
		event.EventData["switch_event"] = (statusWord & 0x0010) != 0 // 开合闸事件
	}
}

// HandleF254Report 处理F254主动上报事件
func HandleF254Report(event *Event3761) {
	zap.S().Infof("处理F254主动上报事件 - 设备: %s, 表号: %s",
		event.DeviceAddr, event.TableNumber)

	if energy, ok := event.EventData["energy"].(float64); ok {
		zap.S().Infof("电能: %.2f kWh", energy)
	}

	if power, ok := event.EventData["power"].(float64); ok {
		zap.S().Infof("功率: %.2f kW", power)
	}

	if voltage, ok := event.EventData["voltage"].(float64); ok {
		zap.S().Infof("电压: %.2f V", voltage)
	}

	if current, ok := event.EventData["current"].(float64); ok {
		zap.S().Infof("电流: %.2f A", current)
	}

	// 这里可以添加具体的业务逻辑
	// 例如：保存数据到数据库、触发告警等
}

// HandleOverloadEvent 处理过载事件
func HandleOverloadEvent(event *Event3761) {
	zap.S().Warnf("处理过载事件 - 设备: %s, 表号: %s",
		event.DeviceAddr, event.TableNumber)

	if power, ok := event.EventData["power"].(float64); ok {
		zap.S().Warnf("功率: %.2f kW", power)
	}

	if current, ok := event.EventData["current"].(float64); ok {
		zap.S().Warnf("电流: %.2f A", current)
	}

	// 这里可以添加具体的业务逻辑
	// 例如：发送过载告警、记录过载事件等
}

// HandleCoverOpenEvent 处理开表盖事件
func HandleCoverOpenEvent(event *Event3761) {
	zap.S().Warnf("处理开表盖事件 - 设备: %s, 表号: %s",
		event.DeviceAddr, event.TableNumber)

	// 这里可以添加具体的业务逻辑
	// 例如：发送开表盖告警、记录开表盖事件等
}

// handle3761Event 处理解析后的3761-BY事件
func handle3761Event(event *Event3761) {
	if event == nil {
		zap.S().Warnln("传入的3761-BY事件为空")
		return
	}

	// 查找并调用对应的事件处理函数
	handler, exists := eventHandlers3761[event.EventType]
	if exists {
		zap.S().Infof("执行3761-BY事件处理函数: %s, 设备: %s", event.EventType, event.DeviceAddr)
		handler(event)
	} else {
		zap.S().Warnf("未找到3761-BY事件处理函数: %s, 设备: %s", event.EventType, event.DeviceAddr)

		// 针对未注册处理函数的事件类型，调用通用处理逻辑
		handleGeneric3761Event(event)
	}
}

// handleGeneric3761Event 处理通用3761-BY事件
func handleGeneric3761Event(event *Event3761) {
	// 打印事件基本信息
	zap.S().Infof("通用处理3761-BY事件: 类型=%s, 设备=%s, 时间=%s",
		event.EventType, event.DeviceAddr, event.EventTime.Format("2006-01-02 15:04:05"))

	// 打印所有事件数据
	for key, value := range event.EventData {
		zap.S().Infof("事件数据: %s = %v", key, value)
	}

	// 这里可以添加通用事件处理逻辑，如：
	// 1. 保存数据到数据库
	// 2. 转发到其他系统
	// 3. 触发告警等
}

// ConvertToDeviceAndPublish3761 将事件转换为设备对象并触发DeviceCollectFinish事件
func ConvertToDeviceAndPublish3761(event *Event3761, deviceRepository repository.DeviceRepository, eventBus *event.EventService) {
	if event == nil {
		zap.S().Warnln("传入的3761-BY事件为空，无法转换为设备对象")
		return
	}

	// 查找设备 - 使用表号查找
	device, err := deviceRepository.FindByAddr(event.DeviceAddr)
	if err != nil {
		zap.S().Errorf("查询设备列表失败: %v", err)
		return
	}

	// 更新设备状态
	device.Online = true
	device.CollectTime = time.Now().Unix()
	device.CollectTotal += 1
	device.CollectSuccess += 1

	// 更新设备属性
	for i, property := range device.Type.Properties {
		switch property.Name {
		case "dev_consumption": // 电能
			if energy, ok := event.EventData["energy"].(float64); ok {
				device.Type.Properties[i].Value = energy
			}
		case "dev_power": // 功率
			if power, ok := event.EventData["power"].(float64); ok {
				device.Type.Properties[i].Value = power
			}
		case "dev_voltage": // 电压
			if voltage, ok := event.EventData["voltage"].(float64); ok {
				device.Type.Properties[i].Value = voltage
			}
		case "dev_current": // 电流
			if current, ok := event.EventData["current"].(float64); ok {
				device.Type.Properties[i].Value = current
			}
		case "dev_status": // 状态
			if status, ok := event.EventData["status_word"].(uint16); ok {
				device.Type.Properties[i].Value = float64(status)
			}
		}
	}

	// 处理告警状态
	alarmReasons := []string{}

	// 根据状态字判断告警
	if statusWord, ok := event.EventData["status_word"].(uint16); ok {
		if (statusWord & 0x0200) != 0 {
			alarmReasons = append(alarmReasons, "开表盖")
		}
		if (statusWord & 0x0010) != 0 {
			alarmReasons = append(alarmReasons, "开合闸事件")
		}
	}

	// 检查过载
	if _, ok := event.EventData["phase_a_overload"]; ok {
		alarmReasons = append(alarmReasons, "A相过载")
	}

	if len(alarmReasons) > 0 {
		device.AlarmStatus = true
		device.AlarmReason = strings.Join(alarmReasons, ",")
		device.AlarmTime = time.Now().Unix()
		device.AlarmTotal += 1
	} else {
		device.AlarmStatus = false
		device.AlarmReason = ""
	}

	// 保存设备状态
	err = deviceRepository.Save(&device)
	if err != nil {
		zap.S().Errorf("保存设备[%s]状态失败: %v", device.Name, err)
		return
	}

	// 触发DeviceCollectFinish事件
	zap.S().Infof("触发设备[%s]采集完成事件", device.Name)
	var e struct {
		Name string
		Data interface{}
	}
	e.Name = "DeviceCollectFinish"
	e.Data = device
	eventBus.Publish(e)
}

// Init3761EventHandlers 初始化3761-BY事件处理函数
func Init3761EventHandlers() {
	// 注册各种事件处理函数
	RegisterEventHandler3761("f254_report", HandleF254Report)
	RegisterEventHandler3761("overload", HandleOverloadEvent)
	RegisterEventHandler3761("cover_open", HandleCoverOpenEvent)

	zap.S().Infoln("3761-BY事件处理函数初始化完成")
}
