package routes

import (
	"go.uber.org/fx"
	"tinyGW/app/api/types"
	"tinyGW/pkg/service/http"
)

var Module = fx.Module("routes",
	fx.Provide(NewRoutes),
	fx.Invoke(accountRouter),
	fx.Invoke(collectorRouter),
	fx.Invoke(collectTaskRouter),
	fx.Invoke(reportTaskRouter),
	fx.Invoke(deviceRouter),
	fx.Invoke(deviceTypeRouter),
	fx.Invoke(userRouter),
)

type Routes struct {
	fx.In
	Http *http.Service
}

func NewRoutes(deps Routes) *types.ApiRouter {
	return &types.ApiRouter{
		deps.Http.Gin,
	}
}
