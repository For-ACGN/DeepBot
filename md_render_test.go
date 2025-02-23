package deepbot

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkdownToHTML(t *testing.T) {
	defer func() { _ = os.RemoveAll("data/chromedp") }()

	md, err := os.ReadFile("testdata/message.md")
	require.NoError(t, err)

	cfg := &Config{}
	cfg.Render.ExecPath = `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
	cfg.Render.Width = 600
	cfg.Render.Height = 300
	deepbot := NewDeepBot(cfg)

	output, err := deepbot.markdownToImage(string(md))
	require.NoError(t, err)

	err = os.WriteFile("testdata/render.jpg", output, 0600)
	require.NoError(t, err)
}
