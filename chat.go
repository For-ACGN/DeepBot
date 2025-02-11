package deepbot

import (
	"fmt"
	"log"
	"strings"

	"github.com/cohesion-org/deepseek-go"
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

func (bot *DeepBot) onReset(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)
	user.setRounds(nil)

	replyMessage(ctx, "重置会话成功")
}
