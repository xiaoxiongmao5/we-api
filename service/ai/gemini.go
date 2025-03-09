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

type Part struct {
	Text string `json:"text"`
}
type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}
type GeminiReq struct {
	Contents []Content `json:"contents"`
}

type GeminiRes struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string  `json:"finishReason"`
		AvgLogprobs  float64 `json:"avgLogprobs"`
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

func (o *GeminiSvr) Do(req OpenAiReq) (*OpenAiRes[Choice], error) {
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

func (o *GeminiSvr) DoStream(req OpenAiReq, resChan chan *OpenAiRes[ChoiceStream], errChan chan error) {
	stream := "generateContent" //非流式传输
	if req.Stream {
		stream = "streamGenerateContent" //流式传输
	}
	apikey := strings.Replace(req.AuthHeader, "Bearer ", "", 1)
	url := fmt.Sprintf("/v1beta/models/%s:%s?key=%s&alt=sse", req.Model, stream, apikey)
	reqGemini := o.trans2GeminiReq(req)
	data := map[string]interface{}{
		"contents": reqGemini.Contents,
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
	geminiReq := &GeminiReq{}
	for _, v := range req.Messages {
		content := Content{
			Role:  v["role"],
			Parts: []Part{},
		}
		content.Parts = append(content.Parts, Part{Text: v["content"]})

		geminiReq.Contents = append(geminiReq.Contents, content)
	}
	return geminiReq
}

func (o *GeminiSvr) trans2OpenAiRes(res *GeminiRes) *OpenAiRes[Choice] {
	openAiRes := &OpenAiRes[Choice]{
		Model: res.ModelVersion,
	}
	for _, v := range res.Candidates {
		choice := Choice{
			Index:        0,
			FinishReason: v.FinishReason,
			Message: Message{
				Role:    v.Content.Role,
				Content: v.Content.Parts[0].Text,
			},
		}
		openAiRes.Choices = append(openAiRes.Choices, choice)
	}
	usageMetadata := res.UsageMetadata
	openAiRes.Usage = Usage{
		PromptTokens:     usageMetadata.PromptTokenCount,
		CompletionTokens: usageMetadata.CandidatesTokenCount,
		TotalTokens:      usageMetadata.TotalTokenCount,
	}
	return openAiRes
}

func (o *GeminiSvr) trans2OpenAiStreamRes(res *GeminiRes) *OpenAiRes[ChoiceStream] {
	openAiRes := &OpenAiRes[ChoiceStream]{
		Model: res.ModelVersion,
	}
	for _, v := range res.Candidates {
		choice := ChoiceStream{
			Index:        0,
			FinishReason: v.FinishReason,
			Message: Message{
				Role:    v.Content.Role,
				Content: v.Content.Parts[0].Text,
			},
		}
		openAiRes.Choices = append(openAiRes.Choices, choice)
	}
	usageMetadata := res.UsageMetadata
	openAiRes.Usage = Usage{
		PromptTokens:     usageMetadata.PromptTokenCount,
		CompletionTokens: usageMetadata.CandidatesTokenCount,
		TotalTokens:      usageMetadata.TotalTokenCount,
	}
	return openAiRes
}
