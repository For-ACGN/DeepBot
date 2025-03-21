package deepbot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wdvxdr1123/ZeroBot"
)

type txt2Image struct {
	Prompt    string `json:"prompt"`
	NegPrompt string `json:"negative_prompt,omitempty"`

	SamplerName string `json:"sampler_name,omitempty"`
	Scheduler   string `json:"scheduler,omitempty"`
	Steps       int    `json:"steps,omitempty"`

	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`

	BatchSize   int     `json:"batch_size,omitempty"`
	BatchCount  int     `json:"n_iter,omitempty"`
	DisCFGScale float64 `json:"distilled_cfg_scale,omitempty"`
	CFGScale    float64 `json:"cfg_scale,omitempty"`

	SendImages bool `json:"send_images"`
	SaveImages bool `json:"save_images"`

	Seed int64 `json:"seed"`
}

func (bot *DeepBot) onDrawImage(ctx *zero.Ctx) {
	if !bot.config.SDWebUI.Enabled {
		bot.sendText(ctx, "画图服务未启用")
		return
	}

	args := textToArgN(ctx.MessageString(), 2)
	if len(args) != 2 {
		bot.sendText(ctx, "非法参数格式")
		return
	}
	prompt := args[1]
	if prompt == " " || prompt == "" {
		bot.sendText(ctx, "非法参数格式")
		return
	}

	bot.sendRandomWait(ctx)
	img, err := bot.drawImage(prompt, 30, 1024, 1024)
	if err != nil {
		log.Println("failed to draw image:", err)
		return
	}
	sendImage(ctx, img)
}

func (bot *DeepBot) onDrawImageWithArgs(ctx *zero.Ctx) {
	if !bot.config.SDWebUI.Enabled {
		bot.sendText(ctx, "画图服务未启用")
		return
	}

	args := textToArgN(ctx.MessageString(), 4)
	if len(args) != 4 {
		bot.sendText(ctx, "非法参数格式")
		return
	}

	width, height, ok := parseResolution(args[1])
	if !ok {
		bot.sendText(ctx, "非法的分辨率参数")
		return
	}
	steps, err := strconv.Atoi(args[2])
	if err != nil || steps < 1 {
		bot.sendText(ctx, "非法的steps参数")
		return
	}
	prompt := args[3]
	if prompt == " " || prompt == "" {
		bot.sendText(ctx, "非法的prompt参数")
		return
	}

	bot.sendRandomWait(ctx)
	img, err := bot.drawImage(prompt, steps, width, height)
	if err != nil {
		log.Println("failed to draw image:", err)
		return
	}
	sendImage(ctx, img)
}

func (bot *DeepBot) sendRandomWait(ctx *zero.Ctx) {
	switch rand.IntN(3) {
	case 0:
		bot.sendText(ctx, "正在画图")
	case 1:
		bot.sendText(ctx, "正在画图ing")
	default:
		bot.sendText(ctx, "等待画图中")
	}
}

func (bot *DeepBot) drawImage(prompt string, steps, width, height int) ([]byte, error) {
	cfg := bot.config.SDWebUI

	timeout := time.Duration(cfg.Timeout) * time.Millisecond
	tr := http.Transport{}
	client := http.Client{
		Transport: &tr,
		Timeout:   timeout,
	}
	defer client.CloseIdleConnections()

	URL, err := url.JoinPath(cfg.URL, "/sdapi/v1/txt2img")
	if err != nil {
		return nil, err
	}
	var arg txt2Image
	if cfg.Config != "" {
		data, err := os.ReadFile(cfg.Config)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &arg)
		if err != nil {
			return nil, err
		}
	} else {
		arg = txt2Image{
			NegPrompt:   "",
			SamplerName: "Euler a",
			Scheduler:   "Normal",
			BatchSize:   1,
			BatchCount:  1,
			DisCFGScale: 3.5,
			CFGScale:    7,
		}
	}
	arg.Prompt = prompt
	arg.Steps = steps
	arg.Width = width
	arg.Height = height
	arg.SendImages = true
	arg.SaveImages = true
	arg.Seed = -1

	data, err := jsonEncode(arg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type result struct {
		Images     []string `json:"images"`
		Parameters any      `json:"parameters"`
		Info       string   `json:"info"`
	}
	var results result
	err = jsonDecode(data, &results)
	if err != nil {
		return nil, err
	}
	img, err := base64.StdEncoding.DecodeString(results.Images[0])
	if err != nil {
		return nil, err
	}
	return img, nil
}

func parseResolution(s string) (int, int, bool) {
	var res []string
	switch {
	case strings.Contains(s, "x"):
		res = strings.Split(s, "x")
	case strings.Contains(s, "*"):
		res = strings.Split(s, "*")
	default:
		return 0, 0, false
	}
	if len(res) != 2 {
		return 0, 0, false
	}
	width, err := strconv.Atoi(res[0])
	if err != nil {
		return 0, 0, false
	}
	height, err := strconv.Atoi(res[1])
	if err != nil {
		return 0, 0, false
	}
	if width < 1 || height < 1 {
		return 0, 0, false
	}
	return width, height, true
}
