package deepbot

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func TestOnSearchWeb(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg := &searchCfg{
		EngineID: os.Getenv("GS_ENGINE_ID"),
		APIKey:   os.Getenv("GS_API_KEY"),
		ProxyURL: os.Getenv("HTTP_PROXY"),
	}
	output, err := onSearchWeb(ctx, cfg, "Golang")
	require.NoError(t, err)
	fmt.Println(output)
}

func TestOnSearchImage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg := &searchCfg{
		EngineID: os.Getenv("GS_ENGINE_ID"),
		APIKey:   os.Getenv("GS_API_KEY"),
		ProxyURL: os.Getenv("HTTP_PROXY"),
	}
	output, err := onSearchImage(ctx, cfg, "Golang", "medium")
	require.NoError(t, err)
	fmt.Println(output)
}

func TestOnBrowseURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(chromePath),
	}
	output, err := onBrowseURL(ctx, opts, "https://www.baidu.com/")
	require.NoError(t, err)
	fmt.Println(output)
}

func TestOnEvalGo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	src := `
package main

import "fmt"

func main() {
	fmt.Println("Hello World!")
}
`
	output, err := onEvalGo(ctx, src)
	require.NoError(t, err)
	fmt.Println(output)
}
