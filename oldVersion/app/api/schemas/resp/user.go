package resp

import "tinyGW/pkg/plugin"

type UserResp struct {
	Username  string        `json:"username" struct:"username"`
	Password  string        `json:"password" struct:"password"`
	CreatedAt plugin.TsTime `json:"createdAt" structs:"createdAt"`
	UpdatedAt plugin.TsTime `json:"updatedAt" structs:"updatedAt"`
}
