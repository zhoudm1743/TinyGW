package resp

import "tinyGW/pkg/plugin"

// Serial 本地串口，主要指两个485串口
type Serial struct {
	Name       string `json:"name" structs:"name"`             // 串口名称
	DeviceName string `json:"deviceName" structs:"deviceName"` // 设备名称
	BaudRate   int    `json:"baudRate" structs:"baudRate"`     // 波特率
	DataBit    int    `json:"dataBit" structs:"dataBit"`       // 数据位
	StopBit    string `json:"stopBit" structs:"stopBit"`       // 停止位
	Check      string `json:"check" structs:"check"`           // 检验
}

// TcpClient 南向网口，主要指接入网桥的设备，即485转tcp协议
type TcpClient struct {
	Name string `json:"name" structs:"name"` // 名称
	Ip   string `json:"ip" structs:"ip"`     // Ip地址
	Port int    `json:"port" structs:"port"` // 端口号
}

// TcpServer 北向网口，主要指北向接入的设备，即北向接入的tcp协议
type TcpServer struct {
	Name string `json:"name" structs:"name"`
	Port int    `json:"port" structs:"port"`
}

type Mqtt struct {
	Name string `json:"name" structs:"name"` // 名称
}

type Channel struct {
	Name string `json:"name" structs:"name"` // 名称
}

type CollectorResp struct {
	Name      string        `json:"name" structs:"name"`
	Type      string        `json:"type" structs:"type"`
	Serial    Serial        `json:"serial" structs:"serial"`
	TcpClient TcpClient     `json:"tcpClient" structs:"tcpClient"`
	TcpServer TcpServer     `json:"tcpServer" structs:"tcpServer"`
	Mqtt      Mqtt          `json:"mqtt" structs:"mqtt"`
	Channel   Channel       `json:"channel" structs:"channel"`
	Timeout   int           `json:"timeout" structs:"timeout"`
	Interval  int           `json:"interval" structs:"interval"`
	CreatedAt plugin.TsTime `json:"createdAt" structs:"createdAt"`
	UpdatedAt plugin.TsTime `json:"updatedAt" structs:"updatedAt"`
}
