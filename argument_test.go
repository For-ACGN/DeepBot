package deepbot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTextToArgN(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		args := textToArgN("deep.选择人设 角色", 2)
		require.Equal(t, "deep.选择人设", args[0])
		require.Equal(t, "角色", args[1])
	})

	t.Run("long tail", func(t *testing.T) {
		args := textToArgN("deep.添加人设 \"角色\" 人设内容 人设内容尾部", 3)
		require.Equal(t, "deep.添加人设", args[0])
		require.Equal(t, "角色", args[1])
		require.Equal(t, "人设内容 人设内容尾部", args[2])
	})
}
