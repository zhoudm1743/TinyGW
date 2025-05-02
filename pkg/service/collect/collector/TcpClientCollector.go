package collector

import (
	"go.uber.org/zap"
	"net"
	"strconv"
	"time"
	"tinyGW/app/models"
)

// TcpClientCollector 【网口采集器】，用于网桥设备
type TcpClientCollector struct {
	models.Collector
	net.Conn
}

var _ Collector = (*TcpClientCollector)(nil)

func (t *TcpClientCollector) Open(device *models.Device) bool {
	var err error
	if t.Conn, err = net.DialTimeout("tcp", t.TcpClient.Ip+":"+strconv.Itoa(t.TcpClient.Port), 500*time.Millisecond); err != nil {
		zap.S().Error("打开Tcp客户端失败!", t.TcpClient)
		t.Conn = nil
		return false
	}

	zap.S().Debug("打开Tcp客户端成功!", t.TcpClient)
	return true
}

func (t *TcpClientCollector) Close() bool {
	if t.Conn != nil {
		if err := t.Conn.Close(); err != nil {
			zap.S().Error("关闭Tcp客户端失败!", t.TcpClient)
			return false
		}
		zap.S().Debug("关闭Tcp客户端成功!", t.TcpClient)
		t.Conn = nil
	}
	return true
}

func (t *TcpClientCollector) Read(data []byte) int {
	if t.Conn != nil {
		cnt, err := t.Conn.Read(data)
		if err != nil {
			return 0
		}
		return cnt
	}
	return 0
}

func (t *TcpClientCollector) Write(data []byte) int {
	if t.Conn != nil {
		cnt, err := t.Conn.Write(data)
		if err != nil {
			t.Close()
			t.Open(nil)
			return 0
		}
		return cnt
	}
	return 0
}

func (t *TcpClientCollector) GetName() string {
	return t.Name
}

func (t *TcpClientCollector) GetTimeout() int {
	return t.Timeout
}

func (t *TcpClientCollector) GetInterval() int {
	return t.Interval
}
