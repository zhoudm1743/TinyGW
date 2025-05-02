package worker

import "encoding/json"

type (
	CommandResponse struct {
		Method string          `json:"method"`
		Params []ResponseParam `json:"params"`
	}

	ResponseParam struct {
		ClientID   string      `json:"clientID"`            // 客户端id
		DeviceName string      `json:"deviceName"`          // 设备名
		CmdName    string      `json:"cmdName"`             // 命令名
		CmdStatus  int         `json:"cmdStatus"`           // 命令状态   返回值？
		CmdResult  interface{} `json:"cmdResult,omitempty"` // 命令结果
	}
)

func (response *CommandResponse) ToJson() []byte {
	result, _ := json.Marshal(response)
	return result
}
