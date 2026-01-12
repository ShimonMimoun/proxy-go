package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"proxy-go/db"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read Request Body
		var requestBody interface{}
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body

			if len(bodyBytes) > 0 {
				json.Unmarshal(bodyBytes, &requestBody) // Try to parse as JSON
				if requestBody == nil {
					requestBody = string(bodyBytes) // Fallback to string
				}
			}
		}

		// Wrap Response Writer
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process Request
		c.Next()

		// Capture Response
		var responseBody interface{}
		respBytes := blw.body.Bytes()
		if len(respBytes) > 0 {
			// For streams, this might be partial or full depending on how the handler flushes.
			// But since we are writing to the buffer in the Write() method, we capture everything that was sent.
            // Note: For SSE (streaming), the body will be a long string of "data: ...".
            // We store it as is (string) or try to parse if it's a single JSON object.
			if err := json.Unmarshal(respBytes, &responseBody); err != nil {
				responseBody = string(respBytes)
			}
		}

        // Determine Provider based on path
        provider := "unknown"
        path := c.Request.URL.Path
        if strings.Contains(path, "azure") {
            provider = "azure"
        } else if strings.Contains(path, "bedrock") {
            provider = "bedrock"
        }

		// Log to Mongo
		entry := db.LogEntry{
			Timestamp:      start,
			Method:         c.Request.Method,
			Path:           c.Request.URL.Path,
			RemoteIP:       c.ClientIP(),
			RequestHeader:  c.Request.Header,
			RequestBody:    requestBody,
			ResponseStatus: c.Writer.Status(),
			ResponseBody:   responseBody,
			DurationMs:     time.Since(start).Milliseconds(),
            Provider:       provider,
		}

		// Run logging in goroutine to not block response
		go db.LogExchange(entry)
	}
}
