package deepbot

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var testBot *DeepBot

func init() {
	cfg := &Config{}
	cfg.Render.ExecPath = `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
	cfg.Render.Width = 600
	cfg.Render.Height = 300
	testBot = NewDeepBot(cfg)
}

func TestMarkdownToHTML(t *testing.T) {
	defer func() { _ = os.RemoveAll(defaultDataDir) }()

	md, err := os.ReadFile("testdata/message.md")
	require.NoError(t, err)

	output, err := testBot.markdownToImage(string(md))
	require.NoError(t, err)

	err = os.WriteFile("testdata/render.jpg", output, 0600)
	require.NoError(t, err)
}

func TestRenderHelpDocument(t *testing.T) {
	defer func() { _ = os.RemoveAll(defaultDataDir) }()

	output, err := testBot.markdownToImage(helpMD)
	require.NoError(t, err)

	err = os.WriteFile("testdata/help.jpg", output, 0600)
	require.NoError(t, err)
}
