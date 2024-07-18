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

func main() {
    // 加载 .env 文件
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    router := routes.SetupRouter()
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

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
