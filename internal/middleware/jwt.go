package middleware

import (
	"TinyGW/internal/controller"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			controller.FailWithUnauthorized(c)
			c.Abort()
			return
		}

		// 检查令牌格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controller.FailWithUnauthorized(c)
			c.Abort()
			return
		}

		// 解析令牌
		token, err := jwt.ParseWithClaims(parts[1], &controller.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(controller.JWTSecret), nil
		})

		if err != nil {
			controller.FailWithUnauthorized(c)
			c.Abort()
			return
		}

		// 验证令牌
		claims, ok := token.Claims.(*controller.CustomClaims)
		if !ok || !token.Valid {
			controller.FailWithUnauthorized(c)
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("userId", claims.UserID)
		c.Set("userName", claims.Name)
		c.Next()
	}
}
