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
		list = "人设列表: " + list
	}

	bot.sendText(ctx, list)
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

	bot.sendText(ctx, char)
}

func (bot *DeepBot) onGetCharacter(ctx *zero.Ctx) {
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

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	content, err := os.ReadFile(file)
	if err != nil {
		log.Printf("failed to read character file: %s\n", err)
		bot.sendText(ctx, "人设不存在")
		return
	}

	output := string(content)
	if output == "" {
		output = "当前人设内容为空"
	}

	file = fmt.Sprintf("data/characters/%d/%s.tpl", user.id, name)
	prompt, err := os.ReadFile(file)
	if err == nil && string(prompt) != "" {
		output += "\n================prompt================\n"
		output += string(prompt)
	}

	bot.sendText(ctx, output)
}

func (bot *DeepBot) onClrCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	file := fmt.Sprintf("data/characters/%d/current.cfg", user.id)
	err := os.WriteFile(file, nil, 0600)
	if err != nil {
		log.Printf("failed to update character config: %s\n", err)
		return
	}

	user.setCharacter("", "", "")

	bot.sendText(ctx, "清除人设成功")
}

func (bot *DeepBot) onSelectCharacter(ctx *zero.Ctx) {
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

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	content, err := os.ReadFile(file)
	if err != nil {
		log.Printf("failed to read character file: %s\n", err)
		bot.sendText(ctx, "人设不存在")
		return
	}
	file = fmt.Sprintf("data/characters/%d/%s.tpl", user.id, name)
	prompt, _ := os.ReadFile(file)

	file = fmt.Sprintf("data/characters/%d/current.cfg", user.id)
	err = os.WriteFile(file, []byte(name), 0600)
	if err != nil {
		log.Printf("failed to update character config: %s\n", err)
		return
	}

	user.setCharacter(name, string(content), string(prompt))

	bot.sendText(ctx, "选择人设成功")
}

func (bot *DeepBot) onSetCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	args := textToArgN(ctx.MessageString(), 3)
	if len(args) != 3 {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	name := args[1]
	if name == " " || name == "" {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	if len(name) > 30 {
		bot.sendText(ctx, "人设名称太长")
		return
	}
	prompt := args[2]
	if prompt == " " || prompt == "" {
		bot.sendText(ctx, "提示词模板为空")
		return
	}

	file := fmt.Sprintf("data/characters/%d/%s.tpl", user.id, name)
	err := os.WriteFile(file, []byte(prompt), 0600)
	if err != nil {
		log.Printf("failed to save character prompt file: %s\n", err)
		return
	}

	bot.sendText(ctx, "添加提示词模板成功")
}

func (bot *DeepBot) onAddCharacter(ctx *zero.Ctx) {
	user := bot.getUser(ctx.Event.UserID)

	args := textToArgN(ctx.MessageString(), 3)
	if len(args) != 3 {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	name := args[1]
	if name == " " || name == "" {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	if len(name) > 30 {
		bot.sendText(ctx, "人设名称太长")
		return
	}

	content := args[2]
	if content == " " || content == "" {
		bot.sendText(ctx, "人设内容为空")
		return
	}

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	err := os.WriteFile(file, []byte(content), 0600)
	if err != nil {
		log.Printf("failed to save character file: %s\n", err)
		return
	}

	bot.sendText(ctx, "添加人设成功")
}

func (bot *DeepBot) onDelCharacter(ctx *zero.Ctx) {
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

	file := fmt.Sprintf("data/characters/%d/%s.txt", user.id, name)
	err := os.Remove(file)
	if err != nil {
		log.Printf("failed to remove character file: %s\n", err)
		bot.sendText(ctx, "人设不存在")
		return
	}

	file = fmt.Sprintf("data/characters/%d/%s.tpl", user.id, name)
	_ = os.Remove(file)

	file = fmt.Sprintf("data/characters/%d/current.cfg", user.id)
	char, err := os.ReadFile(file)
	if err != nil {
		log.Printf("failed to read current character name: %s\n", err)
		return
	}
	if string(char) == name {
		err = os.WriteFile(file, nil, 0600)
		if err != nil {
			log.Printf("failed to update character config: %s\n", err)
			return
		}
	}

	bot.sendText(ctx, "删除人设成功")
}
