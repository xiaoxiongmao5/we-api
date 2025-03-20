package meta

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	Mode           string
	FullMode       string
	APIKey         string
	IsStream       bool
	RequestURLPath string
	StartTime      time.Time
}

func GetByContext(c *gin.Context) *Meta {
	meta := Meta{
		APIKey:         strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		RequestURLPath: c.Request.URL.String(),
		StartTime:      time.Now(),
	}

	return &meta
}
