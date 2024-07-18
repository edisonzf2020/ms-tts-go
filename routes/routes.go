package routes

import (
	"ms-tts-go/handlers"
	"ms-tts-go/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// 加载模板文件
	router.LoadHTMLGlob("templates/*")

	// 公开路由
	router.GET("/", handlers.Index)

	// 受保护的路由
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.GET("/voices", handlers.GetVoiceList)
		protected.POST("/tts", handlers.SynthesizeVoicePost)
		protected.GET("/tts", handlers.SynthesizeVoice)
	}

	// 添加新的兼容 OpenAI API 的路由
	openai := router.Group("/v1")
	openai.Use(middlewares.AuthMiddleware())
	{
		openai.GET("/models", handlers.GetModels)
		openai.POST("/audio/speech", handlers.CreateSpeech)
	}

	return router
}
