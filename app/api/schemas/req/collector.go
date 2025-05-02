package req

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

type CollectorSaveReq struct {
	Name      string    `json:"name" form:"name" binding:"required"`
	Type      string    `json:"type" form:"type" binding:"required"`
	Serial    Serial    `json:"serial" form:"serial"`
	TcpClient TcpClient `json:"tcpClient" form:"tcpClient"`
	TcpServer TcpServer `json:"tcpServer" form:"tcpServer"`
	Mqtt      Mqtt      `json:"mqtt" form:"mqtt"`
	Channel   Channel   `json:"channel" form:"channel"`
	Timeout   int       `json:"timeout" form:"timeout" default:"3000"`
	Interval  int       `json:"interval" form:"interval" default:"200"`
}
