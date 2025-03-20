package ai

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/xiaoxiongmao5/we-api/share/request"
	"github.com/xiaoxiongmao5/we-api/utils"
	"github.com/xiaoxiongmao5/we-api/xlog"
)

type ClaudeSvr struct {
	ctx    context.Context
	logger *xlog.Logger
}

func NewClaudeSvr(ctx context.Context) *ClaudeSvr {
	return &ClaudeSvr{
		ctx:    ctx,
		logger: utils.Log(ctx, "ClaudeSvr"),
	}
}

type ClaudeReq struct {
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

type ClaudeDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeUsage struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}
type ClaudeRes struct {
	Id           string        `json:"id"`
	Type         string        `json:"type"`
	Role         string        `json:"role"`
	Model        string        `json:"model"`
	Content      []ClaudeDelta `json:"content"`
	StopReason   string        `json:"stop_reason"`
	StopSequence string        `json:"stop_sequence"`
	Usage        ClaudeUsage   `json:"usage"`
}

type ClaudeStreamResData struct {
	Type string `json:"type"`
}

// {"id":"msg_01Kk8uTCJsJJY6FodUgLDZSL","type":"message","role":"assistant","model":"claude-3-5-sonnet-20241022","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":1}}
type ClaudeStreamResDataMessage struct {
	Type    string `json:"type"`
	Message struct {
		Id    string `json:"id"`
		Type  string `json:"type"`
		Role  string `json:"role"`
		Model string `json:"model"`
		Usage struct {
			InputTokens              int `json:"input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			OutputTokens             int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// {"type":"text_delta","text":"! How can I help you today?"}
type ClaudeStreamResDataContentBlockDelta struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

// {"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":12}
type ClaudeStreamResDataMessageDelta struct {
	Type  string `json:"type"`
	Delta struct {
		StopReason string `json:"stop_reason"`
		Usage      struct {
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
}

func (o *ClaudeSvr) getFetchOpts(req OpenAiReq) request.FetchOpts {
	myReq := ClaudeReq{
		Model:            req.Model,
		Messages:         req.Messages,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		N:                req.N,
		Stream:           req.Stream,
		Stop:             req.Stop,
		MaxTokens:        req.MaxTokens,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
		LogitBias:        req.LogitBias,
		User:             req.User,
	}

	apikey := strings.Replace(req.AuthHeader, "Bearer ", "", 1)
	headers := map[string]string{
		"x-api-key":         apikey,
		"anthropic-version": "2023-06-01", //time.Now().Format("2006-01-02"),
	}

	return request.FetchOpts{
		Host:     "https://poloai.top",
		Url:      "/v1/messages",
		Method:   "POST",
		PostData: myReq,
		Headers:  headers,
	}
}

func (o *ClaudeSvr) Do(req OpenAiReq) (*OpenAiRes[OpenAiChoice], error) {
	fetchOpts := o.getFetchOpts(req)

	res, err := request.Fetch[ClaudeRes](fetchOpts)
	if err != nil {
		return nil, err
	}

	resOpenAi := o.trans2OpenAiRes(res)
	return resOpenAi, nil
}

func (o *ClaudeSvr) DoStream(req OpenAiReq, resChan chan *OpenAiRes[OpenAiChoiceStream], errChan chan error) {
	fetchOpts := o.getFetchOpts(req)

	myChan := make(chan []byte)
	go func() {
		request.FetchStreamBase(fetchOpts, myChan, errChan)
	}()

	one := &ClaudeRes{
		Content: make([]ClaudeDelta, 0),
		Usage:   ClaudeUsage{},
	}
	for res := range myChan {
		str := string(res)
		if strings.HasPrefix(str, "event:") {
			continue
		}
		if strings.HasPrefix(str, "data:") {
			b := []byte(strings.TrimPrefix(str, "data: "))
			var baseData ClaudeStreamResData
			if err := json.Unmarshal(b, &baseData); err != nil {
				errChan <- err
				return
			}
			switch baseData.Type {
			case "message_start":
				var data ClaudeStreamResDataMessage
				err := json.Unmarshal(b, &data)
				if err != nil {
					errChan <- err
					return
				}
				one.Id = data.Message.Id
				one.Model = data.Message.Model
				one.Role = data.Message.Role
				one.Usage.InputTokens = data.Message.Usage.InputTokens
			case "content_block_delta":
				if len(one.Content) == 0 {
					one.Content = append(one.Content, ClaudeDelta{one.Role, ""})
				}
				var data ClaudeStreamResDataContentBlockDelta
				err := json.Unmarshal(b, &data)
				if err != nil {
					errChan <- err
					return
				}
				one.Content[0].Text = data.Delta.Text
				resChan <- o.trans2OpenAiStreamRes(one)
				one.Content[0].Text = ""
			case "message_delta":
				var data ClaudeStreamResDataMessageDelta
				err := json.Unmarshal(b, &data)
				if err != nil {
					errChan <- err
					return
				}
				one.Usage.OutputTokens = data.Delta.Usage.OutputTokens
				resChan <- o.trans2OpenAiStreamRes(one)
			case "message_stop":
				close(resChan)
				return
			}
		}
	}
}

func (o *ClaudeSvr) trans2OpenAiRes(res *ClaudeRes) *OpenAiRes[OpenAiChoice] {
	result := &OpenAiRes[OpenAiChoice]{
		Id:    res.Id,
		Model: res.Model,
		Usage: OpenAiUsage{
			PromptTokens:     res.Usage.InputTokens,
			CompletionTokens: res.Usage.OutputTokens,
			TotalTokens:      res.Usage.InputTokens + res.Usage.OutputTokens,
		},
		Choices: []OpenAiChoice{},
	}

	for _, v := range res.Content {
		result.Choices = append(result.Choices, OpenAiChoice{
			Message: OpenAiMessage{
				Role:    res.Role,
				Content: v.Text,
			},
		})
	}

	return result
}

func (o *ClaudeSvr) trans2OpenAiStreamRes(res *ClaudeRes) *OpenAiRes[OpenAiChoiceStream] {
	result := &OpenAiRes[OpenAiChoiceStream]{
		Id:    res.Id,
		Model: res.Model,
		Usage: OpenAiUsage{
			PromptTokens:     res.Usage.InputTokens,
			CompletionTokens: res.Usage.OutputTokens,
			TotalTokens:      res.Usage.InputTokens + res.Usage.OutputTokens,
		},
		Choices: []OpenAiChoiceStream{},
	}

	for _, v := range res.Content {
		choice := OpenAiChoiceStream{
			Index: 0,
			Message: OpenAiMessage{
				Role:    v.Type,
				Content: v.Text,
			},
		}
		result.Choices = append(result.Choices, choice)
	}

	return result
}
