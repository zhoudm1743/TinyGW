package models

// Serial 本地串口，主要指两个485串口
type Serial struct {
	Name       string `json:"name"`       // 串口名称
	DeviceName string `json:"deviceName"` // 设备名称
	BaudRate   int    `json:"baudRate"`   // 波特率
	DataBit    int    `json:"dataBit"`    // 数据位
	StopBit    string `json:"stopBit"`    // 停止位
	Check      string `json:"check"`      // 检验
}

// TcpClient 南向网口，主要指接入网桥的设备，即485转tcp协议
type TcpClient struct {
	Name string `json:"name"` // 名称
	Ip   string `json:"ip"`   // Ip地址
	Port int    `json:"port"` // 端口号
}

// TcpServer 北向网口，主要指北向接入的设备，即北向接入的tcp协议
type TcpServer struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type Mqtt struct {
	Name string `json:"name"` // 名称
}

type Channel struct {
	Name string `json:"name"` // 名称
}

type FourGPRS struct {
	Name string `json:"name"` // 名称
}

// Collector 采集接口，目前包含本地串口、TcpClient，Mqtt，Channel
// 其中Serial、TcpClient、TcpServer、Mqtt、Channel为json格式，需要解析为对应的结构体
type Collector struct {
	Model
	Name      string    `gorm:"type:varchar(255);primary_key;" json:"name"`
	Type      string    `json:"type" gorm:"type:varchar(255);not null;index:idx_collector_type"` // 接口类型Serial或TcpClient或TcpServer或Mqtt或Channel
	Address   string    `json:"address" gorm:"type:varchar(255);index:idx_collector_address"`    // 地址
	Serial    Serial    `json:"serial" gorm:"type:json;serializer:json"`                         // 串口
	TcpClient TcpClient `json:"tcpClient" gorm:"type:json;serializer:json"`                      // Tcp客户端
	TcpServer TcpServer `json:"tcpServer" gorm:"type:json;serializer:json"`                      // Tcp服务端
	Mqtt      Mqtt      `json:"mqtt" gorm:"type:json;serializer:json"`                           // Mqtt
	Channel   Channel   `json:"channel" gorm:"type:json;serializer:json"`                        // 通道
	FourGPRS  FourGPRS  `json:"fouGPRS" gorm:"type:json;serializer:json"`                        // 4G模块
	Timeout   int       `json:"timeout"`                                                         // 超时
	Interval  int       `json:"interval"`                                                        // 间隔
}
