package deepbot

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsMarkdown(t *testing.T) {

}

func TestMarkdownToHTML(t *testing.T) {
	md := `
Go语言（也称为Golang）是一种开源的编程语言，由Google的Robert Griesemer、
Rob Pike和Ken Thompson于2007年设计，2009年正式发布。Go语言的设计目标是提供一种
简单、高效、可靠的编程语言，特别适合现代多核和网络应用的开发。

### 主要特点

1. **简洁易学**：Go语言的语法简洁明了，去除了许多其他语言中的复杂特性，如继承和泛型，使得代码易于阅读和维护。

2. **高效性能**：Go语言编译后的代码运行速度快，接近C/C++的性能，适合高性能应用。

3. **并发支持**：Go语言内置了强大的并发支持，通过goroutine和channel，可以轻松实现并发编程。

4. **垃圾回收**：Go语言具有自动垃圾回收机制，减少了内存管理的负担。

5. **跨平台**：Go语言支持多种操作系统和架构，编写的代码可以在不同的平台上编译和运行。

6. **标准库丰富**：Go语言的标准库非常丰富，涵盖了网络、文本处理、加密、数据库等多个领域。

### 应用场景

- **Web开发**：Go语言的高效性能和并发支持使其成为Web服务器和后端服务的理想选择。
- **云计算**：许多云原生工具和平台（如Kubernetes、Docker）都是用Go语言编写的。
- **微服务**：Go语言的轻量级和高效性使其非常适合构建微服务架构。
- **系统编程**：Go语言可以用于编写操作系统工具、网络工具等系统级应用。

### 总结

Go语言以其简洁、高效和强大的并发支持，成为了现代软件开发中的热门选择。
无论是Web开发、云计算还是系统编程，Go语言都能提供高效的解决方案。
`
	cfg := &Config{}
	cfg.Render.ExecPath = `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
	cfg.Render.Width = 600
	cfg.Render.Height = 900
	deepbot := NewDeepBot(cfg)

	output, err := deepbot.markdownToImage(md)
	require.NoError(t, err)

	err = os.WriteFile("testdata/render.jpg", output, 0600)
	require.NoError(t, err)
}
