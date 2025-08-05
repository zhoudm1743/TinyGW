package collector

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"tinyGW/app/models"
	"tinyGW/pkg/service/listener"

	"go.uber.org/zap"
)

// FourGDirectCollector 4G直连采集器，用于4G设备直接连接
type FourGDirectCollector struct {
	models.Collector
	deviceConnections map[string]net.Conn // 设备地址 -> 连接映射
	connectionMutex   sync.RWMutex        // 连接管理锁

	// 命令响应匹配机制
	commandResponses map[string]*CommandResponse // 命令ID -> 响应数据
	responseMutex    sync.RWMutex                // 响应管理锁
	responseCond     *sync.Cond                  // 响应条件变量

	// 命令ID生成
	commandCounter int64      // 命令计数器
	counterMutex   sync.Mutex // 计数器锁
}

// CommandResponse 命令响应结构
type CommandResponse struct {
	CommandID   string    // 命令ID
	Data        []byte    // 响应数据
	ReceiveTime time.Time // 接收时间
	IsComplete  bool      // 是否完整
}

// DeviceConnection 设备连接信息
type DeviceConnection struct {
	DeviceAddr string    // 设备地址
	Conn       net.Conn  // 网络连接
	LastUsed   time.Time // 最后使用时间
}

func (f *FourGDirectCollector) Open(device *models.Device) bool {
	// 初始化连接映射和命令响应映射
	if f.deviceConnections == nil {
		f.deviceConnections = make(map[string]net.Conn)
	}
	if f.commandResponses == nil {
		f.commandResponses = make(map[string]*CommandResponse)
	}
	if f.responseCond == nil {
		f.responseCond = sync.NewCond(&f.responseMutex)
	}

	// 启动定期清理任务
	go f.startPeriodicCleanup()

	// 如果设备参数为空，只进行初始化，不进行具体设备连接
	if device == nil {
		// zap.S().Info("FourGDirectCollector.Open: 设备参数为空，仅进行初始化")
		f.UpdateConnections()
		return true
	}

	// zap.S().Infof("FourGDirectCollector.Open: 打开4G直连采集器，设备: %s", device.Name)

	// 检查设备地址是否为空
	if device.Address == "" {
		zap.S().Error("FourGDirectCollector.Open: 设备地址为空")
		return false
	}

	// 检查设备是否在线
	onlineDevices := listener.GetOnlineDevices()
	deviceAddr := device.Address

	// 处理Q3761-1376协议的地址格式：通讯地址|电表地址
	commAddr := deviceAddr
	if strings.Contains(deviceAddr, "|") {
		parts := strings.Split(deviceAddr, "|")
		if len(parts) >= 2 {
			commAddr = parts[0] // 取通讯地址部分
		}
	}

	if onlineDevice, exists := onlineDevices[commAddr]; exists {
		f.connectionMutex.Lock()
		f.deviceConnections[deviceAddr] = onlineDevice.Conn
		f.connectionMutex.Unlock()

		// 注册数据回调函数（使用通讯地址）
		listener.RegisterDataCallback(commAddr, f.handleDeviceData)

		// zap.S().Infof("FourGDirectCollector.Open: 设备 %s 已在线，连接建立成功", deviceAddr)
		return true
	}
	return false
}

func (f *FourGDirectCollector) Close() bool {
	// zap.S().Infof("FourGDirectCollector.Close: 关闭4G直连采集器")

	f.connectionMutex.Lock()
	defer f.connectionMutex.Unlock()

	// 注销所有设备的数据回调函数
	for deviceAddr := range f.deviceConnections {
		listener.UnregisterDataCallback(deviceAddr)
	}

	// 关闭所有连接
	for deviceAddr, conn := range f.deviceConnections {
		if conn != nil {
			if err := conn.Close(); err != nil {
				zap.S().Errorf("FourGDirectCollector.Close: 关闭设备 %s 连接失败: %v", deviceAddr, err)
			} else {
				// zap.S().Debugf("FourGDirectCollector.Close: 设备 %s 连接关闭成功", deviceAddr)
			}
		}
	}

	// 清空连接映射和命令响应映射
	f.deviceConnections = make(map[string]net.Conn)

	f.responseMutex.Lock()
	f.commandResponses = make(map[string]*CommandResponse)
	f.responseMutex.Unlock()

	return true
}

// generateCommandID 生成唯一的命令ID
func (f *FourGDirectCollector) generateCommandID() string {
	f.counterMutex.Lock()
	defer f.counterMutex.Unlock()

	f.commandCounter++
	return fmt.Sprintf("cmd_%d_%d", time.Now().UnixNano(), f.commandCounter)
}

// handleDeviceData 处理从zdm.go转发过来的设备数据
func (f *FourGDirectCollector) handleDeviceData(commAddr string, data []byte) {
	// 只处理F1透明转发响应，过滤掉其他数据
	if !f.isValidF1Response(data) {
		// zap.S().Debugf("FourGDirectCollector.handleDeviceData: 过滤无效数据，通讯地址: %s, 长度: %d", commAddr, len(data))
		return
	}

	// zap.S().Debugf("FourGDirectCollector.handleDeviceData: 收到有效F1响应，通讯地址: %s, 长度: %d", commAddr, len(data))

	// 创建数据副本
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// 查找对应的完整设备地址
	var fullDeviceAddr string
	for deviceAddr := range f.deviceConnections {
		if strings.Contains(deviceAddr, "|") {
			parts := strings.Split(deviceAddr, "|")
			if len(parts) >= 2 && parts[0] == commAddr {
				fullDeviceAddr = deviceAddr
				break
			}
		} else if deviceAddr == commAddr {
			fullDeviceAddr = deviceAddr
			break
		}
	}

	// 如果没有找到完整地址，使用通讯地址
	if fullDeviceAddr == "" {
		fullDeviceAddr = commAddr
		zap.S().Warnf("FourGDirectCollector.handleDeviceData: 未找到通讯地址 %s 对应的完整设备地址，使用通讯地址", commAddr)
	}

	// 将响应数据存储到命令响应映射中
	f.responseMutex.Lock()

	// 查找最近的命令ID（这里简化处理，使用时间戳作为命令ID）
	commandID := fmt.Sprintf("cmd_%d", time.Now().Unix())

	response := &CommandResponse{
		CommandID:   commandID,
		Data:        dataCopy,
		ReceiveTime: time.Now(),
		IsComplete:  true,
	}

	f.commandResponses[commandID] = response
	f.responseMutex.Unlock()

	// 通知等待的Read方法
	if f.responseCond != nil {
		f.responseCond.Signal()
	}
}

// isValidF1Response 判断是否为有效的F1透明转发响应
func (f *FourGDirectCollector) isValidF1Response(data []byte) bool {
	if len(data) < 18 {
		return false
	}

	// 检查是否是Q3761-1376协议
	if data[0] == 0x68 && len(data) >= 18 {
		// 检查AFN字段
		afn := data[12] // AFN在第13字节（索引12）

		// 只接受AFN=10H的数据转发帧
		if afn == 0x10 && len(data) >= 18 {
			// DT1, DT2 在AFN之后
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

			// 只接受F1透明转发响应
			if fn == 1 {
				// 检查数据完整性
				if len(data) >= 25 && data[len(data)-1] == 0x16 {
					return true
				}
			}
		}
	}

	return false
}

func (f *FourGDirectCollector) Read(data []byte) int {
	timeout := time.Duration(f.GetTimeout()) * time.Second

	// 确保条件变量已初始化
	if f.responseCond == nil {
		f.responseMutex.Lock()
		if f.responseCond == nil {
			f.responseCond = sync.NewCond(&f.responseMutex)
		}
		f.responseMutex.Unlock()
	}

	// 等待响应数据
	f.responseMutex.Lock()
	defer f.responseMutex.Unlock()

	// 查找最新的响应数据
	var latestResponse *CommandResponse
	var latestCommandID string

	for commandID, response := range f.commandResponses {
		if response.IsComplete {
			if latestResponse == nil || response.ReceiveTime.After(latestResponse.ReceiveTime) {
				latestResponse = response
				latestCommandID = commandID
			}
		}
	}

	// 如果找到响应数据，返回并清理
	if latestResponse != nil {
		copyLen := copy(data, latestResponse.Data)
		delete(f.commandResponses, latestCommandID)
		zap.S().Debugf("FourGDirectCollector.Read: 读取到响应数据，命令ID: %s, 长度: %d", latestCommandID, copyLen)
		return copyLen
	}

	// 如果没有找到响应数据，等待
	waitCh := make(chan struct{})
	go func() {
		f.responseCond.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		// 重新查找响应数据
		for commandID, response := range f.commandResponses {
			if response.IsComplete {
				if latestResponse == nil || response.ReceiveTime.After(latestResponse.ReceiveTime) {
					latestResponse = response
					latestCommandID = commandID
				}
			}
		}

		if latestResponse != nil {
			copyLen := copy(data, latestResponse.Data)
			delete(f.commandResponses, latestCommandID)
			zap.S().Debugf("FourGDirectCollector.Read: 等待后读取到响应数据，命令ID: %s, 长度: %d", latestCommandID, copyLen)
			return copyLen
		}

	case <-time.After(timeout):
		// 超时，清空所有响应数据
		cleanedCount := len(f.commandResponses)
		f.commandResponses = make(map[string]*CommandResponse)
		zap.S().Warnf("FourGDirectCollector.Read: 超时，清空了 %d 条响应数据", cleanedCount)
	}

	return 0
}

func (f *FourGDirectCollector) Write(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	// 清空之前的响应数据
	f.responseMutex.Lock()
	f.commandResponses = make(map[string]*CommandResponse)
	f.responseMutex.Unlock()

	zap.S().Debugf("FourGDirectCollector.Write: 清空响应数据，准备发送命令")

	// 获取所有在线设备的连接
	onlineDevices := listener.GetOnlineDevices()
	if len(onlineDevices) == 0 {
		zap.S().Warnf("FourGDirectCollector.Write: 没有在线设备，无法发送数据")
		return 0
	}

	totalWritten := 0

	// 向所有在线设备发送数据
	for deviceAddr, onlineDevice := range onlineDevices {
		if onlineDevice.Conn == nil {
			continue
		}

		// 设置写入超时
		onlineDevice.Conn.SetWriteDeadline(time.Now().Add(time.Duration(f.GetTimeout()) * time.Second))

		cnt, err := onlineDevice.Conn.Write(data)
		if err != nil {
			zap.S().Errorf("FourGDirectCollector.Write: 向设备 %s 写入失败: %v", deviceAddr, err)
			f.reconnectDevice(deviceAddr)
			continue
		}

		// 重置写入超时
		onlineDevice.Conn.SetWriteDeadline(time.Time{})

		if cnt > 0 {
			zap.S().Debugf("FourGDirectCollector.Write: 向设备 %s 写入 %d 字节数据", deviceAddr, cnt)
			totalWritten += cnt

			// 更新连接映射
			f.connectionMutex.Lock()
			f.deviceConnections[deviceAddr] = onlineDevice.Conn
			f.connectionMutex.Unlock()
		}
	}

	return totalWritten
}

func (f *FourGDirectCollector) GetName() string {
	return f.Name
}

func (f *FourGDirectCollector) GetTimeout() int {
	if f.Timeout <= 0 {
		return 30 // 默认30秒超时
	}
	return f.Timeout
}

func (f *FourGDirectCollector) GetInterval() int {
	if f.Interval <= 0 {
		return 1000 // 默认1秒间隔
	}
	return f.Interval
}

// reconnectDevice 重新连接设备
func (f *FourGDirectCollector) reconnectDevice(deviceAddr string) {
	zap.S().Infof("FourGDirectCollector.reconnectDevice: 尝试重新连接设备 %s", deviceAddr)

	// 检查设备是否重新上线
	onlineDevices := listener.GetOnlineDevices()
	if onlineDevice, exists := onlineDevices[deviceAddr]; exists && onlineDevice.Conn != nil {
		f.connectionMutex.Lock()
		f.deviceConnections[deviceAddr] = onlineDevice.Conn
		f.connectionMutex.Unlock()

		zap.S().Infof("FourGDirectCollector.reconnectDevice: 设备 %s 重新连接成功", deviceAddr)
	} else {
		zap.S().Warnf("FourGDirectCollector.reconnectDevice: 设备 %s 不在线，无法重新连接", deviceAddr)
	}
}

// GetDeviceConnection 获取指定设备的连接
func (f *FourGDirectCollector) GetDeviceConnection(deviceAddr string) (net.Conn, bool) {
	f.connectionMutex.RLock()
	defer f.connectionMutex.RUnlock()

	conn, exists := f.deviceConnections[deviceAddr]
	return conn, exists
}

// GetAllDeviceConnections 获取所有设备连接
func (f *FourGDirectCollector) GetAllDeviceConnections() map[string]net.Conn {
	f.connectionMutex.RLock()
	defer f.connectionMutex.RUnlock()

	result := make(map[string]net.Conn)
	for addr, conn := range f.deviceConnections {
		result[addr] = conn
	}
	return result
}

// WriteToDevice 向指定设备写入数据
func (f *FourGDirectCollector) WriteToDevice(deviceAddr string, data []byte) (int, error) {
	conn, exists := f.GetDeviceConnection(deviceAddr)
	if !exists {
		return 0, fmt.Errorf("设备 %s 连接不存在", deviceAddr)
	}

	if conn == nil {
		return 0, fmt.Errorf("设备 %s 连接为空", deviceAddr)
	}

	// 设置写入超时
	conn.SetWriteDeadline(time.Now().Add(time.Duration(f.GetTimeout()) * time.Second))
	defer conn.SetWriteDeadline(time.Time{})

	cnt, err := conn.Write(data)
	if err != nil {
		zap.S().Errorf("FourGDirectCollector.WriteToDevice: 向设备 %s 写入失败: %v", deviceAddr, err)
		return 0, err
	}

	zap.S().Debugf("FourGDirectCollector.WriteToDevice: 向设备 %s 写入 %d 字节数据: [% 2x]", deviceAddr, cnt, data[:cnt])
	return cnt, nil
}

// ReadFromDevice 从指定设备读取数据
func (f *FourGDirectCollector) ReadFromDevice(deviceAddr string, data []byte) (int, error) {
	conn, exists := f.GetDeviceConnection(deviceAddr)
	if !exists {
		return 0, fmt.Errorf("设备 %s 连接不存在", deviceAddr)
	}

	if conn == nil {
		return 0, fmt.Errorf("设备 %s 连接为空", deviceAddr)
	}

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(time.Duration(f.GetTimeout()) * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	cnt, err := conn.Read(data)
	if err != nil {
		zap.S().Debugf("FourGDirectCollector.ReadFromDevice: 从设备 %s 读取失败: %v", deviceAddr, err)
		return 0, err
	}

	if cnt > 0 {
		zap.S().Debugf("FourGDirectCollector.ReadFromDevice: 从设备 %s 读取 %d 字节数据: [% 2x]", deviceAddr, cnt, data[:cnt])
	}

	return cnt, nil
}

// UpdateConnections 更新连接映射（从在线设备列表同步）
func (f *FourGDirectCollector) UpdateConnections() {
	onlineDevices := listener.GetOnlineDevices()

	f.connectionMutex.Lock()
	defer f.connectionMutex.Unlock()

	// 获取当前已注册的设备列表
	currentDevices := make(map[string]bool)
	for deviceAddr := range f.deviceConnections {
		currentDevices[deviceAddr] = true
	}

	// 清空现有连接映射
	f.deviceConnections = make(map[string]net.Conn)

	// 从在线设备列表更新连接
	for deviceAddr, onlineDevice := range onlineDevices {
		if onlineDevice.Conn != nil {
			f.deviceConnections[deviceAddr] = onlineDevice.Conn

			// 如果是新设备，注册数据回调函数
			if !currentDevices[deviceAddr] {
				listener.RegisterDataCallback(deviceAddr, f.handleDeviceData)
				zap.S().Debugf("FourGDirectCollector.UpdateConnections: 新设备 %s 注册数据回调", deviceAddr)
			}

			zap.S().Debugf("FourGDirectCollector.UpdateConnections: 更新设备 %s 连接", deviceAddr)
		}
	}

	zap.S().Infof("FourGDirectCollector.UpdateConnections: 更新了 %d 个设备连接", len(f.deviceConnections))
}

// GetOnlineDeviceCount 获取在线设备数量
func (f *FourGDirectCollector) GetOnlineDeviceCount() int {
	onlineDevices := listener.GetOnlineDevices()
	return len(onlineDevices)
}

// IsDeviceOnline 检查指定设备是否在线
func (f *FourGDirectCollector) IsDeviceOnline(deviceAddr string) bool {
	_, online := listener.GetDeviceStatus(deviceAddr)
	return online
}

// startPeriodicCleanup 启动定期清理任务
func (f *FourGDirectCollector) startPeriodicCleanup() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒清理一次
	defer ticker.Stop()

	for range ticker.C {
		f.cleanupOldResponses()
	}
}

// cleanupOldResponses 清理旧的响应数据
func (f *FourGDirectCollector) cleanupOldResponses() {
	f.responseMutex.Lock()
	defer f.responseMutex.Unlock()

	now := time.Now()
	cleanedCount := 0

	for commandID, response := range f.commandResponses {
		// 清理超过5分钟的响应数据
		if now.Sub(response.ReceiveTime) > 5*time.Minute {
			delete(f.commandResponses, commandID)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		zap.S().Debugf("FourGDirectCollector.cleanupOldResponses: 清理了 %d 条旧响应数据", cleanedCount)
	}
}

var _ Collector = (*FourGDirectCollector)(nil)
