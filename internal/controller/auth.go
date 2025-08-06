package controller

import (
	"TinyGW/internal/repo"
	"TinyGW/internal/schemas/req"
	"TinyGW/internal/schemas/resp"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT密钥
const (
	JWTSecret         = "tinyGW_secret_key"
	TokenExpiration   = 24 * time.Hour     // 访问令牌过期时间
	RefreshExpiration = 7 * 24 * time.Hour // 刷新令牌过期时间
)

// CustomClaims 自定义JWT Claims
type CustomClaims struct {
	UserID uint   `json:"userId"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

type AuthController struct {
	userRepo *repo.UserRepo
}

func NewAuthController(userRepo *repo.UserRepo) *AuthController {
	return &AuthController{
		userRepo: userRepo,
	}
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var loginReq req.LoginReq
	if err := ctx.ShouldBindJSON(&loginReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	// 查询用户
	user, err := c.userRepo.Get(loginReq.UserName)
	if err != nil {
		FailWithUnauthorized(ctx)
		return
	}

	// 验证密码
	if user.Password != loginReq.Password {
		FailWithUnauthorized(ctx)
		return
	}

	// 生成令牌
	accessToken, refreshToken, err := c.generateTokens(user.ID, user.Name)
	if err != nil {
		FailWithServerError(ctx, "生成令牌失败")
		return
	}

	// 构建响应
	userInfo := resp.UserInfo{
		ID:        user.ID,
		Name:      user.Name,
		Avatar:    "",                // 可以根据需要设置默认头像
		Role:      []string{"super"}, // 默认为super角色
		Email:     "",
		Phone:     "",
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	loginResp := resp.LoginResp{
		UserInfo:     userInfo,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	Success(ctx, loginResp)
}

// RefreshToken 刷新令牌
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var refreshReq req.RefreshTokenReq
	if err := ctx.ShouldBindJSON(&refreshReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	// 解析刷新令牌
	token, err := jwt.ParseWithClaims(refreshReq.RefreshToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})

	if err != nil {
		FailWithUnauthorized(ctx)
		return
	}

	// 验证令牌
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		FailWithUnauthorized(ctx)
		return
	}

	// 生成新令牌
	accessToken, refreshToken, err := c.generateTokens(claims.UserID, claims.Name)
	if err != nil {
		FailWithServerError(ctx, "生成令牌失败")
		return
	}

	// 构建响应
	tokenResp := resp.TokenResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	Success(ctx, tokenResp)
}

// GetUserRoutes 获取用户路由
func (c *AuthController) GetUserRoutes(ctx *gin.Context) {
	// 这里可以根据用户ID获取对应的路由权限
	// 但在当前实现中，我们直接返回所有路由
	Success(ctx, []gin.H{}) // 返回空数组，前端会使用静态路由
}

// 生成访问令牌和刷新令牌
func (c *AuthController) generateTokens(userID uint, userName string) (string, string, error) {
	// 生成访问令牌
	accessTokenClaims := CustomClaims{
		UserID: userID,
		Name:   userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "TinyGW",
			Subject:   userName,
			ID:        generateTokenID(userID, userName),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", "", err
	}

	// 生成刷新令牌
	refreshTokenClaims := CustomClaims{
		UserID: userID,
		Name:   userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "TinyGW",
			Subject:   userName,
			ID:        generateTokenID(userID, userName),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// 生成令牌ID
func generateTokenID(userID uint, userName string) string {
	data := fmt.Sprintf("%d-%s-%d", userID, userName, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
