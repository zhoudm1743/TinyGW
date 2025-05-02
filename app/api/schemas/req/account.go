package req

type AccountLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
