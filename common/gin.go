package common

import (
	"io"

	"github.com/gin-gonic/gin"
)

func GetRequestBody(c *gin.Context) ([]byte, error) {
	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	defer c.Request.Body.Close()
	c.Set("requestBytes", bytes)
	return bytes, nil
}

func UnmarshalBody(c *gin.Context, v any) error {
	err := c.ShouldBind(&v)
	return err
}

func SetEventStreamHeaders(c *gin.Context) {
	// 设置响应头部，关键是 Content-Type: text/event-stream
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no")
}
