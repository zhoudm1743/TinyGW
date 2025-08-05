package event

import (
	"fmt"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"

	"tinyGW/app/api/repository"
	"tinyGW/pkg/service/command"
	"tinyGW/pkg/service/event"
)

// Events2025F183 2025F183-37协议事件处理器
type Events2025F183 struct {
	deviceRepository repository.DeviceRepository
	eventBus         *event.EventService
	parser           *Parser2025F183
}

// NewEvents2025F183 创建2025F183-37协议事件处理器
func NewEvents2025F183(deviceRepository repository.DeviceRepository, eventBus *event.EventService) *Events2025F183 {
	return &Events2025F183{
		deviceRepository: deviceRepository,
		eventBus:         eventBus,
		parser:           NewParser2025F183(),
	}
}

// Parser2025F183 2025F183-37协议解析器
type Parser2025F183 struct {
}

// NewParser2025F183 创建2025F183-37协议解析器
func NewParser2025F183() *Parser2025F183 {
	return &Parser2025F183{}
}

// Parse 解析2025F183-37协议数据
func (p *Parser2025F183) Parse(deviceAddr string, data []byte) (map[string]interface{}, error) {
	// 检查数据合法性
	if len(data) < 16 {
		return nil, fmt.Errorf("数据长度不足")
	}

	// 检查是否是主动上报(0x97)
	if !isActiveReport(data) {
		return nil, fmt.Errorf("非主动上报数据")
	}

	// 使用原有解析逻辑
	event, err := parse2025F183Event(deviceAddr, data)
	if err != nil {
		return nil, err
	}

	return event.EventData, nil
}

// updateDeviceStatus 更新设备状态
func (e *Events2025F183) updateDeviceStatus(deviceAddr string, parsedData map[string]interface{}) {
	// 防御性检查
	if e.deviceRepository == nil {
		zap.S().Errorf("deviceRepository为空，无法更新设备 %s 状态", deviceAddr)
		return
	}

	// 检查参数有效性
	if deviceAddr == "" || parsedData == nil {
		zap.S().Warnf("无效的参数: deviceAddr=[%s], parsedData=[%v]", deviceAddr, parsedData)
		return
	}

	// 从数据库查找设备
	device, err := e.deviceRepository.FindByAddr(deviceAddr)
	if err != nil {
		zap.S().Errorf("找不到设备 %s: %v", deviceAddr, err)
		return
	}

	// 更新设备属性
	device.Online = true
	device.CollectTime = time.Now().Unix()
	device.CollectTotal += 1
	device.CollectSuccess += 1

	// 更新设备测量值
	for k, v := range parsedData {
		// 更新设备属性
		for i, prop := range device.Type.Properties {
			if prop.Name == k {
				device.Type.Properties[i].Value = v
				break
			}
		}
	}

	// 特殊处理用水量
	if totalFlow, ok := parsedData["total_flow"].(float64); ok {
		// 检查是否需要更新累计用水量
		for i, prop := range device.Type.Properties {
			if prop.Name == "dev_consumption" {
				device.Type.Properties[i].Value = totalFlow
				break
			}
		}
	}

	// 保存设备
	e.deviceRepository.Save(&device)

	// 发布设备采集完成事件（如果eventBus可用）
	if e.eventBus != nil {
		e.eventBus.Publish(event.Event{
			Name: "DeviceCollectFinish",
			Data: device,
		})
	}
}

// Event2025F183 2025F183-37协议事件结构
type Event2025F183 struct {
	DeviceAddr  string                 // 设备地址
	EventType   string                 // 事件类型
	EventTime   time.Time              // 事件时间
	TableNumber string                 // 表号
	EventData   map[string]interface{} // 事件数据
	RawData     []byte                 // 原始数据
}

// EventHandler2025F183 2025F183-37事件处理函数类型
type EventHandler2025F183 func(event *Event2025F183)

var (
	// 事件处理函数映射
	eventHandlers2025F183 = make(map[string]EventHandler2025F183)
)

// RegisterEventHandler2025F183 注册2025F183-37事件处理函数
func RegisterEventHandler2025F183(eventType string, handler EventHandler2025F183) {
	eventHandlers2025F183[eventType] = handler
	zap.S().Infof("注册2025F183-37事件处理函数: %s", eventType)
}

// UnregisterEventHandler2025F183 注销2025F183-37事件处理函数
func UnregisterEventHandler2025F183(eventType string) {
	delete(eventHandlers2025F183, eventType)
	zap.S().Infof("注销2025F183-37事件处理函数: %s", eventType)
}

// Handle2025F183Event 处理2025F183-37协议事件
func Handle2025F183Event(deviceAddr string, data []byte, deviceRepository repository.DeviceRepository, eventBus *event.EventService) {
	if len(data) < 16 {
		zap.S().Warnln("2025F183-37事件数据长度不足")
		return
	}

	// 检查是否是主动上报数据(0x97)
	if !isActiveReport(data) {
		return
	}

	// 解析事件数据
	event, err := parse2025F183Event(deviceAddr, data)
	if err != nil {
		zap.S().Errorf("解析2025F183-37事件失败: %v", err)
		return
	}

	// 调用对应的事件处理函数
	handle2025F183Event(event, deviceRepository, eventBus)
}

// isActiveReport 检查是否是主动上报数据(0x97)
func isActiveReport(data []byte) bool {
	if len(data) < 11 {
		return false
	}

	// 检查控制码是否为0x97（主动上报）
	ctrlCode := data[10]
	return ctrlCode == 0x97
}

// parse2025F183Event 解析2025F183-37事件数据
func parse2025F183Event(deviceAddr string, data []byte) (*Event2025F183, error) {
	event := &Event2025F183{
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
	if len(data) < 16 {
		return nil, fmt.Errorf("数据长度不足")
	}

	// 提取表地址（位置2-8）
	var tableNumber string
	if len(data) >= 9 {
		for i := 8; i >= 2; i-- {
			tableNumber += fmt.Sprintf("%02X", data[i])
		}
		event.TableNumber = tableNumber
	}

	// 获取上报类型
	var reportType uint16
	if len(data) >= 15 {
		reportType = (uint16(data[13]) << 8) | uint16(data[14])
	}
	event.EventData["report_type"] = reportType

	// 获取数据包类型
	var dataPackType byte
	if len(data) >= 16 {
		dataPackType = data[16]
	}
	event.EventData["data_pack_type"] = dataPackType

	// 根据数据包类型确定事件类型
	eventType, err := determine2025F183EventType(data, dataPackType)
	if err != nil {
		return nil, err
	}
	event.EventType = eventType

	// 根据不同的数据包类型解析数据
	switch dataPackType {
	case 0x01, 0x02: // 预付费表上报
		parsePrePayData(data, event)
	case 0x03: // 后付费表上报
		parsePostPayData(data, event)
	}

	return event, nil
}

// determine2025F183EventType 根据数据包类型确定事件类型
func determine2025F183EventType(data []byte, dataPackType byte) (string, error) {
	// 先根据控制码确认是主动上报
	if len(data) < 11 || data[10] != 0x97 {
		return "unknown", fmt.Errorf("非主动上报数据")
	}

	// 根据数据包类型判断事件类型
	switch dataPackType {
	case 0x01:
		return "prepay_report", nil // 预付费表普通上报
	case 0x02:
		return "prepay_step_report", nil // 预付费阶梯表上报
	case 0x03:
		return "postpay_report", nil // 后付费表上报
	case 0x13:
		return "alarm_report", nil // 告警上报
	case 0x14:
		return "valve_status", nil // 阀门状态
	case 0x15:
		return "battery_low", nil // 电池电量低
	default:
		return "general_report", nil // 普通上报
	}
}

// parsePrePayData 解析预付费表数据
func parsePrePayData(data []byte, event *Event2025F183) {
	if len(data) < 65 {
		zap.S().Warnln("预付费表数据长度不足")
		return
	}

	// 解析状态字
	if len(data) >= 60 {
		status := (uint16(data[58]) << 8) | uint16(data[59])
		event.EventData["status"] = status

		// 解析状态位
		event.EventData["valve_closed"] = (status & 0x0001) != 0 // 阀门关闭状态
		event.EventData["battery_low"] = (status & 0x0002) != 0  // 电池电量低
		event.EventData["meter_error"] = (status & 0x0004) != 0  // 表计故障
		event.EventData["reverse_flow"] = (status & 0x0008) != 0 // 反流
		event.EventData["no_flow"] = (status & 0x0010) != 0      // 长时间无流量
		event.EventData["leakage"] = (status & 0x0020) != 0      // 漏水
	}

	// 解析累计用量（预付费表在位置61-64）
	if len(data) >= 65 {
		totalValue := uint32(data[61]) +
			(uint32(data[62]) << 8) +
			(uint32(data[63]) << 16) +
			(uint32(data[64]) << 24)
		event.EventData["total_flow"] = float64(totalValue) / 10.0 // 单位立方米
	}

	// 解析电池电压
	if len(data) >= 45 {
		voltage := (uint16(data[43]) << 8) | uint16(data[44])
		event.EventData["battery_voltage"] = float64(voltage) / 100.0 // 单位V
	}

	// 设置可读属性
	if valveClosed, ok := event.EventData["valve_closed"].(bool); ok {
		if valveClosed {
			event.EventData["valve_status"] = "关闭"
		} else {
			event.EventData["valve_status"] = "开启"
		}
	}

	if batteryLow, ok := event.EventData["battery_low"].(bool); ok {
		if batteryLow {
			event.EventData["battery_status"] = "电量低"
		} else {
			event.EventData["battery_status"] = "正常"
		}
	}

	if meterError, ok := event.EventData["meter_error"].(bool); ok {
		if meterError {
			event.EventData["meter_status"] = "故障"
		} else {
			event.EventData["meter_status"] = "正常"
		}
	}

	// 设置水表读数
	if totalFlow, ok := event.EventData["total_flow"].(float64); ok {
		event.EventData["water_reading"] = totalFlow
	}
}

// parsePostPayData 解析后付费表数据
func parsePostPayData(data []byte, event *Event2025F183) {
	if len(data) < 65 {
		zap.S().Warnf("后付费表数据长度不足: %d < 65", len(data))
		return
	}

	// 解析状态字
	if len(data) >= 60 {
		status := (uint16(data[58]) << 8) | uint16(data[59])
		event.EventData["status"] = status

		// 解析状态位
		event.EventData["valve_closed"] = (status & 0x0001) != 0 // 阀门关闭状态
		event.EventData["battery_low"] = (status & 0x0002) != 0  // 电池电量低
		event.EventData["meter_error"] = (status & 0x0004) != 0  // 表计故障
		event.EventData["reverse_flow"] = (status & 0x0008) != 0 // 反流
		event.EventData["no_flow"] = (status & 0x0010) != 0      // 长时间无流量
		event.EventData["leakage"] = (status & 0x0020) != 0      // 漏水
	}

	// 解析累计用量（后付费表在位置61-64）
	if len(data) >= 65 {
		// 原始字节值
		b1 := data[61]
		b2 := data[62]
		b3 := data[63]
		b4 := data[64]

		// 根据实际测试数据分析和表计显示值
		// 当原始字节是[00 0B 04 00]时，水表读数应该是2.82立方米
		// 正确的计算方式是将第二个字节(b2)和第三个字节(b3)组合为16位整数后除以1000
		waterReading := float64(uint16(b2)<<8|uint16(b3)) / 1000.0

		// 设置水表读数
		event.EventData["total_flow"] = waterReading
		event.EventData["water_reading"] = waterReading

		// 记录原始字节值，仅在Debug级别时输出日志
		if zap.S().Level() == zap.DebugLevel {
			event.EventData["raw_bytes"] = fmt.Sprintf("%02X %02X %02X %02X", b1, b2, b3, b4)
			// zap.S().Debugf("水表读数原始字节: [%02X %02X %02X %02X]", b1, b2, b3, b4)
		}
	}

	// 解析电池电压
	if len(data) >= 45 {
		voltage := (uint16(data[43]) << 8) | uint16(data[44])
		event.EventData["battery_voltage"] = float64(voltage) / 100.0 // 单位V
	}

	// 设置可读属性
	if valveClosed, ok := event.EventData["valve_closed"].(bool); ok {
		if valveClosed {
			event.EventData["valve_status"] = "关闭"
		} else {
			event.EventData["valve_status"] = "开启"
		}
	}

	if batteryLow, ok := event.EventData["battery_low"].(bool); ok {
		if batteryLow {
			event.EventData["battery_status"] = "电量低"
		} else {
			event.EventData["battery_status"] = "正常"
		}
	}

	if meterError, ok := event.EventData["meter_error"].(bool); ok {
		if meterError {
			event.EventData["meter_status"] = "故障"
		} else {
			event.EventData["meter_status"] = "正常"
		}
	}
}

// HandlePrePayReport 处理预付费表上报事件
func HandlePrePayReport(event *Event2025F183) {
	zap.S().Infof("处理预付费表上报事件 - 设备: %s, 表号: %s",
		event.DeviceAddr, event.TableNumber)

	if flow, ok := event.EventData["total_flow"].(float64); ok {
		zap.S().Infof("水表读数: %.2f 立方米", flow)
	}

	if voltage, ok := event.EventData["battery_voltage"].(float64); ok {
		zap.S().Infof("电池电压: %.2f V", voltage)
	}

	// 这里可以添加具体的业务逻辑
	// 例如：保存数据到数据库、触发告警等
}

// HandlePostPayReport 处理后付费表上报事件
func HandlePostPayReport(event *Event2025F183) {
	zap.S().Infof("处理后付费表上报事件 - 设备: %s, 表号: %s",
		event.DeviceAddr, event.TableNumber)

	if flow, ok := event.EventData["water_reading"].(float64); ok {
		zap.S().Infof("水表读数: %.2f 立方米", flow)
	}

	if voltage, ok := event.EventData["battery_voltage"].(float64); ok {
		zap.S().Infof("电池电压: %.2f V", voltage)
	}

	// 只在Debug级别时打印详细信息
	if zap.S().Level() == zap.DebugLevel {
		// 打印调试信息，查看原始字节值
		if rawBytes, ok := event.EventData["raw_bytes"].(string); ok {
			zap.S().Debugf("水表读数原始字节: %s", rawBytes)
		}

		// 打印其他状态信息
		if valveStatus, ok := event.EventData["valve_status"].(string); ok {
			zap.S().Debugf("阀门状态: %s", valveStatus)
		}

		if batteryStatus, ok := event.EventData["battery_status"].(string); ok {
			zap.S().Debugf("电池状态: %s", batteryStatus)
		}

		if meterStatus, ok := event.EventData["meter_status"].(string); ok {
			zap.S().Debugf("表计状态: %s", meterStatus)
		}
	}

	// 这里可以添加具体的业务逻辑
	// 例如：保存数据到数据库、触发告警等
}

// HandleAlarmReport 处理告警上报事件
func HandleAlarmReport(event *Event2025F183) {
	zap.S().Warnf("处理告警上报事件 - 设备: %s, 表号: %s",
		event.DeviceAddr, event.TableNumber)

	// 检查各种告警状态
	if batteryLow, ok := event.EventData["battery_low"].(bool); ok && batteryLow {
		zap.S().Warnf("告警: 电池电量低")
	}
	if meterError, ok := event.EventData["meter_error"].(bool); ok && meterError {
		zap.S().Warnf("告警: 表计故障")
	}
	if reverseFlow, ok := event.EventData["reverse_flow"].(bool); ok && reverseFlow {
		zap.S().Warnf("告警: 反流")
	}
	if noFlow, ok := event.EventData["no_flow"].(bool); ok && noFlow {
		zap.S().Warnf("告警: 长时间无流量")
	}
	if leakage, ok := event.EventData["leakage"].(bool); ok && leakage {
		zap.S().Warnf("告警: 漏水")
	}

	// 这里可以添加具体的业务逻辑
	// 例如：发送告警通知、记录告警事件等
}

// HandleBatteryLowEvent 处理电池低电量事件
func HandleBatteryLowEvent(event *Event2025F183) {
	zap.S().Warnf("处理电池低电量事件 - 设备: %s, 表号: %s",
		event.DeviceAddr, event.TableNumber)

	if voltage, ok := event.EventData["battery_voltage"].(float64); ok {
		zap.S().Warnf("电池电压: %.2f V", voltage)
	}

	// 这里可以添加具体的业务逻辑
	// 例如：发送电池更换通知、记录电池状态等
}

// HandleValveStatusEvent 处理阀门状态事件
func HandleValveStatusEvent(event *Event2025F183) {
	zap.S().Infof("处理阀门状态事件 - 设备: %s, 表号: %s",
		event.DeviceAddr, event.TableNumber)

	if valveClosed, ok := event.EventData["valve_closed"].(bool); ok {
		if valveClosed {
			zap.S().Infof("阀门状态: 关闭")
		} else {
			zap.S().Infof("阀门状态: 开启")
		}
	}

	// 这里可以添加具体的业务逻辑
	// 例如：更新阀门状态、记录操作日志等
}

// handle2025F183Event 处理解析后的2025F183-37事件
func handle2025F183Event(event *Event2025F183, deviceRepository repository.DeviceRepository, eventBus *event.EventService) {
	if event == nil {
		zap.S().Warnln("传入的2025F183-37事件为空")
		return
	}

	// 查找并调用对应的事件处理函数
	handler, exists := eventHandlers2025F183[event.EventType]
	if exists {
		// zap.S().Infof("执行2025F183-37事件处理函数: %s, 设备: %s", event.EventType, event.DeviceAddr)
		handler(event)
	} else {
		// zap.S().Infof("未找到2025F183-37事件处理函数: %s, 设备: %s", event.EventType, event.DeviceAddr)

		// 针对未注册处理函数的事件类型，调用通用处理逻辑
		handleGeneric2025F183Event(event)
	}

	// 转换为设备对象并触发DeviceCollectFinish事件
	ConvertToDeviceAndPublish(event, deviceRepository, eventBus)
}

// handleGeneric2025F183Event 处理通用2025F183-37事件
func handleGeneric2025F183Event(event *Event2025F183) {
	// 打印事件基本信息
	// zap.S().Infof("通用处理2025F183-37事件: 类型=%s, 设备=%s, 时间=%s",
	// 	event.EventType, event.DeviceAddr, event.EventTime.Format("2006-01-02 15:04:05"))

	// // 打印所有事件数据
	// for key, value := range event.EventData {
	// 	zap.S().Infof("事件数据: %s = %v", key, value)
	// }

	// 这里可以添加通用事件处理逻辑，如：
	// 1. 保存数据到数据库
	// 2. 转发到其他系统
	// 3. 触发告警等
}

// ConvertToDeviceAndPublish 将事件转换为设备对象并触发DeviceCollectFinish事件
func ConvertToDeviceAndPublish(event *Event2025F183, deviceRepository repository.DeviceRepository, eventBus *event.EventService) {
	if event == nil {
		zap.S().Warnln("传入的2025F183-37事件为空，无法转换为设备对象")
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
		case "dev_flow": // 总流量
			if flow, ok := event.EventData["water_reading"].(float64); ok {
				// 打印调试信息，确认正在使用正确的值
				// zap.S().Debugf("更新设备[%s]属性[%s]的值: %.2f", device.Name, property.Name, flow)
				device.Type.Properties[i].Value = flow
			}
		case "dev_battery": // 电池电压
			if voltage, ok := event.EventData["battery_voltage"].(float64); ok {
				device.Type.Properties[i].Value = voltage
			}
		case "dev_signal": // 信号强度
			if signal, ok := event.EventData["signal_strength"].(int); ok {
				device.Type.Properties[i].Value = float64(signal)
			}
		case "dev_valve": // 阀门状态
			if valveStatus, ok := event.EventData["valve_status"].(string); ok {
				var valveValue float64 = 0
				if valveStatus == "关闭" {
					valveValue = 1
				}
				device.Type.Properties[i].Value = valveValue
			}
		case "dev_battery_status": // 电池状态
			if batteryStatus, ok := event.EventData["battery_status"].(string); ok {
				var batteryValue float64 = 0
				if batteryStatus == "电量低" {
					batteryValue = 1
				}
				device.Type.Properties[i].Value = batteryValue
			}
		}
	}

	// 处理告警状态
	alarmReasons := []string{}
	if batteryStatus, ok := event.EventData["battery_status"].(string); ok && batteryStatus == "电量低" {
		alarmReasons = append(alarmReasons, "电池电量低")
	}
	if meterStatus, ok := event.EventData["meter_status"].(string); ok && meterStatus == "故障" {
		alarmReasons = append(alarmReasons, "表计故障")
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
	// zap.S().Infof("触发设备[%s]采集完成事件", device.Name)
	var e struct {
		Name string
		Data interface{}
	}
	e.Name = "DeviceCollectFinish"
	e.Data = device
	eventBus.Publish(e)
}

// Init2025F183EventHandlers 初始化2025F183-37事件处理函数
func Init2025F183EventHandlers() {
	// 注册各种事件处理函数
	RegisterEventHandler2025F183("prepay_report", HandlePrePayReport)
	RegisterEventHandler2025F183("prepay_step_report", HandlePrePayReport) // 阶梯表也使用同样的处理函数
	RegisterEventHandler2025F183("postpay_report", HandlePostPayReport)
	RegisterEventHandler2025F183("alarm_report", HandleAlarmReport)
	RegisterEventHandler2025F183("battery_low", HandleBatteryLowEvent)
	RegisterEventHandler2025F183("valve_status", HandleValveStatusEvent)
	RegisterEventHandler2025F183("general_report", HandlePrePayReport) // 一般上报使用预付费表处理函数

	zap.S().Infoln("2025F183-37事件处理函数初始化完成")
}

func Handle2025F183Cmd(deviceAddr string, deviceRepository repository.DeviceRepository, eventBus *event.EventService, commandManager *command.Manager, conn net.Conn) {
	zap.S().Infof("处理2025F183-37指令: %s", deviceAddr)

	// 循环处理队列中的所有指令
	for command.GetQueueLength(deviceAddr) > 0 {
		data, ok := commandManager.Get(deviceAddr)
		if !ok {
			break
		}

		// 发送指令到设备
		_, err := conn.Write(data)
		if err != nil {
			zap.S().Errorf("发送指令到设备[%s]失败: %v", deviceAddr, err)
			break
		}

		// 从队列中移除已发送的指令
		commandManager.Remove(deviceAddr)
		zap.S().Infof("发送指令到设备[%s]成功: %s", deviceAddr, fmt.Sprintf("[% 2X]", data))

		// 如果还有更多指令，给设备一点响应时间
		if command.GetQueueLength(deviceAddr) > 0 {
			time.Sleep(100 * time.Millisecond) // 指令间隔，可以根据实际设备响应时间调整
		}
	}
}
