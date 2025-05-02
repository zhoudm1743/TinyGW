package collector

import (
	"github.com/tarm/serial"
	"go.uber.org/zap"
	"time"
	"zsxagw/api/domain"
)

// SerialCollector 【串口采集器】
type SerialCollector struct {
	domain.Collector
	*serial.Port
}

var _ Collector = (*SerialCollector)(nil)

func (s *SerialCollector) Open(device *domain.Device) bool {
	// 采集接口通用设置
	serialCopy := &s.Serial

	// 某个设备需要特殊设置
	if device != nil && device.Alone {
		serialCopy = &device.Serial
	}

	var serialParity serial.Parity
	switch serialCopy.Check {
	case "N":
		serialParity = serial.ParityNone
	case "O":
		serialParity = serial.ParityOdd
	case "E":
		serialParity = serial.ParityEven
	}

	var serialStop serial.StopBits
	switch serialCopy.StopBit {
	case "1":
		serialStop = serial.Stop1
	case "1.5":
		serialStop = serial.Stop1Half
	case "2":
		serialStop = serial.Stop2
	}
	serialConfig := &serial.Config{
		Name:        s.Serial.DeviceName,
		Baud:        serialCopy.BaudRate,
		ReadTimeout: time.Millisecond + 1,
		Parity:      serialParity,
		StopBits:    serialStop,
	}

	var err error
	s.Port, err = serial.OpenPort(serialConfig)
	if err != nil {
		zap.S().Errorf("打开串口[%s]失败 %v", s.Name, err)
		return false
	}

	zap.S().Debugf("打开串口[%s]成功！", s.Name)
	return true
}

func (s *SerialCollector) Close() bool {
	if s.Port == nil {
		zap.S().Errorf("串口[%s]文件句柄不存在", s.Name)
		return false
	}

	err := s.Port.Close()
	if err != nil {
		zap.S().Errorf("关闭串口[%s]失败, %v", s.Name, err)
	}

	zap.S().Debugf("关闭串口[%s]成功", s.Name)

	s.Port = nil
	return true
}

func (s *SerialCollector) Read(data []byte) int {
	if s.Port == nil {
		return 0
	}

	cnt, _ := s.Port.Read(data)

	return cnt
}

func (s *SerialCollector) Write(data []byte) int {
	if s.Port == nil {
		return 0
	}

	cnt, _ := s.Port.Write(data)

	return cnt
}

func (s *SerialCollector) GetName() string {
	return s.Name
}

func (s *SerialCollector) GetTimeout() int {
	return s.Timeout
}

func (s *SerialCollector) GetInterval() int {
	return s.Interval
}
