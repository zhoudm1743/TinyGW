package routes

func registerUserRoutes(r Routes) {
	api := r.Engine.Group("/api")

	// 需要认证的路由
	auth := api.Group("/")
	// TODO: 添加认证中间件
	// auth.Use(middleware.JWTAuth())

	// 用户相关路由
	auth.POST("/user", r.UserController.Create)
	auth.PUT("/user", r.UserController.Update)
	auth.DELETE("/user/:name", r.UserController.Delete)
	auth.GET("/user/:name", r.UserController.Get)
	auth.GET("/users", r.UserController.List)
	auth.GET("/user/count", r.UserController.Count)
}
