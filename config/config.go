package config

import (
	"app-frame-work/filters"
)

type AppConfig struct {
	AppName         string
	ServiceDiscover struct {
		Registry struct {
			Network  string
			BindAddr string
			Enable   bool
		}
	}
	ServerConfig struct {
		Network  string
		BindAddr string
	}
	ServiceType   string
	Filters       filters.Filters
	IgnoreRouters []string
}

func NewDefaultAppConfig() *AppConfig {
	return &AppConfig{
		ServiceDiscover: struct {
			Registry struct {
				Network  string
				BindAddr string
				Enable   bool
			}
		}{Registry: struct {
			Network  string
			BindAddr string
			Enable   bool
		}{Network: "tcp", BindAddr: "127.0.0.1:8848", Enable: false}},
		ServiceType: "backend",
		ServerConfig: struct {
			Network  string
			BindAddr string
		}{Network: "tcp", BindAddr: "127.0.0.1:8080"},
		AppName: "app",
	}
}
