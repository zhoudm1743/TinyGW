package cache

import (
	"github.com/patrickmn/go-cache"
	"go.uber.org/fx"
)

func NewCache() *cache.Cache {
	return cache.New(cache.NoExpiration, cache.NoExpiration)
}

var Module = fx.Provide(NewCache)
