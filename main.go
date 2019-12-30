package main

import (
	"context"
	"net/http"
	"time"
	"websocket-gate/booster"
	"websocket-gate/common"
	"websocket-gate/config"
	"websocket-gate/logger"
	"websocket-gate/router"

	"github.com/judwhite/go-svc/svc"
)

type BaseServer struct {
	server *http.Server
}

func (s *BaseServer) Init(env svc.Environment) error {
	config.InitConfig()
	println("RunMode:", config.CURMODE)
	for key, val := range config.GetSection("system") {
		println(key, val)
	}

	// init log
	logger.ConfigLogger(
		config.CURMODE,
		config.GetConfig("system", "app_name"),
		config.GetConfig("logs", "dir"),
		config.GetConfig("logs", "file_name"),
		config.GetConfigInt("logs", "keep_days"),
		config.GetConfigInt("logs", "rate_hours"),
	)
	println("logger init success")

	// init mysql
	dbInfo := config.GetSection("dbInfo")
	for name, info := range dbInfo {
		if err := common.AddDB(
			name,
			info,
			config.GetConfigInt("mysql", "maxConn"),
			config.GetConfigInt("mysql", "idleConn"),
			time.Hour*time.Duration(config.GetConfigInt("mysql", "maxLeftTime"))); err != nil {
			return err
		}
	}
	println("mysql init success")

	// init redis
	if err := common.AddRedisInstance(
		"",
		config.GetConfig("redis", "addr"),
		config.GetConfig("redis", "port"),
		config.GetConfig("redis", "password"),
		config.GetConfigInt("redis", "db_num")); err != nil {
		return err
	}
	println("redis init success")

	booster.InitBooster()
	println("booster init success")

	return nil
}

func (s *BaseServer) Start() error {
	s.server = &http.Server{
		Addr:    ":" + config.GetConfig("system", "http_listen_port"),
		Handler: router.NewEngine(),
	}
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()
	println("http service start")

	return nil
}

func (s *BaseServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		println(err.Error())
	}

	// close booster
	booster.CloseBooster()
	println("close booster success")

	// release source
	common.ReleaseMysqlDBPool()
	common.ReleaseRedisPool()

	return nil
}

func main() {
	if err := svc.Run(&BaseServer{}); err != nil {
		println(err.Error())
	}
}
