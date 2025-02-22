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

type toolArgument struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

var defaultTools = []deepseek.Tool{toolGetTime, toolEvalGo}

var toolGetTime = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name:        "GetTime",
		Description: "获取当前的日期以及时间，返回的时间字符串格式为RFC3339。",
	},
}

var toolEvalGo = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name: "EvalGo",
		Description: "传入Go语言的源码，返回该程序运行时产生的输出，" +
			"如果模型需要借助外部程序，可以调用这个函数。" +
			"注意，请将参数放入源码中，这个函数只有一个参数用来接收源码，" +
			"如果EvalGo运行有问题，将会返回以\"Go Error: \"开头的错误信息，" +
			"否则正常返回程序的输出，即使这个程序运行时产生了错误。",
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

func onEvalGo(ctx context.Context, src string) (string, error) {
	fmt.Println("================Eval================")
	fmt.Println(src)
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

func onFetchURL(ctx context.Context, url string) (string, error) {
	fmt.Println("fetch:", url)

	tempDir, err := os.MkdirTemp("", "chromedp-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`),
		chromedp.UserDataDir(tempDir),
	)
	ctx, cancel := chromedp.NewExecAllocator(ctx, options...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	// var output string
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
		// chromedp.TextContent(".h", &output, chromedp.NodeVisible),
	}
	err = chromedp.Run(ctx, tasks...)
	if err != nil {
		return "", err
	}
	return "", nil
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
