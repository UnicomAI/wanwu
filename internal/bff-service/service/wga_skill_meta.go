package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/grpc/status"
)

func findGeneratedSkillFrontMatter(customSkillID string) (*util.SkillFrontMatter, error) {
	store, err := NewGeneralAgentSkillWorkspaceStore(customSkillID)
	if err != nil {
		return nil, err
	}

	workspaceDir := GetWgaWorkspaceThreadDir(store)
	if workspaceDir == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill workspace not found: %s", customSkillID))
	}
	if stat, err := os.Stat(workspaceDir); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill workspace not found: %s", customSkillID))
	} else if !stat.IsDir() {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill workspace is not a directory: %s", workspaceDir))
	}

	var skillDirs []string
	err = filepath.WalkDir(workspaceDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "tmp" || name == "input" || name == "output" {
			if path != workspaceDir {
				return filepath.SkipDir
			}
		}
		if path == workspaceDir {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, "SKILL.md")); err == nil {
			skillDirs = append(skillDirs, path)
			return filepath.SkipDir
		} else if !os.IsNotExist(err) {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to scan custom skill workspace %s: %v", workspaceDir, err))
	}

	switch len(skillDirs) {
	case 0:
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("generated skill not found in workspace: %s", customSkillID))
	case 1:
		return readGeneratedSkillFrontMatter(skillDirs[0])
	default:
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("expected exactly one generated skill in workspace %s, got %d", workspaceDir, len(skillDirs)))
	}
}

func readGeneratedSkillFrontMatter(skillDir string) (*util.SkillFrontMatter, error) {
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to read generated skill metadata %s: %v", skillDir, err))
	}
	fm, err := util.ParseSkillFrontMatter(string(data))
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to parse generated skill metadata skillDir=%s frontMatter=%s err=%v", skillDir, summarizeSkillFrontMatterForLog(string(data)), err))
	}
	return fm, nil
}

func summarizeSkillFrontMatterForLog(content string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return "<missing front matter>"
	}

	rest := content[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return "<unterminated front matter>"
	}

	frontMatter := strings.TrimSpace(rest[:endIdx])
	if frontMatter == "" {
		return "<empty front matter>"
	}

	frontMatter = strings.ReplaceAll(frontMatter, "\r\n", "\n")
	lines := strings.Split(frontMatter, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	summary := strings.Join(lines, " | ")
	runes := []rune(summary)
	if len(runes) > 512 {
		return string(runes[:512]) + "...(truncated)"
	}
	return summary
}

func formatGeneratedSkillMetaError(err error) string {
	if err == nil {
		return ""
	}

	st, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}

	var parts []string
	for _, detail := range st.Details() {
		statusDetail, ok := detail.(*errs.Status)
		if !ok {
			continue
		}
		if args := statusDetail.GetArgs(); len(args) > 0 {
			parts = append(parts, args...)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, " | ")
	}

	if msg := strings.TrimSpace(st.Message()); msg != "" {
		return msg
	}
	return err.Error()
}
