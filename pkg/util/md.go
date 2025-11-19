package util

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Linkify,
		extension.TaskList,
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	))

// markdown->html
func Md2html(markdown []byte) (string, error) {
	var buf bytes.Buffer
	// 转换为 HTML [EN] Convert to HTML
	if err := md.Convert(markdown, &buf); err != nil {
		return "", err
	}
	// 输出 HTML [EN] Output HTML
	return buf.String(), nil
}
