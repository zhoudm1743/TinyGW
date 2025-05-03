package models

// DeviceProperty 设备属性
type DeviceProperty struct {
	Name        string      `json:"name"`        // 名称，英文标识
	Description string      `json:"description"` // 描述，中文涵义
	Type        string      `json:"type"`        // 类型:int,long,double,string
	Length      int         `json:"length"`      // 长度
	Decimal     int         `json:"decimal"`     // 小数位
	Unit        string      `json:"unit"`        // 计量单位
	Value       interface{} `json:"value"`       // 数值
	Reported    bool        `json:"reported"`    // 是否上报
	IsAlarm     bool        `json:"isAlarm"`     // 是否报警
	Threshold   float64     `json:"threshold"`   // 预警值 +-区间
	Used        float64     `json:"used"`
	AutoCalc    bool        `json:"autoCalc"`
	Scale       float64     `json:"scale"` // 倍率
}

// DeviceType 设备类型
type DeviceType struct {
	Model
	Name       string           `json:"name" gorm:"type:varchar(255);unique;index:idx_device_name"` // 名称，英文标识
	Driver     string           `json:"driver"`                                                     // 驱动程序目录、文件名
	Properties []DeviceProperty `json:"properties" gorm:"type:json;serializer:json"`
}
