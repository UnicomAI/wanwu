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
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
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
