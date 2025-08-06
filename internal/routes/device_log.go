package routes

func registerDeviceLogRoutes(r Routes) {
	api := r.Engine.Group("/api")

	// 需要认证的路由
	auth := api.Group("/")
	// TODO: 添加认证中间件
	// auth.Use(middleware.JWTAuth())

	// 设备日志相关路由
	auth.POST("/device-log", r.DeviceLogController.Create)
	auth.DELETE("/device-log/:id", r.DeviceLogController.Delete)
	auth.GET("/device-logs", r.DeviceLogController.List)
	auth.GET("/device-log/count", r.DeviceLogController.Count)

	// 清除历史日志
	auth.POST("/device-logs/clear", r.DeviceLogController.ClearBefore)
}
