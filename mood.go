package deepbot

import (
	"fmt"
	"log"

	"github.com/cohesion-org/deepseek-go"
	"github.com/wdvxdr1123/ZeroBot"
)

const promptGetMood = `
根据上下文的对话，请选择以下最合适的心情/情绪/状态，注意只需要回答两个字，请暂时忽略之前已经要求的回答格式。
可选项: 厌恶、困惑、嫉妒、害羞、尴尬、平静、快乐、恐惧、悲伤、惊讶、愤怒、期待、温暖、感动、生气。
`

var validMoods = map[string]struct{}{
	"厌恶": {}, "困惑": {}, "嫉妒": {}, "害羞": {},
	"尴尬": {}, "平静": {}, "快乐": {}, "恐惧": {},
	"悲伤": {}, "惊讶": {}, "愤怒": {}, "期待": {},
	"温暖": {}, "感动": {}, "生气": {},
}

var moodPrompt = map[string]string{
	"厌恶": "disgusted", "困惑": "confused", "嫉妒": "jealous", "害羞": "shy",
	"尴尬": "embarrassed", "平静": "calm", "快乐": "happy", "恐惧": "afraid",
	"悲伤": "sad", "惊讶": "surprised", "愤怒": "angry", "期待": "expectant",
	"温暖": "warm", "感动": "touched", "生气": "angry", "通用": "",
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

	bot.sendText(ctx, mood)
}

func (bot *DeepBot) onUpdateMood(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	mood, err := bot.updateMood(user)
	if err != nil {
		log.Printf("failed to update mood: %s\n", err)
		bot.sendText(ctx, "更新心情失败")
		return
	}

	bot.sendText(ctx, mood)
}

func (bot *DeepBot) updateMood(user *user) (string, error) {
	req := &ChatRequest{
		Model:       deepseek.DeepSeekChat,
		Temperature: 1,
		TopP:        1,
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
