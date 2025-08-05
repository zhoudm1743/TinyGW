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

type device struct {
	fx.In
	Srv service.DeviceService
}

func deviceRouter(t device, r *types.ApiRouter) {
	api := r.Group("/api", middleware.JWTAuth())
	api.POST("/device", t.add)
	api.PUT("/device", t.update)
	api.DELETE("/device/:name", t.delete)
	api.GET("/device/:name", t.find)
	api.GET("/devices", t.findAll)
	api.GET("/device/list", t.list)
}

func (t *device) add(c *gin.Context) {
	var saveReq req.DeviceReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Add(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *device) update(c *gin.Context) {
	var saveReq req.DeviceReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Update(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *device) delete(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	err := t.Srv.Delete(name)
	response.CheckAndResp(c, err)
}

func (t *device) find(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	res, err := t.Srv.Find(name)
	response.CheckAndRespWithData(c, res, err)
}

func (t *device) findAll(c *gin.Context) {
	res, err := t.Srv.FindAll()
	response.CheckAndRespWithData(c, res, err)
}

func (t *device) list(c *gin.Context) {
	var pageReq req.PageReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &pageReq)) {
		return
	}
	res, err := t.Srv.List(&pageReq)
	response.CheckAndRespWithData(c, res, err)
}
