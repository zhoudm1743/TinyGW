package resp

// LoginResp 登录响应
type LoginResp struct {
	UserInfo     UserInfo `json:"userInfo"`     // 用户信息
	AccessToken  string   `json:"accessToken"`  // 访问令牌
	RefreshToken string   `json:"refreshToken"` // 刷新令牌
}

// UserInfo 用户信息
type UserInfo struct {
	ID        uint     `json:"id"`        // 用户ID
	Name      string   `json:"name"`      // 用户名
	Avatar    string   `json:"avatar"`    // 头像
	Role      []string `json:"role"`      // 角色
	Email     string   `json:"email"`     // 邮箱
	Phone     string   `json:"phone"`     // 手机号
	CreatedAt int64    `json:"createdAt"` // 创建时间
	UpdatedAt int64    `json:"updatedAt"` // 更新时间
}

// TokenResp 刷新令牌响应
type TokenResp struct {
	AccessToken  string `json:"accessToken"`  // 访问令牌
	RefreshToken string `json:"refreshToken"` // 刷新令牌
}
