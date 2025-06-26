package config

import (
	"app-frame-work/filters"
	"app-frame-work/logger"
	"app-frame-work/util"
	"flag"
	"fmt"
	"os"
)

var myLogger = logger.BuildMyLogger()

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

func LoadAppConfig() *AppConfig {
	// 定义命令行参数
	configPath := flag.String("configPath", "config.properties", "指定文件路径")

	// 解析命令行参数
	flag.Parse()
	// 使用参数
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		panic(fmt.Errorf("config.properties not found in: %s", configPath))
	}
	config, err := util.LoadProperties(*configPath)
	if err != nil {
		panic(fmt.Sprintf("Error loading config: %v\n", err))
	}
	appConfig := AppConfig{}
	if err := util.MapToStruct(config, &appConfig); err != nil {
		panic(fmt.Sprintf("Error mapping config: %v\n", err))
	}

	myLogger.Info("load prop file info %+v\n", appConfig)
	return &appConfig
}
