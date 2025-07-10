package event

import (
	"net"
	"strings"

	"go.uber.org/zap"
)

// Handle2025F183Event 处理2025F183-37水表事件
func (e *Events2025F183) Handle2025F183Event(deviceAddr string, conn net.Conn, data []byte) {
	zap.S().Infof("处理2025F183-37水表事件，设备地址: %s, 数据长度: %d", deviceAddr, len(data))

	// 首先检查是否有待发送的指令
	// if cmdData, exists := command.Get(deviceAddr); exists {
	// 	// 找到待发送的指令，立即发送
	// 	zap.S().Infof("检测到设备 %s 有待发送指令，立即发送: %X", deviceAddr, cmdData)

	// 	// 设置写入超时
	// 	conn.SetWriteDeadline(time.Now().Add(30 * time.Second)) // 30秒超时

	// 	// 发送指令
	// 	cnt, err := conn.Write(cmdData)
	// 	if err != nil {
	// 		zap.S().Errorf("向设备 %s 发送指令失败: %v", deviceAddr, err)
	// 	} else {
	// 		zap.S().Infof("向设备 %s 发送了 %d 字节的指令: %X", deviceAddr, cnt, cmdData)

	// 		// 发送成功后从队列中移除指令
	// 		// command.Remove(deviceAddr)
	// 		zap.S().Infof("成功处理设备 %s 的待发送指令，已从队列移除", deviceAddr)
	// 	}

	// 	// 重置写入超时
	// 	conn.SetWriteDeadline(time.Time{})
	// } else {
	// 	// 检查系列号变体
	// 	// 2025F183-37水表的设备地址有时候可能有多种形式:
	// 	// 1. 原始地址，如82522506190001
	// 	// 2. 特定格式地址，加前缀或后缀等

	// 	// 尝试一些可能的变体
	// 	possibleAddrs := []string{
	// 		deviceAddr,                     // 原始地址
	// 		"82522506190001",               // 固定地址（测试水表）
	// 		"zsxagwA010_" + deviceAddr,     // 前缀变体
	// 		deviceAddr + "_82522506190001", // 后缀变体
	// 	}

	// 	// 遍历所有可能的地址变体
	// 	for _, addr := range possibleAddrs {
	// 		if addr == deviceAddr {
	// 			continue // 已经检查过原始地址了
	// 		}

	// 		if cmdData, exists := command.Get(addr); exists {
	// 			zap.S().Infof("检测到设备 %s 有指令（通过变体地址 %s），立即发送: %X", deviceAddr, addr, cmdData)

	// 			// 设置写入超时
	// 			conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// 			// 发送指令
	// 			cnt, err := conn.Write(cmdData)
	// 			if err != nil {
	// 				zap.S().Errorf("向设备 %s (变体地址 %s) 发送指令失败: %v", deviceAddr, addr, err)
	// 			} else {
	// 				zap.S().Infof("向设备 %s (变体地址 %s) 发送了 %d 字节的指令: %X", deviceAddr, addr, cnt, cmdData)

	// 				// 发送成功后从队列中移除指令
	// 				command.Remove(addr)
	// 				zap.S().Infof("成功处理设备 %s (变体地址 %s) 的待发送指令，已从队列移除", deviceAddr, addr)
	// 			}

	// 			// 重置写入超时
	// 			conn.SetWriteDeadline(time.Time{})
	// 			break // 找到一个匹配的地址就退出循环
	// 		}
	// 	}

	// 	zap.S().Infof("设备 %s 没有待发送的指令", deviceAddr)
	// }

	// 按正常事件处理数据上报
	e.HandleEvent(deviceAddr, conn, data)
}

// HandleEvent 正常事件处理，解析事件数据并进行业务处理
func (e *Events2025F183) HandleEvent(deviceAddr string, conn net.Conn, data []byte) {
	if data == nil || len(data) < 18 {
		zap.S().Warnf("收到的2025F183-37水表数据不完整: %X", data)
		return
	}

	// 解析数据帧
	parsedData, err := e.parser.Parse(deviceAddr, data)
	if err != nil {
		zap.S().Errorf("解析2025F183-37水表数据失败: %v, data: %X", err, data)
		return
	}

	// 处理业务逻辑
	if len(parsedData) > 0 {
		// 设备通讯地址|电表地址
		fullDeviceAddr := deviceAddr
		if !strings.Contains(deviceAddr, "|") && parsedData["meter_addr"] != nil {
			// 如果设备地址不包含"|"，并且解析出了电表地址，则拼接完整地址
			meterAddr, ok := parsedData["meter_addr"].(string)
			if ok && meterAddr != "" {
				fullDeviceAddr = deviceAddr + "|" + meterAddr
			}
		}

		// 更新设备状态（如果deviceRepository可用）
		if e.deviceRepository != nil {
			e.updateDeviceStatus(fullDeviceAddr, parsedData)
		} else {
			zap.S().Warnf("无法更新设备 %s 状态: deviceRepository为空", fullDeviceAddr)
		}
	}
}
