package deepbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/cohesion-org/deepseek-go"
	"github.com/wdvxdr1123/ZeroBot"
)

const promptToolCall = `
[外部函数调用指南]
   你可以使用浏览器来访问原先你访问不到的外部资源，具体请使用FetchURL工具函数。
   你可以生成并且执行Go语言代码，来访问原先你访问不到的外部资源，具体请使用EvalGo工具函数。
   请注意如果你只是需要浏览网页，请优先使用FetchURL，而不是生成相关代码使用EvalGo来访问。
   一般来说，不要重复地访问同一个URL，以及不要递归访问网站内容中的出现URL。
   一般来说，仅当你需要访问实时信息时才应该使用FetchURL工具函数。
   禁止多次来回调用FetchURL工具函数，一轮对话中只允许使用一次FetchURL。
`

type chatResp struct {
	Answer    string
	Reasoning string
}

func (cr *chatResp) String() string {
	return cr.Answer
}

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

	bot.replyMessage(ctx, user, resp.Answer)
}

func (bot *DeepBot) onChatX(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "chatx ", "", 1)
	fmt.Println("chatx", ctx.Event.GroupID, msg)
	user := bot.getUser(ctx.Event.UserID)

	req := &ChatRequest{
		Model:       deepseek.DeepSeekChat,
		Temperature: 1.3,
		MaxTokens:   8192,
	}
	resp, err := bot.chat(req, user, msg)
	if err != nil {
		log.Printf("%s, failed to chat: %s\n", resp, err)
		return
	}

	bot.replyMessage(ctx, user, resp.Answer)
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

	bot.replyMessage(ctx, user, resp.Answer)
}

func (bot *DeepBot) onReasoning(ctx *zero.Ctx) {
	msg := ctx.MessageString()
	msg = strings.Replace(msg, "aix ", "", 1)
	fmt.Println("aix", ctx.Event.GroupID, msg)
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

	tpl := `
<h3>思考过程</h3>
<div>%s</div>

<h3>回复内容</h3>
<div>%s</div>
`
	reasoning := resp.Reasoning
	if isMarkdown(reasoning) {
		reasoning = markdownToHTML(reasoning)
	}
	answer := resp.Answer
	if isMarkdown(answer) {
		answer = markdownToHTML(answer)
	}
	output := fmt.Sprintf(tpl, reasoning, answer)

	img, err := bot.htmlToImage(output)
	if err != nil {
		log.Println(err)
		return
	}
	sendImage(ctx, img)
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

	bot.replyMessage(ctx, user, resp.Answer)
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
		bot.replyResponse(ctx, "非法模型名称")
		return
	}
	resp, err := bot.chat(req, user, msg)
	if err != nil {
		log.Printf("%s, failed to chat: %s\n", resp, err)
		return
	}

	bot.replyMessage(ctx, user, resp.Answer)
}

func (bot *DeepBot) onGetModel(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)
	model := user.getModel()

	bot.replyResponse(ctx, "当前模型: "+model)
}

func (bot *DeepBot) onSetModel(ctx *zero.Ctx) {
	msg := textToArgN(ctx.MessageString(), 2)
	if len(msg) != 2 {
		bot.replyResponse(ctx, "非法参数格式")
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
		bot.replyResponse(ctx, "非法模型名称")
		return
	}

	user := bot.getUser(ctx.Event.UserID)
	user.setModel(model)

	bot.replyResponse(ctx, "设置模型成功")
}

func (bot *DeepBot) onEnableToolCall(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)
	user.setToolCall(true)

	bot.replyResponse(ctx, "全局启用函数")
}

func (bot *DeepBot) onDisableToolCall(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)
	user.setToolCall(false)

	bot.replyResponse(ctx, "全局禁用函数")
}

func (bot *DeepBot) onReset(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)
	user.setRounds(nil)

	bot.replyResponse(ctx, "重置会话成功")
}

func (bot *DeepBot) onPoke(ctx *zero.Ctx) {
	event := ctx.Event
	if !event.IsToMe {
		return
	}
	if event.NoticeType != "notify" || event.SubType != "poke" {
		return
	}

	switch rand.IntN(8) {
	case 0:
		bot.replyResponse(ctx, "?")
	case 1:
		bot.replyResponse(ctx, "??")
	case 2:
		bot.replyResponse(ctx, "???")
	case 3:
		bot.replyResponse(ctx, "¿¿¿")
	case 4:
		bot.replyResponse(ctx, "别戳了")
	case 5:
		bot.replyResponse(ctx, "再戳我就要爆了")
	default:
		bot.replyEmoticon(ctx, nil)
	}
}

func (bot *DeepBot) chat(req *ChatRequest, user *user, msg string) (*chatResp, error) {
	if !user.canToolCall() {
		req.Tools = nil
		req.ToolChoice = nil
	}
	var err error
	for i := 0; i < 3; i++ {
		var resp *chatResp
		resp, err = bot.tryChat(req, user, msg)
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
			fmt.Printf("[warning] retry send chat request with %d times\n", i+1)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}
	return nil, err
}

func (bot *DeepBot) tryChat(req *ChatRequest, user *user, msg string) (*chatResp, error) {
	var messages []ChatMessage
	// build and append system prompt
	character := user.getCharacter()
	if len(req.Tools) > 0 && req.Model != deepseek.DeepSeekReasoner {
		character += "\n\n" + promptToolCall
	}
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
	// reset usage counter before process tool calls
	user.setContext(usageGetTime, 0)
	user.setContext(usageFetchURL, 0)
	user.setContext(usageEvalGo, 0)
	resp, err = bot.doToolCalls(req, resp, user)
	if err != nil {
		return nil, fmt.Errorf("failed to process tool call: %s", err)
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
	answer := ChatMessage{
		Role:    deepseek.ChatMessageRoleAssistant,
		Content: content,
	}
	rounds = append(rounds, &round{
		Question: question,
		Answer:   answer,
	})
	user.setRounds(rounds)
	cr := &chatResp{
		Answer:    content,
		Reasoning: cm.ReasoningContent,
	}
	fmt.Println("==================chat response=================")
	fmt.Println(content)
	fmt.Println("================================================")
	return cr, nil
}

func (bot *DeepBot) doToolCalls(req *ChatRequest, resp *ChatResponse, user *user) (*ChatResponse, error) {
	toolCalls := resp.Choices[0].Message.ToolCalls
	numCalls := len(toolCalls)
	if numCalls == 0 {
		return resp, nil
	}
	fmt.Println("num calls:", numCalls)

	question := ChatMessage{
		Role:      deepseek.ChatMessageRoleAssistant,
		ToolCalls: toolCalls,
	}
	var answers []ChatMessage
	for i := 0; i < numCalls; i++ {
		toolCall := toolCalls[i]
		answer, err := bot.doToolCall(toolCall, user)
		if err != nil {
			return nil, err
		}
		answers = append(answers, ChatMessage{
			Role:       "tool",
			Content:    answer,
			ToolCallID: toolCall.ID,
		})
		fmt.Println(answer)
	}

	// 只允许使用一次 FetchURL
	// tools := bot.tools
	// fmt.Println(tools)
	// for i := 0; i < numCalls; i++ {
	// 	if toolCalls[i].Function.Name == fnFetchURL {
	// 		for j := 0; j < len(tools); j++ {
	// 			if tools[j].Function.Name == fnFetchURL {
	// 				tools = append(tools[:j], tools[j+1:]...)
	// 			}
	// 		}
	// 		fmt.Println(tools)
	// 	}
	// }

	messages := req.Messages
	messages = append(messages, question)
	messages = append(messages, answers...)
	toolReq := &ChatRequest{
		Model:       deepseek.DeepSeekChat,
		Messages:    messages,
		Temperature: 1.3,
		MaxTokens:   8192,
		// Tools:       tools,
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
	// for i := 0; i < len(answers); i++ {
	// 	rounds = append(rounds, &round{
	// 		Answer: answer[i],
	// 	})
	// }
	// user.setRounds(rounds)
	return bot.doToolCalls(toolReq, resp, user)
}

func (bot *DeepBot) doToolCall(toolCall deepseek.ToolCall, user *user) (string, error) {
	decoder := json.NewDecoder(strings.NewReader(toolCall.Function.Arguments))
	decoder.DisallowUnknownFields()

	fnName := toolCall.Function.Name
	var answer string
	switch fnName {
	case fnGetTime:
		usage := user.getContext(usageGetTime).(int) + 1
		if usage >= 5 {
			return "", fmt.Errorf("too many calls about %s", fnGetTime)
		}
		user.setContext(usageGetTime, usage)

		answer = onGetTime()
	case fnFetchURL:
		usage := user.getContext(usageFetchURL).(int) + 1
		if usage >= 3 {
			return "", fmt.Errorf("too many calls about %s", fnFetchURL)
		}
		user.setContext(usageFetchURL, usage)

		args := struct {
			URL string `json:"url"`
		}{}
		err := decoder.Decode(&args)
		if err != nil {
			return "", err
		}

		timeout := time.Duration(bot.config.FetchURL.Timeout) * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var options []chromedp.ExecAllocatorOption
		cfg := bot.config.FetchURL
		if cfg.ProxyURL != "" {
			options = append(options, chromedp.ProxyServer(cfg.ProxyURL))
		}
		if cfg.ExecPath != "" {
			options = append(options, chromedp.ExecPath(cfg.ExecPath))
		}
		output, err := onFetchURL(ctx, options, args.URL)
		if err != nil {
			return "Chromedp Error: " + err.Error(), nil
		}
		answer = output
	case fnEvalGo:
		usage := user.getContext(usageEvalGo).(int) + 1
		if usage >= 5 {
			return "", fmt.Errorf("too many calls about %s", fnEvalGo)
		}
		user.setContext(usageEvalGo, usage)

		args := struct {
			Src string `json:"src"`
		}{}
		err := decoder.Decode(&args)
		if err != nil {
			return "", err
		}

		timeout := time.Duration(bot.config.EvalGo.Timeout) * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		output, err := onEvalGo(ctx, args.Src)
		if err != nil {
			return "Go Error: " + err.Error(), nil
		}
		answer = output
	default:
		return "", fmt.Errorf("unknown function: %s", fnName)
	}
	return answer, nil
}

// case "GetLocation":
// 	answer = "当前城市是: 汉堡王"
// case "GetTemperature":
// 	answer = "当前温度是: 8℃"
// case "GetRelativeHumidity":
// 	answer = "当前相对湿度是: 32%"

// func chatStream(client *deepseek.Client, request *deepseek.StreamChatCompletionRequest) (string, error) {
// 	stream, err := client.CreateChatCompletionStream(context.Background(), request)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create chat completion stream: %s", err)
// 	}
// 	defer func() { _ = stream.Close() }()
// 	var response string
// 	for {
// 		var resp *deepseek.StreamChatCompletionResponse
// 		resp, err = stream.Recv()
// 		if err == io.EOF {
// 			err = nil
// 			break
// 		}
// 		if err != nil {
// 			err = fmt.Errorf("failed to receive chat completion response: %s", err)
// 			break
// 		}
// 		for _, choice := range resp.Choices {
// 			response += choice.Delta.Content

// 			fmt.Print(choice.Delta.Content)
// 		}
// 	}
// 	if response == "" {
// 		return "", errors.New("receive empty response")
// 	}
// 	return response, err
//
