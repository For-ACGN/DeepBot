package deepbot

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/wdvxdr1123/ZeroBot"
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

	bot.replyMessage(ctx, list)
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

	bot.replyMessage(ctx, char)
}

func (bot *DeepBot) onGetCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	msg := textToArgN(ctx.MessageString(), 2)
	if len(msg) != 2 {
		bot.replyMessage(ctx, "非法参数格式")
		return
	}
	name := msg[1]
	if name == " " || name == "" {
		bot.replyMessage(ctx, "非法参数格式")
		return
	}

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	content, err := os.ReadFile(file)
	if err != nil {
		log.Printf("failed to read character file: %s\n", err)
		bot.replyMessage(ctx, "人设不存在")
		return
	}

	char := string(content)
	if char == "" {
		char = "当前人设内容为空"
	}

	bot.replyMessage(ctx, char)
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

	bot.replyMessage(ctx, "清除人设成功")
}

func (bot *DeepBot) onSetCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	msg := textToArgN(ctx.MessageString(), 2)
	if len(msg) != 2 {
		bot.replyMessage(ctx, "非法参数格式")
		return
	}
	name := msg[1]
	if name == " " || name == "" {
		bot.replyMessage(ctx, "非法参数格式")
		return
	}

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	content, err := os.ReadFile(file)
	if err != nil {
		log.Printf("failed to read character file: %s\n", err)
		bot.replyMessage(ctx, "人设不存在")
		return
	}
	file = fmt.Sprintf("data/characters/%d/current.cfg", user.id)
	err = os.WriteFile(file, []byte(name), 0644)
	if err != nil {
		log.Printf("failed to update character config: %s\n", err)
		return
	}

	user.setCharacter(string(content))

	bot.replyMessage(ctx, "选择人设成功")
}

func (bot *DeepBot) onAddCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	msg := textToArgN(ctx.MessageString(), 3)
	if len(msg) != 3 {
		bot.replyMessage(ctx, "非法参数格式")
		return
	}
	name := msg[1]
	if name == " " || name == "" {
		bot.replyMessage(ctx, "非法参数格式")
		return
	}
	if len(name) > 30 {
		bot.replyMessage(ctx, "人设名称太长")
		return
	}

	content := msg[2]
	if content == " " || content == "" {
		bot.replyMessage(ctx, "人设内容为空")
		return
	}

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		log.Printf("failed to save character file: %s\n", err)
		return
	}

	bot.replyMessage(ctx, "添加人设成功")
}

func (bot *DeepBot) onDelCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	msg := textToArgN(ctx.MessageString(), 2)
	if len(msg) != 2 {
		bot.replyMessage(ctx, "非法参数格式")
		return
	}
	name := msg[1]
	if name == " " || name == "" {
		bot.replyMessage(ctx, "非法参数格式")
		return
	}

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	err := os.Remove(file)
	if err != nil {
		log.Printf("failed to remove character file: %s\n", err)
		bot.replyMessage(ctx, "人设不存在")
		return
	}

	file = fmt.Sprintf("data/characters/%d/current.cfg", user.id)
	char, err := os.ReadFile(file)
	if err != nil {
		log.Printf("failed to read current character name: %s\n", err)
		return
	}
	if string(char) == name {
		err = os.WriteFile(file, nil, 0644)
		if err != nil {
			log.Printf("failed to update character config: %s\n", err)
			return
		}
	}

	bot.replyMessage(ctx, "删除人设成功")
}
