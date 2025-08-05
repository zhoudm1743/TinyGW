package resp

type AccountLoginResp struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}
