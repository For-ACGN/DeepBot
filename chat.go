package deepbot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"github.com/wdvxdr1123/ZeroBot"
)

func (bot *DeepBot) onChat(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "chat ", "", 1)
	fmt.Println("chat", ctx.Event.GroupID, msg)

	user := bot.getUser(ctx.Event.UserID)
	character := user.getCharacter()
	rounds := user.getRounds()

	req := &deepseek.StreamChatCompletionRequest{
		Model:       deepseek.DeepSeekChat,
		Messages:    buildMessages(character, rounds, msg),
		Temperature: 1.3,
		MaxTokens:   8192,
	}
	resp, err := chat(bot.client, req)
	if err != nil {
		log.Printf("%s, failed to send deepseek request: %s\n", resp, err)
		return
	}

	rounds = append(rounds, &round{
		Question: msg,
		Answer:   resp,
	})
	user.setRounds(rounds)

	replyMessage(ctx, resp)
}

func (bot *DeepBot) onCoder(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "coder ", "", 1)
	fmt.Println("coder", ctx.Event.GroupID, msg)

	user := bot.getUser(ctx.Event.UserID)
	character := user.getCharacter()
	rounds := user.getRounds()

	req := &deepseek.StreamChatCompletionRequest{
		Model:       deepseek.DeepSeekCoder,
		Messages:    buildMessages(character, rounds, msg),
		Temperature: 0,
		MaxTokens:   8192,
	}
	resp, err := chat(bot.client, req)
	if err != nil {
		log.Printf("%s, failed to send deepseek request: %s\n", resp, err)
		return
	}

	rounds = append(rounds, &round{
		Question: msg,
		Answer:   resp,
	})
	user.setRounds(rounds)

	replyMessage(ctx, resp)
}

func (bot *DeepBot) onReasoner(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "ai ", "", 1)
	fmt.Println("ai", ctx.Event.GroupID, msg)

	user := bot.getUser(ctx.Event.UserID)
	character := user.getCharacter()
	rounds := user.getRounds()

	req := &deepseek.StreamChatCompletionRequest{
		Model:       deepseek.DeepSeekReasoner,
		Messages:    buildMessages(character, rounds, msg),
		Temperature: 1.2,
		MaxTokens:   8192,
	}
	resp, err := chat(bot.client, req)
	if err != nil {
		log.Printf("%s, failed to send deepseek request: %s\n", resp, err)
		return
	}

	rounds = append(rounds, &round{
		Question: msg,
		Answer:   resp,
	})
	user.setRounds(rounds)

	replyMessage(ctx, resp)
}

func (bot *DeepBot) onMessage(ctx *zero.Ctx) {
	if !ctx.Event.IsToMe {
		return
	}
	msg := ctx.MessageString()

	user := bot.getUser(ctx.Event.UserID)
	character := user.getCharacter()
	rounds := user.getRounds()

	messages := buildMessages(character, rounds, msg)
	var (
		resp string
		err  error
	)
	switch user.getModel() {
	case deepseek.DeepSeekChat:
		req := &deepseek.StreamChatCompletionRequest{
			Model:       deepseek.DeepSeekChat,
			Messages:    messages,
			Temperature: 1.3,
			MaxTokens:   8192,
		}
		resp, err = chat(bot.client, req)
	case deepseek.DeepSeekCoder:
		req := &deepseek.StreamChatCompletionRequest{
			Model:       deepseek.DeepSeekCoder,
			Messages:    messages,
			Temperature: 0,
			MaxTokens:   8192,
		}
		resp, err = chat(bot.client, req)
	case deepseek.DeepSeekReasoner:
		req := &deepseek.StreamChatCompletionRequest{
			Model:       deepseek.DeepSeekReasoner,
			Messages:    messages,
			Temperature: 1.2,
			MaxTokens:   8192,
		}
		resp, err = chat(bot.client, req)
	default:
		log.Println("[error] get invalid model")
	}
	if err != nil {
		log.Printf("%s, failed to send deepseek request: %s\n", resp, err)
		return
	}

	rounds = append(rounds, &round{
		Question: msg,
		Answer:   resp,
	})
	user.setRounds(rounds)

	replyMessage(ctx, resp)
}

func (bot *DeepBot) onSetModel(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "bot.设置模型 ", "", 1)

	var model string
	switch msg {
	case "r1":
		model = deepseek.DeepSeekReasoner
	case "chat":
		model = deepseek.DeepSeekChat
	case "coder":
		model = deepseek.DeepSeekCoder
	case "8b":
		model = "deepseek-r1:8b" // 联合测试使用
	default:
		replyMessage(ctx, "非法模型名称")
		return
	}

	user := bot.getUser(ctx.Event.UserID)
	user.setModel(model)

	replyMessage(ctx, "设置模型成功")
}

func (bot *DeepBot) onReset(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)
	user.setRounds(nil)

	replyMessage(ctx, "重置会话成功")
}

func buildMessages(character string, rounds []*round, msg string) []deepseek.ChatCompletionMessage {
	var messages []deepseek.ChatCompletionMessage
	if character != "" {
		messages = append(messages, deepseek.ChatCompletionMessage{
			Role:    constants.ChatMessageRoleSystem,
			Content: character,
		})
	}
	for i := 0; i < len(rounds); i++ {
		messages = append(messages, deepseek.ChatCompletionMessage{
			Role:    constants.ChatMessageRoleUser,
			Content: rounds[i].Question,
		})
		messages = append(messages, deepseek.ChatCompletionMessage{
			Role:    constants.ChatMessageRoleAssistant,
			Content: rounds[i].Answer,
		})
	}
	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: msg,
	})
	return messages
}

func chat(client *deepseek.Client, request *deepseek.StreamChatCompletionRequest) (string, error) {
	stream, err := client.CreateChatCompletionStream(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion stream: %s", err)
	}
	defer func() { _ = stream.Close() }()
	var response string
	for {
		var resp *deepseek.StreamChatCompletionResponse
		resp, err = stream.Recv()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			err = fmt.Errorf("failed to receive chat completion response: %s", err)
			break
		}
		for _, choice := range resp.Choices {
			response += choice.Delta.Content

			fmt.Print(choice.Delta.Content)
		}
	}
	if response == "" {
		return "", errors.New("receive empty response")
	}
	// // TODO move it
	// response = response[strings.Index(response, "</think>")+8+2:]
	return response, err
}
