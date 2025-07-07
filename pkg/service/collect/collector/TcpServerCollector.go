package collector

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"
	"tinyGW/app/models"
	"tinyGW/pkg/service/listener"

	"go.uber.org/zap"
)

// TcpServerCollector 【网口采集器】，用于网桥设备，支持4G水表等设备
type TcpServerCollector struct {
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

	// 当前使用的设备地址
	currentDeviceAddr string
}

// NewTcpServerCollector 创建新的TCP服务器采集器
func NewTcpServerCollector(collector models.Collector) *TcpServerCollector {
	t := &TcpServerCollector{
		Collector:         collector,
		deviceConnections: make(map[string]net.Conn),
		commandResponses:  make(map[string]*CommandResponse),
	}
	t.responseCond = sync.NewCond(&t.responseMutex)

	// 启动定期清理任务
	go t.startPeriodicCleanup()

	return t
}

func (t *TcpServerCollector) Open(device *models.Device) bool {
	// 初始化连接映射和命令响应映射
	if t.deviceConnections == nil {
		t.deviceConnections = make(map[string]net.Conn)
	}
	if t.commandResponses == nil {
		t.commandResponses = make(map[string]*CommandResponse)
	}
	if t.responseCond == nil {
		t.responseCond = sync.NewCond(&t.responseMutex)
	}

	// 如果设备参数为空，只进行初始化，不进行具体设备连接
	if device == nil {
		t.UpdateConnections()
		return true
	}

	// 检查设备地址是否为空
	if device.Address == "" {
		zap.S().Error("TcpServerCollector.Open: 设备地址为空")
		return false
	}

	// 保存当前设备地址
	t.currentDeviceAddr = device.Address

	// 从Echo连接池获取连接
	conn := t.check()
	if conn != nil {
		t.connectionMutex.Lock()
		t.deviceConnections[device.Address] = conn
		t.connectionMutex.Unlock()

		// 注册数据回调函数
		listener.RegisterDataCallback(device.Address, t.handleDeviceData)

		zap.S().Debugf("TcpServerCollector.Open: 设备 %s 连接成功", device.Address)
		return true
	}

	zap.S().Warnf("TcpServerCollector.Open: 设备 %s 连接失败", device.Address)
	return false
}

func (t *TcpServerCollector) Close() bool {
	t.connectionMutex.Lock()
	defer t.connectionMutex.Unlock()

	// 注销所有设备的数据回调函数
	for deviceAddr := range t.deviceConnections {
		listener.UnregisterDataCallback(deviceAddr)
	}

	// 清空连接映射和命令响应映射
	t.deviceConnections = make(map[string]net.Conn)

	t.responseMutex.Lock()
	t.commandResponses = make(map[string]*CommandResponse)
	t.responseMutex.Unlock()

	// 关闭Echo连接
	conn := t.check()
	if conn != nil {
		if err := conn.Close(); err != nil {
			zap.S().Error("关闭Tcp客户端失败!", t.TcpServer)
			return false
		}
		zap.S().Debug("关闭Tcp客户端成功!", t.TcpServer)
	}

	return true
}

func (t *TcpServerCollector) Read(data []byte) int {
	timeout := time.Duration(t.GetTimeout()) * time.Second

	// 确保条件变量已初始化
	if t.responseCond == nil {
		t.responseMutex.Lock()
		if t.responseCond == nil {
			t.responseCond = sync.NewCond(&t.responseMutex)
		}
		t.responseMutex.Unlock()
	}

	// 首先尝试直接从连接读取数据
	conn := t.check()
	if conn != nil {
		conn.SetReadDeadline(time.Now().Add(timeout))
		cnt, err := conn.Read(data)
		if err == nil && cnt > 0 {
			if string(data[:cnt]) == "1234567890" || bytes.Equal(data[:cnt], []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30}) {
				return 0
			}
			zap.S().Debugf("TcpServerCollector.Read: 直接读取数据成功，长度: %d", cnt)
			return cnt
		}
	}

	// 等待响应数据
	t.responseMutex.Lock()
	defer t.responseMutex.Unlock()

	// 查找最新的响应数据
	var latestResponse *CommandResponse
	var latestCommandID string

	for commandID, response := range t.commandResponses {
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
		delete(t.commandResponses, latestCommandID)
		zap.S().Debugf("TcpServerCollector.Read: 读取到响应数据，命令ID: %s, 长度: %d", latestCommandID, copyLen)
		return copyLen
	}

	// 如果没有找到响应数据，等待
	waitCh := make(chan struct{})
	go func() {
		t.responseCond.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		// 重新查找响应数据
		for commandID, response := range t.commandResponses {
			if response.IsComplete {
				if latestResponse == nil || response.ReceiveTime.After(latestResponse.ReceiveTime) {
					latestResponse = response
					latestCommandID = commandID
				}
			}
		}

		if latestResponse != nil {
			copyLen := copy(data, latestResponse.Data)
			delete(t.commandResponses, latestCommandID)
			zap.S().Debugf("TcpServerCollector.Read: 等待后读取到响应数据，命令ID: %s, 长度: %d", latestCommandID, copyLen)
			return copyLen
		}

	case <-time.After(timeout):
		// 超时，清空所有响应数据
		cleanedCount := len(t.commandResponses)
		t.commandResponses = make(map[string]*CommandResponse)
		zap.S().Warnf("TcpServerCollector.Read: 超时，清空了 %d 条响应数据", cleanedCount)
	}

	return 0
}

func (t *TcpServerCollector) Write(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	// 清空之前的响应数据
	t.responseMutex.Lock()
	t.commandResponses = make(map[string]*CommandResponse)
	t.responseMutex.Unlock()

	conn := t.check()
	if conn != nil {
		// 设置写入超时
		conn.SetWriteDeadline(time.Now().Add(time.Duration(t.GetTimeout()) * time.Second))

		cnt, err := conn.Write(data)
		if err != nil {
			t.Close()
			t.Open(nil)
			zap.S().Errorf("TcpServerCollector.Write: 写入失败: %v", err)
			return 0
		}

		// 重置写入超时
		conn.SetWriteDeadline(time.Time{})

		zap.S().Debugf("TcpServerCollector.Write: 写入 %d 字节数据", cnt)
		return cnt
	}

	return 0
}

func (t *TcpServerCollector) GetName() string {
	return t.Name
}

func (t *TcpServerCollector) GetTimeout() int {
	if t.Timeout <= 0 {
		return 30 // 默认30秒超时
	}
	return t.Timeout
}

func (t *TcpServerCollector) GetInterval() int {
	if t.Interval <= 0 {
		return 1000 // 默认1秒间隔
	}
	return t.Interval
}

// 检查连接
func (t *TcpServerCollector) check() net.Conn {
	echo := listener.GetEcho(t.TcpServer.Name)
	if echo.Conn == nil {
		return nil
	}
	return echo.Conn
}

// generateCommandID 生成唯一的命令ID
func (t *TcpServerCollector) generateCommandID() string {
	t.counterMutex.Lock()
	defer t.counterMutex.Unlock()

	t.commandCounter++
	return fmt.Sprintf("cmd_%d_%d", time.Now().UnixNano(), t.commandCounter)
}

// handleDeviceData 处理设备数据回调
func (t *TcpServerCollector) handleDeviceData(deviceAddr string, data []byte) {
	// 创建数据副本
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// 为619协议数据特殊处理
	is619Protocol := false
	if len(data) > 10 && data[0] == 0x68 && (data[10] == 0x97 || data[10] == 0x81 || data[10] == 0x84) {
		is619Protocol = true
		zap.S().Debugf("TcpServerCollector.handleDeviceData: 检测到619协议数据，地址: %s, 控制码: 0x%02X", deviceAddr, data[10])
	}

	// 将响应数据存储到命令响应映射中
	t.responseMutex.Lock()

	// 生成命令ID
	commandID := t.generateCommandID()

	response := &CommandResponse{
		CommandID:   commandID,
		Data:        dataCopy,
		ReceiveTime: time.Now(),
		IsComplete:  true,
	}

	t.commandResponses[commandID] = response
	t.responseMutex.Unlock()

	// 通知等待的Read方法
	if t.responseCond != nil {
		t.responseCond.Signal()
	}

	// 对于619协议，可以在此进行特殊处理
	if is619Protocol {
		// 这里可以根据需要进行额外处理
		zap.S().Debugf("TcpServerCollector.handleDeviceData: 619协议数据处理完成")
	}
}

// UpdateConnections 更新连接映射
func (t *TcpServerCollector) UpdateConnections() {
	conn := t.check()
	if conn != nil && t.currentDeviceAddr != "" {
		t.connectionMutex.Lock()
		t.deviceConnections[t.currentDeviceAddr] = conn
		t.connectionMutex.Unlock()

		// 确保已注册回调
		listener.RegisterDataCallback(t.currentDeviceAddr, t.handleDeviceData)
		zap.S().Debugf("TcpServerCollector.UpdateConnections: 更新设备 %s 连接", t.currentDeviceAddr)
	}
}

// startPeriodicCleanup 启动定期清理任务
func (t *TcpServerCollector) startPeriodicCleanup() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒清理一次
	defer ticker.Stop()

	for range ticker.C {
		t.cleanupOldResponses()
	}
}

// cleanupOldResponses 清理旧的响应数据
func (t *TcpServerCollector) cleanupOldResponses() {
	t.responseMutex.Lock()
	defer t.responseMutex.Unlock()

	now := time.Now()
	cleanedCount := 0

	for commandID, response := range t.commandResponses {
		// 清理超过5分钟的响应数据
		if now.Sub(response.ReceiveTime) > 5*time.Minute {
			delete(t.commandResponses, commandID)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		zap.S().Debugf("TcpServerCollector.cleanupOldResponses: 清理了 %d 条旧响应数据", cleanedCount)
	}
}

// 确保实现了Collector接口
var _ Collector = (*TcpServerCollector)(nil)
