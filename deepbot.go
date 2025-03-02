package deepbot

import (
	"embed"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
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

	Chromedp struct {
		ExecPath string `toml:"exec_path"`
		ProxyURL string `toml:"proxy_url"`
	} `toml:"chromedp"`

	Renderer struct {
		Enabled bool  `toml:"enabled"`
		Width   int64 `toml:"width"`
		Height  int64 `toml:"height"`
		Timeout int   `toml:"timeout"`
	} `toml:"renderer"`

	Emoticon struct {
		Rate int `toml:"rate"`
	} `toml:"emoticon"`

	SearchAPI struct {
		Enabled bool `toml:"enabled"`
		Timeout int  `toml:"timeout"`
	} `toml:"search_api"`

	Browser struct {
		Enabled bool `toml:"enabled"`
		Timeout int  `toml:"timeout"`
	} `toml:"browser"`

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
	if config.SearchAPI.Enabled {
		tools = append(tools, toolSearchWeb)
		tools = append(tools, toolSearchImage)
	}
	if config.Browser.Enabled {
		tools = append(tools, toolBrowseURL)
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
	zero.OnCommand("chatx ", filter).SetBlock(true).Handle(bot.onChatX)
	zero.OnCommand("ai ", filter).SetBlock(true).Handle(bot.onReasoner)
	zero.OnCommand("aix ", filter).SetBlock(true).Handle(bot.onReasoning)
	zero.OnCommand("coder ", filter).SetBlock(true).Handle(bot.onCoder)
	zero.OnCommand("coderx ", filter).SetBlock(true).Handle(bot.onCoderX)
	zero.OnCommand("deep.当前模型", filter).SetBlock(true).Handle(bot.onGetModel)
	zero.OnCommand("deep.设置模型 ", filter).SetBlock(true).Handle(bot.onSetModel)
	zero.OnCommand("deep.启用函数", filter).SetBlock(true).Handle(bot.onEnableToolCall)
	zero.OnCommand("deep.禁用函数", filter).SetBlock(true).Handle(bot.onDisableToolCall)
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
	zero.OnCommand("deep.读取心情", filter).SetBlock(true).Handle(bot.onGetMood)
	zero.OnCommand("deep.当前心情", filter).SetBlock(true).Handle(bot.onUpdateMood)
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

func (bot *DeepBot) getChromedpOptions() []chromedp.ExecAllocatorOption {
	var options []chromedp.ExecAllocatorOption
	cfg := bot.config.Chromedp
	if cfg.ExecPath != "" {
		options = append(options, chromedp.ExecPath(cfg.ExecPath))
	}
	if cfg.ProxyURL != "" {
		options = append(options, chromedp.ProxyServer(cfg.ProxyURL))
	}
	return options
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

//go:embed template/help.md
var helpMD string

func (bot *DeepBot) onHelp(ctx *zero.Ctx) {
	bot.reply(ctx, nil, helpMD)
}

// process command about chat.
func (bot *DeepBot) reply(ctx *zero.Ctx, user *user, msg string) {
	defer bot.postProcess(ctx, user, msg)
	if !bot.config.Renderer.Enabled {
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
	bot.sendText(ctx, msg)
}

// process command about get status.
func (bot *DeepBot) sendText(ctx *zero.Ctx, msg string) {
	if len(msg) < 1024 {
		sendText(ctx, msg)
		return
	}
	// renderer long text to image
	sections := strings.Split(msg, "\n")
	builder := strings.Builder{}
	builder.Grow(len(msg))
	for _, section := range sections {
		builder.WriteString("<div>")
		builder.WriteString(section)
		builder.WriteString("</div>")
	}
	msg = builder.String()
	img, err := bot.htmlToImage(msg)
	if err != nil {
		log.Println(err)
		return
	}
	sendImage(ctx, img)
}

func (bot *DeepBot) sendImage(ctx *zero.Ctx, path string) {
	fmt.Println("===============reply image==============")
	fmt.Println(path)
	fmt.Println("========================================")

	img, err := os.ReadFile(path)
	if err != nil {
		log.Println("failed to load image:", err)
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
