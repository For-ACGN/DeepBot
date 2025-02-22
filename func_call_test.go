package deepbot

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnEvalGo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	src := `
package main

import "fmt"

func main() {
	fmt.Println("Hello World")
}
`
	output, err := onEvalGo(ctx, src)
	require.NoError(t, err)
	fmt.Println(output)
}

func TestOnFetchURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := onFetchURL(ctx, "https://www.baidu.com/")
	require.NoError(t, err)
	fmt.Println(output)
}
