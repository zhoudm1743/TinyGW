package listener

import (
	"net"
	"sync"
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

type Message struct {
	code    string
	message []byte
}
