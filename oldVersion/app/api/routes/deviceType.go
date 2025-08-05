package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"go.uber.org/fx"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/service"
	"tinyGW/app/api/types"
	"tinyGW/pkg/plugin/response"
	"tinyGW/pkg/service/http/middleware"
	"tinyGW/pkg/util"
)

type deviceType struct {
	fx.In
	Srv service.DeviceTypeService
}

func deviceTypeRouter(t deviceType, r *types.ApiRouter) {
	api := r.Group("/api", middleware.JWTAuth())
	api.POST("/device-type", t.add)
	api.PUT("/device-type", t.update)
	api.DELETE("/device-type/:name", t.delete)
	api.GET("/device-type/:name", t.find)
	api.GET("/device-types", t.findAll)
	api.GET("/device-type/list", t.list)
	api.POST("/device-type/upload/:name", t.upload)

	// id为设备类型id，propertyId为属性id
	api.POST("/device-property/:name", t.addProperties)
	api.PUT("/device-property/:name/:propertyid", t.updateProperties)
	api.DELETE("/device-property/:name/:propertyid", t.deleteProperties)
	api.GET("/device-property/:name/:propertyid", t.findProperty)
	api.GET("/device-property/:name", t.findAllProperties)
}

func (t *deviceType) add(c *gin.Context) {
	var saveReq req.DeviceTypeReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Add(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *deviceType) update(c *gin.Context) {
	var saveReq req.DeviceTypeReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &saveReq)) {
		return
	}
	err := t.Srv.Update(&saveReq)
	response.CheckAndResp(c, err)
}

func (t *deviceType) delete(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	err := t.Srv.Delete(name)
	response.CheckAndResp(c, err)
}

func (t *deviceType) find(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	res, err := t.Srv.Find(name)
	response.CheckAndRespWithData(c, res, err)
}

func (t *deviceType) findAll(c *gin.Context) {
	res, err := t.Srv.FindAll()
	response.CheckAndRespWithData(c, res, err)
}

func (t *deviceType) list(c *gin.Context) {
	var pageReq req.PageReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &pageReq)) {
		return
	}
	res, err := t.Srv.List(&pageReq)
	response.CheckAndRespWithData(c, res, err)
}

func (t *deviceType) addProperties(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	var dpReq req.DeviceProperty
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &dpReq)) {
		return
	}
	err := t.Srv.AddProperties(name, &dpReq)
	response.CheckAndResp(c, err)
}

func (t *deviceType) updateProperties(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	propertyId := c.Param("propertyid")
	if propertyId == "" {
		response.FailWithMsg(c, response.ParamsValidError, "propertyid不能为空")
		return
	}
	var dpReq req.DeviceProperty
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, &dpReq)) {
		return
	}
	err := t.Srv.UpdateProperties(name, cast.ToInt(propertyId), &dpReq)
	response.CheckAndResp(c, err)
}

func (t *deviceType) deleteProperties(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	propertyId := c.Param("propertyid")
	if propertyId == "" {
		response.FailWithMsg(c, response.ParamsValidError, "propertyid不能为空")
		return
	}
	err := t.Srv.DeleteProperties(name, cast.ToInt(propertyId))
	response.CheckAndResp(c, err)
}

func (t *deviceType) findProperty(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	propertyId := c.Param("propertyid")
	if propertyId == "" {
		response.FailWithMsg(c, response.ParamsValidError, "propertyid不能为空")
		return
	}
	res, err := t.Srv.FindProperty(name, cast.ToInt(propertyId))
	response.CheckAndRespWithData(c, res, err)
}

func (t *deviceType) findAllProperties(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}
	res, err := t.Srv.FindAllProperties(name)
	response.CheckAndRespWithData(c, res, err)
}

func (t *deviceType) upload(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.FailWithMsg(c, response.ParamsValidError, "name不能为空")
		return
	}

	res, err := t.Srv.Upload(name, c)
	response.CheckAndRespWithData(c, res, err)
}
