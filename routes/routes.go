package routes

import (
    "github.com/gin-gonic/gin"
    "ms-tts-go/handlers"
    "ms-tts-go/middlewares"
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

    return router
}
