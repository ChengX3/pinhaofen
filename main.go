package main

import (
	"fmt"
	"log"

	"zufen/config"
	"zufen/database"
	"zufen/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := database.Init(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	r := gin.Default()

	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	api := r.Group("/api")
	{
		api.POST("/register", handler.Register)
		api.GET("/status/:uuid", handler.GetStatus)
		api.GET("/config", handler.GetConfig)
	}

	port := config.Get().Server.Port
	addr := fmt.Sprintf(":%d", port)
	log.Printf("服务启动在 http://localhost%s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
