package deepbot

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/cohesion-org/deepseek-go"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const (
	fnGetTime  = "GetTime"
	fnFetchURL = "FetchURL"
	fnEvalGo   = "EvalGo"
)

const (
	usageGetTime  = "Usage_" + fnGetTime
	usageFetchURL = "Usage_" + fnFetchURL
	usageEvalGo   = "Usage_" + fnEvalGo
)

type toolArgument struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

var toolGetTime = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name:        fnGetTime,
		Description: "获取当前的日期以及时间，返回的时间字符串格式为RFC3339。",
	},
}

var toolFetchURL = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name: fnFetchURL,
		Description: "" +
			"使用浏览器去访问指定的URL，返回的结果是过滤后的可见文本内容," +
			"注意这个函数只会返回可见的文本内容，所以返回的内容格式会有点怪异，这是正常情况。" +
			"你只需要分析文本内容即可，不用在意这些内容里出现的类似让你登录、注册、使用前必读等操作。" +
			"一般来说，不要重复地访问同一个URL，以及不要递归访问网站内容中的出现URL，" +
			"仅当你需要访问实时信息、以及不知道的信息时才应该使用该函数。" +
			"禁止多次来回调用该工具函数，一轮对话中只允许使用3次该函数。",
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

func onFetchURL(ctx context.Context, opts []chromedp.ExecAllocatorOption, url string) (string, error) {
	fmt.Println("================FetchURL================")
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
