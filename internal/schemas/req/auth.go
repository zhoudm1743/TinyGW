package req

// LoginReq 登录请求
type LoginReq struct {
	UserName string `json:"userName" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// RefreshTokenReq 刷新令牌请求
type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"` // 刷新令牌
}
