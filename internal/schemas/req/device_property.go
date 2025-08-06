package req

type DevicePropertyReq struct {
	Name        string      `json:"name" binding:"required"`        // 名称，英文标识
	Description string      `json:"description" binding:"required"` // 描述，中文涵义
	Type        string      `json:"type" binding:"required"`        // 类型:int,long,double,string
	Length      int         `json:"length"`                         // 长度
	Decimal     int         `json:"decimal"`                        // 小数位
	Unit        string      `json:"unit"`                           // 计量单位
	Value       interface{} `json:"value"`                          // 数值
	Reported    bool        `json:"reported"`                       // 是否上报
	IsAlarm     bool        `json:"isAlarm"`                        // 是否报警
	Threshold   float64     `json:"threshold"`                      // 预警值 +-区间
	Used        float64     `json:"used"`
	AutoCalc    bool        `json:"autoCalc"`
	Scale       float64     `json:"scale"` // 倍率
}
