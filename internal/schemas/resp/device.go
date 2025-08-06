package resp

type DeviceResp struct {
	Name           string  `json:"name"`
	TypeID         string  `json:"typeID"`
	Address        string  `json:"address"`
	CollectorID    string  `json:"collectorID"`
	Online         bool    `json:"online"`
	CollectTime    int64   `json:"collectTime"`
	CollectTotal   int     `json:"collectTotal"`
	CollectSuccess int     `json:"collectSuccess"`
	ReportTime     int64   `json:"reportTime"`
	ReportTotal    int     `json:"reportTotal"`
	ReportSuccess  int     `json:"reportSuccess"`
	AlarmStatus    bool    `json:"alarmStatus"`
	AlarmReason    string  `json:"alarmReason"`
	AlarmTime      int64   `json:"alarmTime"`
	AlarmTotal     int     `json:"alarmTotal"`
	InitialVal     float64 `json:"initialVal"`
	Scale          float64 `json:"scale"`
	CreatedAt      int64   `json:"createdAt"`
	UpdatedAt      int64   `json:"updatedAt"`
}
