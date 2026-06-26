package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type SkillFrontMatter struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	ClosedSource bool   `yaml:"closed_source"`
}

// IsSkillClosedSource 从 markdown 内容判断 skill 是否闭源。
// 不含 frontmatter 或解析失败时返回 false（向后兼容：老数据视为开源）。
func IsSkillClosedSource(markdown string) bool {
	markdown = strings.TrimSpace(markdown)
	if !strings.HasPrefix(markdown, "---") {
		return false
	}
	rest := markdown[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return false
	}
	frontMatterStr := strings.TrimSpace(rest[:endIdx])
	var fm struct {
		ClosedSource bool `yaml:"closed_source"`
	}
	if err := yaml.Unmarshal([]byte(frontMatterStr), &fm); err != nil {
		// yaml 解析失败时（如 description 含 C1 控制字符），fallback 到逐行字符串匹配
		return isClosedSourceByString(frontMatterStr)
	}
	return fm.ClosedSource
}

// isClosedSourceByString 从 frontmatter 字符串中逐行查找 closed_source 标志。
func isClosedSourceByString(frontMatterStr string) bool {
	for _, line := range strings.Split(frontMatterStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "closed_source:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "closed_source:"))
			return val == "true" || val == "True" || val == "TRUE"
		}
	}
	return false
}

// InjectClosedSourceFlag 在 markdown frontmatter 中注入或移除 closed_source 标记。
// 不含 frontmatter 时原样返回。保留原有字段和正文，仅增删 closed_source 行。
func InjectClosedSourceFlag(markdown string, closed bool) string {
	markdown = strings.TrimSpace(markdown)
	if !strings.HasPrefix(markdown, "---") {
		return markdown
	}
	rest := markdown[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return markdown
	}
	frontMatterStr := rest[:endIdx]
	body := rest[endIdx+4:] // "\n---" 之后的内容

	lines := strings.Split(frontMatterStr, "\n")
	var result []string
	found := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "closed_source:") {
			found = true
			if closed {
				result = append(result, "closed_source: true")
			}
		} else {
			result = append(result, line)
		}
	}
	if closed && !found {
		result = append(result, "closed_source: true")
	}
	return "---" + strings.Join(result, "\n") + "\n---" + body
}

// skillNameRegex allows kebab-case names with Unicode letters (e.g. Chinese characters),
// digits, and hyphens. Must start with a letter, end with a letter or digit,
// and not contain consecutive hyphens.
var skillNameRegex = regexp.MustCompile(`^[\p{L}][\p{L}\p{N}]*(-[\p{L}\p{N}]+)*$`)

// ParseSkillFrontMatter 解析技能的Markdown内容，提取FrontMatter
func ParseSkillFrontMatter(content string) (*SkillFrontMatter, error) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("SKILL.md file must start with front matter delimiters")
	}

	rest := content[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return nil, fmt.Errorf("SKILL.md file must end with front matter delimiters")
	}

	frontMatterStr := strings.TrimSpace(rest[:endIdx])
	frontMatterStr = strings.ToValidUTF8(frontMatterStr, "")

	var fm SkillFrontMatter
	if err := yaml.Unmarshal([]byte(frontMatterStr), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %v", err)
	}
	if fm.Name == "" || fm.Description == "" {
		return nil, fmt.Errorf("SKILL.md file must contain both name and description in front matter")
	}
	if !skillNameRegex.MatchString(fm.Name) {
		return nil, fmt.Errorf("SKILL.md file name must be in kebab-case")
	}

	return &fm, nil
}

// ExtractSkillMarkdownFromZip 从ZIP文件中提取SKILL.md文件的内容，返回完整markdown内容、名称、描述
func ExtractSkillMarkdownFromZip(zipData []byte) (string, *SkillFrontMatter, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", nil, fmt.Errorf("failed to read zip file: %v", err)
	}

	var skillMdFile *zip.File
	for _, file := range reader.File {
		fileName := filepath.Base(file.Name)
		if fileName == "SKILL.md" {
			skillMdFile = file
			break
		}
	}

	if skillMdFile == nil {
		return "", nil, fmt.Errorf("SKILL.md file not found in the zip archive")
	}

	rc, err := skillMdFile.Open()
	if err != nil {
		return "", nil, fmt.Errorf("failed to open SKILL.md file: %v", err)
	}
	defer func() { _ = rc.Close() }()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read SKILL.md file: %v", err)
	}

	markdownContent := string(content)
	fm, err := ParseSkillFrontMatter(markdownContent)
	if err != nil {
		return "", nil, err
	}

	return markdownContent, fm, nil
}
