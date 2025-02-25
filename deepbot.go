package deepbot

import (
	"embed"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// just for prevent [import _ "embed"] :)
var _ embed.FS

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

	Render struct {
		Enabled  bool   `toml:"enabled"`
		Width    int64  `toml:"width"`
		Height   int64  `toml:"height"`
		ExecPath string `toml:"exec_path"`
		DataDir  string `toml:"data_dir"`
	} `toml:"md_render"`

	FetchURL struct {
		Enabled  bool   `toml:"enabled"`
		Timeout  int    `toml:"timeout"`
		ProxyURL string `toml:"proxy_url"`
		ExecPath string `toml:"exec_path"`
	} `toml:"fetch_url"`

	EvalGo struct {
		Enabled bool `toml:"enabled"`
		Timeout int  `toml:"timeout"`
	} `toml:"eval_go"`
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
	// build tools from config
	var tools []deepseek.Tool
	tools = append(tools, toolGetTime)
	if config.FetchURL.Enabled {
		tools = append(tools, toolFetchURL)
	}
	if config.EvalGo.Enabled {
		tools = append(tools, toolEvalGo)
	}
	bot := DeepBot{
		config: config,
		client: client,
		tools:  tools,
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
	zero.OnCommand("air ", filter).SetBlock(true).Handle(bot.onReasoning)
	zero.OnCommand("deep.当前模型", filter).SetBlock(true).Handle(bot.onGetModel)
	zero.OnCommand("deep.设置模型 ", filter).SetBlock(true).Handle(bot.onSetModel)
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
	zero.OnCommand("help", filter).SetBlock(true).Handle(bot.onHelp)
	zero.OnCommand("deep.help", filter).SetBlock(true).Handle(bot.onHelp)
	zero.OnCommand("deep.帮助文档", filter).SetBlock(true).Handle(bot.onHelp)
	zero.OnCommand("deep.帮助信息", filter).SetBlock(true).Handle(bot.onHelp)
	zero.OnMessage(filter).SetBlock(true).Handle(bot.onMessage)
	zero.OnNotice(filter).SetBlock(true).Handle(bot.onPoke)
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

//go:embed help.md
var helpMD string

func (bot *DeepBot) onHelp(ctx *zero.Ctx) {
	bot.replyMessage(ctx, helpMD)
}

func (bot *DeepBot) onPoke(ctx *zero.Ctx) {
	event := ctx.Event
	if !event.IsToMe {
		return
	}
	if event.NoticeType != "notify" || event.SubType != "poke" {
		return
	}

	switch rand.IntN(2) {
	case 0:
		sendText(ctx, "别戳了")
	case 1:
		sendText(ctx, "再戳我就要爆了")
	}
}

func (bot *DeepBot) replyMessage(ctx *zero.Ctx, msg string) {
	if !bot.config.Render.Enabled {
		sendText(ctx, msg)
		return
	}
	if isMarkdown(msg) {
		img, err := bot.markdownToImage(msg)
		if err != nil {
			log.Println(err)
			return
		}
		sendImage(ctx, img)
		return
	}
	if len(msg) < 2048 {
		sendText(ctx, msg)
		return
	}
	img, err := bot.htmlToImage(msg)
	if err != nil {
		log.Println(err)
		return
	}
	sendImage(ctx, img)
}

func sendText(ctx *zero.Ctx, msg string) {
	// wait random time before send
	time.Sleep(time.Duration(500+rand.IntN(2000)) * time.Millisecond)
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

func sendImage(ctx *zero.Ctx, img []byte) {
	// process private chat
	if ctx.Event.GroupID == 0 {
		ctx.Send(message.ImageBytes(img))
		return
	}
	// random send message with at
	if ctx.Event.IsToMe && rand.IntN(3) == 0 {
		array := message.Message{}
		array = append(array, message.At(ctx.Event.UserID))
		array = append(array, message.ImageBytes(img))
		ctx.Send(array)
		return
	}
	ctx.Send(message.ImageBytes(img))
}
