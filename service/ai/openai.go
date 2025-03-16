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
	AuthHeader string `json:"-"`
	Model      string `json:"model"`
	// 以聊天格式生成聊天完成的消息。
	Messages []map[string]string `json:"messages"`
	// 使用什么采样温度，介于 0 和 2 之间。较高的值（如 0.8）将使输出更加随机，而较低的值（如 0.2）将使输出更加集中和确定。 我们通常建议改变这个或top_p但不是两者。
	Temperature float64 `json:"temperature"`
	// 若设置，将发送部分消息增量，就像在 ChatGPT 中一样。当令牌可用时，令牌将作为纯数据服务器发送事件   data: [DONE]发送，流由消息终止。
	Stream bool `json:"stream"`
	// 聊天完成时生成的最大令牌数。 输入标记和生成标记的总长度受模型上下文长度的限制。
	MaxTokens int `json:"max_tokens"`
}

type OpenAiCompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

type OpenAiUsage struct {
	PromptTokens            int                           `json:"prompt_tokens"`
	CompletionTokens        int                           `json:"completion_tokens"`
	TotalTokens             int                           `json:"total_tokens"`
	CompletionTokensDetails OpenAiCompletionTokensDetails `json:"completion_tokens_details"`
}

type OpenAiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAiChoice struct {
	Message      OpenAiMessage `json:"message"`
	Logprobs     interface{}   `json:"logprobs"`
	FinishReason interface{}   `json:"finish_reason"`
	Index        int           `json:"index"`
}
type OpenAiChoiceStream struct {
	Message      OpenAiMessage `json:"delta"`
	Logprobs     interface{}   `json:"logprobs"`
	FinishReason interface{}   `json:"finish_reason"`
	Index        int           `json:"index"`
}

type OpenAiRes[T OpenAiChoice | OpenAiChoiceStream] struct {
	Id      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Usage   OpenAiUsage `json:"usage"`
	Choices []T         `json:"choices"`
}

func (o *OpenAiSvr) Do(req OpenAiReq) (*OpenAiRes[OpenAiChoice], error) {
	headers := make(map[string]string)
	headers["Authorization"] = req.AuthHeader

	data := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      req.Stream,
	}
	res, err := request.Fetch[OpenAiRes[OpenAiChoice]](request.FetchOpts{
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

func (o *OpenAiSvr) DoStream(req OpenAiReq, resChan chan *OpenAiRes[OpenAiChoiceStream], errChan chan error) {
	headers := make(map[string]string)
	headers["Authorization"] = req.AuthHeader

	data := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      req.Stream,
	}

	go func() {
		request.FetchStream[OpenAiRes[OpenAiChoiceStream]](request.FetchOpts{
			Host:    "https://api-u999v356s7k9v8f0.aistudio-app.com",
			Url:     "/v1/chat/completions",
			Method:  "POST",
			Data:    data,
			Headers: headers,
		}, resChan, errChan)
	}()
}
