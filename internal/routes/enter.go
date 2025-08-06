package routes

import (
	"TinyGW/core/web"
	"TinyGW/internal/controller"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Routes struct {
	Engine                *gin.Engine
	UserController        *controller.UserController
	DeviceController      *controller.DeviceController
	CollectorController   *controller.CollectorController
	CollectTaskController *controller.CollectTaskController
	DeviceTypeController  *controller.DeviceTypeController
	DeviceLogController   *controller.DeviceLogController
	AuthController        *controller.AuthController
}

type RouteDeps struct {
	fx.In
	WebService            *web.Service
	UserController        *controller.UserController
	DeviceController      *controller.DeviceController
	CollectorController   *controller.CollectorController
	CollectTaskController *controller.CollectTaskController
	DeviceTypeController  *controller.DeviceTypeController
	DeviceLogController   *controller.DeviceLogController
	AuthController        *controller.AuthController
}

var Module = fx.Options(
	fx.Invoke(RegisterRoutes),
)

func RegisterRoutes(deps RouteDeps) {
	// 使用WebService中的Gin引擎
	engine := deps.WebService.Gin

	// 创建路由参数
	routes := Routes{
		Engine:                engine,
		UserController:        deps.UserController,
		DeviceController:      deps.DeviceController,
		CollectorController:   deps.CollectorController,
		CollectTaskController: deps.CollectTaskController,
		DeviceTypeController:  deps.DeviceTypeController,
		DeviceLogController:   deps.DeviceLogController,
		AuthController:        deps.AuthController,
	}

	// 注册所有路由
	registerUserRoutes(routes)
	registerDeviceRoutes(routes)
	registerCollectorRoutes(routes)
	registerCollectTaskRoutes(routes)
	registerDeviceTypeRoutes(routes)
	registerDeviceLogRoutes(routes)
	registerAuthRoutes(routes)
}
