package resp

import "TinyGW/models"

type CollectorResp struct {
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	Address   string           `json:"address"`
	Serial    models.Serial    `json:"serial"`
	TcpClient models.TcpClient `json:"tcpClient"`
	TcpServer models.TcpServer `json:"tcpServer"`
	Mqtt      models.Mqtt      `json:"mqtt"`
	Channel   models.Channel   `json:"channel"`
	FourGPRS  models.FourGPRS  `json:"fourGPRS"`
	Timeout   int              `json:"timeout"`
	Interval  int              `json:"interval"`
	Enable    bool             `json:"enable"`
	CreatedAt int64            `json:"createdAt"`
	UpdatedAt int64            `json:"updatedAt"`
}
