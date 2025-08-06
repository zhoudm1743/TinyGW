package routes

import (
	"TinyGW/internal/middleware"
)

func registerAuthRoutes(r Routes) {
	api := r.Engine.Group("/api")

	// 登录和刷新令牌不需要认证
	api.POST("/login", r.AuthController.Login)
	api.POST("/updateToken", r.AuthController.RefreshToken)

	// 获取用户路由，需要认证
	auth := api.Group("/")
	auth.Use(middleware.JWTAuth())

	auth.GET("/getUserRoutes", r.AuthController.GetUserRoutes)
}
