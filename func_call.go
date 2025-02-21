package deepbot

import (
	"bytes"
	"context"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var defaultTools = []deepseek.Tool{toolGetTime, toolEvalGo}

var toolGetTime = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name:        "GetTime",
		Description: "获取当前的日期以及时间，返回的时间字符串格式为RFC3339。",
	},
}

type goSrc struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

var toolEvalGo = deepseek.Tool{
	Type: "function",
	Function: deepseek.Function{
		Name: "EvalGo",
		Description: "传入Go语言的源码，返回该程序运行时产生的输出，" +
			"如果模型需要借助外部程序，可以调用这个函数。" +
			"注意，请将参数放入源码中，这个函数只有一个参数用来接收源码，" +
			"如果运行有问题，将会返回以\"Go Error: \"开头的错误信息，否则正常返回程序的输出。",
		Parameters: &deepseek.FunctionParameters{
			Type: "object",
			Properties: map[string]interface{}{
				"src": &goSrc{
					Type:        "string",
					Description: "传入的Go语言源码",
				},
			},
			Required: []string{"src"},
		},
	},
}

func onGetTime() string {
	return time.Now().Format(time.RFC3339)
}

func onEvalGo(src string) string {
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
		return "Go Error: " + err.Error()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_, err = interpreter.EvalWithContext(ctx, src)
	if err != nil {
		return "Go Error: " + err.Error()
	}
	return output.String()
}
