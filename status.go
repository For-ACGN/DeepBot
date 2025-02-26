package deepbot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/wdvxdr1123/ZeroBot"
)

const promptGetMood = `
根据上下文的对话，请选择以下最合适的心情/情绪/状态，注意只需要回答两个字，请暂时忽略之前已经要求的回答格式。
可选项: 厌恶、困惑、嫉妒、害羞、尴尬、平静、快乐、恐惧、悲伤、惊讶、愤怒、期待、温暖、生气。
`

var validMoods = map[string]struct{}{
	"厌恶": {}, "困惑": {}, "嫉妒": {}, "害羞": {},
	"尴尬": {}, "平静": {}, "快乐": {}, "恐惧": {},
	"悲伤": {}, "惊讶": {}, "愤怒": {}, "期待": {},
	"温暖": {}, "生气": {},
}

func isValidMood(mood string) bool {
	_, ok := validMoods[mood]
	return ok
}

func (bot *DeepBot) onGetMood(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	mood := user.getMood()
	if mood == "" {
		mood = "平静"
	}

	bot.replyMessage(ctx, mood)
}

func (bot *DeepBot) onUpdateMood(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	mood, err := bot.updateMood(user)
	if err != nil {
		log.Printf("failed to update mood: %s\n", err)
		bot.replyMessage(ctx, "更新心情失败")
		return
	}

	bot.replyMessage(ctx, mood)
}

func (bot *DeepBot) updateMood(user *user) (string, error) {
	req := &ChatRequest{
		Model:       deepseek.DeepSeekChat,
		Temperature: 1,
		MaxTokens:   8192,
	}
	resp, err := bot.seek(req, user, promptGetMood)
	if err != nil {
		return "", fmt.Errorf("failed to get mood: %s", err)
	}
	mood := resp.Answer
	if !isValidMood(mood) {
		return "", fmt.Errorf("get invalid mood: %s", mood)
	}
	user.setMood(mood)
	return mood, nil
}

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
	cr := &chatResp{
		Answer:    content,
		Reasoning: cm.ReasoningContent,
	}
	fmt.Println("==================seek response=================")
	fmt.Println(content)
	fmt.Println("================================================")
	return cr, nil
}
