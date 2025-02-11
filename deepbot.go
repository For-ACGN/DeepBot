package deepbot

import (
	"sync"

	"github.com/cohesion-org/deepseek-go"
	"github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Config struct {
	APIKey  string
	BaseURL string
	BotCfg  *zero.Config
	GroupID []int64
}

type DeepBot struct {
	config *zero.Config
	client *deepseek.Client

	users   map[int64]*user
	usersMu sync.Mutex
}

func NewDeepBot(cfg *Config) *DeepBot {
	client := deepseek.NewClient(cfg.APIKey)
	if cfg.BaseURL != "" {
		client.BaseURL = cfg.BaseURL
	}
	bot := DeepBot{
		config: cfg.BotCfg,
		client: client,
		users:  make(map[int64]*user),
	}
	filter := func(ctx *zero.Ctx) bool {
		if ctx.Event.GroupID == 0 {
			return true
		}
		for i := 0; i < len(cfg.GroupID); i++ {
			if ctx.Event.GroupID == cfg.GroupID[i] {
				return true
			}
		}
		return false
	}
	zero.OnCommand("chat ", filter).SetBlock(true).Handle(bot.onChat)
	zero.OnCommand("coder ", filter).SetBlock(true).Handle(bot.onCoder)
	zero.OnCommand("ai ", filter).SetBlock(true).Handle(bot.onReasoner)
	zero.OnCommand("deep.reset", filter).SetBlock(true).Handle(bot.onReset)
	zero.OnCommand("deep.重置", filter).SetBlock(true).Handle(bot.onReset)
	zero.OnCommand("deep.重置会话", filter).SetBlock(true).Handle(bot.onReset)
	zero.OnCommand("deep.列出人设", filter).SetBlock(true).Handle(bot.onListCharacter)
	zero.OnCommand("deep.当前人设", filter).SetBlock(true).Handle(bot.onCurCharacter)
	zero.OnCommand("deep.清除人设", filter).SetBlock(true).Handle(bot.onClrCharacter)
	zero.OnCommand("deep.查看人设 ", filter).SetBlock(true).Handle(bot.onGetCharacter)
	zero.OnCommand("deep.选择人设 ", filter).SetBlock(true).Handle(bot.onSetCharacter)
	zero.OnCommand("deep.添加人设 ", filter).SetBlock(true).Handle(bot.onAddCharacter)
	zero.OnCommand("deep.删除人设 ", filter).SetBlock(true).Handle(bot.onDelCharacter)
	zero.OnCommand("deep.设置模型 ", filter).SetBlock(true).Handle(bot.onSetModel)
	zero.OnCommand("deep.help", filter).SetBlock(true).Handle(bot.onHelp)
	zero.OnMessage(filter).SetBlock(true).Handle(bot.onMessage)
	return &bot
}

func (bot *DeepBot) getUser(uid int64) *user {
	bot.usersMu.Lock()
	defer bot.usersMu.Unlock()
	user := bot.users[uid]
	if user == nil {
		user = newUser(uid)
		bot.users[uid] = user
	}
	return user
}

func (bot *DeepBot) onHelp(ctx *zero.Ctx) {
	var help string
	if ctx.Event.GroupID != 0 {
		help += "\n"
	}
	help += "[at]    使用当前模型进行对话\n"
	help += "chat    使用deepseek-chat模型，适合通用对话\n"
	help += "coder   使用deepseek-coder模型，适合编程\n"
	help += "ai      使用deepseek-r1模型，带有推理功能\n"
	help += "deep.设置模型  设置当前模型: [r1、chat、coder]"
	help += "deep.重置会话  重置当前对话上下文，可用reset、重置代替\n"
	help += "deep.列出人设  列出所有人设\n"
	help += "deep.当前人设  查看当前人设\n"
	help += "deep.清除人设  清除当前人设\n"
	help += "deep.查看人设  查看人设内容: [角色A]\n"
	help += "deep.选择人设  选择一个人设: [角色A]\n"
	help += "deep.添加人设  添加一个人设: [角色A] [人设内容]\n"
	help += "deep.删除人设  删除一个人设: [角色A]\n"
	replyMessage(ctx, help)
}

func replyMessage(ctx *zero.Ctx, msg string) {
	if ctx.Event.GroupID == 0 {
		ctx.Send(message.Text(msg))
		return
	}
	// array := message.Message{}
	// array = append(array, message.At(ctx.Event.UserID))
	// array = append(array, message.Text(" "+msg))
	// ctx.Send(array)
	ctx.Send(message.Text(msg))
}

func (bot *DeepBot) Run() {
	zero.RunAndBlock(bot.config, nil)
}
