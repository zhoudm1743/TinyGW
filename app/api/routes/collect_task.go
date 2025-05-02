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

type collectTask struct {
	fx.In
	Srv service.CollectTaskService
}

func collectTaskRouter(t collectTask, r *types.ApiRouter) {
	api := r.Group("/api", middleware.JWTAuth())
	api.POST("/collect-task", t.add)
	api.PUT("/collect-task", t.update)
	api.DELETE("/collect-task/:name", t.delete)
	api.GET("/collect-task/:name", t.find)
	api.GET("/collect-tasks", t.findAll)
	api.GET("/collect-task/list", t.list)
}

func (t *collectTask) add(c *gin.Context) {
	var saveReq req.CollectTaskReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Add(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *collectTask) update(c *gin.Context) {
	var saveReq req.CollectTaskReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Update(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *collectTask) delete(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	err := t.Srv.Delete(name)
	response.CheckAndResp(c, err)
}

func (t *collectTask) find(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	res, err := t.Srv.Find(name)
	response.CheckAndRespWithData(c, res, err)
}

func (t *collectTask) findAll(c *gin.Context) {
	res, err := t.Srv.FindAll()
	response.CheckAndRespWithData(c, res, err)
}

func (t *collectTask) list(c *gin.Context) {
	var pageReq req.PageReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &pageReq)) {
		return
	}
	res, err := t.Srv.List(&pageReq)
	response.CheckAndRespWithData(c, res, err)
}
