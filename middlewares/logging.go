// middlewares/logging.go

package middlewares

import (
    "bytes"
    "io"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

func LoggingMiddleware(log *logrus.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 开始时间
        start := time.Now()

        // 读取请求体
        var bodyBytes []byte
        if c.Request.Body != nil {
            bodyBytes, _ = io.ReadAll(c.Request.Body)
        }
        // 恢复请求体，因为读取后 body 会被消耗
        c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

        // 处理请求
        c.Next()

        // 结束时间
        end := time.Now()

        // 日志字段
        fields := logrus.Fields{
            "client_ip":  c.ClientIP(),
            "duration":   end.Sub(start),
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "query":      c.Request.URL.RawQuery,
            "status":     c.Writer.Status(),
            "user_agent": c.Request.UserAgent(),
            "headers":    c.Request.Header,
        }

        // 如果是 POST 请求，记录请求体
        if c.Request.Method == "POST" {
            fields["body"] = string(bodyBytes)
        }

        // 记录日志
        log.WithFields(fields).Info("Request processed")
    }
}
