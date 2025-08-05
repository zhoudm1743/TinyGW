package util

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type jwtUtil struct{}

var JwtUtil = jwtUtil{}

var jwtSecret = []byte("inzj.cn")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (jwtUtil) GenerateToken(username string) (string, error) {
	claims := Claims{
		username,
		jwt.RegisteredClaims{
			Issuer:    "tinyGW",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (jwtUtil) ParseToken(tokenString string) (*Claims, error) {
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("解析错误！")
	}
}
