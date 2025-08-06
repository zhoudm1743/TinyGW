package routes

func registerCollectorRoutes(r Routes) {
	api := r.Engine.Group("/api")

	// 需要认证的路由
	auth := api.Group("/")
	// TODO: 添加认证中间件
	// auth.Use(middleware.JWTAuth())

	// 采集器相关路由
	auth.POST("/collector", r.CollectorController.Create)
	auth.PUT("/collector", r.CollectorController.Update)
	auth.DELETE("/collector/:name", r.CollectorController.Delete)
	auth.GET("/collector/:name", r.CollectorController.Get)
	auth.GET("/collectors", r.CollectorController.List)
	auth.GET("/collector/count", r.CollectorController.Count)
}
