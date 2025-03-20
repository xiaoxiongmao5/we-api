package controller

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/xiaoxiongmao5/we-api/common"
	"github.com/xiaoxiongmao5/we-api/meta"
	"github.com/xiaoxiongmao5/we-api/relay/model"
	"github.com/xiaoxiongmao5/we-api/service/adaptor"
	"github.com/xiaoxiongmao5/we-api/service/adaptor/openai"
	"github.com/xiaoxiongmao5/we-api/utils"
	"github.com/xiaoxiongmao5/we-api/xlog"
)

func RelayTextHander(c *gin.Context) {
	meta := meta.GetByContext(c)
	ctx := c.Request.Context()
	logger := utils.Log(ctx, "RelayTextHander")
	var textRequest *model.GeneralOpenAIRequest
	err := common.UnmarshalBody(c, &textRequest)
	if err != nil {
		logger.Error("UnmarshalBody err", xlog.Err(err), xlog.String("model", textRequest.Model))
		c.JSON(400, gin.H{"message": "param invalid"})
		return
	}
	meta.IsStream = textRequest.Stream
	meta.FullMode = textRequest.Model

	// 获取适配器
	adaptorImpl := GetAdaptor(textRequest.Model)
	if adaptorImpl == nil {
		return
	}

	// get request body
	requestBody, err := getRequestBody(c, meta, textRequest, adaptorImpl)
	if err != nil {
		c.JSON(410, gin.H{"message": err.Error()})
		return
	}

	// do request
	resp, err := adaptor.DoRequest(c, adaptorImpl, meta, requestBody)
	if err != nil {
		logger.Error("DoRequest failed", xlog.Err(err))
		c.JSON(411, gin.H{"message": err.Error()})
		return
	}

	// do response
	usage, respErr := adaptorImpl.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Error("respErr is not nil", xlog.Any("respErr", respErr))
		return
	}

	logger.Info("usage", xlog.Any("usage", usage))
}

func GetAdaptor(model string) adaptor.Adaptor {
	var svr adaptor.Adaptor
	switch model {
	// case "gpt-3.5-turbo", "gpt-4o":
	// 	svr = &openai.Adaptor{}
	// case "gemini-2.0-flash-exp":
	// 	return ai.NewGeminiSvr(ctx)
	// case "claude-3-5-sonnet-20241022":
	// 	return ai.NewClaudeSvr(ctx)
	default:
		svr = &openai.Adaptor{}
	}
	return svr
}

func getRequestBody(c *gin.Context, meta *meta.Meta, textRequest *model.GeneralOpenAIRequest, adaptorImpl adaptor.Adaptor) (io.Reader, error) {
	ctx := c.Request.Context()
	logger := utils.Log(ctx, "getRequestBody")

	// 转换请求参数到具体适配器格式
	convertedRequest, err := adaptorImpl.ConvertRequest(c, meta, textRequest)
	if err != nil {
		logger.Error("ConvertRequest err", xlog.Err(err), xlog.String("model", textRequest.Model))
		return nil, err
	}

	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		logger.Error("json.Marshal(convertedRequest) err", xlog.Err(err))
		return nil, err
	}

	logger.Info("", xlog.String("jsonData", string(jsonData)))

	requestBody := bytes.NewReader(jsonData)
	return requestBody, nil

}
