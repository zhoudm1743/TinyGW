package internal

import (
	"TinyGW/internal/controller"
	"TinyGW/internal/repo"
	"TinyGW/internal/routes"

	"go.uber.org/fx"
)

var Module = fx.Module("internal",
	repo.Module,
	controller.Module,
	routes.Module,
)
