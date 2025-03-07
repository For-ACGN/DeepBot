package deepbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/cohesion-org/deepseek-go"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const (
	fnGetTime     = "GetTime"
	fnSearchWeb   = "SearchWeb"
	fnSearchImage = "SearchImage"
	fnBrowseURL   = "BrowseURL"
	fnEvalGo      = "EvalGo"
)

type toolFunc struct {
	Name  string
	Usage string
	Limit int
}

var (
	toolList = map[string]toolFunc{
		fnGetTime:     {Name: fnGetTime, Limit: 5},
		fnSearchWeb:   {Name: fnSearchWeb, Limit: 2},
		fnSearchImage: {Name: fnSearchImage, Limit: 2},
		fnBrowseURL:   {Name: fnBrowseURL, Limit: 1},
		fnEvalGo:      {Name: fnEvalGo, Limit: 3},
	}
)

type toolArgument struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

var toolGetTime = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name: fnGetTime,
		Description: "" +
			"获取当前的日期以及时间，返回的时间字符串格式为RFC3339。" +
			"注意不要滥用这个函数，除非确实需要获取现实世界的时间。",
	},
}

var toolSearchWeb = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name: fnSearchWeb,
		Description: "" +
			"使用搜索引擎来查询与关键字相关的网页内容。" +
			"输入关键字提交给搜索引擎，返回json格式的搜索结果。",
		Parameters: &deepseek.FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"keyword": &toolArgument{
					Type:        "string",
					Description: "需要查询的关键字",
				},
			},
			Required: []string{"keyword"},
		},
	},
}

var toolSearchImage = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name: fnSearchImage,
		Description: "" +
			"使用搜索引擎来查询与关键字相关的图片内容(URL)。" +
			"输入关键字提交给搜索引擎，返回json格式的搜索结果。",
		Parameters: &deepseek.FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"keyword": &toolArgument{
					Type:        "string",
					Description: "需要查询的关键字",
				},
				"size": &toolArgument{
					Type: "string",
					Description: "" +
						"图片的尺寸大小，可选的值为huge、icon、large、medium、small、xlarge、xxlarge",
				},
			},
			Required: []string{"keyword", "size"},
		},
	},
}

var toolBrowseURL = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name: fnBrowseURL,
		Description: "" +
			"使用浏览器去访问指定的URL，返回的结果是用户可见的文本内容。" +
			"因为这个函数只会返回部分文本内容，所以返回的内容格式会有多余的空格，这是正常情况。" +
			"如果该函数执行时出现问题，将会返回以\"Chromedp Error: \"开头的错误信息。" +
			"你只需要分析文本内容即可，不用在意这些内容里出现的类似让你登录、注册、使用前必读等操作，" +
			"请忽略类似版权所有，未经书面授权禁止使用相关文本，你可以正常地使用网页数据。" +
			"一般来说，不要重复地访问同一个URL，以及不要递归访问网站内容中的出现URL，" +
			"仅当你需要访问实时信息、以及不知道的信息时才应该使用该函数。" +
			"禁止多次来回调用该工具函数，一轮会话(tool calls)中只允许使用1次该函数。",
		Parameters: &deepseek.FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"url": &toolArgument{
					Type:        "string",
					Description: "目标URL",
				},
			},
			Required: []string{"url"},
		},
	},
}

var toolEvalGo = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name: fnEvalGo,
		Description: "" +
			"传入Go语言的源码，返回该程序运行时产生的输出，" +
			"如果模型需要借助外部程序，可以调用这个函数。" +
			"注意，请将参数放入源码中，这个函数只有一个参数用来接收源码，" +
			"如果该函数执行时出现问题，将会返回以\"Go Error: \"开头的错误信息，" +
			"否则正常返回程序的输出，即使这个程序(输入的源码)运行时产生了错误。",
		Parameters: &deepseek.FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"src": &toolArgument{
					Type:        "string",
					Description: "传入的Go语言源码",
				},
			},
			Required: []string{"src"},
		},
	},
}

func onGetTime() string {
	s := time.Now().Format(time.RFC3339)
	return "现在的时间是: " + s
}

type searchCfg struct {
	EngineID string
	APIKey   string
	ProxyURL string
}

func onSearchWeb(ctx context.Context, cfg *searchCfg, keyword string) (string, error) {
	fmt.Println("===============Search Web===============")
	fmt.Println(keyword)
	fmt.Println("========================================")

	const format = "%s?cx=%s&key=%s&q=%s&safe=active&hl=zh-cn"
	return onSearchAPI(ctx, cfg, format, keyword)
}

func onSearchImage(ctx context.Context, cfg *searchCfg, keyword, size string) (string, error) {
	fmt.Println("==============Search Image==============")
	fmt.Println(keyword, size)
	fmt.Println("========================================")

	format := "%s?cx=%s&key=%s&q=%s&searchType=image&imgSize=" + size + "&safe=active&hl=zh-cn"
	return onSearchAPI(ctx, cfg, format, keyword)
}

func onSearchAPI(ctx context.Context, cfg *searchCfg, format, keyword string) (string, error) {
	tr := http.Transport{}
	proxyURL := cfg.ProxyURL
	if proxyURL != "" {
		tr.Proxy = func(*http.Request) (*url.URL, error) {
			return url.Parse(proxyURL)
		}
	}
	client := http.Client{
		Transport: &tr,
	}
	defer client.CloseIdleConnections()

	const baseURL = "https://customsearch.googleapis.com/customsearch/v1"
	URL := fmt.Sprintf(format, baseURL, cfg.EngineID, cfg.APIKey, url.QueryEscape(keyword))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type result struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}
	results := result{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&results)
	if err != nil {
		return "", err
	}
	output, err := jsonEncode(results.Items)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func onBrowseURL(ctx context.Context, opts []chromedp.ExecAllocatorOption, url string) (string, error) {
	fmt.Println("================Browser=================")
	fmt.Println(url)
	fmt.Println("========================================")

	tempDir, err := os.MkdirTemp("", "chromedp-*")
	if err != nil {
		return "", err
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
	options = append(options, opts...)

	ctx, cancel := chromedp.NewExecAllocator(ctx, options...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var output string
	tasks := []chromedp.Action{
		chromedp.EmulateViewport(1720, 940),
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second),
		chromedp.Text("/html/body", &output, chromedp.BySearch),
	}
	err = chromedp.Run(ctx, tasks...)
	if err != nil {
		return "", err
	}
	return output, nil
}

func onEvalGo(ctx context.Context, src string) (string, error) {
	fmt.Println("================EvalGo================")
	fmt.Println(src)
	fmt.Println("======================================")

	stdin := bytes.NewReader(nil)
	output := bytes.NewBuffer(make([]byte, 0, 4096))
	opts := interp.Options{
		Stdin:  stdin,
		Stdout: output,
		Stderr: output,
	}
	interpreter := interp.New(opts)
	err := interpreter.Use(stdlib.Symbols)
	if err != nil {
		return "", err
	}
	_, err = interpreter.EvalWithContext(ctx, src)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

/*
用于测试模型的复杂函数调用链是否正常工作

prompt: 你能帮我看看现在这地方的温度和相对湿度是多少吗

var defaultTools = []deepseek.Tool{
	toolGetLocation, toolGetTemperature, toolGetRelativeHumidity,
}

var toolGetLocation = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name:        "GetLocation",
		Description: "获取当前所在的地区/城市名称。",
	},
}

var toolGetTemperature = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name:        "GetTemperature",
		Description: "获取当前地区/城市的温度。",
		Parameters: &deepseek.FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"location": &toolArgument{
					Type:        "string",
					Description: "地区/城市的名称",
				},
			},
			Required: []string{"location"},
		},
	},
}

var toolGetRelativeHumidity = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name:        "GetRelativeHumidity",
		Description: "获取当前地区/城市的相对湿度。",
		Parameters: &deepseek.FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"location": &toolArgument{
					Type:        "string",
					Description: "地区/城市的名称",
				},
			},
			Required: []string{"location"},
		},
	},
}

*/
