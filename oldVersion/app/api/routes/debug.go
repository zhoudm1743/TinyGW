package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/service"
	"tinyGW/app/api/types"
	"tinyGW/pkg/plugin/response"
	"tinyGW/pkg/service/http/middleware"
	"tinyGW/pkg/util"
)

type debug struct {
	fx.In
	Srv service.DebugService
}

func debugRouter(d debug, r *types.ApiRouter) {
	api := r.Group("/api", middleware.JWTAuth())
	api.POST("/debug/test", d.test)
	api.POST("/debug/upgrade", d.upgrade)
	api.POST("/debug/system-status", d.systemStatus)
	api.POST("/debug/system-reboot", d.systemReboot)
	api.POST("/debug/system-ntp", d.systemNtp)
}

func (d debug) test(c *gin.Context) {
	var test req.DebugTestReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &test)) {
		return
	}
	res, err := d.Srv.Test(&test)
	response.CheckAndRespWithData(c, res, err)
}

func (d debug) upgrade(c *gin.Context) {
	err := d.Srv.Upgrade(c)
	response.CheckAndResp(c, err)
}

func (d debug) systemStatus(c *gin.Context) {
	res, err := d.Srv.SystemStatus()
	response.CheckAndRespWithData(c, res, err)
}

func (d debug) systemReboot(c *gin.Context) {
	err := d.Srv.SystemReboot()
	response.CheckAndResp(c, err)
}

func (d debug) systemNtp(c *gin.Context) {
	res, err := d.Srv.SystemNtp()
	response.CheckAndRespWithData(c, res, err)
}
