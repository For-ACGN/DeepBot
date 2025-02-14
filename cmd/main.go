package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/For-ACGN/DeepBot"
	"github.com/pelletier/go-toml/v2"
)

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "cfg", "config.toml", "set the configuration file path")
	flag.Parse()
}

func main() {
	data, err := os.ReadFile(cfgPath)
	checkError(err)

	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var config deepbot.Config
	err = decoder.Decode(&config)
	checkError(err)

	bot := deepbot.NewDeepBot(&config)
	bot.Run()
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
