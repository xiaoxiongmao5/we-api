package adaptor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaoxiongmao5/we-api/meta"
	"github.com/xiaoxiongmao5/we-api/relay/model"
)

type Adaptor interface {
	GetRequestURL(meta *meta.Meta) (string, error)
	SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error
	ConvertRequest(c *gin.Context, meta *meta.Meta, request *model.GeneralOpenAIRequest) (any, error)
	// ConvertImageRequest(request *model.ImageRequest) (any, error)
	DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode)
}
