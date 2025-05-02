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

type user struct {
	fx.In
	Srv service.UserService
}

func userRouter(t user, r *types.ApiRouter) {
	api := r.Group("/api", middleware.JWTAuth())
	api.POST("/user", t.add)
	api.PUT("/user", t.update)
	api.DELETE("/user/:name", t.delete)
	api.GET("/user/:name", t.find)
	api.GET("/users", t.findAll)
	api.GET("/user/list", t.list)
}

func (t *user) add(c *gin.Context) {
	var saveReq req.UserReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Add(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *user) update(c *gin.Context) {
	var saveReq req.UserReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Update(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *user) delete(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	err := t.Srv.Delete(name)
	response.CheckAndResp(c, err)
}

func (t *user) find(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	res, err := t.Srv.Find(name)
	response.CheckAndRespWithData(c, res, err)
}

func (t *user) findAll(c *gin.Context) {
	res, err := t.Srv.FindAll()
	response.CheckAndRespWithData(c, res, err)
}

func (t *user) list(c *gin.Context) {
	var pageReq req.PageReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &pageReq)) {
		return
	}
	res, err := t.Srv.List(&pageReq)
	response.CheckAndRespWithData(c, res, err)
}
