package model

/*
[INFO] 如果一个结构体的某字段可能是多种类型，比如string array nil等，就设置为 any 结构，
然后对该结构体本身提供方法，检测是什么类型，统一转成一个类型即可
*/

type Message struct {
	Role             string  `json:"role,omitempty"` //消息作者的角色
	Content          any     `json:"content,omitempty"`
	ReasoningContent any     `json:"reasoning_content,omitempty"`
	Name             *string `json:"name,omitempty"` //参与者的可选名称。提供模型信息以区分同一角色的参与者
	// ToolCalls        []Tool  `json:"tool_calls,omitempty"`
	// ToolCallId       string  `json:"tool_call_id,omitempty"`
}

type ImageURL struct {
	Url    string `json:"url,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type MessageContent struct {
	Type     string    `json:"type,omitempty"`
	Text     string    `json:"text"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

func (m Message) IsStringContent() bool {
	_, ok := m.Content.(string)
	return ok
}

func (m Message) StringContent() string {
	content, ok := m.Content.(string)
	if ok {
		return content
	}

	contentList, ok := m.Content.([]any)
	if ok {
		var contentStr string
		for _, contentItem := range contentList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			if contentType, ok := contentMap["type"].(string); ok && contentType == ContentTypeText {
				if subStr, ok := contentMap["text"].(string); ok {
					contentStr += subStr
				}
			}
		}
		return contentStr
	}

	return ""
}

func (m Message) ParseContent() []MessageContent {
	conentList := make([]MessageContent, 0)
	contentStr, ok := m.Content.(string)
	if ok {
		conentList = append(conentList, MessageContent{
			Type: ContentTypeText,
			Text: contentStr,
		})
		return conentList
	}

	anyList, ok := m.Content.([]any)
	if ok {
		for _, contentItem := range anyList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			switch contentMap["type"] {
			case ContentTypeText:
				conentList = append(conentList, MessageContent{
					Type: ContentTypeText,
					Text: contentMap["text"].(string),
				})
			case ContentTypeImageURL:
				// [TODO]测试这里是否能映射成功
				image_url, ok := contentMap["image_url"].(ImageURL)
				if !ok {
					break
				}
				conentList = append(conentList, MessageContent{
					Type:     ContentTypeImageURL,
					ImageURL: &image_url,
				})
			}
		}
	}

	return conentList
}
