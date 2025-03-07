package deepbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/wdvxdr1123/ZeroBot"
)

func (bot *DeepBot) onListConversation(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	dir, err := os.ReadDir(fmt.Sprintf("data/conversation/%d", user.id))
	if err != nil {
		log.Printf("failed to list conversation: %s\n", err)
		return
	}

	var list string
	for _, file := range dir {
		name := file.Name()
		list += strings.ReplaceAll(name, ".json", "") + " "
	}

	if list == "" {
		list = "会话列表为空"
	} else {
		list = "会话列表: " + list
	}

	bot.sendText(ctx, list)
}

func (bot *DeepBot) onSaveConversation(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	args := textToArgN(ctx.MessageString(), 2)
	if len(args) != 2 {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	name := args[1]
	if name == " " || name == "" {
		bot.sendText(ctx, "非法参数格式")
		return
	}

	rounds := user.getRounds()
	if len(rounds) == 0 {
		bot.sendText(ctx, "当前会话内容为空")
		return
	}

	output, err := jsonEncode(&rounds)
	if err != nil {
		log.Println("failed to encode conversation:", err)
		return
	}
	path := fmt.Sprintf("data/conversation/%d/%s.json", user.id, name)
	err = os.WriteFile(path, output, 0600)
	if err != nil {
		log.Println("failed to save conversation:", err)
		return
	}

	bot.sendText(ctx, "保存会话成功")
}

func (bot *DeepBot) onLoadConversation(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	args := textToArgN(ctx.MessageString(), 2)
	if len(args) != 2 {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	name := args[1]
	if name == " " || name == "" {
		bot.sendText(ctx, "非法参数格式")
		return
	}

	path := fmt.Sprintf("data/conversation/%d/%s.json", user.id, name)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("failed to read conversation:", err)
		return
	}
	var rounds []*round
	err = json.NewDecoder(bytes.NewReader(data)).Decode(&rounds)
	if err != nil {
		log.Println("failed to decode conversation:", err)
		return
	}

	user.setRounds(rounds)

	bot.sendText(ctx, "加载会话成功")
}

func (bot *DeepBot) onPreviewConversation(ctx *zero.Ctx) {

}

func (bot *DeepBot) onCopyConversation(ctx *zero.Ctx) {

}

func (bot *DeepBot) onDeleteConversation(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	args := textToArgN(ctx.MessageString(), 2)
	if len(args) != 2 {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	name := args[1]
	if name == " " || name == "" {
		bot.sendText(ctx, "非法参数格式")
		return
	}

	path := fmt.Sprintf("data/conversation/%d/%s.json", user.id, name)
	exists, err := isFileExists(path)
	if err != nil {
		log.Println("failed to check conversation:", err)
		return
	}
	if !exists {
		bot.sendText(ctx, "会话不存在")
		return
	}

	err = os.Remove(path)
	if err != nil {
		log.Println("failed to delete conversation:", err)
		return
	}

	bot.sendText(ctx, "删除会话成功")
}
