package render

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func StringData(c *gin.Context, str string) {
	str = strings.TrimPrefix(str, "data: ")
	str = strings.TrimSuffix(str, "\r")

	// SSE 格式: data: your_data\n\n
	fmt.Fprintf(c.Writer, "data: %s\n\n", str)

	// 强制刷新缓冲区，确保数据立即发送到客户端
	c.Writer.(http.Flusher).Flush()
}

func ObjectData(c *gin.Context, object interface{}) error {
	jsonData, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("json.Marshal(object) err: %w", err)
	}

	StringData(c, string(jsonData))
	return nil
}

func Done(c *gin.Context) {
	StringData(c, "[DONE]\n\n")
	// fmt.Fprintf(c.Writer, "[DONE]\n\n")
	// c.Writer.(http.Flusher).Flush()
}
