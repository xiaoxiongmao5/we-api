package openai

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaoxiongmao5/we-api/meta"
	"github.com/xiaoxiongmao5/we-api/relay/model"
)

type Adaptor struct {
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	url := "https://api.damser.xyz/v1/chat/completions"

	if meta.IsStream {
		return url, nil
	}

	return url, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)

	if meta.IsStream {
		return nil
	}

	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, meta *meta.Meta, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return request, errors.New("request is nil")
	}

	if request.Stream {
		// always return usage in stream mode
		if request.StreamOptions == nil {
			request.StreamOptions = &model.StreamOptions{
				IncludeUsage: true,
			}
		}
	}
	return request, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, _, usage = StreamHandler(c, resp)
	} else {
		err, usage = Handler(c, resp)
	}

	return
}
