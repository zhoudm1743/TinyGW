package req

import "TinyGW/models"

type CollectorReq struct {
	Name      string           `json:"name" binding:"required"`
	Type      string           `json:"type" binding:"required"`
	Address   string           `json:"address" binding:"required"`
	Serial    models.Serial    `json:"serial"`
	TcpClient models.TcpClient `json:"tcpClient"`
	TcpServer models.TcpServer `json:"tcpServer"`
	Mqtt      models.Mqtt      `json:"mqtt"`
	Channel   models.Channel   `json:"channel"`
	FourGPRS  models.FourGPRS  `json:"fourGPRS"`
	Timeout   int              `json:"timeout"`
	Interval  int              `json:"interval"`
	Enable    bool             `json:"enable"`
}
