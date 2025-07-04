package listener

import (
	"testing"
	"time"
)

func Test3761EventHandling(t *testing.T) {
	// 初始化事件处理函数
	Init3761EventHandlers()

	// 测试开合闸事件数据（基于文档示例）
	// 68 DA 00 DA 00 68 CA 01 89 29 A9 00 10 71 00 00 20 1F 00 00 00 00 00 01 00 00 FF 01 1E 00 FD 00 1A 00 00 00 00 03 00 01 05 67 89 01 23 45 00 02 21 06 09 03 15 17 00 00 14 00 10 FE 01 16
	switchEventData := []byte{
		0x68, 0xDA, 0x00, 0xDA, 0x00, 0x68, 0xCA, 0x01, 0x89, 0x29, 0xA9, 0x00, // 帧头
		0x10,                   // AFN=10H
		0x71,                   // 控制域
		0x00, 0x00, 0x20, 0x1F, // 地址域
		0x00,                                           // 应用层数据
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0xFF, // 模块维测数据
		0x01,       // 数据协议类型：事件上报
		0x1E, 0x00, // 透明转发内容字节数
		0xFD,       // 透明转发标识符
		0x00, 0x1A, // 帧长度
		0x00, 0x00, // 控制字
		0x00, 0x00, // 时间控制字
		0x03,       // TAG个数
		0x00, 0x01, // TAG1：表号
		0x05, 0x67, 0x89, 0x01, 0x23, 0x45, // 表号数据
		0x00, 0x02, // TAG2：时间
		0x21, 0x06, 0x09, 0x03, 0x15, 0x17, 0x00, // 时间数据
		0x00, 0x14, // TAG3：状态字
		0x00, 0x10, // 状态字数据（开闸）
		0xFE,       // 结束符
		0x01, 0x16, // 校验和
	}

	// 测试过载事件数据
	overloadEventData := []byte{
		0x68, 0x6E, 0x01, 0x6E, 0x01, 0x68, 0xCA, 0x01, 0x89, 0x29, 0xA9, 0x00, // 帧头
		0x10,                   // AFN=10H
		0x72,                   // 控制域
		0x00, 0x00, 0x20, 0x1F, // 地址域
		0x00,                                           // 应用层数据
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0xFF, // 模块维测数据
		0x01,       // 数据协议类型：事件上报
		0x43, 0x00, // 透明转发内容字节数
		0xFD,       // 透明转发标识符
		0x00, 0x3F, // 帧长度
		0x00, 0x00, // 控制字
		0x00, 0x00, // 时间控制字
		0x0B,       // TAG个数
		0x00, 0x01, // TAG1：表号
		0x05, 0x67, 0x89, 0x01, 0x23, 0x45, // 表号数据
		0x00, 0x02, // TAG2：时间
		0x21, 0x06, 0x09, 0x03, 0x16, 0x31, 0x37, // 时间数据
		0x00, 0x93, // TAG3：A相过载
		0x00, 0x20, // A相过载数据
		0x00, 0x15, // TAG4：电压
		0x26, 0x00, // 电压数据
		0x00, 0x18, // TAG5：电流
		0x00, 0x00, 0x00, // 电流数据
		0x00, 0x1B, // TAG6：有功功率
		0x29, 0x41, 0x00, // 有功功率数据
		0x00, 0x1C, // TAG7：视在功率
		0x00, 0x00, 0x00, // 视在功率数据
		0x00, 0x1F, // TAG8：功率因数
		0x29, 0x41, 0x00, // 功率因数数据
		0x00, 0x20, // TAG9：正向有功
		0x00, 0x00, 0x00, // 正向有功数据
		0x00, 0x23, // TAG10：反向有功
		0x07, 0x08, // 反向有功数据
		0x00, 0x24, // TAG11：状态字4
		0x10, 0x00, // 状态字4数据
		0x0F,       // 校验和
		0x49, 0x16, // 结束符
	}

	// 测试开表盖事件数据
	coverEventData := []byte{
		0x68, 0xDA, 0x00, 0xDA, 0x00, 0x68, 0xCA, 0x01, 0x89, 0x29, 0xA9, 0x00, // 帧头
		0x10,                   // AFN=10H
		0x72,                   // 控制域
		0x00, 0x00, 0x20, 0x1F, // 地址域
		0x00,                                           // 应用层数据
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0xFF, // 模块维测数据
		0x01,       // 数据协议类型：事件上报
		0x1E, 0x00, // 透明转发内容字节数
		0xFD,       // 透明转发标识符
		0x00, 0x1A, // 帧长度
		0x00, 0x00, // 控制字
		0x00, 0x00, // 时间控制字
		0x03,       // TAG个数
		0x00, 0x01, // TAG1：表号
		0x05, 0x67, 0x89, 0x01, 0x23, 0x45, // 表号数据
		0x00, 0x02, // TAG2：时间
		0x21, 0x06, 0x09, 0x03, 0x17, 0x03, 0x00, // 时间数据
		0x00, 0x96, // TAG3：开表盖
		0x02, 0x00, // 开表盖数据
		0x60, // 校验和
		0x16, // 结束符
	}

	// 测试TAG=FFFFH自定义事件数据
	customEventData := []byte{
		0x68, 0x42, 0x02, 0x42, 0x02, 0x68, 0xCA, 0x12, 0x12, 0x6E, 0xA4, 0xF2, // 帧头
		0x10,                   // AFN=10H
		0x71,                   // 控制域
		0x00, 0x00, 0x20, 0x1F, // 地址域
		0x00,                                           // 应用层数据
		0x12, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0xFF, // 模块维测数据
		0x00,       // 数据协议类型：周期上报
		0x78, 0x00, // 透明转发内容字节数
		0xFD,       // 透明转发标识符
		0x00, 0x74, // 帧长度
		0x00, 0x00, // 控制字
		0x00, 0x00, // 时间控制字
		0x04,       // TAG个数
		0x00, 0x01, // TAG1：表号
		0x20, 0x23, 0x12, 0x12, 0x11, 0x34, // 表号数据
		0x00, 0x02, // TAG2：时间
		0x24, 0x01, 0x08, 0x01, 0x17, 0x44, 0x00, // 时间数据
		0xFF, 0xFF, // TAG3：自定义数据
		0x01, 0x01, 0xFF, 0x00, // 基表ID
		0x28, // 数据长度
		// 自定义数据内容（40字节）
		0x00, 0x00, 0x00, 0x29, 0x17, 0x08, 0x01, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x29, 0x17, 0x08, 0x01, 0x24, 0x00, 0x00, 0x00, 0x09,
		0x00, 0x05, 0x08, 0x03,
		0xFF, 0xFF, // TAG4：另一个自定义数据
		0x01, 0x01, 0xFF, 0x01, // 基表ID
		0x28, // 数据长度
		// 自定义数据内容（40字节）
		0x96, 0x93, 0x01, 0x47, 0x12, 0x19, 0x12, 0x23, 0x96, 0x93, 0x01, 0x47, 0x12, 0x19, 0x12, 0x23,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x97, 0x32, 0x00, 0x10, 0x18, 0x19, 0x12, 0x23,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xD0, 0xDC, // 校验和
		0x16, // 结束符
	}

	// 测试用例
	testCases := []struct {
		name     string
		data     []byte
		expected string
	}{
		{"开合闸事件", switchEventData, "switch_event"},
		{"过载事件", overloadEventData, "overload_event"},
		{"开表盖事件", coverEventData, "cover_event"},
		{"自定义事件", customEventData, "custom_event"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 测试F254检测
			if !isF254ActiveReport(tc.data) {
				t.Errorf("F254检测失败: %s", tc.name)
			}

			// 测试事件解析
			event, err := parse3761Event("test_device", tc.data)
			if err != nil {
				t.Errorf("事件解析失败: %v", err)
				return
			}

			// 验证事件类型
			if event.EventType != tc.expected {
				t.Errorf("事件类型不匹配，期望: %s, 实际: %s", tc.expected, event.EventType)
			}

			// 验证设备地址
			if event.DeviceAddr != "test_device" {
				t.Errorf("设备地址不匹配，期望: test_device, 实际: %s", event.DeviceAddr)
			}

			// 验证表号
			if event.TableNumber == "" {
				t.Errorf("表号为空")
			}

			// 验证事件时间
			if event.EventTime.IsZero() {
				t.Errorf("事件时间为零")
			}

			t.Logf("事件解析成功: 类型=%s, 表号=%s, 时间=%s",
				event.EventType, event.TableNumber,
				event.EventTime.Format("2006-01-02 15:04:05"))
		})
	}
}

func TestTagDataParsing(t *testing.T) {
	// 测试表号解析
	tableNumberData := []byte{0x05, 0x67, 0x89, 0x01, 0x23, 0x45}
	value, offset, err := parseTagData(0x0001, tableNumberData, 0)
	if err != nil {
		t.Errorf("表号解析失败: %v", err)
	}
	if offset != 6 {
		t.Errorf("表号解析偏移量错误，期望: 6, 实际: %d", offset)
	}
	if tableNumber, ok := value.(string); !ok || tableNumber != "056789012345" {
		t.Errorf("表号解析结果错误，期望: 056789012345, 实际: %v", value)
	}

	// 测试时间解析
	timeData := []byte{0x21, 0x06, 0x09, 0x03, 0x15, 0x17, 0x00}
	value, offset, err = parseTagData(0x0002, timeData, 0)
	if err != nil {
		t.Errorf("时间解析失败: %v", err)
	}
	if offset != 7 {
		t.Errorf("时间解析偏移量错误，期望: 7, 实际: %d", offset)
	}
	if eventTime, ok := value.(time.Time); !ok {
		t.Errorf("时间解析结果类型错误")
	} else {
		expectedTime := time.Date(2021, 6, 9, 3, 21, 23, 0, time.Local)
		if !eventTime.Equal(expectedTime) {
			t.Errorf("时间解析结果错误，期望: %v, 实际: %v", expectedTime, eventTime)
		}
	}

	// 测试状态字解析
	statusData := []byte{0x00, 0x10}
	value, offset, err = parseTagData(0x0014, statusData, 0)
	if err != nil {
		t.Errorf("状态字解析失败: %v", err)
	}
	if offset != 2 {
		t.Errorf("状态字解析偏移量错误，期望: 2, 实际: %d", offset)
	}
	if status, ok := value.(uint16); !ok || status != 16 {
		t.Errorf("状态字解析结果错误，期望: 16, 实际: %v", value)
	}

	// 测试自定义数据解析
	customData := []byte{0x01, 0x01, 0xFF, 0x00, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05}
	value, offset, err = parseTagData(0xFFFF, customData, 0)
	if err != nil {
		t.Errorf("自定义数据解析失败: %v", err)
	}
	if offset != 10 {
		t.Errorf("自定义数据解析偏移量错误，期望: 10, 实际: %d", offset)
	}
	if custom, ok := value.(map[string]interface{}); !ok {
		t.Errorf("自定义数据解析结果类型错误")
	} else {
		if baseTableID, ok := custom["base_table_id"].(string); !ok || baseTableID != "0101FF00" {
			t.Errorf("自定义数据基表ID错误，期望: 0101FF00, 实际: %v", baseTableID)
		}
		if dataLength, ok := custom["data_length"].(int); !ok || dataLength != 5 {
			t.Errorf("自定义数据长度错误，期望: 5, 实际: %v", dataLength)
		}
	}
}

func TestEventTypeDetermination(t *testing.T) {
	// 测试开合闸事件判断
	switchData := map[string]interface{}{
		"status_word": uint16(0x0010),
	}
	eventType, err := determineEventType(switchData)
	if err != nil {
		t.Errorf("开合闸事件类型判断失败: %v", err)
	}
	if eventType != "switch_event" {
		t.Errorf("开合闸事件类型错误，期望: switch_event, 实际: %s", eventType)
	}

	// 测试过载事件判断
	overloadData := map[string]interface{}{
		"phase_a_overload": uint16(0x0020),
	}
	eventType, err = determineEventType(overloadData)
	if err != nil {
		t.Errorf("过载事件类型判断失败: %v", err)
	}
	if eventType != "overload_event" {
		t.Errorf("过载事件类型错误，期望: overload_event, 实际: %s", eventType)
	}

	// 测试开表盖事件判断
	coverData := map[string]interface{}{
		"cover_opened": uint16(0x0002),
	}
	eventType, err = determineEventType(coverData)
	if err != nil {
		t.Errorf("开表盖事件类型判断失败: %v", err)
	}
	if eventType != "cover_event" {
		t.Errorf("开表盖事件类型错误，期望: cover_event, 实际: %s", eventType)
	}

	// 测试自定义事件判断
	customData := map[string]interface{}{
		"custom_data": map[string]interface{}{
			"base_table_id": "0101FF00",
			"data_length":   5,
		},
	}
	eventType, err = determineEventType(customData)
	if err != nil {
		t.Errorf("自定义事件类型判断失败: %v", err)
	}
	if eventType != "custom_event" {
		t.Errorf("自定义事件类型错误，期望: custom_event, 实际: %s", eventType)
	}

	// 测试定时上报判断
	periodicData := map[string]interface{}{
		"event_time": time.Now(),
	}
	eventType, err = determineEventType(periodicData)
	if err != nil {
		t.Errorf("定时上报类型判断失败: %v", err)
	}
	if eventType != "periodic_report" {
		t.Errorf("定时上报类型错误，期望: periodic_report, 实际: %s", eventType)
	}
}
