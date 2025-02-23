package deepbot

import (
	"fmt"
	"regexp"
	"strings"
)

// ================================ 以下代码由DeepSeek-R1生成 ================================

type markdownDetector struct {
	features []featureRule
	maxScore float64
}

type featureRule struct {
	pattern *regexp.Regexp
	weight  float64
	name    string
}

func newMarkdownDetector() *markdownDetector {
	features := []featureRule{
		// 标题: # 后跟空格和文本
		{regexp.MustCompile(`(?m)^#{1,6}\s+.+$`), 2.0, "header"},
		// 列表项: */-/数字. 后跟空格
		{regexp.MustCompile(`(?m)^(\*|-|\d+\.)\s+`), 1.0, "list"},
		// 链接/图片
		{regexp.MustCompile(`\[.*?\]\(.+?\)`), 2.0, "link"},
		// 粗体/斜体
		{regexp.MustCompile(`(\*\*.*?\*\*|\*.*?\*)`), 1.0, "emphasis"},
		// 引用块
		{regexp.MustCompile(`(?m)^>\s+`), 1.0, "blockquote"},
		// 水平线
		{regexp.MustCompile(`(?m)^-{3,}$|^_{3,}$|^\*{3,}$`), 1.0, "hr"},
		// 表格
		{regexp.MustCompile(`(?m)^\|.*\|$`), 3.0, "table"},
	}
	return &markdownDetector{features: features}
}

func (md *markdownDetector) Analyze(text string) float64 {
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

func (md *markdownDetector) preprocess(text string) []string {
	stripped := regexp.MustCompile("(?s)```.*?```").ReplaceAllString(text, "")
	return strings.Split(stripped, "\n")
}

func (md *markdownDetector) normalizeScore(score float64) float64 {
	if md.maxScore == 0 {
		return 0.0
	}
	probability := (score / md.maxScore) * 100
	if probability > 100 {
		return 100.0
	}
	return probability
}

func isMarkdown(text string) bool {
	detector := newMarkdownDetector()
	score := detector.Analyze(text)
	fmt.Println("markdown score:", score)
	return score >= 10
}
