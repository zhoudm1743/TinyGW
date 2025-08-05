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

type reportTask struct {
	fx.In
	Srv service.ReportTaskService
}

func reportTaskRouter(t reportTask, r *types.ApiRouter) {
	api := r.Group("/api", middleware.JWTAuth())
	api.POST("/report-task", t.add)
	api.PUT("/report-task", t.update)
	api.DELETE("/report-task/:name", t.delete)
	api.GET("/report-task/:name", t.find)
	api.GET("/report-tasks", t.findAll)
	api.GET("/report-task/list", t.list)
	api.GET("/report-task/start/:name", t.start)
	api.GET("/report-task/stop/:name", t.stop)
}

func (t *reportTask) add(c *gin.Context) {
	var saveReq req.ReportTaskReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Add(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *reportTask) update(c *gin.Context) {
	var saveReq req.ReportTaskReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Update(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *reportTask) delete(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	err := t.Srv.Delete(name)
	response.CheckAndResp(c, err)
}

func (t *reportTask) find(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	res, err := t.Srv.Find(name)
	response.CheckAndRespWithData(c, res, err)
}

func (t *reportTask) findAll(c *gin.Context) {
	res, err := t.Srv.FindAll()
	response.CheckAndRespWithData(c, res, err)
}

func (t *reportTask) list(c *gin.Context) {
	var pageReq req.PageReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &pageReq)) {
		return
	}
	res, err := t.Srv.List(&pageReq)
	response.CheckAndRespWithData(c, res, err)
}

func (t *reportTask) start(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	err := t.Srv.Start(name)
	response.CheckAndResp(c, err)
}

func (t *reportTask) stop(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	err := t.Srv.Stop(name)
	response.CheckAndResp(c, err)
}
