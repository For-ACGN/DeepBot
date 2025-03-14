package deepbot

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/cohesion-org/deepseek-go"
	"github.com/wdvxdr1123/ZeroBot"
)

const promptGetEmoticon = `
根据上下文的对话，请生成一串合适的AI画图prompt，以", "隔开每一个prompt，注意不要超过30个prompt。
请注意用英文，因为我需要用NovelAI提供的模型来生成。
`

func (bot *DeepBot) replyEmoticon(ctx *zero.Ctx, user *user) {
	if user == nil || user.getCharacter() == "" {
		dir := "data/emoticon/通用"
		cat := selectRandomItem(dir)
		img := selectRandomItem(cat)
		bot.sendImage(ctx, img)
		return
	}

	role := user.getRole()
	category := user.getMood()
	if category == "" {
		category = "通用"
	}
	if bot.config.SDWebUI.Enabled {
		prompt := user.getPrompt()
		mood := moodPrompt[category]
		prompt = strings.ReplaceAll(prompt, "{{.mood}}", mood)

		// append prompt from model
		req := &ChatRequest{
			Model:       deepseek.DeepSeekChat,
			Temperature: 0.5,
			TopP:        1,
			MaxTokens:   4096,
		}
		resp, err := bot.seek(req, user, promptGetEmoticon)
		if err == nil {
			prompt += ", " + resp.Answer
		} else {
			log.Println("failed to get emoticon prompt:", err)
		}
		fmt.Println("draw image prompt:", prompt)

		img, err := bot.drawImage(prompt, 30, 1024, 1024)
		if err == nil {
			sendImage(ctx, img)
			return
		}
	}

	dir := fmt.Sprintf("data/emoticon/%s/%s", role, category)
	img := selectRandomItem(dir)
	bot.sendImage(ctx, img)
}

func selectRandomItem(dir string) string {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	entry, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	if len(entry) == 0 {
		return ""
	}
	idx := rand.IntN(len(entry))
	item := entry[idx]
	return filepath.Join(dir, item.Name())
}
