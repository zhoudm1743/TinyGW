package collector

import (
	"bytes"
	"net"
	"tinyGW/app/models"
	"tinyGW/pkg/service/listener"

	"go.uber.org/zap"
)

// TcpServerCollector 【网口采集器】，用于网桥设备
type TcpServerCollector struct {
	models.Collector
}

func (t TcpServerCollector) Open(device *models.Device) bool {
	//TODO implement me
	return t.check() != nil
}

func (t TcpServerCollector) Close() bool {
	//TODO implement me
	conn := t.check()
	if conn != nil {
		if err := conn.Close(); err != nil {
			zap.S().Error("关闭Tcp客户端失败!", t.TcpServer)
			return false
		}
		zap.S().Debug("关闭Tcp客户端成功!", t.TcpServer)
		conn = nil
	}
	return true
}

func (t TcpServerCollector) Read(data []byte) int {
	conn := t.check()
	if conn != nil {
		cnt, err := conn.Read(data)
		if err != nil {
			return 0
		}
		if string(data[cnt]) == "1234567890" || bytes.Equal(data[:cnt], []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30}) {
			data = []byte{}
			return 0
		}
		zap.S().Debug("TcpServerCollector读取数据成功!", string(data[:cnt]))
		return cnt
	}
	return 0
}

func (t TcpServerCollector) Write(data []byte) int {
	//TODO implement me
	conn := t.check()
	if conn != nil {
		cnt, err := conn.Write(data)
		if err != nil {
			t.Close()
			t.Open(nil)
			return 0
		}
		return cnt
	}
	return 0
}

func (t TcpServerCollector) GetName() string {
	//TODO implement me
	return t.Name
}

func (t TcpServerCollector) GetTimeout() int {
	//TODO implement me
	return t.Timeout
}

func (t TcpServerCollector) GetInterval() int {
	//TODO implement me
	return t.Interval
}

var _ Collector = (*TcpServerCollector)(nil)

func (t TcpServerCollector) check() net.Conn {
	echo := listener.GetEcho(t.TcpServer.Name)
	if echo.Conn == nil {
		return nil
	}
	return echo.Conn
}
