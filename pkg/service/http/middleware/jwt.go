package middleware

import (
	"github.com/gin-gonic/gin"
	"time"
	"tinyGW/pkg/plugin/response"
	"tinyGW/pkg/util"
)

func JWTAuth() gin.HandlerFunc {
	return func(context *gin.Context) {
		token := context.GetHeader("token")

		if token == "" {
			response.Fail(context, response.TokenEmpty)
			context.Abort()
			return
		} else {
			claims, err := util.JwtUtil.ParseToken(token)
			if err != nil {
				response.Fail(context, response.TokenInvalid)
				context.Abort()
				return
			} else if time.Now().Unix() > claims.ExpiresAt.Unix() {
				response.Fail(context, response.TokenExpired)
				context.Abort()
				return
			}
		}
	}
}
