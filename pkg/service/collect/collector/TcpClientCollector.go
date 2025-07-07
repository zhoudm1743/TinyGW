package collector

import (
	"bytes"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"syscall"
	"time"
	"tinyGW/app/models"

	"go.uber.org/zap"
)

// TcpClientCollector 【网口采集器】，用于网桥设备
type TcpClientCollector struct {
	models.Collector
	net.Conn
}

var _ Collector = (*TcpClientCollector)(nil)

// 常见的无效字节模式
var (
	// 心跳包模式
	heartbeatPattern = []byte("1234567890")
	// DTU常见噪声标志
	dtuNoisePatterns = [][]byte{
		[]byte("AT"),
		[]byte("OK"),
		[]byte("ERROR"),
		[]byte("+CGATT"),
		[]byte("+CSQ"),
	}
	// 纯数字序列正则，用于检测连续多个数字的情况
	numericSequenceRegex = regexp.MustCompile(`^\d{5,}$`)
	// 缓存无效协议请求模式
	invalidProtocolPattern = regexp.MustCompile(`^(GET|POST|PUT|DELETE|HEAD|OPTIONS) /.*HTTP/\d\.\d`)
)

func (t *TcpClientCollector) Open(device *models.Device) bool {
	var err error
	if t.Conn, err = net.DialTimeout("tcp", t.TcpClient.Ip+":"+strconv.Itoa(t.TcpClient.Port), 500*time.Millisecond); err != nil {
		zap.S().Error("打开Tcp客户端失败!", t.TcpClient)
		t.Conn = nil
		return false
	}

	// 设置TCP连接参数
	if tcpConn, ok := t.Conn.(*net.TCPConn); ok {
		// 启用TCP keepalive
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		// 禁用Nagle算法，提高小包传输效率
		tcpConn.SetNoDelay(true)
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
		// 设置读取超时，避免长时间阻塞
		t.Conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		cnt, err := t.Conn.Read(data)

		if err != nil {
			// 超时不是错误，只是暂时没有数据
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return 0
			}

			// 连接已关闭，需要重新连接
			if err == io.EOF || isConnReset(err) {
				t.Close()
				if t.Open(nil) {
					// 重连成功，重试读取
					cnt, err = t.Conn.Read(data)
					if err != nil {
						return 0
					}
				} else {
					return 0
				}
			}
			return 0
		}
		return cnt
	}
	return 0
}

// isConnReset 检测连接重置错误
func isConnReset(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(*os.SyscallError); ok {
			return syscallErr.Err == syscall.ECONNRESET
		}
	}
	return false
}

// isValidCommand 检查命令数据是否有效，过滤掉常见的无效数据
func isValidCommand(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// 检测是否为心跳包
	if bytes.Contains(data, heartbeatPattern) {
		return false
	}

	// 检测是否包含DTU噪声标志
	for _, pattern := range dtuNoisePatterns {
		if bytes.Contains(data, pattern) {
			return false
		}
	}

	// 检测是否为纯数字序列(5位以上)
	if len(data) <= 32 && numericSequenceRegex.Match(data) {
		return false
	}

	// 过滤HTTP请求（可能是误连接）
	if invalidProtocolPattern.Match(data) {
		return false
	}

	// 检查是否全部为非打印字符
	allNonPrintable := true
	for _, b := range data {
		// 可打印字符范围通常是32-126
		if b >= 32 && b <= 126 {
			allNonPrintable = false
			break
		}
	}

	// 如果全是非打印字符，可能是损坏数据
	if allNonPrintable && len(data) > 3 {
		return false
	}

	return true
}

func (t *TcpClientCollector) Write(data []byte) int {
	if t.Conn != nil {
		// 检查命令是否有效，过滤无效命令
		if !isValidCommand(data) {
			// 降级为Debug级别，减少日志量
			zap.S().Debug("检测到无效命令数据，已过滤", data)
			return 0
		}

		// 设置写入超时，避免长时间阻塞
		t.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		cnt, err := t.Conn.Write(data)

		if err != nil {
			// 连接已关闭，尝试重连后重试
			if err == io.EOF || isConnReset(err) {
				t.Close()
				if t.Open(nil) {
					// 重连成功，重试写入
					cnt, err = t.Conn.Write(data)
					if err != nil {
						zap.S().Error("重连后写入数据失败:", err)
						return 0
					}
					return cnt
				}
			}
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
