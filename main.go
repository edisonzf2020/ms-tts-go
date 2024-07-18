package main

import (
    "context"
    "ms-tts-go/routes"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/joho/godotenv"
    "github.com/sirupsen/logrus"
)

var log = logrus.New()

func loadConfig() {
    // 尝试加载 .env 文件，如果失败则从环境变量中读取
    if err := godotenv.Load(); err != nil {
        log.Warn("Error loading .env file. Using environment variables.")
    }

    // 设置默认值或从环境变量中读取
    if os.Getenv("PORT") == "" {
        os.Setenv("PORT", "8070")
    }
    if os.Getenv("CACHE_DURATION") == "" {
        os.Setenv("CACHE_DURATION", "3600")
    }
    
    // 检查必需的环境变量
    if os.Getenv("SECRET_TOKEN") == "" {
        log.Fatal("SECRET_TOKEN is not set. This is required.")
    }
}

func main() {
    loadConfig()

    // 配置 logger
    log.SetFormatter(&logrus.JSONFormatter{})
    log.SetLevel(logrus.InfoLevel)

    router := routes.SetupRouter(log)
    port := os.Getenv("PORT")

    srv := &http.Server{
        Addr:    ":" + port,
        Handler: router,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Info("Shutdown Server ...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server Shutdown:", err)
    }
    log.Info("Server exiting")
}
