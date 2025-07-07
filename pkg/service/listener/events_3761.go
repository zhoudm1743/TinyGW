package listener

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
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

// Handle3761Event 处理3761-BY协议事件
func Handle3761Event(deviceAddr string, data []byte) {
	if len(data) < 18 {
		zap.S().Warnln("3761-BY事件数据长度不足")
		return
	}

	// 检查是否是F254主动上报数据
	if !isF254ActiveReport(data) {
		return
	}

	// 解析事件数据
	event, err := parse3761Event(deviceAddr, data)
	if err != nil {
		zap.S().Errorf("解析3761-BY事件失败: %v", err)
		return
	}

	// 调用对应的事件处理函数
	if handler, exists := eventHandlers3761[event.EventType]; exists {
		handler(event)
	} else {
		zap.S().Warnf("未找到事件处理函数: %s", event.EventType)
	}
}

// isF254ActiveReport 检查是否是F254主动上报数据
func isF254ActiveReport(data []byte) bool {
	if len(data) < 18 {
		return false
	}

	// 检查AFN=10H和Fn=254
	afn := data[12] // AFN在第13字节（索引12）
	if afn != 0x10 {
		return false
	}

	// 检查DT1和DT2计算Fn值
	dt1 := data[16] // DT1
	dt2 := data[17] // DT2

	// 计算Fn值
	fn := 0
	if dt1 != 0 {
		for i := 0; i < 8; i++ {
			if (dt1 & (1 << i)) != 0 {
				fn = int(dt2)*8 + (i + 1)
				break
			}
		}
	}

	return fn == 254
}

// parse3761Event 解析3761-BY事件数据
func parse3761Event(deviceAddr string, data []byte) (*Event3761, error) {
	event := &Event3761{
		DeviceAddr: deviceAddr,
		EventData:  make(map[string]interface{}, 8), // 预分配合理容量，减少扩容
	}

	// 只在Debug模式下复制原始数据
	if zap.S().Level() == zap.DebugLevel {
		event.RawData = make([]byte, len(data))
		copy(event.RawData, data)
	}

	// 解析透明转发内容 - 使用更高效的提取方式
	transparentData, err := extractTransparentDataOptimized(data)
	if err != nil {
		return nil, fmt.Errorf("提取透明转发数据失败: %v", err)
	}

	// 解析DLT645数据域
	dlt645Data, err := parseDLT645DataOptimized(transparentData)
	if err != nil {
		return nil, fmt.Errorf("解析DLT645数据失败: %v", err)
	}

	// 根据TAG判断事件类型
	eventType, err := determineEventType(dlt645Data)
	if err != nil {
		return nil, fmt.Errorf("判断事件类型失败: %v", err)
	}

	event.EventType = eventType
	event.EventData = dlt645Data

	// 提取表号和时间
	if tableNumber, ok := dlt645Data["table_number"]; ok {
		event.TableNumber = fmt.Sprintf("%v", tableNumber)
	}

	if eventTime, ok := dlt645Data["event_time"]; ok {
		if t, ok := eventTime.(time.Time); ok {
			event.EventTime = t
		}
	}

	return event, nil
}

// extractTransparentDataOptimized 提取透明转发数据的优化版本
func extractTransparentDataOptimized(data []byte) ([]byte, error) {
	// 快速检查数据长度
	if len(data) < 18 {
		return nil, fmt.Errorf("数据长度不足")
	}

	// 查找0xFD标识符 - 使用更高效的内存搜索
	for i := 18; i < len(data)-1; i++ {
		if data[i] == 0xFD {
			// 直接返回切片引用，避免额外内存分配
			return data[i:], nil
		}
	}

	return nil, fmt.Errorf("未找到透明转发标识符0xFD")
}

// parseDLT645DataOptimized 解析DLT645数据域的优化版本
func parseDLT645DataOptimized(transparentData []byte) (map[string]interface{}, error) {
	if len(transparentData) < 8 {
		return nil, fmt.Errorf("透明转发数据长度不足")
	}

	// 预分配合理容量，减少map扩容
	result := make(map[string]interface{}, 16)

	// 跳过0xFD标识符和长度域
	offset := 1 + 2 + 2 + 2 // 0xFD + 帧长度 + 控制字 + 时间控制字

	// 读取TAG个数
	if offset >= len(transparentData) {
		return nil, fmt.Errorf("数据长度不足，无法读取TAG个数")
	}
	tagCount := int(transparentData[offset])
	offset++

	// 限制最大TAG数量，防止异常数据
	if tagCount > 32 {
		zap.S().Warnf("TAG数量异常: %d，限制为32", tagCount)
		tagCount = 32
	}

	// 解析每个TAG数据对
	for i := 0; i < tagCount && offset+3 < len(transparentData); i++ {
		// 读取TAG（2字节）
		if offset+1 >= len(transparentData) {
			break
		}
		tag := binary.BigEndian.Uint16(transparentData[offset:])
		offset += 2

		// 根据TAG解析数据
		dataValue, newOffset, err := parseTagDataOptimized(tag, transparentData, offset)
		if err != nil {
			// 降级为Debug日志，减少日志量
			zap.S().Debugf("解析TAG 0x%04X数据失败: %v", tag, err)
			// 尝试跳过这个TAG而不是终止整个解析
			offset += 2 // 尝试跳过数据，移动到下一个可能的TAG位置
			continue
		}

		// 存储解析结果 - 使用预定义的TAG键名
		tagKey := getTagKey(tag)
		result[tagKey] = dataValue

		// 特殊处理常用TAG
		switch tag {
		case 0x0001: // 表号
			result["table_number"] = dataValue
		case 0x0002: // 时间
			if timeValue, ok := dataValue.(time.Time); ok {
				result["event_time"] = timeValue
			}
		case 0x0014: // 状态字
			result["status_word"] = dataValue
		case 0x0093: // A相过载
			result["phase_a_overload"] = dataValue
		case 0x0096: // 开表盖
			result["cover_opened"] = dataValue
		case 0xFFFF: // 自定义数据
			result["custom_data"] = dataValue
		}

		offset = newOffset
	}

	return result, nil
}

// getTagKey 返回TAG的预定义键名，避免重复的字符串格式化
func getTagKey(tag uint16) string {
	switch tag {
	case 0x0001:
		return "tag_0001"
	case 0x0002:
		return "tag_0002"
	case 0x0014:
		return "tag_0014"
	case 0x0093:
		return "tag_0093"
	case 0x0096:
		return "tag_0096"
	case 0xFFFF:
		return "tag_FFFF"
	default:
		// 针对其他TAG，缓存格式化后的字符串
		return fmt.Sprintf("tag_%04X", tag)
	}
}

// parseTagDataOptimized 优化的TAG数据解析函数
func parseTagDataOptimized(tag uint16, data []byte, offset int) (interface{}, int, error) {
	switch tag {
	case 0x0001: // 表号（6字节）
		if offset+6 > len(data) {
			return nil, offset, fmt.Errorf("数据长度不足，无法读取表号")
		}
		// 使用预分配的缓冲区和更高效的字符串构建
		var sb strings.Builder
		sb.Grow(12) // 预分配空间：6个字节，每个字节2个十六进制字符
		for i := 0; i < 6; i++ {
			fmt.Fprintf(&sb, "%02X", data[offset+i])
		}
		return sb.String(), offset + 6, nil

	case 0x0002: // 时间（7字节）
		if offset+7 > len(data) {
			return nil, offset, fmt.Errorf("数据长度不足，无法读取时间")
		}
		// 时间格式：YY MM DD WW hh mm ss
		year := 2000 + int(data[offset])
		month := time.Month(data[offset+1])
		day := int(data[offset+2])
		// weekday := time.Weekday(data[offset+3]) // 暂时不使用星期信息
		hour := int(data[offset+4])
		minute := int(data[offset+5])
		second := int(data[offset+6])

		// 检查日期有效性，避免无效日期导致panic
		if month < 1 || month > 12 || day < 1 || day > 31 || hour > 23 || minute > 59 || second > 59 {
			return nil, offset, fmt.Errorf("无效的日期时间数据")
		}

		// 创建时间对象
		t := time.Date(year, month, day, hour, minute, second, 0, time.Local)
		return t, offset + 7, nil

	case 0x0014: // 状态字（2字节）
		if offset+2 > len(data) {
			return nil, offset, fmt.Errorf("数据长度不足，无法读取状态字")
		}
		statusWord := binary.BigEndian.Uint16(data[offset:])
		return statusWord, offset + 2, nil

	case 0x0093: // A相过载（2字节）
		if offset+2 > len(data) {
			return nil, offset, fmt.Errorf("数据长度不足，无法读取A相过载")
		}
		overload := binary.BigEndian.Uint16(data[offset:])
		return overload, offset + 2, nil

	case 0x0096: // 开表盖（2字节）
		if offset+2 > len(data) {
			return nil, offset, fmt.Errorf("数据长度不足，无法读取开表盖")
		}
		coverStatus := binary.BigEndian.Uint16(data[offset:])
		return coverStatus, offset + 2, nil

	case 0xFFFF: // 自定义数据
		if offset+5 > len(data) {
			return nil, offset, fmt.Errorf("数据长度不足，无法读取自定义数据")
		}
		// TAG=FFFFH格式：4字节基表ID + 1字节数据长度 + N字节数据内容
		baseTableID := fmt.Sprintf("%02X%02X%02X%02X",
			data[offset], data[offset+1], data[offset+2], data[offset+3])
		dataLength := int(data[offset+4])
		offset += 5

		if offset+dataLength > len(data) {
			return nil, offset, fmt.Errorf("自定义数据长度不足")
		}

		customData := data[offset : offset+dataLength]
		result := map[string]interface{}{
			"base_table_id": baseTableID,
			"data_length":   dataLength,
			"data_content":  customData,
		}

		return result, offset + dataLength, nil

	default:
		// 对于未知TAG，尝试读取2字节作为默认值
		if offset+2 > len(data) {
			return nil, offset, fmt.Errorf("数据长度不足，无法读取默认TAG数据")
		}
		value := binary.BigEndian.Uint16(data[offset:])
		return value, offset + 2, nil
	}
}

// determineEventType 根据解析的数据判断事件类型
func determineEventType(data map[string]interface{}) (string, error) {
	// 检查状态字
	if statusWord, ok := data["status_word"]; ok {
		if status, ok := statusWord.(uint16); ok {
			// 根据状态字判断事件类型
			if (status & 0x0010) != 0 {
				return "switch_event", nil // 开合闸事件
			}
			if (status & 0x0200) != 0 {
				return "cover_event", nil // 开表盖事件
			}
		}
	}

	// 检查A相过载
	if _, ok := data["phase_a_overload"]; ok {
		return "overload_event", nil // 过载事件
	}

	// 检查开表盖
	if _, ok := data["cover_opened"]; ok {
		return "cover_event", nil // 开表盖事件
	}

	// 检查自定义数据
	if _, ok := data["custom_data"]; ok {
		return "custom_event", nil // 自定义事件
	}

	// 检查是否有时间但没有特定事件标识，可能是定时上报
	if _, ok := data["event_time"]; ok {
		return "periodic_report", nil // 定时上报
	}

	return "unknown_event", nil
}

// 具体事件处理函数

// HandleSwitchEvent 处理开合闸事件
func HandleSwitchEvent(event *Event3761) {
	zap.S().Infof("处理开合闸事件 - 设备: %s, 表号: %s, 时间: %s",
		event.DeviceAddr, event.TableNumber, event.EventTime.Format("2006-01-02 15:04:05"))

	// 这里可以添加具体的业务逻辑
	// 例如：记录到数据库、发送告警、更新设备状态等
}

// HandleOverloadEvent 处理过载事件
func HandleOverloadEvent(event *Event3761) {
	zap.S().Infof("处理过载事件 - 设备: %s, 表号: %s, 时间: %s",
		event.DeviceAddr, event.TableNumber, event.EventTime.Format("2006-01-02 15:04:05"))

	// 这里可以添加具体的业务逻辑
	// 例如：记录过载信息、发送告警、记录过载持续时间等
}

// HandleCoverEvent 处理开表盖事件
func HandleCoverEvent(event *Event3761) {
	zap.S().Infof("处理开表盖事件 - 设备: %s, 表号: %s, 时间: %s",
		event.DeviceAddr, event.TableNumber, event.EventTime.Format("2006-01-02 15:04:05"))

	// 这里可以添加具体的业务逻辑
	// 例如：记录开表盖事件、发送安全告警、记录操作人员等
}

// HandlePowerOutageEvent 处理停电事件
func HandlePowerOutageEvent(event *Event3761) {
	zap.S().Infof("处理停电事件 - 设备: %s, 表号: %s, 时间: %s",
		event.DeviceAddr, event.TableNumber, event.EventTime.Format("2006-01-02 15:04:05"))

	// 这里可以添加具体的业务逻辑
	// 例如：记录停电时间、计算停电时长、发送停电通知等
}

// HandleCustomEvent 处理自定义事件（TAG=FFFFH）
func HandleCustomEvent(event *Event3761) {
	zap.S().Infof("处理自定义事件 - 设备: %s, 表号: %s, 时间: %s",
		event.DeviceAddr, event.TableNumber, event.EventTime.Format("2006-01-02 15:04:05"))

	if customData, ok := event.EventData["custom_data"]; ok {
		if data, ok := customData.(map[string]interface{}); ok {
			zap.S().Infof("自定义事件数据 - 基表ID: %v, 数据长度: %v",
				data["base_table_id"], data["data_length"])
		}
	}

	// 这里可以添加具体的业务逻辑
	// 例如：解析自定义数据、执行特定操作等
}

// HandlePeriodicReport 处理定时上报
func HandlePeriodicReport(event *Event3761) {
	zap.S().Infof("处理定时上报 - 设备: %s, 表号: %s, 时间: %s",
		event.DeviceAddr, event.TableNumber, event.EventTime.Format("2006-01-02 15:04:05"))

	// 这里可以添加具体的业务逻辑
	// 例如：存储定时数据、更新设备状态、生成报表等
}

// HandleModuleEvent 处理模块事件上报
func HandleModuleEvent(event *Event3761) {
	zap.S().Infof("处理模块事件上报 - 设备: %s, 表号: %s, 时间: %s",
		event.DeviceAddr, event.TableNumber, event.EventTime.Format("2006-01-02 15:04:05"))

	// 这里可以添加具体的业务逻辑
	// 例如：更新模块状态、记录模块事件、发送模块告警等
}

// Init3761EventHandlers 初始化3761-BY事件处理函数
func Init3761EventHandlers() {
	// 注册各种事件处理函数
	RegisterEventHandler3761("switch_event", HandleSwitchEvent)
	RegisterEventHandler3761("overload_event", HandleOverloadEvent)
	RegisterEventHandler3761("cover_event", HandleCoverEvent)
	RegisterEventHandler3761("power_outage_event", HandlePowerOutageEvent)
	RegisterEventHandler3761("custom_event", HandleCustomEvent)
	RegisterEventHandler3761("periodic_report", HandlePeriodicReport)
	RegisterEventHandler3761("module_event", HandleModuleEvent)

	zap.S().Infoln("3761-BY事件处理函数初始化完成")
}
