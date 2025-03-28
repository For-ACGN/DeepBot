package deepbot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// ================================ 以下代码由DeepSeek-R1生成 ================================

type mdDetector struct {
	features []mdRule
	maxScore float64
}

type mdRule struct {
	name    string
	pattern *regexp.Regexp
	weight  float64
}

func newMDDetector() *mdDetector {
	features := []mdRule{
		// 标题: # 后跟空格和文本
		{"header", regexp.MustCompile(`(?m)^#{1,6}\s+.+$`), 2.0},
		// 列表项: */-/数字. 后跟空格
		{"list", regexp.MustCompile(`(?m)^(\*|-|\d+\.)\s+`), 1.0},
		// 链接/图片
		{"link", regexp.MustCompile(`\[.*?]\(.+?\)`), 2.0},
		// 粗体/斜体
		{"emphasis", regexp.MustCompile(`(\*\*.*?\*\*|\*.*?\*)`), 1.0},
		// 引用块
		{"blockquote", regexp.MustCompile(`(?m)^>\s+`), 1.0},
		// 水平线
		{"hr", regexp.MustCompile(`(?m)^-{3,}$|^_{3,}$|^\*{3,}$`), 1.0},
		// 表格
		{"table", regexp.MustCompile(`(?m)^\|.*\|$`), 3.0},
	}
	return &mdDetector{features: features}
}

func (md *mdDetector) Analyze(text string) float64 {
	lines := md.preprocess(text)
	score := 0.0
	inCodeBlock := false
	// 代码块检测需要特殊处理
	codeBlockRegex := regexp.MustCompile("(?s)```.*?```| {4}.*")
	codeBlocks := codeBlockRegex.FindAllString(text, -1)
	score += float64(len(codeBlocks)) * 3.0
	// 计算最大可能得分（按行数*2估算）
	md.maxScore = float64(len(lines)) * 2.0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// 跳过代码块内的内容
		if inCodeBlock {
			if strings.Contains(line, "```") {
				inCodeBlock = false
			}
			continue
		}
		if strings.Contains(line, "```") {
			inCodeBlock = true
			continue
		}
		// 计算其他特征
		for _, rule := range md.features {
			if matches := rule.pattern.FindAllString(line, -1); len(matches) > 0 {
				score += rule.weight * float64(len(matches))
			}
		}
	}
	return md.normalizeScore(score)
}

func (md *mdDetector) preprocess(text string) []string {
	stripped := regexp.MustCompile("(?s)```.*?```").ReplaceAllString(text, "")
	return strings.Split(stripped, "\n")
}

func (md *mdDetector) normalizeScore(score float64) float64 {
	if md.maxScore == 0 {
		return 0.0
	}
	probability := (score / md.maxScore) * 100
	if probability > 100 {
		return 100.0
	}
	return probability
}

// ====================================================================================

func isMarkdown(text string) bool {
	if strings.Count(text, "```") >= 2 {
		return true
	}
	detector := newMDDetector()
	score := detector.Analyze(text)
	fmt.Println("markdown score:", score)
	return score >= 10
}

func markdownToHTML(md string) string {
	// create Markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))
	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return string(markdown.Render(doc, renderer))
}
