package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	"go.uber.org/fx"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/service"
	"tinyGW/app/api/types"
	"tinyGW/pkg/plugin/response"
	"tinyGW/pkg/service/http/middleware"
	"tinyGW/pkg/util"
)

type collector struct {
	fx.In
	Srv service.CollectorService
}

func collectorRouter(t collector, r *types.ApiRouter) {
	api := r.Group("/api", middleware.JWTAuth())
	api.POST("/collector", t.add)
	api.PUT("/collector", t.update)
	api.DELETE("/collector/:name", t.delete)
	api.GET("/collector/:name", t.find)
	api.GET("/collectors", t.findAll)
	api.GET("/collector/list", t.list)
}

func (t *collector) add(c *gin.Context) {
	var saveReq req.CollectorSaveReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Add(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *collector) update(c *gin.Context) {
	var saveReq req.CollectorSaveReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Update(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *collector) delete(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	err := t.Srv.Delete(name)
	response.CheckAndResp(c, err)
}

func (t *collector) find(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	res, err := t.Srv.Find(name)
	response.CheckAndRespWithData(c, res, err)
}

func (t *collector) findAll(c *gin.Context) {
	color.Infoln("collector findAll")
	res, err := t.Srv.FindAll()
	response.CheckAndRespWithData(c, res, err)
}

func (t *collector) list(c *gin.Context) {
	var pageReq req.PageReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &pageReq)) {
		return
	}
	res, err := t.Srv.List(&pageReq)
	response.CheckAndRespWithData(c, res, err)
}
