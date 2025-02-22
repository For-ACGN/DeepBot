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

const promptToolCall = "" +
	"你可以生成并且执行Go语言代码，来访问原先你访问不到的外部资源，具体请使用EvalGo工具函数。"

func (bot *DeepBot) onChat(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "chat ", "", 1)
	fmt.Println("chat", ctx.Event.GroupID, msg)
	user := bot.getUser(ctx.Event.UserID)

	req := &ChatRequest{
		Model:       deepseek.DeepSeekChat,
		Temperature: 1.3,
		MaxTokens:   8192,
		Tools:       bot.tools,
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

	req := &ChatRequest{
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

	req := &ChatRequest{
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
	model := user.getModel()

	req := &ChatRequest{
		MaxTokens: 8192,
		Model:     model,
	}
	switch model {
	case deepseek.DeepSeekChat:
		req.Temperature = 1.3
		req.Tools = bot.tools
	case deepseek.DeepSeekCoder:
		req.Temperature = 0
	case deepseek.DeepSeekReasoner:
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

func (bot *DeepBot) onGetModel(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)
	model := user.getModel()

	replyMessage(ctx, "当前模型: "+model)
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

func (bot *DeepBot) chat(req *ChatRequest, user *user, msg string) (string, error) {
	var messages []ChatMessage
	// build and append system prompt
	character := user.getCharacter()
	if len(bot.tools) > 0 {
		character += "\n" + promptToolCall
	}
	if character != "" {
		messages = append(messages, ChatMessage{
			Role:    constants.ChatMessageRoleSystem,
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
	resp, err = bot.doToolCall(req, resp, user)
	if err != nil {
		return "", fmt.Errorf("failed to process tool call: %s", err)
	}
	cm := resp.Choices[0].Message
	if cm.Role != constants.ChatMessageRoleAssistant {
		return "", errors.New("invalid response role: " + cm.Role)
	}
	response := cm.Content
	if response == "" {
		return "", errors.New("receive empty response")
	}
	answer := ChatMessage{
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

func (bot *DeepBot) doToolCall(req *ChatRequest, resp *ChatResponse, user *user) (*ChatResponse, error) {
	toolCalls := resp.Choices[0].Message.ToolCalls
	numCalls := len(toolCalls)
	if numCalls == 0 {
		return resp, nil
	}
	fmt.Println("num calls:", numCalls)

	question := ChatMessage{
		Role:      constants.ChatMessageRoleAssistant,
		ToolCalls: toolCalls,
	}
	var answer []ChatMessage
	for i := 0; i < numCalls; i++ {
		toolCall := toolCalls[i]
		fnName := toolCall.Function.Name

		var result string
		switch fnName {
		case "GetTime":
			result = onGetTime()
		case "EvalGo":
			decoder := json.NewDecoder(strings.NewReader(toolCall.Function.Arguments))
			decoder.DisallowUnknownFields()
			args := struct {
				Src string `json:"src"`
			}{}
			err := decoder.Decode(&args)
			if err != nil {
				return nil, err
			}
			fmt.Println(args.Src)
			result = onEvalGo(args.Src)
		// case "GetLocation":
		// 	result = "当前城市是: 汉堡王"
		// case "GetTemperature":
		// 	result = "当前温度是: 8℃"
		// case "GetRelativeHumidity":
		// 	result = "当前相对湿度是: 32%"
		default:
			return nil, fmt.Errorf("unknown function: %s", fnName)
		}
		fmt.Println(fnName, result)

		answer = append(answer, ChatMessage{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCall.ID,
		})
	}

	messages := req.Messages
	messages = append(messages, question)
	messages = append(messages, answer...)
	toolReq := &ChatRequest{
		Model:       deepseek.DeepSeekChat,
		Messages:    messages,
		Temperature: 1.3,
		MaxTokens:   8192,
		Tools:       defaultTools,
	}
	resp, err := bot.client.CreateChatCompletion(context.Background(), toolReq)
	if err != nil {
		return nil, err
	}

	// 2025/02/22 经过测试，模型暂时不会将工具函数的返回结果应用在全局上下文，只有当前一轮的问答。
	// 可能是为了避免不及时的函数结果，但是后期可以加一个标志，使其可以应用在全局上下文。
	// rounds := user.getRounds()
	// rounds = append(rounds, &round{
	// 	Question: question,
	// })
	// for i := 0; i < len(answer); i++ {
	// 	rounds = append(rounds, &round{
	// 		Answer: answer[i],
	// 	})
	// }
	// user.setRounds(rounds)

	return bot.doToolCall(toolReq, resp, user)
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
