package zhipuai

import (
	adaptercommon "chat/adapter/common"
	"chat/globals"
	"chat/utils"
	"fmt"
	"strings"
)

func (c *ChatInstance) GetChatEndpoint(model string) string {
	return fmt.Sprintf("%s/api/paas/v3/model-api/%s/sse-invoke", c.GetEndpoint(), c.GetModel(model))
}

func (c *ChatInstance) GetModel(model string) string {
	switch model {
	case globals.ZhiPuChatGLMTurbo:
		return ChatGLMTurbo
	case globals.ZhiPuChatGLMPro:
		return ChatGLMPro
	case globals.ZhiPuChatGLMStd:
		return ChatGLMStd
	case globals.ZhiPuChatGLMLite:
		return ChatGLMLite
	default:
		return ChatGLMStd
	}
}

func (c *ChatInstance) FormatMessages(messages []globals.Message) []globals.Message {
	messages = utils.DeepCopy[[]globals.Message](messages)
	for i := range messages {
		if messages[i].Role == globals.Tool {
			continue
		}

		if messages[i].Role == globals.System {
			messages[i].Role = globals.User
		}
	}
	return messages
}

func (c *ChatInstance) GetBody(props *adaptercommon.ChatProps) ChatRequest {
	return ChatRequest{
		Prompt:      c.FormatMessages(props.Message),
		TopP:        props.TopP,
		Temperature: props.Temperature,
	}
}

func (c *ChatInstance) CreateStreamChatRequest(props *adaptercommon.ChatProps, hook globals.Hook) error {
	return utils.EventSource(
		"POST",
		c.GetChatEndpoint(props.Model),
		map[string]string{
			"Content-Type":  "application/json",
			"Accept":        "text/event-stream",
			"Authorization": c.GetToken(),
		},
		ChatRequest{
			Prompt: c.FormatMessages(props.Message),
		},
		func(data string) error {
			if !strings.HasPrefix(data, "data:") {
				return nil
			}

			data = strings.TrimPrefix(data, "data:")
			return hook(&globals.Chunk{Content: data})
		},
	)
}
