package deepbot

import (
	"context"
	"embed"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

//go:embed asset
var asset embed.FS

//go:embed template/renderer.html
var renderer string

func (bot *DeepBot) markdownToImage(content string) ([]byte, error) {
	output := markdownToHTML(content)
	return bot.htmlToImage(output)
}

func (bot *DeepBot) htmlToImage(content string) ([]byte, error) {
	// insert code about js and css for renderer code block
	document := strings.ReplaceAll(renderer, "{{data}}", content)
	fmt.Println(document)

	// deploy a http server for headless browser
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer func() { _ = listener.Close() }()
	randomName := fmt.Sprintf("%d.html", rand.Uint())
	serveMux := http.ServeMux{}
	serveMux.HandleFunc("/"+randomName, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(document))
	})
	serveMux.Handle("/asset/", http.FileServerFS(asset))
	server := http.Server{
		Handler: &serveMux,
	}
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Close() }()
	targetURL := fmt.Sprintf("http://%s/%s", listener.Addr(), randomName)

	// start headless browser to renderer it
	tempDir, err := os.MkdirTemp("", "chromedp-*")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

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

		chromedp.UserDataDir(tempDir),
	}
	options = append(options, bot.getChromedpOptions()...)

	cfg := bot.config.Renderer
	timeout := time.Duration(cfg.Timeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ctx, cancel = chromedp.NewExecAllocator(ctx, options...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var image []byte
	tasks := []chromedp.Action{
		chromedp.EmulateViewport(cfg.Width, cfg.Height, chromedp.EmulateScale(4)),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(time.Second),
		chromedp.FullScreenshot(&image, 100),
	}
	err = chromedp.Run(ctx, tasks...)
	if err != nil {
		return nil, err
	}
	return image, nil
}
