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

/*
Messages // 以聊天格式生成聊天完整的消息。

Temperature // 使用什么采样温度，介于 0 和 2 之间。较高的值（如 0.8）将使输出更加随机，而较低的值（如 0.2）将使输出更加集中和确定。 我们通常建议改变这个或top_p但不是两者。

TopP // 一种替代温度采样的方法，称为核采样，其中模型考虑具有 top_p 概率质量的标记的结果。所以 0.1 意味着只考虑构成前 10% 概率质量的标记。 我们通常建议改变这个或temperature但不是两者。

N int // 为每个输入消息生成多少个聊天完成选项。

Stream // 若设置，将发送部分消息增量，就像在 ChatGPT 中一样。当令牌可用时，令牌将作为纯数据服务器发送事件   data: [DONE]发送，流由消息终止。

Stop // API 将停止生成更多令牌的最多 4 个序列。

MaxTokens // 聊天完成时生成的最大令牌数。 输入标记和生成标记的总长度受模型上下文长度的限制。

PresencePenalty // -2.0 和 2.0 之间的数字。正值会根据新标记在文本中的现有频率对其进行惩罚，从而降低模型逐字重复同一行的可能性。

LogitBias // 修改指定标记出现在完成中的可能性。 接受一个 json 对象，该对象将标记（由标记器中的标记 ID 指定）映射到从 -100 到 100 的关联偏差值。从数学上讲，偏差会在采样之前添加到模型生成的 logits 中。确切的效果因模型而异，但 -1 和 1 之间的值应该会减少或增加选择的可能性；像 -100 或 100 这样的值应该导致相关令牌的禁止或独占选择。

User string // 代表您的最终用户的唯一标识符，可以帮助 OpenAI 监控和检测滥用行为。
*/
type OpenAiReq struct {
	AuthHeader       string              `json:"-"`
	Model            string              `json:"model"`
	Messages         []map[string]string `json:"messages"`
	Temperature      float64             `json:"temperature"`
	TopP             float64             `json:"top_p"`
	N                int                 `json:"n"`
	Stream           bool                `json:"stream"`
	Stop             string              `json:"stop"`
	MaxTokens        int                 `json:"max_tokens"`
	PresencePenalty  int                 `json:"presence_penalty"`
	FrequencyPenalty int                 `json:"frequency_penalty"`
	LogitBias        string              `json:"logit_bias"`
	User             string              `json:"user"`
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

func (o *OpenAiSvr) getFetchOpts(req OpenAiReq) request.FetchOpts {
	headers := make(map[string]string)
	headers["Authorization"] = req.AuthHeader

	return request.FetchOpts{
		Host:     "https://api.damser.xyz",
		Url:      "/v1/chat/completions",
		Method:   "POST",
		PostData: req,
		Headers:  headers,
	}
}

func (o *OpenAiSvr) Do(req OpenAiReq) (*OpenAiRes[OpenAiChoice], error) {
	fetchOpts := o.getFetchOpts(req)

	res, err := request.Fetch[OpenAiRes[OpenAiChoice]](fetchOpts)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *OpenAiSvr) DoStream(req OpenAiReq, resChan chan *OpenAiRes[OpenAiChoiceStream], errChan chan error) {
	fetchOpts := o.getFetchOpts(req)

	go func() {
		request.FetchStream[OpenAiRes[OpenAiChoiceStream]](fetchOpts, resChan, errChan)
	}()
}
