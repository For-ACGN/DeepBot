package deepbot

import (
	"math/rand/v2"

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
	cfg := bot.config.Emoticon
	if !cfg.Enabled {
		return
	}
	rate := cfg.Rate
	if rate < 1 {
		return
	}
	if rate < rand.IntN(100) {
		return
	}
	bot.replyEmoticon(ctx, user)
}
