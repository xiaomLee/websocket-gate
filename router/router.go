package router

import (
	"websocket-gate/booster"
	"websocket-gate/middleware"

	"github.com/gin-gonic/gin"
)

func NewEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// use middleware here
	engine.Use(middleware.Recover)
	engine.Use(middleware.RequestStart)
	engine.Use(middleware.RequestOut)

	// router here
	engine.Any("/", HealthCheck)
	engine.Any("/ws", booster.GetBoosterInstance().WsHandler)

	return engine
}

func HealthCheck(c *gin.Context) {
	c.String(200, "hello world")
	return
}
