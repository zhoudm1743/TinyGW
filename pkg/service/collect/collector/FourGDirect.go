package collector

import (
	"fmt"
	"net"
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
}

// DeviceConnection 设备连接信息
type DeviceConnection struct {
	DeviceAddr string    // 设备地址
	Conn       net.Conn  // 网络连接
	LastUsed   time.Time // 最后使用时间
}

func (f *FourGDirectCollector) Open(device *models.Device) bool {
	// 初始化连接映射
	if f.deviceConnections == nil {
		f.deviceConnections = make(map[string]net.Conn)
	}

	// 如果设备参数为空，只进行初始化，不进行具体设备连接
	if device == nil {
		zap.S().Info("FourGDirectCollector.Open: 设备参数为空，仅进行初始化")

		// 同步所有在线设备的连接
		f.UpdateConnections()
		return true
	}

	zap.S().Infof("FourGDirectCollector.Open: 打开4G直连采集器，设备: %s", device.Name)

	// 检查设备地址是否为空
	if device.Address == "" {
		zap.S().Error("FourGDirectCollector.Open: 设备地址为空")
		return false
	}

	// 检查设备是否在线
	onlineDevices := listener.GetOnlineDevices()
	deviceAddr := device.Address

	if onlineDevice, exists := onlineDevices[deviceAddr]; exists {
		f.connectionMutex.Lock()
		f.deviceConnections[deviceAddr] = onlineDevice.Conn
		f.connectionMutex.Unlock()

		zap.S().Infof("FourGDirectCollector.Open: 设备 %s 已在线，连接建立成功", deviceAddr)
		return true
	}

	zap.S().Warnf("FourGDirectCollector.Open: 设备 %s 不在线，无法建立连接", deviceAddr)
	return false
}

func (f *FourGDirectCollector) Close() bool {
	zap.S().Infof("FourGDirectCollector.Close: 关闭4G直连采集器")

	f.connectionMutex.Lock()
	defer f.connectionMutex.Unlock()

	// 关闭所有连接
	for deviceAddr, conn := range f.deviceConnections {
		if conn != nil {
			if err := conn.Close(); err != nil {
				zap.S().Errorf("FourGDirectCollector.Close: 关闭设备 %s 连接失败: %v", deviceAddr, err)
			} else {
				zap.S().Debugf("FourGDirectCollector.Close: 设备 %s 连接关闭成功", deviceAddr)
			}
		}
	}

	// 清空连接映射
	f.deviceConnections = make(map[string]net.Conn)
	return true
}

func (f *FourGDirectCollector) Read(data []byte) int {
	// 获取所有在线设备的连接
	onlineDevices := listener.GetOnlineDevices()
	if len(onlineDevices) == 0 {
		return 0
	}

	// 尝试从任意一个在线设备读取数据
	for deviceAddr, onlineDevice := range onlineDevices {
		if onlineDevice.Conn == nil {
			continue
		}

		// 设置读取超时
		onlineDevice.Conn.SetReadDeadline(time.Now().Add(time.Duration(f.GetTimeout()) * time.Second))

		cnt, err := onlineDevice.Conn.Read(data)
		if err != nil {
			zap.S().Debugf("FourGDirectCollector.Read: 从设备 %s 读取失败: %v", deviceAddr, err)
			continue
		}

		// 重置读取超时
		onlineDevice.Conn.SetReadDeadline(time.Time{})

		if cnt > 0 {
			zap.S().Debugf("FourGDirectCollector.Read: 从设备 %s 读取 %d 字节数据: [% 2x]", deviceAddr, cnt, data[:cnt])

			// 更新连接映射
			f.connectionMutex.Lock()
			f.deviceConnections[deviceAddr] = onlineDevice.Conn
			f.connectionMutex.Unlock()

			return cnt
		}
	}

	return 0
}

func (f *FourGDirectCollector) Write(data []byte) int {
	if len(data) == 0 {
		return 0
	}

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

			// 如果连接断开，尝试重新连接
			f.reconnectDevice(deviceAddr)
			continue
		}

		// 重置写入超时
		onlineDevice.Conn.SetWriteDeadline(time.Time{})

		if cnt > 0 {
			zap.S().Debugf("FourGDirectCollector.Write: 向设备 %s 写入 %d 字节数据: [% 2x]", deviceAddr, cnt, data[:cnt])
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

	// 清空现有连接映射
	f.deviceConnections = make(map[string]net.Conn)

	// 从在线设备列表更新连接
	for deviceAddr, onlineDevice := range onlineDevices {
		if onlineDevice.Conn != nil {
			f.deviceConnections[deviceAddr] = onlineDevice.Conn
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

var _ Collector = (*FourGDirectCollector)(nil)
