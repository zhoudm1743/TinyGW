package req

type ReportTaskReq struct {
	Name       string `json:"name" form:"name" binding:"required"`
	ReportName string `json:"reportName" form:"reportName"`                // 平台名称
	Ip         string `json:"ip" form:"ip" binding:"required"`             // ip地址
	Port       int    `json:"port" form:"port" binding:"required"`         // 端口
	ClientID   string `json:"clientID" form:"clientID" binding:"required"` // 网关编号
	Username   string `json:"username" form:"username" binding:"required"` // 用户名
	Password   string `json:"password" form:"password" binding:"required"` // 密码
	Cron       string `json:"cron" form:"cron" binding:"required"`         // cron表达式
	Status     int8   `json:"status" form:"status" binding:"required"`     // 运行状态 0：停止 1：运行
}
