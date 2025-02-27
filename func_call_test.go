package deepbot

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func TestOnFetchURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(chromePath),
	}
	output, err := onFetchURL(ctx, opts, "https://www.baidu.com/")
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
