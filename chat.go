package deepbot

import (
	"context"
	"encoding/json"
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

	req := &deepseek.ChatCompletionRequest{
		Model:       deepseek.DeepSeekChat,
		Temperature: 1.3,
		MaxTokens:   8192,
		// Tools:       bot.tools,
	}
	resp, err := bot.chat(req, user, msg)
	if err != nil {
		log.Printf("%s, failed to chat: %s\n", resp, err)
		return
	}

	replyMessage(ctx, resp)
}

func (bot *DeepBot) onCoder(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "coder ", "", 1)
	fmt.Println("coder", ctx.Event.GroupID, msg)
	user := bot.getUser(ctx.Event.UserID)

	req := &deepseek.ChatCompletionRequest{
		Model:       deepseek.DeepSeekCoder,
		Temperature: 0,
		MaxTokens:   8192,
	}
	resp, err := bot.chat(req, user, msg)
	if err != nil {
		log.Printf("%s, failed to chat: %s\n", resp, err)
		return
	}

	replyMessage(ctx, resp)
}

func (bot *DeepBot) onReasoner(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "ai ", "", 1)
	fmt.Println("ai", ctx.Event.GroupID, msg)
	user := bot.getUser(ctx.Event.UserID)

	req := &deepseek.ChatCompletionRequest{
		Model:       deepseek.DeepSeekReasoner,
		Temperature: 1.2,
		MaxTokens:   8192,
	}
	resp, err := bot.chat(req, user, msg)
	if err != nil {
		log.Printf("%s, failed to chat: %s\n", resp, err)
		return
	}

	replyMessage(ctx, resp)
}

func (bot *DeepBot) onMessage(ctx *zero.Ctx) {
	if !ctx.Event.IsToMe {
		return
	}
	msg := ctx.MessageString()
	user := bot.getUser(ctx.Event.UserID)

	req := &deepseek.ChatCompletionRequest{
		MaxTokens: 8192,
	}
	switch user.getModel() {
	case deepseek.DeepSeekChat:
		req.Model = deepseek.DeepSeekChat
		req.Temperature = 1.3
		req.Tools = bot.tools
	case deepseek.DeepSeekCoder:
		req.Model = deepseek.DeepSeekCoder
		req.Temperature = 0
	case deepseek.DeepSeekReasoner:
		req.Model = deepseek.DeepSeekCoder
		req.Temperature = 1.2
	default:
		replyMessage(ctx, "非法模型名称")
		return
	}
	resp, err := bot.chat(req, user, msg)
	if err != nil {
		log.Printf("%s, failed to chat: %s\n", resp, err)
		return
	}

	replyMessage(ctx, resp)
}

func (bot *DeepBot) onSetModel(ctx *zero.Ctx) {
	msg := textToArgN(ctx.MessageString(), 2)
	if len(msg) != 2 {
		replyMessage(ctx, "非法参数格式")
		return
	}

	model := msg[1]
	switch model {
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

func (bot *DeepBot) chat(req *deepseek.ChatCompletionRequest, user *user, msg string) (string, error) {
	var messages []deepseek.ChatCompletionMessage
	// append system prompt
	character := user.getCharacter()
	if character != "" {
		messages = append(messages, deepseek.ChatCompletionMessage{
			Role:    constants.ChatMessageRoleSystem,
			Content: character,
		})
	}
	// append user past round message
	rounds := user.getRounds()
	for i := 0; i < len(rounds); i++ {
		messages = append(messages, rounds[i].Question)
		messages = append(messages, rounds[i].Answer)
	}
	// append user question
	question := deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: msg,
	}
	messages = append(messages, question)
	// send request
	req.Messages = messages
	resp, err := bot.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %s", err)
	}
	// err = processToolCall(client, req, resp)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to process tool call: %s", err)
	// }
	cm := resp.Choices[0].Message
	if cm.Role != constants.ChatMessageRoleAssistant {
		return "", errors.New("invalid response role: " + cm.Role)
	}
	response := cm.Content
	if response == "" {
		return "", errors.New("receive empty response")
	}
	answer := deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleAssistant,
		Content: response,
	}
	rounds = append(rounds, &round{
		Question: question,
		Answer:   answer,
	})
	user.setRounds(rounds)
	return response, nil
}

func processToolCall(client *deepseek.Client, request *deepseek.ChatCompletionRequest, resp *deepseek.ChatCompletionResponse) error {
	toolCalls := resp.Choices[0].Message.ToolCalls
	if len(toolCalls) == 0 {

		fmt.Println("debug: exit processToolCall")
		return nil
	}
	tc := toolCalls[0]
	fmt.Println(tc)
	fmt.Println(tc.Function.Name)

	msg := resp.Choices[0].Message
	fmt.Println(msg.Role)
	fmt.Println(msg.Content)

	newMsg := request.Messages
	newMsg = append(newMsg, deepseek.ChatCompletionMessage{
		Role:       msg.Role,
		Content:    msg.Content,
		ToolCallID: tc.ID,
		ToolCalls:  toolCalls,
	})

	var result string
	// switch tc.Function.Name {
	// case "GetTime":
	// 	result = onGetTime()
	// case "EvalGo":
	// 	decoder := json.NewDecoder(strings.NewReader(tc.Function.Arguments))
	// 	decoder.DisallowUnknownFields()
	// 	args := struct {
	// 		Src string `json:"src"`
	// 	}{}
	// 	err := decoder.Decode(&args)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	result = onEvalGo(args.Src)
	// }

	decoder := json.NewDecoder(strings.NewReader(tc.Function.Arguments))
	decoder.DisallowUnknownFields()
	args := struct {
		Src string `json:"src"`
	}{}
	err := decoder.Decode(&args)
	if err != nil {
		fmt.Println("panic::::!!!!!!!!!!!!!", err)
		return err
	}
	result = onEvalGo(args.Src)
	fmt.Println("onEvalGo result:", result)

	newMsg = append(newMsg, deepseek.ChatCompletionMessage{
		Role:       "tool",
		Content:    result,
		ToolCallID: tc.ID,
	})

	toolReq := &deepseek.ChatCompletionRequest{
		Model:       deepseek.DeepSeekChat,
		Messages:    newMsg,
		Temperature: 1.3,
		MaxTokens:   8192,
		Tools:       defaultTools,
	}
	resp, err = client.CreateChatCompletion(context.Background(), toolReq)
	if err != nil {
		return err
	}
	return processToolCall(client, toolReq, resp)
}

func chatStream(client *deepseek.Client, request *deepseek.StreamChatCompletionRequest) (string, error) {
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
	return response, err
}
