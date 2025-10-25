package server

import (
	"review-service/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewConsulRegistrar)

// 服务注册
func NewConsulRegistrar(rc *conf.Registry) *consul.Registry {
	cfg := api.DefaultConfig()
	cfg.Address = rc.Addr
	cfg.Scheme = rc.Scheme
	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	reg := consul.New(client, consul.WithHealthCheck(true))
	return reg
}
