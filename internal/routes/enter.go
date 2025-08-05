package routes

import (
	"TinyGW/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterAllRoutes(r *gin.Engine, services struct {
	Collector   service.CollectorService
	Device      service.DeviceService
	User        service.UserService
	CollectTask service.CollectTaskService
	ReportTask  service.ReportTaskService
	DeviceType  service.DeviceTypeService
}) {
	RegisterCollectorRoutes(r, services.Collector)
	RegisterDeviceRoutes(r, services.Device)
	RegisterUserRoutes(r, services.User)
	RegisterCollectTaskRoutes(r, services.CollectTask)
	RegisterReportTaskRoutes(r, services.ReportTask)
	RegisterDeviceTypeRoutes(r, services.DeviceType)
}
