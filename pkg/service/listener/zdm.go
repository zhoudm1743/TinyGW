package listener

import (
	"bytes"
	"go.uber.org/zap"
	"net"
	"time"
)

var (
	Listener net.Listener
	stopChan = make(chan struct{}) // 新增停止信号通道
)

func Start() {
	defer func() {
		if r := recover(); r != nil {
			zap.S().Infoln("监听器异常重启:", r)
			time.Sleep(time.Second)
			Start()
		}
	}()

	var err error
	Listener, err = net.Listen("tcp", "0.0.0.0:51483")
	if err != nil {
		zap.S().Infoln("Listen error:", err)
		time.Sleep(time.Second)
		Start()
		return
	}

	zap.S().Infoln("监听成功: ", Listener.Addr().String())

loop: // 添加循环标签
	for {
		select {
		case <-stopChan: // 监听停止信号
			break loop
		default:
			conn, err := Listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					zap.S().Infoln("Accept temp error:", ne)
					time.Sleep(time.Second)
					continue
				}
				zap.S().Errorln("Accept error:", err)
				break loop // 遇到非临时错误时退出循环
			}
			go handleConn(conn)
		}
	}
}

// Stop 关闭监听
func Stop() {
	close(stopChan) // 发送停止信号
	if Listener != nil {
		if err := Listener.Close(); err == nil {
			Listener = nil // 清空监听器实例
		}
	}
}

func handleConn(conn net.Conn) {
	buf := make([]byte, 64)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				zap.S().Infoln("Read temp error:", ne)
				time.Sleep(time.Second)
				continue
			}
			continue
		}
		if string(buf[0:n]) == "1234567890" {
			continue
		}
		if bytes.Equal(buf[:2], []byte{0x32, 0x30}) {
			register(string(buf[0:n]), conn)
			break
		}
	}
	return
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
