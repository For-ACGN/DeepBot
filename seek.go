package deepbot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
)

// seek is same as chat, but it will not append response after call.
func (bot *DeepBot) seek(req *ChatRequest, user *user, msg string) (*chatResp, error) {
	var err error
	for i := 0; i < 3; i++ {
		var resp *chatResp
		resp, err = bot.trySeek(req, user, msg)
		if err == nil {
			return resp, nil
		}
		var retry bool
		errStr := err.Error()
		for _, es := range []string{
			"failed to create chat completion",
			"receive empty message content",
		} {
			if strings.Contains(errStr, es) {
				retry = true
				break
			}
		}
		if retry {
			fmt.Printf("[warning] retry send seek request with %d times\n", i+1)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}
	return nil, err
}

func (bot *DeepBot) trySeek(req *ChatRequest, user *user, msg string) (*chatResp, error) {
	var messages []ChatMessage
	// build and append system prompt
	character := user.getCharacter()
	if character != "" {
		messages = append(messages, ChatMessage{
			Role:    deepseek.ChatMessageRoleSystem,
			Content: character,
		})
	}
	// append user past round message
	rounds := user.getRounds()
	for i := 0; i < len(rounds); i++ {
		question := rounds[i].Question
		if question.Role != "" {
			messages = append(messages, question)
		}
		answer := rounds[i].Answer
		if answer.Role != "" {
			messages = append(messages, answer)
		}
	}

	// fmt.Println("================================================")
	// fmt.Println(messages)
	// fmt.Println("================================================")

	// append user question
	question := ChatMessage{
		Role:    deepseek.ChatMessageRoleUser,
		Content: msg,
	}
	messages = append(messages, question)
	// send request
	req.Messages = messages
	resp, err := bot.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %s", err)
	}
	// process response
	cm := resp.Choices[0].Message
	if cm.Role != deepseek.ChatMessageRoleAssistant {
		return nil, errors.New("invalid message role: " + cm.Role)
	}
	content := cm.Content
	if content == "" {
		return nil, errors.New("receive empty message content")
	}

	fmt.Println("==================seek response=================")
	fmt.Println(content)
	fmt.Println("------------------------------------------------")
	usage := resp.Usage
	fmt.Println("prompt token:", usage.PromptTokens, "completion token:", usage.CompletionTokens)
	fmt.Println("cache hit:", usage.PromptCacheHitTokens, "cache miss:", usage.PromptCacheMissTokens)
	fmt.Println("================================================")

	cr := &chatResp{
		Answer:    content,
		Reasoning: cm.ReasoningContent,
	}
	return cr, nil
}
