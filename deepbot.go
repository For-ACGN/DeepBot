package deepbot

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type ChatRequest = deepseek.ChatCompletionRequest
type ChatResponse = deepseek.ChatCompletionResponse
type ChatMessage = deepseek.ChatCompletionMessage

type Config struct {
	GroupID []int64 `toml:"group_id"`

	DeepSeek struct {
		APIKey  string `toml:"api_key"`
		BaseURL string `toml:"base_url"`
		Timeout int    `toml:"timeout"`
	} `toml:"deepseek"`

	OneBot struct {
		WSClient struct {
			Enabled bool   `toml:"enabled"`
			URL     string `toml:"url"`
			Token   string `toml:"token"`
		} `toml:"ws_client"`

		WSServer struct {
			Enabled bool   `toml:"enabled"`
			URL     string `toml:"url"`
			Token   string `toml:"token"`
		} `toml:"ws_server"`
	} `toml:"onebot"`
}

type DeepBot struct {
	config *Config
	client *deepseek.Client
	tools  []deepseek.Tool

	users   map[int64]*user
	usersMu sync.Mutex
}

func NewDeepBot(config *Config) *DeepBot {
	client := deepseek.NewClient(config.DeepSeek.APIKey)
	baseURL := config.DeepSeek.BaseURL
	if baseURL != "" {
		client.BaseURL = baseURL
	}
	timeout := config.DeepSeek.Timeout
	if timeout != 0 {
		client.Timeout = time.Duration(timeout) * time.Millisecond
	}
	bot := DeepBot{
		config: config,
		client: client,
		tools:  defaultTools,
		users:  make(map[int64]*user),
	}
	// register message handler
	groupID := config.GroupID
	filter := func(ctx *zero.Ctx) bool {
		if ctx.Event.GroupID == 0 {
			return true
		}
		for i := 0; i < len(groupID); i++ {
			if ctx.Event.GroupID == groupID[i] {
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
	zero.OnCommand("deep.人设列表", filter).SetBlock(true).Handle(bot.onListCharacter)
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

func (bot *DeepBot) Run() {
	var drivers []zero.Driver
	onebot := bot.config.OneBot
	client := onebot.WSClient
	if client.Enabled {
		drivers = append(drivers, driver.NewWebSocketClient(client.URL, client.Token))
	}
	server := onebot.WSServer
	if server.Enabled {
		drivers = append(drivers, driver.NewWebSocketServer(1, server.URL, server.Token))
	}
	cfg := zero.Config{
		NickName: []string{"deepbot"},
		Driver:   drivers,
	}
	zero.RunAndBlock(&cfg, nil)
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
	help += "deep.列出人设  列出所有人设，可用人设列表代替\n"
	help += "deep.当前人设  查看当前人设\n"
	help += "deep.清除人设  清除当前人设\n"
	help += "deep.查看人设  查看人设内容: [角色A]\n"
	help += "deep.选择人设  选择一个人设: [角色A]\n"
	help += "deep.添加人设  添加一个人设: [角色A] [人设内容]\n"
	help += "deep.删除人设  删除一个人设: [角色A]\n"
	replyMessage(ctx, help)
}

func replyMessage(ctx *zero.Ctx, msg string) {
	// wait random time
	time.Sleep(time.Duration(500 + rand.IntN(2000)))
	// process private chat
	if ctx.Event.GroupID == 0 {
		ctx.Send(message.Text(msg))
		return
	}
	// random send message with at
	if ctx.Event.IsToMe && rand.IntN(3) == 0 {
		array := message.Message{}
		array = append(array, message.At(ctx.Event.UserID))
		array = append(array, message.Text(" "+msg))
		ctx.Send(array)
		return
	}
	ctx.Send(message.Text(msg))
}
