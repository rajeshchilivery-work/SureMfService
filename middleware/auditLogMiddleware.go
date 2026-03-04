package middleware

import (
	"SureCommonService/clients"
	"SureCommonService/queue"
	"SureCommonService/structs"
	"SureCommonService/utils"
	"bytes"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		rbw := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = rbw

		c.Next()

		uid, _ := c.Get("uid")
		userID := ""
		if uid != nil {
			userID = uid.(string)
		}

		auditLog := structs.AuditLog{
			ServiceName:  "SureMFService",
			EndpointURL:  c.Request.URL.String(),
			HTTPMethod:   c.Request.Method,
			RequestBody:  string(bodyBytes),
			ResponseBody: rbw.body.String(),
			UserID:       userID,
			IPAddress:    c.ClientIP(),
			CreatedAt:    time.Now(),
		}
		log.Printf("[AUDIT] %s %s %d", auditLog.HTTPMethod, auditLog.EndpointURL, c.Writer.Status())

		b, err := utils.StructToBytes(auditLog)
		if err != nil {
			log.Printf("Failed to marshal audit log: %v", err)
			return
		}
		client := clients.NewClient()
		client.PublishMessage("SureMFService", queue.GetQueueConfig(queue.AUDIT_LOG).QueueName, b)
	}
}
