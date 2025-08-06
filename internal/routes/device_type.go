package routes

func registerDeviceTypeRoutes(r Routes) {
	api := r.Engine.Group("/api")

	// 需要认证的路由
	auth := api.Group("/")
	// TODO: 添加认证中间件
	// auth.Use(middleware.JWTAuth())

	// 设备类型相关路由
	auth.POST("/device-type", r.DeviceTypeController.Create)
	auth.PUT("/device-type", r.DeviceTypeController.Update)
	auth.DELETE("/device-type/:name", r.DeviceTypeController.Delete)
	auth.GET("/device-type/:name", r.DeviceTypeController.Get)
	auth.GET("/device-types", r.DeviceTypeController.List)
	auth.GET("/device-types/all", r.DeviceTypeController.GetAll)
	auth.GET("/device-type/count", r.DeviceTypeController.Count)

	// 设备类型属性管理
	auth.GET("/device-type/:name/properties", r.DeviceTypeController.GetProperties)
	auth.PUT("/device-type/:name/properties", r.DeviceTypeController.UpdateProperties)

	// 单个属性管理
	auth.POST("/device-type/:name/property", r.DeviceTypeController.AddProperty)
	auth.PUT("/device-type/:name/property/:propertyName", r.DeviceTypeController.UpdateProperty)
	auth.DELETE("/device-type/:name/property/:propertyName", r.DeviceTypeController.DeleteProperty)
	auth.GET("/device-type/:name/property/:propertyName", r.DeviceTypeController.GetProperty)

	// 文件上传
	auth.POST("/device-type/:name/upload", r.DeviceTypeController.Upload)
}
