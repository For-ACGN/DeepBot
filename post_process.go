package deepbot

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"

	"github.com/wdvxdr1123/ZeroBot"
)

func (bot *DeepBot) postProcess(ctx *zero.Ctx, user *user, msg string) {
	if user == nil {
		return
	}
	bot.mayUpdateMood(user)
	bot.randomEmoticon(ctx, user)

	_ = msg
}

func (bot *DeepBot) mayUpdateMood(user *user) {
	if rand.IntN(100) < 75 {
		return
	}
	_, _ = bot.updateMood(user)
}

func (bot *DeepBot) randomEmoticon(ctx *zero.Ctx, user *user) {
	rate := bot.config.Emoticon.Rate
	if rate < 1 {
		return
	}
	if rate < rand.IntN(100) {
		return
	}
	bot.replyEmoticon(ctx, user)
}

func (bot *DeepBot) replyEmoticon(ctx *zero.Ctx, user *user) {
	if user == nil || user.getCharacter() == "" {
		dir := "data/emoticon/通用"
		cat := selectRandomItem(dir)
		img := selectRandomItem(cat)
		bot.replyImage(ctx, img)
		return
	}
	role := user.getRole()
	category := user.getMood()
	if category == "" {
		category = "通用"
	}
	dir := fmt.Sprintf("data/emoticon/%s/%s", role, category)
	img := selectRandomItem(dir)
	bot.replyImage(ctx, img)
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
