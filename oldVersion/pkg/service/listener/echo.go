package listener

import (
	"net"
	"sync"
	"time"
)

type EchoClient struct {
	Code      string   `json:"code"`
	Conn      net.Conn `json:"conn"`
	Addr      string   `json:"addr"`
	CreatedAt int64    `json:"created_at"`
}

var (
	echoClients = make([]EchoClient, 0)
	echoLock    sync.Mutex
)

func delEcho(code string) {
	echoLock.Lock()
	defer echoLock.Unlock()
	for i, client := range echoClients {
		if client.Code == code {
			echoClients = append(echoClients[:i], echoClients[i+1:]...)
			return
		}
	}
	return
}

func GetEchoByAddr(addr string) EchoClient {
	echoLock.Lock()
	defer echoLock.Unlock()
	if len(echoClients) == 0 {
		return EchoClient{}
	}
	for _, client := range echoClients {
		if client.Addr == addr {
			return client
		}
	}
	return EchoClient{}
}

// GetEcho 获取客户端
func GetEcho(code string) EchoClient {
	echoLock.Lock()
	defer echoLock.Unlock()
	if len(echoClients) == 0 {
		return EchoClient{}
	}
	for _, client := range echoClients {
		if client.Code == code {
			return client
		}
	}
	return EchoClient{}
}

// GetEchoClients
func GetEchoClients() []EchoClient {
	echoLock.Lock()
	defer echoLock.Unlock()
	return append([]EchoClient(nil), echoClients...)
}

// RegisterEchoClient 注册一个新的Echo客户端连接
func RegisterEchoClient(code string, conn net.Conn) bool {
	echoLock.Lock()
	defer echoLock.Unlock()

	// 检查是否已存在相同代码的客户端
	for i, client := range echoClients {
		if client.Code == code {
			// 如果连接不同，关闭旧连接
			if client.Conn != conn && client.Conn != nil {
				client.Conn.Close()
			}
			// 更新连接
			echoClients[i].Conn = conn
			echoClients[i].Addr = conn.RemoteAddr().String()
			return true
		}
	}

	// 添加新客户端
	newClient := EchoClient{
		Code:      code,
		Conn:      conn,
		Addr:      conn.RemoteAddr().String(),
		CreatedAt: time.Now().Unix(),
	}
	echoClients = append(echoClients, newClient)
	return true
}

type Message struct {
	code    string
	message []byte
}
