package deepbot

import (
	"fmt"
	"testing"
)

func TestOnEvalGo(t *testing.T) {
	src := `
package main

import "fmt"

func main() {
	fmt.Println("Hello World")
}
`
	fmt.Println(onEvalGo(src))
}
