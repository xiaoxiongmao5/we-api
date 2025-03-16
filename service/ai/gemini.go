package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiaoxiongmao5/we-api/share/request"
	"github.com/xiaoxiongmao5/we-api/utils"
	"github.com/xiaoxiongmao5/we-api/xlog"
)

type GeminiSvr struct {
	ctx    context.Context
	logger *xlog.Logger
}

func NewGeminiSvr(ctx context.Context) *GeminiSvr {
	return &GeminiSvr{
		ctx:    ctx,
		logger: utils.Log(ctx, "GeminiSvr"),
	}
}

type GeminiPart struct {
	Text string `json:"text"`
}
type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}
type GeminiReq struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiRes struct {
	Candidates []struct {
		Content      GeminiContent `json:"content"`
		FinishReason string        `json:"finishReason"`
		AvgLogprobs  float64       `json:"avgLogprobs"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
		PromptTokensDetails  []struct {
			Modality   string `json:"modality"`
			TokenCount int    `json:"tokenCount"`
		} `json:"promptTokensDetails"`
		CandidatesTokensDetails []struct {
			Modality   string `json:"modality"`
			TokenCount int    `json:"tokenCount"`
		} `json:"candidatesTokensDetails"`
	} `json:"usageMetadata"`
	ModelVersion string `json:"modelVersion"`
}

func (o *GeminiSvr) Do(req OpenAiReq) (*OpenAiRes[OpenAiChoice], error) {
	stream := "generateContent" //非流式传输
	if req.Stream {
		stream = "streamGenerateContent" //流式传输
	}
	apikey := strings.Replace(req.AuthHeader, "Bearer ", "", 1)
	url := fmt.Sprintf("/v1beta/models/%s:%s?key=%s", req.Model, stream, apikey)
	reqGemini := o.trans2GeminiReq(req)
	data := map[string]interface{}{
		"contents": reqGemini.Contents,
	}
	res, err := request.Fetch[GeminiRes](request.FetchOpts{
		Host:   "https://generativelanguage.googleapis.com",
		Url:    url,
		Method: "POST",
		Data:   data,
	})
	if err != nil {
		return nil, err
	}

	resOpenAi := o.trans2OpenAiRes(res)
	return resOpenAi, nil
}

func (o *GeminiSvr) DoStream(req OpenAiReq, resChan chan *OpenAiRes[OpenAiChoiceStream], errChan chan error) {
	stream := "generateContent" //非流式传输
	if req.Stream {
		stream = "streamGenerateContent" //流式传输
	}
	apikey := strings.Replace(req.AuthHeader, "Bearer ", "", 1)
	url := fmt.Sprintf("/v1beta/models/%s:%s?key=%s&alt=sse", req.Model, stream, apikey)
	myReq := o.trans2GeminiReq(req)
	data := map[string]interface{}{
		"contents": myReq.Contents,
	}

	myChan := make(chan *GeminiRes)
	go func() {
		request.FetchStream[GeminiRes](request.FetchOpts{
			Host:   "https://generativelanguage.googleapis.com",
			Url:    url,
			Method: "POST",
			Data:   data,
		}, myChan, errChan)
	}()

	for res := range myChan {
		resChan <- o.trans2OpenAiStreamRes(res)
	}
	close(resChan)
}

func (o *GeminiSvr) trans2GeminiReq(req OpenAiReq) *GeminiReq {
	myReq := &GeminiReq{}
	for _, v := range req.Messages {
		content := GeminiContent{
			Role:  v["role"],
			Parts: []GeminiPart{},
		}
		content.Parts = append(content.Parts, GeminiPart{Text: v["content"]})

		myReq.Contents = append(myReq.Contents, content)
	}
	return myReq
}

func (o *GeminiSvr) trans2OpenAiRes(res *GeminiRes) *OpenAiRes[OpenAiChoice] {
	result := &OpenAiRes[OpenAiChoice]{
		Model: res.ModelVersion,
		Usage: OpenAiUsage{
			PromptTokens:     res.UsageMetadata.PromptTokenCount,
			CompletionTokens: res.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      res.UsageMetadata.TotalTokenCount,
		},
		Choices: []OpenAiChoice{},
	}

	for _, v := range res.Candidates {
		result.Choices = append(result.Choices, OpenAiChoice{
			FinishReason: v.FinishReason,
			Message: OpenAiMessage{
				Role:    v.Content.Role,
				Content: v.Content.Parts[0].Text,
			},
		})
	}

	return result
}

func (o *GeminiSvr) trans2OpenAiStreamRes(res *GeminiRes) *OpenAiRes[OpenAiChoiceStream] {
	result := &OpenAiRes[OpenAiChoiceStream]{
		Model: res.ModelVersion,
		Usage: OpenAiUsage{
			PromptTokens:     res.UsageMetadata.PromptTokenCount,
			CompletionTokens: res.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      res.UsageMetadata.TotalTokenCount,
		},
		Choices: []OpenAiChoiceStream{},
	}

	for _, v := range res.Candidates {
		choice := OpenAiChoiceStream{
			Index:        0,
			FinishReason: v.FinishReason,
			Message: OpenAiMessage{
				Role:    v.Content.Role,
				Content: v.Content.Parts[0].Text,
			},
		}
		result.Choices = append(result.Choices, choice)
	}

	return result
}
