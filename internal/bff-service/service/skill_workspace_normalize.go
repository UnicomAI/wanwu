package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	git_util "github.com/UnicomAI/wanwu/pkg/git-util"
)

// applyWgaWorkspaceDirSkillPolicy 对 Skill 生成场景应用工作区目录策略。
func applyWgaWorkspaceDirSkillPolicy(dirs *WgaWorkspaceDirs) (*WgaWorkspaceDirs, error) {
	if dirs == nil || dirs.OutputDir == "" {
		return dirs, nil
	}

	skillDir := filepath.Join(dirs.OutputDir, generalAgentWorkspaceSkillDirName)
	info, err := os.Stat(skillDir)
	if os.IsNotExist(err) {
		return dirs, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat skill workspace dir %s: %w", skillDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("skill workspace path is not a directory: %s", skillDir)
	}

	return &WgaWorkspaceDirs{
		InputDir:  filepath.Clean(skillDir),
		OutputDir: skillDir,
	}, nil
}

// normalizeCustomSkillWorkspaceNestedSkill 归并工作区内多余嵌套的 skill 目录。
func normalizeCustomSkillWorkspaceNestedSkill(customSkillID string) error {
	customSkillID = strings.TrimSpace(customSkillID)
	if customSkillID == "" {
		return nil
	}

	skillRoot, err := getSkillDir(customSkillID)
	if err != nil {
		return err
	}
	if skillRoot == "" {
		return nil
	}

	repo := git_util.Open(skillRoot)
	mu := repo.GetMutex()
	mu.Lock()
	defer mu.Unlock()

	workspaceDir := filepath.Join(skillRoot, generalAgentWorkspaceSkillDirName)
	nestedDir := filepath.Join(workspaceDir, generalAgentWorkspaceSkillDirName)
	info, err := os.Lstat(nestedDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat nested skill workspace dir %s: %w", nestedDir, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("nested skill workspace dir is symlink: %s", nestedDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("nested skill workspace path is not a directory: %s", nestedDir)
	}

	if err := mergeCustomSkillNestedDirContents(nestedDir, workspaceDir); err != nil {
		return err
	}
	if err := os.Remove(nestedDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove nested skill workspace dir %s: %w", nestedDir, err)
	}
	return nil
}

// mergeCustomSkillNestedDirContents 将嵌套目录内容合并到目标目录。
func mergeCustomSkillNestedDirContents(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read nested skill workspace dir %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		if entry.Name() == ".git" {
			if err := os.RemoveAll(filepath.Join(srcDir, entry.Name())); err != nil {
				return fmt.Errorf("remove nested git metadata: %w", err)
			}
			continue
		}
		if err := mergeCustomSkillNestedEntry(srcDir, dstDir, entry); err != nil {
			return err
		}
	}
	return nil
}

// mergeCustomSkillNestedEntry 合并单个嵌套文件或目录条目。
func mergeCustomSkillNestedEntry(srcDir, dstDir string, entry os.DirEntry) error {
	srcPath := filepath.Join(srcDir, entry.Name())
	dstPath := filepath.Join(dstDir, entry.Name())

	srcInfo, err := os.Lstat(srcPath)
	if err != nil {
		return fmt.Errorf("stat nested skill workspace entry %s: %w", srcPath, err)
	}
	if filepath.Clean(dstPath) == filepath.Clean(srcDir) {
		if !srcInfo.IsDir() || srcInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refuse to merge nested skill workspace entry into itself: %s", srcPath)
		}
		if err := mergeCustomSkillNestedDirContents(srcPath, dstDir); err != nil {
			return err
		}
		if err := os.Remove(srcPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove merged nested skill workspace dir %s: %w", srcPath, err)
		}
		return nil
	}

	dstInfo, err := os.Lstat(dstPath)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("create target parent dir %s: %w", filepath.Dir(dstPath), err)
		}
		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("move nested skill workspace entry %s to %s: %w", srcPath, dstPath, err)
		}
		return nil
	case err != nil:
		return fmt.Errorf("stat target skill workspace entry %s: %w", dstPath, err)
	}

	srcIsDir := srcInfo.IsDir() && srcInfo.Mode()&os.ModeSymlink == 0
	dstIsDir := dstInfo.IsDir() && dstInfo.Mode()&os.ModeSymlink == 0
	if srcIsDir && dstIsDir {
		if err := mergeCustomSkillNestedDirContents(srcPath, dstPath); err != nil {
			return err
		}
		if err := os.Remove(srcPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove merged nested skill workspace dir %s: %w", srcPath, err)
		}
		return nil
	}

	if err := os.RemoveAll(dstPath); err != nil {
		return fmt.Errorf("remove conflicting skill workspace entry %s: %w", dstPath, err)
	}
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("move nested skill workspace entry %s to %s: %w", srcPath, dstPath, err)
	}
	return nil
}
