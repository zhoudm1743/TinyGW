package routes

func registerDeviceRoutes(r Routes) {
	api := r.Engine.Group("/api")

	// 需要认证的路由
	auth := api.Group("/")
	// TODO: 添加认证中间件
	// auth.Use(middleware.JWTAuth())

	// 设备相关路由
	auth.POST("/device", r.DeviceController.Create)
	auth.PUT("/device", r.DeviceController.Update)
	auth.DELETE("/device/:name", r.DeviceController.Delete)
	auth.GET("/device/:name", r.DeviceController.Get)
	auth.GET("/devices", r.DeviceController.List)
	auth.GET("/device/count", r.DeviceController.Count)
}
