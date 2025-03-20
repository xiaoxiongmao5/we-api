package openai

import "github.com/xiaoxiongmao5/we-api/relay/model"

type TextResponseChoice struct {
	Index        int           `json:"index"`
	Message      model.Message `json:"message"`
	FinishReason string        `json:"finish_reason"`
	Logprobs     interface{}   `json:"logprobs"`
}

type TextResponse struct {
	Id      string               `json:"id"`
	Model   string               `json:"model,omitempty"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Choices []TextResponseChoice `json:"choices"`
	Usage   model.Usage          `json:"usage"`
}

type ChatCompletionsStreamResponseChoice struct {
	Index        int           `json:"index"`
	Message      model.Message `json:"delta"`
	FinishReason *string       `json:"finish_reason,omitempty"`
	Logprobs     interface{}   `json:"logprobs"`
}

type ChatCompletionsStreamResponse struct {
	Id      string                                `json:"id"`
	Model   string                                `json:"model"`
	Object  string                                `json:"object"`
	Created int64                                 `json:"created"`
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
	Usage   *model.Usage                          `json:"usage,omitempty"`
}
