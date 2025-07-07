package listener

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"tinyGW/app/models"
	"tinyGW/pkg/service/conf"
	"tinyGW/pkg/service/script"

	"go.uber.org/zap"
)

// OnlineDevice 在线设备信息
type OnlineDevice struct {
	DeviceAddr    string    // 设备地址
	Protocol      string    // 协议类型
	Conn          net.Conn  // 网络连接
	LastHeartbeat time.Time // 最后心跳时间
	LoginTime     time.Time // 登录时间
	RemoteAddr    string    // 远程地址
}

// DataCallback 数据回调函数类型
type DataCallback func(deviceAddr string, data []byte)

var (
	listeners []net.Listener  // 修改为监听器切片，支持多端口
	stopChans []chan struct{} // 每个监听器对应一个停止信号通道
	wg        sync.WaitGroup  // 用于等待所有监听器结束

	// 传统设备管理（保持兼容性）
	onlineDevices = make(map[string]*OnlineDevice) // key: 设备地址
	deviceMutex   sync.RWMutex                     // 设备管理锁

	// 优化版设备管理
	optimizedDevices sync.Map // 使用sync.Map替代mutex+map
	stats            = &PerformanceStats{StartTime: time.Now().Unix()}

	// 是否使用优化版本
	useOptimized = true

	// 连接断开统计
	disconnectionCount int64

	// 数据回调函数管理
	dataCallbacks     = make(map[string]DataCallback) // 设备地址 -> 回调函数
	dataCallbackMutex sync.RWMutex

	// 错误日志限流器
	temporaryErrorMap     = make(map[string]int64) // 地址 -> 上次记录时间
	temporaryErrorMapLock sync.Mutex               // 错误映射锁

	// 连续超时后断开连接的阈值
	maxConsecutiveTimeouts = 100        // 连续超时次数阈值
	temporaryErrorInterval = int64(300) // 错误日志间隔(秒)
)

func Start() {
	defer func() {
		if r := recover(); r != nil {
			zap.S().Infoln("监听器异常重启:", r)
			time.Sleep(time.Second)
			Start()
		}
	}()

	config := conf.NewConfig()
	tcpPorts := config.Server.TcpPorts
	if tcpPorts == "" {
		tcpPorts = "51483" // 默认端口
	}

	// 解析端口列表
	portStrs := strings.Split(tcpPorts, ",")
	ports := make([]int, 0, len(portStrs))

	for _, portStr := range portStrs {
		portStr = strings.TrimSpace(portStr)
		if portStr == "" {
			continue
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			zap.S().Errorln("无效的端口号:", portStr)
			continue
		}
		ports = append(ports, port)
	}

	if len(ports) == 0 {
		zap.S().Errorln("没有有效的端口配置")
		return
	}

	// 清理之前的监听器
	Stop()

	// 启动设备健康检查
	go startDeviceHealthCheck()

	// 启动性能统计
	if useOptimized {
		go startPerformanceStats()
	}

	// 初始化3761-BY事件处理函数
	Init3761EventHandlers()

	// 初始化监听器和停止信号通道
	listeners = make([]net.Listener, 0, len(ports))
	stopChans = make([]chan struct{}, 0, len(ports))

	// 为每个端口创建监听器
	for _, port := range ports {
		listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			zap.S().Errorln("Listen error on port", port, ":", err)
			continue
		}

		listeners = append(listeners, listener)
		stopChan := make(chan struct{})
		stopChans = append(stopChans, stopChan)

		zap.S().Infoln("监听成功: ", listener.Addr().String())

		// 为每个端口启动一个goroutine
		wg.Add(1)
		go func(l net.Listener, stopCh chan struct{}) {
			defer wg.Done()
			startListener(l, stopCh)
		}(listener, stopChan)
	}
}

func startListener(listener net.Listener, stopChan chan struct{}) {
loop:
	for {
		select {
		case <-stopChan:
			break loop
		default:
			conn, err := listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					zap.S().Infoln("Accept temp error:", ne)
					time.Sleep(time.Second)
					continue
				}
				zap.S().Errorln("Accept error:", err)
				break loop
			}
			go handleConn(conn)
		}
	}
}

// Stop 关闭所有监听器
func Stop() {
	// 发送停止信号给所有监听器
	for _, stopChan := range stopChans {
		close(stopChan)
	}

	// 关闭所有监听器
	for _, listener := range listeners {
		if listener != nil {
			listener.Close()
		}
	}

	// 等待所有监听器结束
	wg.Wait()

	// 清空切片
	listeners = nil
	stopChans = nil
}

func handleConn(conn net.Conn) {
	// 从对象池获取缓冲区
	buf := GetDataBuffer()
	defer ReturnDataBuffer(buf)

	// 确保连接关闭时清理相关设备
	defer func() {
		cleanupDeviceByConnection(conn)
		conn.Close()
		// zap.S().Infof("连接已关闭: %s", conn.RemoteAddr().String())
	}()

	// 设置连接超时
	conn.SetReadDeadline(time.Now().Add(10 * time.Minute))

	// 连接信息和统计
	remoteAddr := conn.RemoteAddr().String()
	consecutiveTimeouts := 0

	for {
		n, err := conn.Read(buf)
		if err != nil {
			// 检查是否是连接断开相关的错误
			if isConnectionClosed(err) {
				// 降低日志级别，减少日志量
				zap.S().Debugf("连接断开: %s, 错误: %v", remoteAddr, err)
				return
			}

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				// 增加连续超时计数
				consecutiveTimeouts++

				// 如果连续超时次数超过阈值，断开连接
				if consecutiveTimeouts > maxConsecutiveTimeouts {
					zap.S().Warnf("连接 %s 连续超时 %d 次，断开连接", remoteAddr, consecutiveTimeouts)
					return
				}

				// 使用限流函数决定是否记录日志
				if shouldLogTemporaryError(remoteAddr) {
					zap.S().Debugf("连接 %s 临时读取错误 (已发生 %d 次): %v", remoteAddr, consecutiveTimeouts, ne)
				}

				time.Sleep(time.Second)
				continue
			}

			// 其他读取错误，视为连接断开
			zap.S().Warnf("读取错误，断开连接: %s, 错误: %v", remoteAddr, err)
			return
		}

		// 重置连续超时计数
		if consecutiveTimeouts > 0 {
			consecutiveTimeouts = 0
		}

		// 重置读取超时
		conn.SetReadDeadline(time.Now().Add(10 * time.Minute))

		// 心跳包检查，快速路径，避免记录和处理心跳包数据
		if n == 10 && string(buf[0:n]) == "1234567890" {
			continue
		}

		// 登录包检查，使用BytePrefix比较而不是完整字符串转换
		if n >= 2 && buf[0] == 0x32 && buf[1] == 0x30 {
			register(string(buf[0:n]), conn)
			break
		}

		// 减少Hex日志记录，提高性能
		if zap.S().Level() == zap.DebugLevel {
			zap.S().Debugln("TcpServer接收数据: ", fmt.Sprintf("[% 2x]", buf[:n]))
		} else {
			zap.S().Infof("TcpServer接收数据: %d字节", n)
		}

		// 复制一份数据，避免在异步处理中buf被覆盖
		dataCopy := make([]byte, n)
		copy(dataCopy, buf[:n])
		go handleData(dataCopy, conn)
	}
}

// 是设备主动上报的，需要解析数据，处理登录，心跳，设备事件
func handleData(b []byte, conn net.Conn) {
	if len(b) < 4 {
		zap.S().Warnln("数据长度不足，忽略处理")
		return
	}

	// 根据数据特征判断协议类型
	protocol := detectProtocol(b)
	if protocol == "" {
		zap.S().Warnln("无法识别的协议类型，数据:", fmt.Sprintf("[% 2x]", b))
		return
	}

	// zap.S().Infoln("检测到协议类型:", protocol, "数据长度:", len(b))

	// 针对2025F183-37协议的特殊处理：简单提取地址并加入echo系统
	if protocol == "2025F183-37" {
		// 提取设备地址
		deviceAddr := extractDeviceAddress(b, protocol, conn)
		// zap.S().Debugf("标准提取设备地址: %s", deviceAddr)

		// 添加到在线设备
		addOnlineDevice(deviceAddr, protocol, conn)

		// 同时将设备地址和连接添加到Echo系统
		RegisterEchoClient(deviceAddr, conn)
		// zap.S().Infof("2025F183-37设备已注册到Echo系统, 地址: %s", deviceAddr)

		// 通知数据回调函数
		notifyDataCallbacks(deviceAddr, b)

		return
	}

	// 以下是其他协议的处理流程...

	// 创建 Lua 脚本运行器
	runner := script.NewLuaRunner()
	err := runner.Open(protocol)
	if err != nil {
		zap.S().Errorln("打开Lua脚本失败:", err)
		return
	}
	defer runner.Close()

	// 提取设备地址
	var deviceAddr string
	if useOptimized && protocol == "Q3761-1376" {
		deviceAddr = FastExtractDeviceAddr(b)
		// zap.S().Debugf("快速提取设备地址: %s (原始数据: [% 2x])", deviceAddr, b[:12])
	} else {
		deviceAddr = extractDeviceAddress(b, protocol, conn)
		zap.S().Debugf("标准提取设备地址: %s", deviceAddr)
	}

	// 调用 Lua 脚本解析数据
	tempVariables := make([]models.DeviceProperty, 0)
	success := runner.AnalysisRx(deviceAddr, []models.DeviceProperty{}, b, len(b), &tempVariables)

	if success {
		// zap.S().Infoln("数据解析成功，协议:", protocol, "设备地址:", deviceAddr)

		// 直接从原始数据检测帧类型（更可靠）
		isLogin := false
		isHeartbeat := false

		if protocol == "Q3761-1376" && len(b) >= 18 {
			// 检查AFN和DT来判断帧类型
			afn := b[12] // AFN在第13字节（索引12）
			dt1 := b[16] // DT1在第17字节（索引16）
			dt2 := b[17] // DT2在第18字节（索引17）

			if afn == 0x02 { // 链路接口检测
				// DT按位解析：DT1的第0位代表F1，第2位代表F3
				if (dt1 & 0x01) != 0 { // F1登录
					isLogin = true
					zap.S().Infof("检测到F1登录帧: AFN=%02X, DT1=%02X, DT2=%02X", afn, dt1, dt2)
				} else if (dt1 & 0x04) != 0 { // F3心跳
					isHeartbeat = true
					// zap.S().Infof("检测到F3心跳帧: AFN=%02X, DT1=%02X, DT2=%02X", afn, dt1, dt2)
				}
			} else if afn == 0x10 { // 数据转发
				// AFN=10H的帧也需要确认回复，特别是F254主动上报数据
				// zap.S().Infof("收到数据转发帧: AFN=%02X, DT1=%02X, DT2=%02X", afn, dt1, dt2)
				// 将数据转发帧视为心跳，需要确认回复
				isHeartbeat = true
			}
		}

		// 打印解析结果（用于调试）
		for _, variable := range tempVariables {
			zap.S().Infof("解析数据 - %s(%s): %v", variable.Name, variable.Description, variable.Value)
		}

		// 处理登录请求
		if isLogin {
			// zap.S().Infof("收到设备登录请求: %s", deviceAddr)
			addOnlineDevice(deviceAddr, protocol, conn)

			// 启动连接健康监控
			monitorConnectionHealth(conn, deviceAddr)

			// 发送登录确认响应 (AFN=00H F1 全部确认)
			go func() {
				// 添加延迟，避免与数据响应冲突
				time.Sleep(50 * time.Millisecond)

				var response []byte
				if useOptimized && protocol == "Q3761-1376" {
					response = GenerateFastConfirmResponse(deviceAddr)
					// zap.S().Debugf("生成快速确认响应，设备: %s, 响应长度: %d, 内容: [% 2x]", deviceAddr, len(response), response)
					defer ReturnResponseToPool(response)
				} else {
					response = generateConfirmResponse(deviceAddr, protocol)
					zap.S().Debugf("生成标准确认响应，设备: %s, 响应长度: %d", deviceAddr, len(response))
				}
				if response != nil {
					err := sendResponse(conn, response)
					if err != nil {
						zap.S().Errorf("发送登录确认失败: %v", err)
						// 如果发送失败且是连接断开，立即清理设备
						if isConnectionClosed(err) {
							removeDeviceByAddr(deviceAddr)
						}
					} else {
						// zap.S().Infof("已发送登录确认响应: %s", deviceAddr)
					}
				} else {
					zap.S().Errorf("生成登录确认响应失败: %s", deviceAddr)
				}
			}()
		}

		// 处理心跳请求
		if isHeartbeat {
			// zap.S().Debugf("收到设备心跳: %s", deviceAddr)
			updateHeartbeat(deviceAddr)

			// 发送心跳确认响应 (AFN=00H F1 全部确认)
			go func() {
				// 添加延迟，避免与数据响应冲突
				time.Sleep(50 * time.Millisecond)

				var response []byte
				if useOptimized && protocol == "Q3761-1376" {
					response = GenerateFastConfirmResponse(deviceAddr)
					// zap.S().Debugf("生成快速心跳确认响应，设备: %s, 响应长度: %d, 设备Addr, len(response))
					defer ReturnResponseToPool(response)
				} else {
					response = generateConfirmResponse(deviceAddr, protocol)
					// zap.S().Debugf("生成标准心跳确认响应，设备: %s, 响应长度: %d", deviceAddr, len(response))
				}
				if response != nil {
					err := sendResponse(conn, response)
					if err != nil {
						zap.S().Errorf("发送心跳确认失败: %v", err)
						// 如果发送失败且是连接断开，立即清理设备
						if isConnectionClosed(err) {
							removeDeviceByAddr(deviceAddr)
						}
					}
				} else {
					zap.S().Errorf("生成心跳确认响应失败: %s", deviceAddr)
				}
			}()
		}

		// 存储到数据库或上报到云端等其他处理
		// 目前只是记录日志，后续可以根据需要扩展

		// 处理3761-BY协议的F254主动上报事件
		if protocol == "Q3761-1376" && len(b) >= 18 {
			afn := b[12]
			dt1 := b[16]
			dt2 := b[17]

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

			// 如果是F254主动上报数据，调用事件处理函数
			if afn == 0x10 && fn == 254 {
				zap.S().Infof("处理3761-BY F254主动上报事件: 设备=%s", deviceAddr)
				Handle3761Event(deviceAddr, b)
			}
		}

		// 通知数据回调函数（过滤掉确认响应）
		if !isConfirmResponse(b) {
			notifyDataCallbacks(deviceAddr, b)
		} else {
			zap.S().Debugf("过滤确认响应，不通知数据回调函数: 设备=%s, AFN=%02X", deviceAddr, b[12])
		}
	} else {
		zap.S().Warnln("数据解析失败，协议:", protocol, "设备地址:", deviceAddr)

		// 即使解析失败也通知回调函数，让上层决定如何处理（但也要过滤确认响应）
		if !isConfirmResponse(b) {
			notifyDataCallbacks(deviceAddr, b)
		} else {
			zap.S().Debugf("过滤确认响应，不通知数据回调函数: 设备=%s, AFN=%02X", deviceAddr, b[12])
		}
	}
}

// isConfirmResponse 判断是否为确认响应帧或主动上报数据
func isConfirmResponse(data []byte) bool {
	if len(data) < 13 {
		return false
	}

	// 检查是否是Q3761-1376协议
	if data[0] == 0x68 && len(data) >= 18 {
		// 检查AFN字段
		afn := data[12] // AFN在第13字节（索引12）

		// AFN=00H是确认/否认帧
		if afn == 0x00 {
			return true
		}

		// AFN=10H是数据转发帧，需要检查Fn
		if afn == 0x10 && len(data) >= 18 {
			// DA1, DA2, DT1, DT2 在AFN之后
			da1 := data[14] // DA1
			da2 := data[15] // DA2
			dt1 := data[16] // DT1
			dt2 := data[17] // DT2

			// 根据3761-BY.md文档计算Fn值
			fn := 0
			if dt1 != 0 {
				// 找到DT1中为1的位
				for i := 0; i < 8; i++ {
					if (dt1 & (1 << i)) != 0 {
						fn = int(dt2)*8 + (i + 1)
						break
					}
				}
			}

			// F254 (0xFE) 是主动上报数据，需要特殊处理
			if fn == 254 {
				zap.S().Debugf("isConfirmResponse: 检测到F254主动上报数据，AFN=%02X, DA1=%02X, DA2=%02X, DT1=%02X, DT2=%02X, Fn=%d", afn, da1, da2, dt1, dt2, fn)
				// 不返回true，让数据继续处理，在handleData中会调用事件处理函数
				return false
			}
		}
	}

	return false
}

// detectProtocol 根据数据特征检测协议类型
func detectProtocol(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	// Q3761-1376 协议特征检测：起始帧 0x68
	if data[0] == 0x68 && len(data) >= 16 {
		// 检查是否有第二个 0x68（Q3761-1376 协议特征）
		if len(data) >= 6 && data[5] == 0x68 {
			return "Q3761-1376"
		}

		// 2025F183-37 协议特征检测：起始帧 0x68 + 表类型 0x10
		if len(data) >= 2 && data[1] == 0x10 {
			return "2025F183-37"
		}
	}

	return ""
}

// extractDeviceAddress 从数据中提取设备地址
func extractDeviceAddress(data []byte, protocol string, conn net.Conn) string {
	switch protocol {
	case "Q3761-1376":
		// Q3761-1376 协议地址域在第7-11字节（5字节地址域）
		if len(data) >= 12 {
			// 地址域: [行政区划码2字节BCD] [终端地址2字节二进制] [主站地址1字节]

			// 位置7-8: 行政区划码（BCD码，小端序）
			// data[7]是低字节, data[8]是高字节
			regionLow := data[7]  // 例如：0x30 (BCD: 30)
			regionHigh := data[8] // 例如：0x05 (BCD: 05)

			// BCD转换为十进制：0x30 -> 30, 0x05 -> 05，组合为0530
			region1 := int((regionHigh>>4)*10 + (regionHigh & 0x0F)) // 高字节BCD转换
			region2 := int((regionLow>>4)*10 + (regionLow & 0x0F))   // 低字节BCD转换
			regionCode := region1*100 + region2                      // 0530

			// 位置9-10: 终端地址（二进制，小端序）
			// data[9]是低字节, data[10]是高字节
			terminalAddr := int(data[9]) + int(data[10])*256 // 40961

			// 组合完整地址：行政区划码(4位) + 终端地址(5位，前面补0)
			return fmt.Sprintf("%04d%05d", regionCode, terminalAddr)
		}
	case "2025F183-37":
		// 2025F183-37 协议地址在第3-9字节（BCD码，7字节地址，低字节在前）
		if len(data) >= 9 {
			// 619协议规定：BCD码，低字节在前
			addr := ""
			// 从后往前读取地址（第9字节到第3字节）
			for i := 8; i >= 2; i-- {
				addr += fmt.Sprintf("%02X", data[i])
			}
			zap.S().Debugf("提取2025F183-37协议地址: %s, 原始数据: [% 2X]", addr, data[2:9])
			return addr
		}
	}

	// 默认返回连接地址作为设备地址
	return conn.RemoteAddr().String()
}

// 设备管理函数 -------------------------------------------------------

// addOnlineDevice 添加在线设备
func addOnlineDevice(deviceAddr, protocol string, conn net.Conn) {
	if useOptimized {
		addOnlineDeviceOptimized(deviceAddr, protocol, conn)
	} else {
		deviceMutex.Lock()
		defer deviceMutex.Unlock()

		now := time.Now()
		onlineDevices[deviceAddr] = &OnlineDevice{
			DeviceAddr:    deviceAddr,
			Protocol:      protocol,
			Conn:          conn,
			LastHeartbeat: now,
			LoginTime:     now,
			RemoteAddr:    conn.RemoteAddr().String(),
		}

		zap.S().Infof("设备上线: %s, 协议: %s, 地址: %s", deviceAddr, protocol, conn.RemoteAddr().String())
	}
}

// updateHeartbeat 更新设备心跳时间
func updateHeartbeat(deviceAddr string) {
	if useOptimized {
		updateHeartbeatOptimized(deviceAddr)
	} else {
		deviceMutex.Lock()
		defer deviceMutex.Unlock()

		if device, exists := onlineDevices[deviceAddr]; exists {
			device.LastHeartbeat = time.Now()
			zap.S().Debugf("更新设备心跳: %s", deviceAddr)
		}
	}
}

// removeOfflineDevice 移除离线设备
func removeOfflineDevice(deviceAddr string) {
	deviceMutex.Lock()
	defer deviceMutex.Unlock()

	if device, exists := onlineDevices[deviceAddr]; exists {
		if device.Conn != nil {
			device.Conn.Close()
		}
		delete(onlineDevices, deviceAddr)
		zap.S().Infof("设备离线: %s", deviceAddr)
	}
}

// isDeviceOnline 检查设备是否在线
func isDeviceOnline(deviceAddr string) bool {
	if useOptimized {
		return isDeviceOnlineOptimized(deviceAddr)
	} else {
		deviceMutex.RLock()
		defer deviceMutex.RUnlock()

		device, exists := onlineDevices[deviceAddr]
		if !exists {
			return false
		}

		// 检查是否超过5分钟没有心跳
		return time.Since(device.LastHeartbeat) < 5*time.Minute
	}
}

// GetOnlineDevices 获取所有在线设备列表
func GetOnlineDevices() map[string]*OnlineDevice {
	if useOptimized {
		return getOnlineDevicesOptimized()
	} else {
		return getOnlineDevicesTraditional()
	}
}

// getOnlineDevicesOptimized 获取优化模式下的在线设备列表
func getOnlineDevicesOptimized() map[string]*OnlineDevice {
	result := make(map[string]*OnlineDevice)
	now := time.Now().Unix()

	optimizedDevices.Range(func(key, value interface{}) bool {
		addr := key.(string)
		device := value.(*FastDeviceInfo)
		lastHeartbeat := atomic.LoadInt64(&device.LastHeartbeat)

		// 检查设备是否在线（5分钟内有活动）
		if now-lastHeartbeat < 300 {
			// 转换为传统格式以保持兼容性
			result[addr] = &OnlineDevice{
				DeviceAddr:    device.Addr,
				Protocol:      device.Protocol,
				Conn:          device.Conn,
				LastHeartbeat: time.Unix(lastHeartbeat, 0),
				LoginTime:     time.Unix(device.LoginTime, 0),
				RemoteAddr:    device.RemoteAddr,
			}
		}
		return true
	})

	return result
}

// getOnlineDevicesTraditional 获取传统模式下的在线设备列表
func getOnlineDevicesTraditional() map[string]*OnlineDevice {
	deviceMutex.RLock()
	defer deviceMutex.RUnlock()

	result := make(map[string]*OnlineDevice)
	for addr, device := range onlineDevices {
		if isDeviceOnline(addr) {
			result[addr] = device
		}
	}
	return result
}

// sendResponse 向设备发送响应
func sendResponse(conn net.Conn, data []byte) error {
	if conn == nil {
		atomic.AddInt64(&stats.Errors, 1)
		return fmt.Errorf("连接为空")
	}

	// 设置写入超时
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetWriteDeadline(time.Time{}) // 重置超时

	_, err := conn.Write(data)
	if err != nil {
		atomic.AddInt64(&stats.Errors, 1)
		// 检查是否是连接断开错误
		if isConnectionClosed(err) {
			zap.S().Warnf("连接已断开，发送响应失败: %v", err)
		} else {
			zap.S().Errorf("发送响应失败: %v", err)
		}
		return err
	}

	// 更新响应发送统计
	atomic.AddInt64(&stats.ResponseSent, 1)
	zap.S().Debugf("发送响应成功: 长度=%d, 内容=[% 2x]", len(data), data)
	return nil
}

// generateResponse 生成协议响应
func generateResponse(protocol, deviceAddr, responseType string) ([]byte, error) {
	runner := script.NewLuaRunner()
	err := runner.Open(protocol)
	if err != nil {
		return nil, fmt.Errorf("打开Lua脚本失败: %v", err)
	}
	defer runner.Close()

	// 根据响应类型生成相应的命令
	var cmdName string
	switch responseType {
	case "login":
		cmdName = "login_response" // 登录响应
	case "heartbeat":
		cmdName = "heartbeat_response" // 心跳响应
	default:
		return nil, fmt.Errorf("未知的响应类型: %s", responseType)
	}

	// 调用设备自定义命令生成响应
	cmdBytes, success, _ := runner.DeviceCustomCmd(deviceAddr, cmdName, "", 0)
	if !success {
		return nil, fmt.Errorf("生成响应失败")
	}

	return cmdBytes, nil
}

// generateConfirmResponse 生成确认响应（AFN=00H F1 全部确认）
func generateConfirmResponse(deviceAddr, protocol string) []byte {
	// 对于Q3761-1376协议，手动构建确认响应帧
	if protocol == "Q3761-1376" {
		// 解析设备地址
		if len(deviceAddr) != 9 {
			return nil
		}

		// 构建确认响应帧
		frame := make([]byte, 20) // 基本确认帧长度

		// 帧头
		frame[0] = 0x68

		// 长度域 (用户数据长度*4 + 协议标识2)
		// 用户数据长度=8 (控制域1+地址域5+AFN1+SEQ1) + 4 (DA2+DT2) = 12
		frameLen := (12 * 4) + 2        // 50
		frame[1] = byte(frameLen % 256) // 50 -> 0x32
		frame[2] = byte(frameLen / 256) // 0
		frame[3] = frame[1]             // 重复
		frame[4] = frame[2]             // 重复

		// 帧头
		frame[5] = 0x68

		// 控制域C (下行帧：DIR=0, PRM=0, 功能码=0认可)
		frame[6] = 0x00

		// 地址域A (从deviceAddr解析)
		digits := make([]int, 9)
		for i := 0; i < 9; i++ {
			if i < len(deviceAddr) {
				digits[i] = int(deviceAddr[i] - '0')
			}
		}

		// 行政区划码 (BCD编码): 0530
		region_d1, region_d2 := digits[0], digits[1]  // 0, 5
		region_d3, region_d4 := digits[2], digits[3]  // 3, 0
		frame[7] = byte((region_d3 * 16) + region_d4) // 30H
		frame[8] = byte((region_d1 * 16) + region_d2) // 05H

		// 终端地址 (二进制编码): 40961
		terminal := digits[4]*10000 + digits[5]*1000 + digits[6]*100 + digits[7]*10 + digits[8]
		frame[9] = byte(terminal % 256)  // 低字节
		frame[10] = byte(terminal / 256) // 高字节

		// 主站地址
		frame[11] = 0x01 // 非零主站地址

		// 应用功能码AFN = 00H (确认/否认)
		frame[12] = 0x00

		// 帧序列域SEQ (下行帧：FIR=1, FIN=1, CON=0, PSEQ=0)
		frame[13] = 0xC0 // 11000000B

		// 数据单元标识 DA (p0)
		frame[14] = 0x00 // DA1
		frame[15] = 0x00 // DA2

		// 数据单元标识 DT (F1 = 1)
		frame[16] = 0x01 // DT1
		frame[17] = 0x00 // DT2

		// 校验和 (从控制域到数据区末尾)
		sum := 0
		for i := 6; i < 18; i++ {
			sum += int(frame[i])
		}
		frame[18] = byte(sum % 256)

		// 结束符
		frame[19] = 0x16

		return frame
	}

	return nil
}

func register(code string, c net.Conn) {
	for _, client := range echoClients {
		if client.Code == code {
			zap.S().Infoln("设备重连: ", code, " Addr: ", c.RemoteAddr().String(), " Time: ", time.Now().Format("2006-01-02 15:04:05"))
			delEcho(code)
			echoClients = append(echoClients, EchoClient{
				Code:      code,
				Conn:      c,
				Addr:      c.RemoteAddr().String(),
				CreatedAt: time.Now().Unix(),
			})
			return
		}
	}
	zap.S().Infoln("注册设备: ", code, " Time:", time.Now().Format("2006-01-02 15:04:05"), c.RemoteAddr().String())
	echoClients = append(echoClients, EchoClient{
		Code:      code,
		Conn:      c,
		Addr:      c.RemoteAddr().String(),
		CreatedAt: time.Now().Unix(),
	})
}

// startDeviceHealthCheck 启动设备健康检查
func startDeviceHealthCheck() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkOfflineDevices()
			cleanupTemporaryErrorMap() // 定期清理过期的错误记录
			checkTimeoutConnections()  // 检查长时间超时的连接
		}
	}
}

// checkOfflineDevices 检查并清理离线设备
func checkOfflineDevices() {
	if useOptimized {
		checkOfflineDevicesOptimized()
	} else {
		checkOfflineDevicesTraditional()
	}
}

// checkOfflineDevicesOptimized 优化模式下的设备清理
func checkOfflineDevicesOptimized() {
	now := time.Now().Unix()
	timeout := int64(300) // 5分钟超时
	offlineCount := 0

	optimizedDevices.Range(func(key, value interface{}) bool {
		addr := key.(string)
		device := value.(*FastDeviceInfo)
		lastHeartbeat := atomic.LoadInt64(&device.LastHeartbeat)

		if now-lastHeartbeat > timeout {
			if device.Conn != nil {
				device.Conn.Close()
			}
			optimizedDevices.Delete(addr)
			atomic.AddInt64(&stats.OnlineDevices, -1)
			offlineCount++
			zap.S().Infof("清理离线设备: %s", addr)
		}
		return true
	})

	if offlineCount > 0 {
		zap.S().Infof("共清理 %d 个离线设备", offlineCount)
	}
}

// checkOfflineDevicesTraditional 传统模式下的设备清理
func checkOfflineDevicesTraditional() {
	deviceMutex.Lock()
	defer deviceMutex.Unlock()

	now := time.Now()
	offlineDevices := make([]string, 0)

	for addr, device := range onlineDevices {
		// 如果超过5分钟没有心跳，认为设备离线
		if now.Sub(device.LastHeartbeat) > 5*time.Minute {
			offlineDevices = append(offlineDevices, addr)
		}
	}

	// 清理离线设备
	for _, addr := range offlineDevices {
		if device, exists := onlineDevices[addr]; exists {
			if device.Conn != nil {
				device.Conn.Close()
			}
			delete(onlineDevices, addr)
			zap.S().Infof("清理离线设备: %s", addr)
		}
	}

	if len(offlineDevices) > 0 {
		zap.S().Infof("共清理 %d 个离线设备", len(offlineDevices))
	}
}

// SendToDevice 向指定设备发送数据
func SendToDevice(deviceAddr string, data []byte) error {
	deviceMutex.RLock()
	device, exists := onlineDevices[deviceAddr]
	deviceMutex.RUnlock()

	if !exists {
		return fmt.Errorf("设备 %s 不在线", deviceAddr)
	}

	if !isDeviceOnline(deviceAddr) {
		return fmt.Errorf("设备 %s 已离线", deviceAddr)
	}

	return sendResponse(device.Conn, data)
}

// GetDeviceStatus 获取设备状态信息
func GetDeviceStatus(deviceAddr string) (*OnlineDevice, bool) {
	deviceMutex.RLock()
	defer deviceMutex.RUnlock()

	device, exists := onlineDevices[deviceAddr]
	if !exists {
		return nil, false
	}

	return device, isDeviceOnline(deviceAddr)
}

// 优化版设备管理函数
func addOnlineDeviceOptimized(deviceAddr, protocol string, conn net.Conn) {
	now := time.Now().Unix()
	device := &FastDeviceInfo{
		Addr:          deviceAddr,
		Protocol:      protocol,
		LastHeartbeat: now,
		LoginTime:     now,
		Conn:          conn,
		RemoteAddr:    conn.RemoteAddr().String(),
	}

	optimizedDevices.Store(deviceAddr, device)
	atomic.AddInt64(&stats.OnlineDevices, 1)
	atomic.AddInt64(&stats.LoginRequests, 1)

	zap.S().Infof("设备上线: %s, 协议: %s, 地址: %s", deviceAddr, protocol, conn.RemoteAddr().String())
}

func updateHeartbeatOptimized(deviceAddr string) {
	if value, ok := optimizedDevices.Load(deviceAddr); ok {
		device := value.(*FastDeviceInfo)
		atomic.StoreInt64(&device.LastHeartbeat, time.Now().Unix())
		atomic.AddInt64(&stats.HeartbeatCount, 1)
		// zap.S().Debugf("更新设备心跳: %s", deviceAddr)
	}
}

func isDeviceOnlineOptimized(deviceAddr string) bool {
	if value, ok := optimizedDevices.Load(deviceAddr); ok {
		device := value.(*FastDeviceInfo)
		lastHeartbeat := atomic.LoadInt64(&device.LastHeartbeat)
		return time.Now().Unix()-lastHeartbeat < 300 // 5分钟超时
	}
	return false
}

// 性能统计输出
func startPerformanceStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		uptime := time.Now().Unix() - stats.StartTime
		onlineCount := atomic.LoadInt64(&stats.OnlineDevices)
		disconnects := atomic.LoadInt64(&disconnectionCount)

		zap.S().Infof("设备统计 - 在线:%d, 登录:%d, 心跳:%d, 响应:%d, 错误:%d, 断开:%d, 运行时间:%ds",
			onlineCount,
			atomic.LoadInt64(&stats.LoginRequests),
			atomic.LoadInt64(&stats.HeartbeatCount),
			atomic.LoadInt64(&stats.ResponseSent),
			atomic.LoadInt64(&stats.Errors),
			disconnects,
			uptime)

		// // 详细显示在线设备信息（当设备数量较少时）
		// if onlineCount > 0 && onlineCount <= 10 {
		// 	onlineDevs := GetOnlineDevices()
		// 	zap.S().Infof("在线设备详情：")
		// 	for addr, device := range onlineDevs {
		// 		if useOptimized {
		// 			if value, ok := optimizedDevices.Load(addr); ok {
		// 				fastDev := value.(*FastDeviceInfo)
		// 				lastHeartbeat := atomic.LoadInt64(&fastDev.LastHeartbeat)
		// 				zap.S().Infof("  设备: %s, 协议: %s, 最后心跳: %ds前, 地址: %s",
		// 					addr, fastDev.Protocol, time.Now().Unix()-lastHeartbeat, fastDev.RemoteAddr)
		// 			}
		// 		} else {
		// 			zap.S().Infof("  设备: %s, 协议: %s, 最后心跳: %s, 地址: %s",
		// 				addr, device.Protocol, device.LastHeartbeat.Format("15:04:05"), device.RemoteAddr)
		// 		}
		// 	}
		// }
	}
}

// GetPerformanceStats 获取性能统计信息
func GetPerformanceStats() *PerformanceStats {
	return &PerformanceStats{
		OnlineDevices:  atomic.LoadInt64(&stats.OnlineDevices),
		LoginRequests:  atomic.LoadInt64(&stats.LoginRequests),
		HeartbeatCount: atomic.LoadInt64(&stats.HeartbeatCount),
		ResponseSent:   atomic.LoadInt64(&stats.ResponseSent),
		Errors:         atomic.LoadInt64(&stats.Errors),
		StartTime:      stats.StartTime,
	}
}

// isConnectionClosed 检查错误是否表示连接已关闭
func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是EOF错误
	if err == io.EOF {
		return true
	}

	// 检查是否是连接重置错误
	if opErr, ok := err.(*net.OpError); ok {
		if opErr.Err == syscall.ECONNRESET {
			return true
		}
		if opErr.Err == syscall.EPIPE {
			return true
		}
	}

	// 检查错误消息中是否包含连接关闭的关键字
	errStr := err.Error()
	if strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "use of closed network connection") {
		return true
	}

	return false
}

// cleanupDeviceByConnection 根据连接对象清理相关设备
func cleanupDeviceByConnection(conn net.Conn) {
	if conn == nil {
		return
	}

	// remoteAddr := conn.RemoteAddr().String()
	// zap.S().Infof("开始清理连接相关设备: %s", remoteAddr)

	// 更新断开连接统计
	atomic.AddInt64(&disconnectionCount, 1)

	if useOptimized {
		cleanupDeviceByConnectionOptimized(conn)
	} else {
		cleanupDeviceByConnectionTraditional(conn)
	}
}

// cleanupDeviceByConnectionOptimized 优化模式下根据连接清理设备
func cleanupDeviceByConnectionOptimized(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	cleanupCount := 0

	optimizedDevices.Range(func(key, value interface{}) bool {
		addr := key.(string)
		device := value.(*FastDeviceInfo)

		// 检查是否是相同的连接对象或远程地址
		if device.Conn == conn || device.RemoteAddr == remoteAddr {
			optimizedDevices.Delete(addr)
			atomic.AddInt64(&stats.OnlineDevices, -1)
			cleanupCount++
			// zap.S().Infof("清理断开连接的设备: %s (地址: %s)", addr, device.RemoteAddr)
		}
		return true
	})

	if cleanupCount > 0 {
		// zap.S().Infof("共清理 %d 个断开连接的设备", cleanupCount)
	}
}

// cleanupDeviceByConnectionTraditional 传统模式下根据连接清理设备
func cleanupDeviceByConnectionTraditional(conn net.Conn) {
	deviceMutex.Lock()
	defer deviceMutex.Unlock()

	remoteAddr := conn.RemoteAddr().String()
	disconnectedDevices := make([]string, 0)

	for addr, device := range onlineDevices {
		// 检查是否是相同的连接对象或远程地址
		if device.Conn == conn || device.RemoteAddr == remoteAddr {
			disconnectedDevices = append(disconnectedDevices, addr)
		}
	}

	// 清理断开连接的设备
	for _, addr := range disconnectedDevices {
		if device, exists := onlineDevices[addr]; exists {
			delete(onlineDevices, addr)
			zap.S().Infof("清理断开连接的设备: %s (地址: %s)", addr, device.RemoteAddr)
		}
	}

	if len(disconnectedDevices) > 0 {
		zap.S().Infof("共清理 %d 个断开连接的设备", len(disconnectedDevices))
	}
}

// monitorConnectionHealth 监控连接健康状态
func monitorConnectionHealth(conn net.Conn, deviceAddr string) {
	// 在goroutine中定期检查连接状态
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
		defer ticker.Stop()

		for range ticker.C {
			// 尝试设置读取截止时间来检测连接状态
			if err := conn.SetReadDeadline(time.Now().Add(1 * time.Microsecond)); err != nil {
				zap.S().Debugf("连接 %s 已断开: %v", deviceAddr, err)
				return
			}

			// 立即重置截止时间
			conn.SetReadDeadline(time.Time{})
		}
	}()
}

// removeDeviceByAddr 根据设备地址移除设备
func removeDeviceByAddr(deviceAddr string) {
	if useOptimized {
		removeDeviceByAddrOptimized(deviceAddr)
	} else {
		removeDeviceByAddrTraditional(deviceAddr)
	}
}

// removeDeviceByAddrOptimized 优化模式下根据设备地址移除设备
func removeDeviceByAddrOptimized(deviceAddr string) {
	if value, ok := optimizedDevices.Load(deviceAddr); ok {
		device := value.(*FastDeviceInfo)
		if device.Conn != nil {
			device.Conn.Close()
		}
		optimizedDevices.Delete(deviceAddr)
		atomic.AddInt64(&stats.OnlineDevices, -1)
		zap.S().Infof("移除设备: %s", deviceAddr)
	}
}

// removeDeviceByAddrTraditional 传统模式下根据设备地址移除设备
func removeDeviceByAddrTraditional(deviceAddr string) {
	deviceMutex.Lock()
	defer deviceMutex.Unlock()

	if device, exists := onlineDevices[deviceAddr]; exists {
		if device.Conn != nil {
			device.Conn.Close()
		}
		delete(onlineDevices, deviceAddr)
		zap.S().Infof("移除设备: %s", deviceAddr)
	}
}

// RegisterDataCallback 注册数据回调函数
func RegisterDataCallback(deviceAddr string, callback DataCallback) {
	dataCallbackMutex.Lock()
	defer dataCallbackMutex.Unlock()

	dataCallbacks[deviceAddr] = callback
	// zap.S().Debugf("注册数据回调函数: %s", deviceAddr)
}

// UnregisterDataCallback 注销数据回调函数
func UnregisterDataCallback(deviceAddr string) {
	dataCallbackMutex.Lock()
	defer dataCallbackMutex.Unlock()

	delete(dataCallbacks, deviceAddr)
	// zap.S().Debugf("注销数据回调函数: %s", deviceAddr)
}

// notifyDataCallbacks 通知数据回调函数
func notifyDataCallbacks(deviceAddr string, data []byte) {
	dataCallbackMutex.RLock()
	defer dataCallbackMutex.RUnlock()

	// zap.S().Infof("尝试通知数据回调函数: %s, 数据长度: %d, 已注册回调数量: %d", deviceAddr, len(data), len(dataCallbacks))

	// 打印所有已注册的设备地址
	// for addr := range dataCallbacks {
	// 	zap.S().Debugf("已注册回调的设备: %s", addr)
	// }

	if callback, exists := dataCallbacks[deviceAddr]; exists {
		// 在goroutine中异步调用回调函数，避免阻塞
		go func() {
			defer func() {
				if r := recover(); r != nil {
					zap.S().Errorf("数据回调函数执行异常: %v", r)
				}
			}()
			callback(deviceAddr, data)
		}()
		// zap.S().Infof("成功通知数据回调函数: %s, 数据长度: %d", deviceAddr, len(data))
	} else {
		// zap.S().Warnf("设备 %s 没有注册数据回调函数", deviceAddr)
	}
}

// shouldLogTemporaryError 决定是否应该记录临时错误
// 使用错误抑制算法，避免频繁记录同一地址的相同错误
func shouldLogTemporaryError(addr string) bool {
	temporaryErrorMapLock.Lock()
	defer temporaryErrorMapLock.Unlock()

	now := time.Now().Unix()
	lastTime, exists := temporaryErrorMap[addr]

	// 如果不存在记录或已过足够时间，允许记录
	if !exists || now-lastTime >= temporaryErrorInterval {
		temporaryErrorMap[addr] = now
		return true
	}

	return false
}

// cleanupTemporaryErrorMap 清理过期的错误记录
// 定期调用此函数以防止内存泄漏
func cleanupTemporaryErrorMap() {
	temporaryErrorMapLock.Lock()
	defer temporaryErrorMapLock.Unlock()

	now := time.Now().Unix()
	expirationTime := now - temporaryErrorInterval*2

	for addr, lastTime := range temporaryErrorMap {
		if lastTime < expirationTime {
			delete(temporaryErrorMap, addr)
		}
	}
}

// checkTimeoutConnections 检查并关闭长时间超时的连接
func checkTimeoutConnections() {
	var timeoutAddrs []string

	temporaryErrorMapLock.Lock()
	now := time.Now().Unix()
	for addr, lastTime := range temporaryErrorMap {
		// 如果超时超过30分钟，尝试强制关闭连接
		if now-lastTime > 1800 {
			timeoutAddrs = append(timeoutAddrs, addr)
		}
	}
	temporaryErrorMapLock.Unlock()

	// 在锁外处理连接关闭，避免长时间持有锁
	for _, addr := range timeoutAddrs {
		// 从设备映射中查找连接
		if useOptimized {
			// 查找并关闭优化版设备的连接
			if value, ok := optimizedDevices.Load(addr); ok {
				device, _ := value.(*OnlineDevice)
				if device != nil && device.Conn != nil {
					zap.S().Warnf("关闭长时间超时的连接: %s, 最后活动时间: %v", addr, time.Unix(now-1800, 0))
					device.Conn.Close()
				}
			}
		} else {
			// 查找并关闭传统版设备的连接
			deviceMutex.RLock()
			if device, ok := onlineDevices[addr]; ok && device.Conn != nil {
				zap.S().Warnf("关闭长时间超时的连接: %s, 最后活动时间: %v", addr, time.Unix(now-1800, 0))
				device.Conn.Close()
			}
			deviceMutex.RUnlock()
		}

		// 从错误映射中移除
		temporaryErrorMapLock.Lock()
		delete(temporaryErrorMap, addr)
		temporaryErrorMapLock.Unlock()
	}

	// 记录统计信息
	if len(timeoutAddrs) > 0 {
		zap.S().Infof("已关闭 %d 个长时间超时的连接", len(timeoutAddrs))
	}
}
