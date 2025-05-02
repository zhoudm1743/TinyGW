package listener

import (
	"bytes"
	"fmt"
	"github.com/gookit/color"
	"net"
	"time"
)

var Listener net.Listener

func Start() {
	var err error
	Listener, err = net.Listen("tcp", "0.0.0.0:51483")
	if err != nil {
		fmt.Println("Listen error:", err)
		time.Sleep(time.Second)
		Start()
	}
	fmt.Println("监听成功: ", Listener.Addr().String())
	for {
		conn, err := Listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				fmt.Println("Accept temp error:", ne)
				time.Sleep(time.Second)
				continue
			}
			fmt.Println("Accept error:", err)
			break
		}
		go handleConn(conn)
	}
}

// Stop 关闭监听
func Stop() {
	err := Listener.Close()
	if err != nil {
		return
	}
}

func handleConn(conn net.Conn) {
	buf := make([]byte, 64)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				fmt.Println("Read temp error:", ne)
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
			color.Println("设备重连: ", code, " Addr: ", c.RemoteAddr().String(), " Time: ", time.Now().Format("2006-01-02 15:04:05"))
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
	color.Greenln("注册设备: ", code, " Time:", time.Now().Format("2006-01-02 15:04:05"), c.RemoteAddr().String())
	echoClients = append(echoClients, EchoClient{
		Code:      code,
		Conn:      c,
		Addr:      c.RemoteAddr().String(),
		CreatedAt: time.Now().Unix(),
	})
}
