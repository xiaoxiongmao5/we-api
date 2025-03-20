package model

/*
[INFO]对于第三方接口的参数，如果有默认值，那我们在定义结构体字段类型时，可以使用指针类型，
然后设置omitempty，如果客户端传参没传，则该值是个空指针nil，并且在json序列化时不会出现该字段，
这样传递给第三方时就不会带上该参数，第三方就使用默认值了，如果传参了，就正常传递给第三方。
问题：指针类型的话，如何将客户端参数正确赋值呢？感觉只要设置omitempty就可以了，为什么要指针？


问题：对于map类型的，无论值是何类型，都整体使用any类型接收，为什么呢？

对于任何结构体内嵌，目前都是要单独定义结构体的，不会在父结构体中直接定义子结构体
*/

type Audio struct { //音频输出的参数。当使用以下方式请求音频输出时，必填 modalities: ["audio"]
	Voice  string `json:"voice,omitempty"`  //模型用于响应的语音。支持的声音包括 alloy 、 ash 、 ballad 、 coral 、 echo 、 sage和shimmer
	Format string `json:"format,omitempty"` //指定输出音频格式。必须是wav 、 mp3 、 flac之一， opus或pcm16
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"` //如果设置，则会在data: [DONE] 消息。此块上的usage字段显示整个请求的令牌使用情况统计信息，而choices字段将始终为空数组。所有其他块也将包含一个usage字段，但值为空。注意：如果流中断，您可能无法收到包含请求的总令牌使用量的最终使用块。
}

type ResponseFormat struct { //响应格式
	Type       string      `json:"type,omitempty"` //text json_schema json_object
	JsonSchema *JSONSchema `json:"json_schema,omitempty"`
}

type JSONSchema struct {
	Description string                 `json:"description,omitempty"`
	Name        string                 `json:"name"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Strict      *bool                  `json:"strict,omitempty"`
}

type GeneralOpenAIRequest struct {
	// https://platform.openai.com/docs/api-reference/chat/create
	Messages            []Message       `json:"messages,omitempty"`
	Model               string          `json:"model,omitempty"`
	Store               *bool           `json:"store,omitempty"`
	ReasoningEffort     *string         `json:"reasoning_effort,omitempty"`
	Metadata            any             `json:"metadata,omitempty"`
	FrequencyPenalty    *float64        `json:"frequency_penalty,omitempty"`
	LogitBias           any             `json:"logit_bias,omitempty"`
	Logprobs            *bool           `json:"logprobs,omitempty"`
	TopLogprobs         *int            `json:"top_logprobs,omitempty"`
	MaxTokens           int             `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"`
	N                   int             `json:"n,omitempty"`
	Modalities          []string        `json:"modalities,omitempty"`
	Prediction          any             `json:"prediction,omitempty"`
	Audio               *Audio          `json:"audio,omitempty"`
	PresencePenalty     *float64        `json:"presence_penalty,omitempty"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
	Seed                float64         `json:"seed,omitempty"`
	ServiceTier         *string         `json:"service_tier,omitempty"`
	Stop                any             `json:"stop,omitempty"`
	Stream              bool            `json:"stream,omitempty"`
	StreamOptions       *StreamOptions  `json:"stream_options,omitempty"`
	Temperature         *float64        `json:"temperature,omitempty"`
	TopP                *float64        `json:"top_p,omitempty"`
	TopK                int             `json:"top_k,omitempty"`
	Tools               []Tool          `json:"tools,omitempty"`
	ToolChoice          any             `json:"tool_choice,omitempty"`
	ParallelTooCalls    *bool           `json:"parallel_tool_calls,omitempty"`
	User                string          `json:"user,omitempty"`
	FunctionCall        any             `json:"function_call,omitempty"`
	Functions           any             `json:"functions,omitempty"`
	// // https://platform.openai.com/docs/api-reference/embeddings/create
	// Input any `json:"input,omitempty"`
	// EncodingFormat string `json:"encoding_format,omitempty"`
	// Dimensions     int    `json:"dimensions,omitempty"`
	// // https://platform.openai.com/docs/api-reference/images/create
	// Prompt  any     `json:"prompt,omitempty"`
	// Quality *string `json:"quality,omitempty"`
	// Size    string  `json:"size,omitempty"`
	// Style   *string `json:"style,omitempty"`
	// // Others
	// Instruction string `json:"instruction,omitempty"`
	// NumCtx      int    `json:"num_ctx,omitempty"`
}

// func (r GeneralOpenAIRequest) ParseInput() []string {
// 	if r.Input == nil {
// 		return nil
// 	}

// 	var input []string

// 	switch r.Input.(type) {
// 	case string:
// 		input = []string{r.Input.(string)}
// 	case []any:
// 		input = make([]string, len(r.Input.([]any)))
// 		for _, item := range r.Input.([]any) {
// 			if str, ok := item.(string); ok {
// 				input = append(input, str)
// 			}
// 		}
// 	}

// 	return input
// }
