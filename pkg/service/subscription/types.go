package subscription

import "encoding/json"

type registerRequest struct {
	Fver  string `json:"fver"`  // 协议版本
	Iccid string `json:"iccid"` // 终端ICCID号
	Imei  string `json:"imei"`  // 终端IMEI号
	Csq   int    `json:"csq"`   // 信号强度
}

// String
func (r registerRequest) String() string {
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}
