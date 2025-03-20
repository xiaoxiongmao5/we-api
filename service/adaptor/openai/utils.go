package openai

import (
	"github.com/xiaoxiongmao5/we-api/relay/model"
)

func ErrorWrapper(err error, code string, statusCode int) *model.ErrorWithStatusCode {
	return &model.ErrorWithStatusCode{
		Error: model.Error{
			Message: err.Error(),
			Type:    "we_api_error",
			Code:    code,
		},
		StatusCode: statusCode,
	}
}
