package worker

import (
	"encoding/json"
)

type (
	// CommandRequest 命令请求：来自物联平台的RPC请求
	CommandRequest struct {
		Method            string             `json:"method"` // 方法：gateway | device
		RequestID         string             `json:"requestID"`
		Params            []RequestParam     `json:"params"`
		ResponseParamChan chan ResponseParam `json:"-"`
	}

	// RequestParam 请求参数
	RequestParam struct {
		ClientID   string                 `json:"clientID"`   // 设备名称
		DeviceName string                 `json:"deviceName"` // 设备名
		CmdName    string                 `json:"cmdName"`    // 命令名称
		CmdParams  map[string]interface{} `json:"cmdParams"`  // 命令参数
	}
)

func (request *CommandRequest) FromJson(payload []byte) error {
	return json.Unmarshal(payload, request)
}
