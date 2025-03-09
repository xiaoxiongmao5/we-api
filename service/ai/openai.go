package ai

import (
	"context"

	"github.com/xiaoxiongmao5/we-api/share/request"
	"github.com/xiaoxiongmao5/we-api/utils"
	"github.com/xiaoxiongmao5/we-api/xlog"
)

type OpenAiSvr struct {
	ctx    context.Context
	logger *xlog.Logger
}

func NewOpenAiSvr(ctx context.Context) *OpenAiSvr {
	return &OpenAiSvr{
		ctx:    ctx,
		logger: utils.Log(ctx, "OpenAiSvr"),
	}
}

type OpenAiReq struct {
	AuthHeader  string              `json:"-"`
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	Temperature float64             `json:"temperature"`
	Stream      bool                `json:"stream"`
}

type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Message      Message     `json:"message"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason interface{} `json:"finish_reason"`
	Index        int         `json:"index"`
}
type ChoiceStream struct {
	Message      Message     `json:"delta"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason interface{} `json:"finish_reason"`
	Index        int         `json:"index"`
}

type OpenAiRes[T Choice | ChoiceStream] struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Usage   Usage  `json:"usage"`
	Choices []T    `json:"choices"`
}

func (o *OpenAiSvr) Do(req OpenAiReq) (*OpenAiRes[Choice], error) {
	headers := make(map[string]string)
	headers["Authorization"] = req.AuthHeader

	data := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      req.Stream,
	}
	res, err := request.Fetch[OpenAiRes[Choice]](request.FetchOpts{
		Host:    "https://api-u999v356s7k9v8f0.aistudio-app.com",
		Url:     "/v1/chat/completions",
		Method:  "POST",
		Data:    data,
		Headers: headers,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *OpenAiSvr) DoStream(req OpenAiReq, resChan chan *OpenAiRes[ChoiceStream], errChan chan error) {
	headers := make(map[string]string)
	headers["Authorization"] = req.AuthHeader

	data := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      req.Stream,
	}

	go func() {
		request.FetchStream[OpenAiRes[ChoiceStream]](request.FetchOpts{
			Host:    "https://api-u999v356s7k9v8f0.aistudio-app.com",
			Url:     "/v1/chat/completions",
			Method:  "POST",
			Data:    data,
			Headers: headers,
		}, resChan, errChan)
	}()
}
