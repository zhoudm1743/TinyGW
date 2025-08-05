package internal

import (
	"TinyGW/internal/repository"
	"TinyGW/internal/routes"
	"TinyGW/internal/schemas"
	"TinyGW/internal/service"

	"go.uber.org/fx"
)

var Module = fx.Options(
	repository.Module,
	service.Module,
	routes.Module,
	schemas.Module,
)
