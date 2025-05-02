package middleware

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"zsxagw/core/response"
)

var jwtSecret = []byte("inzj.cn")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

func GenerateToken(username string) (string, error) {
	claims := Claims{
		username,
		jwt.RegisteredClaims{
			Issuer:    "zsxa.com.cn",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("解析错误！")
	}
}

func JWTAuth() gin.HandlerFunc {
	return func(context *gin.Context) {
		token := context.GetHeader("token")

		if token == "" {
			response.Fail(context, response.TokenEmpty)
			context.Abort()
			return
		} else {
			claims, err := ParseToken(token)
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
