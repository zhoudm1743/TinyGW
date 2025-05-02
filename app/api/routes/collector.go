package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/service"
	"tinyGW/app/api/types"
	"tinyGW/pkg/plugin/response"
	"tinyGW/pkg/util"
)

type account struct {
	fx.In
	Srv service.AccountService
}

func accountRouter(t account, r *types.ApiRouter) {
	r.POST("login", t.login)
}

func (t account) login(c *gin.Context) {
	var loginReq req.AccountLoginReq
	if response.IsFailWithResp(c, util.VerifyUtil.Verify(c, "POST", &loginReq)) {
		return
	}
	res, err := t.Srv.Login(&loginReq)
	response.CheckAndRespWithData(c, res, err)
}
