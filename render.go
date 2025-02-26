package deepbot

import (
	"context"
	"embed"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

//go:embed asset
var asset embed.FS

func (bot *DeepBot) markdownToImage(md string) ([]byte, error) {
	output := markdownToHTML(md)
	return bot.htmlToImage(output)
}

func (bot *DeepBot) htmlToImage(content string) ([]byte, error) {
	// insert code about js and css for render code block
	document := `
<!DOCTYPE html>
<html>

<head>
  <meta charset="UTF-8">

  <link rel="stylesheet" href="asset/github-dark.min.css">
  <link rel="stylesheet" href="asset/katex.min.css">

  <style>
    @font-face {
        font-family: 'Noto Sans SC';
        src: url('asset/font/NotoSansSC-VariableFont_wght.ttf') format('truetype');
        font-style: normal;
    }

    @font-face {
        font-family: 'Roboto Mono';
        src: url('asset/font/RobotoMono-VariableFont_wght.ttf') format('truetype');
        font-style: normal;
    }

    @font-face {
        font-family: 'NotoColorEmoji';
        src: url('asset/font/NotoColorEmoji-Regular.ttf') format('truetype');
        font-style: normal;
    }

    * {
      font-family: 'Noto Sans SC', 'NotoColorEmoji';
    }

    pre {
      max-width: 100%%;
      overflow-x: auto;
      white-space: pre-wrap;
      word-wrap: break-word;
    }

    code {
      font-family: 'Roboto Mono', 'Noto Sans SC', 'NotoColorEmoji';
      background: #3C3D3E;
      padding: 3px;
      border-radius: 4px;
      white-space: pre-wrap;
      word-wrap: break-word;
    }

    table {
      table-layout: auto;
      border-collapse: collapse;
    }

    th, td {
      padding: 8px;
      text-align: left;
      border: 1px solid black;
    }

    li {
      padding: 4px;
    }

    tr:nth-child(even) {
      background-color: #1D1F20;
    }

    tr:nth-child(odd) {
      background-color: #262C36;
    }

    body {
      padding: 16px;
    }
  </style>
</head>

<body>
%s
</body>

<script src="asset/dark-reader.min.js"></script>
<script src="asset/highlight.min.js"></script>
<script src="asset/katex.min.js"></script>
<script src="asset/auto-render.min.js" onload="renderMathInElement(document.body);"></script>

<script>
    DarkReader.enable({
        brightness: 100,
        contrast:   95,
        sepia:      0
    });
    hljs.highlightAll();
</script>

</html>`
	document = fmt.Sprintf(document, content)
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

	// start headless browser to render it
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
	cfg := bot.config.Render
	if cfg.ExecPath != "" {
		options = append(options, chromedp.ExecPath(cfg.ExecPath))
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
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
