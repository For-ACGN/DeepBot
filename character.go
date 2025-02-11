package deepbot

import (
	"fmt"
	"log"
	"os"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func (bot *DeepBot) onListCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	dir, err := os.ReadDir(fmt.Sprintf("data/characters/%d", user.id))
	if err != nil {
		log.Printf("failed to list character: %s\n", err)
		return
	}

	var list string
	for _, file := range dir {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if name == "current.cfg" {
			continue
		}
		list += strings.ReplaceAll(name, ".txt", "") + " "
	}

	if list == "" {
		list = "人设列表为空"
	} else {
		list = "当前人设列表: " + list
	}

	replyMessage(ctx, list)
}

func (bot *DeepBot) onCurCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	file := fmt.Sprintf("data/characters/%d/current.cfg", user.id)
	data, err := os.ReadFile(file)
	if err != nil {
		log.Printf("failed to read character config: %s\n", err)
		return
	}

	char := string(data)
	if char == "" {
		char = "当前无人设"
	} else {
		char = "当前人设: " + char
	}

	replyMessage(ctx, char)
}

func (bot *DeepBot) onGetCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	msg := ctx.MessageString()
	msg = strings.Replace(msg, "bot.查看人设 ", "", 1)

	section := strings.Split(msg, " ")
	if len(section) < 1 {
		replyMessage(ctx, "非法参数格式")
		return
	}
	name := section[0]
	if name == " " || name == "" {
		replyMessage(ctx, "非法参数格式")
		return
	}

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	content, err := os.ReadFile(file)
	if err != nil {
		log.Printf("failed to read character file: %s\n", err)
		return
	}

	char := string(content)
	if char == "" {
		char = "当前人设内容为空"
	}

	replyMessage(ctx, char)
}

func (bot *DeepBot) onClrCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	file := fmt.Sprintf("data/characters/%d/current.cfg", user.id)
	err := os.WriteFile(file, nil, 0644)
	if err != nil {
		log.Printf("failed to update character config: %s\n", err)
		return
	}

	user.setCharacter("")

	replyMessage(ctx, "清除人设成功")
}
