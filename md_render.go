package deepbot

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const defaultDataDir = "data/chromedp"

func (bot *DeepBot) markdownToImage(md string) ([]byte, error) {
	// create Markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))
	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	output := markdown.Render(doc, renderer)

	// deploy a http server for headless browser
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer func() { _ = listener.Close() }()
	randomName := strconv.Itoa(int(rand.Uint32()))
	serveMux := http.ServeMux{}
	serveMux.HandleFunc("/"+randomName, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(output)
	})
	server := http.Server{
		Handler: &serveMux,
	}
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Close() }()
	targetURL := fmt.Sprintf("http://%s/%s", listener.Addr(), randomName)

	// start headless browser to render it
	options := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,

		// After Puppeteer's default behavior.
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
	}
	cfg := bot.config.Render
	if cfg.ExecPath != "" {
		options = append(options, chromedp.ExecPath(cfg.ExecPath))
	}
	dataDir := cfg.DataDir
	if dataDir == "" {
		dataDir = defaultDataDir
	}
	dataDir, err = filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	options = append(options, chromedp.UserDataDir(dataDir))

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	var image []byte
	tasks := []chromedp.Action{
		chromedp.EmulateViewport(cfg.Width, cfg.Height),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(time.Second),
		chromedp.FullScreenshot(&image, 100),
		// chromedp.Sleep(time.Minute),
	}
	err = chromedp.Run(ctx, tasks...)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func isMarkdown(text string) bool {
	return true
}
