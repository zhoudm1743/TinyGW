package routes

func registerCollectTaskRoutes(r Routes) {
	api := r.Engine.Group("/api")

	// 需要认证的路由
	auth := api.Group("/")
	// TODO: 添加认证中间件
	// auth.Use(middleware.JWTAuth())

	// 采集任务相关路由
	auth.POST("/collect-task", r.CollectTaskController.Create)
	auth.PUT("/collect-task", r.CollectTaskController.Update)
	auth.DELETE("/collect-task/:name", r.CollectTaskController.Delete)
	auth.GET("/collect-task/:name", r.CollectTaskController.Get)
	auth.GET("/collect-tasks", r.CollectTaskController.List)
	auth.GET("/collect-task/count", r.CollectTaskController.Count)

	// 任务控制
	auth.POST("/collect-task/:name/start", r.CollectTaskController.Start)
	auth.POST("/collect-task/:name/stop", r.CollectTaskController.Stop)
}
