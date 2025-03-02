package deepbot

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var testBot *DeepBot

func init() {
	cfg := &Config{}
	cfg.Chromedp.ExecPath = chromePath
	cfg.Renderer.Width = 600
	cfg.Renderer.Height = 300
	cfg.Renderer.Timeout = 15000
	testBot = NewDeepBot(cfg)
}

func TestMarkdownToImage(t *testing.T) {
	md, err := os.ReadFile("testdata/message.md")
	require.NoError(t, err)

	output, err := testBot.markdownToImage(string(md))
	require.NoError(t, err)

	err = os.WriteFile("testdata/markdown.jpg", output, 0600)
	require.NoError(t, err)
}

func TestHTMLToImage(t *testing.T) {
	data, err := os.ReadFile("testdata/message.html")
	require.NoError(t, err)

	output, err := testBot.htmlToImage(string(data))
	require.NoError(t, err)

	err = os.WriteFile("testdata/html.jpg", output, 0600)
	require.NoError(t, err)
}

func TestRendererHelpDocument(t *testing.T) {
	output, err := testBot.markdownToImage(helpMD)
	require.NoError(t, err)

	err = os.WriteFile("testdata/help.jpg", output, 0600)
	require.NoError(t, err)
}
