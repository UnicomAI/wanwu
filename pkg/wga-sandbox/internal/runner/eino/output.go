package eino

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"context"

	"github.com/UnicomAI/wanwu/pkg/log"
)

// scrubDirs 是在 copyOutput 阶段需要从宿主 OutputDir 中删除的目录名（沙箱内的中间产物 / 输入）。
var scrubDirs = map[string]bool{
	"skills": true,
	"input":  true,
	"tmp":    true,
}

// copyOutput 把沙箱内的输出目录复制到宿主 OutputDir，并清理隐藏文件与中间目录。
// 沙箱内的 output 子目录会被「拍平」到宿主 OutputDir 根部。
//
// 替换策略：先复制到临时目录并清理，再用 backup+swap 原子替换 OutputDir。
// 这样可以确保沙箱中已删除/重命名的文件不会在 OutputDir 中残留，
// 同时保证任何失败都不会破坏 OutputDir 的现有内容（workspace 链不断）。
func (r *Runner) copyOutput(ctx context.Context) error {
	log.Infof("%s copyOutput start", r.logPrefix)

	outputDir := r.req.OutputDir
	tmpDir := filepath.Join(filepath.Dir(outputDir), ".swaptmp-"+filepath.Base(outputDir))
	bakDir := filepath.Join(filepath.Dir(outputDir), ".swapbak-"+filepath.Base(outputDir))

	// 清理上次失败留下的临时/备份目录
	_ = os.RemoveAll(tmpDir)
	_ = os.RemoveAll(bakDir)

	// 1. 复制沙箱内容到临时目录（失败 → OutputDir 不变，链安全）
	if err := r.sb.CopyFromSandbox(ctx, tmpDir); err != nil {
		log.Errorf("%s copyOutput CopyFromSandbox failed: %v", r.logPrefix, err)
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to copy output from workspace: %w", err)
	}
	// 确保 tmpDir 存在（CopyFromSandbox 在沙箱只有隐藏文件时会跳过）
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// 2. 清理临时目录：删除隐藏文件、中间目录（skills/input/tmp）、拍平 output/
	if err := r.scrubOutputDir(tmpDir); err != nil {
		log.Errorf("%s copyOutput scrub failed: %v", r.logPrefix, err)
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to scrub output directory: %w", err)
	}

	// 3. 原子替换：备份 OutputDir → 安装 tmpDir → 清理备份
	//    （失败 → 从备份恢复，链安全）
	if err := swapDirWithBackup(tmpDir, outputDir, bakDir); err != nil {
		log.Errorf("%s copyOutput swap failed: %v", r.logPrefix, err)
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to swap output directory: %w", err)
	}

	log.Infof("%s copyOutput completed", r.logPrefix)
	return nil
}

// scrubOutputDir 清理输出目录：删除隐藏文件、中间目录（skills/input/tmp）拍平 output/ 子目录。
func (r *Runner) scrubOutputDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Errorf("%s copyOutput ReadDir failed: %v", r.logPrefix, err)
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())

		switch {
		case strings.HasPrefix(entry.Name(), "."):
			log.Infof("%s copyOutput removing hidden file: %s", r.logPrefix, entry.Name())
			if err := os.RemoveAll(entryPath); err != nil {
				log.Errorf("%s copyOutput remove hidden file failed: %s, err: %v", r.logPrefix, entry.Name(), err)
				return fmt.Errorf("failed to remove hidden file %s: %w", entry.Name(), err)
			}

		case entry.IsDir() && scrubDirs[entry.Name()]:
			log.Infof("%s copyOutput removing %s dir", r.logPrefix, entry.Name())
			if err := os.RemoveAll(entryPath); err != nil {
				log.Errorf("%s copyOutput remove %s dir failed: %v", r.logPrefix, entry.Name(), err)
				return fmt.Errorf("failed to remove %s directory: %w", entry.Name(), err)
			}

		case entry.IsDir() && entry.Name() == "output":
			log.Infof("%s copyOutput flattening output subdir", r.logPrefix)
			if err := flattenDir(entryPath, dir); err != nil {
				log.Errorf("%s copyOutput flatten failed: %v", r.logPrefix, err)
				return fmt.Errorf("failed to flatten output directory: %w", err)
			}
		}
	}

	log.Infof("%s copyOutput completed", r.logPrefix)
	return nil
}

// flattenDir 把 src 目录中的所有顶层条目移动到 dst，然后删除空的 src 目录。
func flattenDir(src, dst string) error {
	log.Infof("[flattenDir] start src=%s dst=%s", src, dst)

	subEntries, err := os.ReadDir(src)
	if err != nil {
		log.Errorf("[flattenDir] ReadDir failed: %v", err)
		return fmt.Errorf("failed to read dir %s: %w", src, err)
	}

	for _, sub := range subEntries {
		srcPath := filepath.Join(src, sub.Name())
		dstPath := filepath.Join(dst, sub.Name())
		if err := os.Rename(srcPath, dstPath); err != nil {
			log.Errorf("[flattenDir] move %s failed: %v", sub.Name(), err)
			return fmt.Errorf("failed to move %s: %w", sub.Name(), err)
		}
	}

	if err := os.Remove(src); err != nil {
		log.Errorf("[flattenDir] remove src dir failed: %v", err)
		return err
	}

	log.Infof("[flattenDir] completed")
	return nil
}

// swapDirWithBackup 用备份目录实现原子替换 targetDir 为 tmpDir。
//
// 流程：
//  1. Rename targetDir → bakDir（备份当前内容）
//  2. Rename tmpDir → targetDir（安装新内容）
//  3. 如果步骤 2 失败，恢复 bakDir → targetDir（尽力恢复）
//  4. 删除 bakDir（清理）
//
// 三个目录必须在同一文件系统上，os.Rename 才能工作。
// 如果 targetDir 不存在，函数返回错误，不修改 tmpDir。
//
// 安全保证：任意步骤失败时，targetDir 要么包含原始内容（从备份恢复），
// 要么从未被修改。tmpDir 在失败时不会丢失。
func swapDirWithBackup(tmpDir, targetDir, bakDir string) error {
	if err := os.Rename(targetDir, bakDir); err != nil {
		return fmt.Errorf("failed to backup %s to %s: %w", targetDir, bakDir, err)
	}

	if err := os.Rename(tmpDir, targetDir); err != nil {
		_ = os.Rename(bakDir, targetDir)
		return fmt.Errorf("failed to install %s to %s: %w", tmpDir, targetDir, err)
	}

	_ = os.RemoveAll(bakDir)
	return nil
}
