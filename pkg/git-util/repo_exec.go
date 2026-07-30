package git_util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	path_util "github.com/UnicomAI/wanwu/pkg/path-util"
)

const (
	defaultTimeout         = 30 * time.Second
	maxGitTextOutputBytes  = 4 * 1024 * 1024
	maxGitFileOutputBytes  = 1 * 1024 * 1024
	maxGitArchiveSizeBytes = 20 * 1024 * 1024
)

var (
	muMap sync.Map
	// commitRefRE 是针对命令注入的净化白名单：所有来自外部输入的 commit/ref
	// 在进入 exec.CommandContext 的 argv 之前都必须经 ValidateCommitRef 校验。
	// 仅放行 HEAD、十六进制 hash(7-40 位 SHA-1 或 64 位 SHA-256) 及 ~N/^N 后缀，
	// 从源头拒绝以 - 开头的参数注入(如 --exec=)和 shell 元字符。
	commitRefRE = regexp.MustCompile(`^(HEAD|[0-9a-fA-F]{7,40}|[0-9a-fA-F]{64})([~^][0-9]*)*$`)
	tagNameRE   = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
)

// sanitizeGitArgs 是 git 命令执行出口的兜底过滤器，在所有 exec.CommandContext
// 调用前对 argv 逐项检查。本包以 argv 数组方式传参、不经 shell，故 shell 元字符
// 本身不构成注入；此过滤器用于兜底拦截漏校验路径中混入的危险控制字符(\0 及除
// \t/\n/\r 之外的 ASCII 控制字符、0x7f DEL)。放行 \t/\n/\r 是为了与 bff 层
// GitCommitReq.Check() 的控制字符白名单保持一致——后者明确支持多行 commit message，
// git -m 也接受多行信息；二者语义对齐可避免合法多行 message 在执行出口被误杀。
// 它与 commitRefRE 互补：commitRefRE 在入口约束 commit 位置不能变成选项，
// 此处约束任意位置不能含危险控制字符。
func sanitizeGitArgs(args []string) error {
	for _, a := range args {
		for i := 0; i < len(a); i++ {
			c := a[i]
			if c == '\t' || c == '\n' || c == '\r' {
				continue
			}
			if c < 0x20 || c == 0x7f {
				return fmt.Errorf("arg contains control character: %q", a)
			}
		}
	}
	return nil
}

type limitedBuffer struct {
	buf      bytes.Buffer
	limit    int
	overflow bool
}

// Write 写入受大小限制的缓冲区。
func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		return len(p), nil
	}
	remaining := b.limit - b.buf.Len()
	if remaining > 0 {
		if len(p) <= remaining {
			_, _ = b.buf.Write(p)
		} else {
			_, _ = b.buf.Write(p[:remaining])
			b.overflow = true
		}
	} else {
		b.overflow = true
	}
	return len(p), nil
}

// bytes 返回缓冲区内容，并在超限时返回错误。
func (b *limitedBuffer) bytes() ([]byte, error) {
	if b.overflow {
		return b.buf.Bytes(), fmt.Errorf("git output exceeded %d bytes", b.limit)
	}
	return b.buf.Bytes(), nil
}

// getMu 获取仓库目录对应的全局互斥锁。
func getMu(dir string) *sync.Mutex {
	mu, _ := muMap.LoadOrStore(dir, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// GetMutex 返回指定仓库目录对应的互斥锁。
func GetMutex(dir string) *sync.Mutex {
	return getMu(dir)
}

// gitEnv 构造执行 Git 命令时使用的环境变量。
func gitEnv(dir string) []string {
	safeDir, err := filepath.Abs(dir)
	if err != nil {
		safeDir = dir
	}
	return append(os.Environ(),
		"GIT_CONFIG_COUNT=2",
		"GIT_CONFIG_KEY_0=safe.directory",
		"GIT_CONFIG_VALUE_0="+safeDir,
		"GIT_CONFIG_KEY_1=core.quotePath",
		"GIT_CONFIG_VALUE_1=false",
	)
}

// commandContext 创建带默认取消逻辑的命令上下文。
func commandContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// runGit 使用默认超时时间执行 Git 命令。
func runGit(dir string, args ...string) ([]byte, error) {
	return runGitWithTimeout(dir, defaultTimeout, args...)
}

// runGitWithTimeout 使用指定超时时间执行 Git 命令。
func runGitWithTimeout(dir string, timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := commandContext(timeout)
	defer cancel()
	return runGitCombined(ctx, dir, maxGitTextOutputBytes, args...)
}

// runGitCombined 执行 Git 命令并限制合并输出大小。
func runGitCombined(ctx context.Context, dir string, limit int, args ...string) ([]byte, error) {
	if err := sanitizeGitArgs(args); err != nil {
		return nil, fmt.Errorf("git args rejected: %w", err)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv(dir)

	stdout := &limitedBuffer{limit: limit}
	stderr := &limitedBuffer{limit: maxGitTextOutputBytes}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	data, limitErr := stdout.bytes()
	if limitErr != nil {
		return data, limitErr
	}
	if err != nil {
		errOut, stderrLimitErr := stderr.bytes()
		if stderrLimitErr != nil {
			return data, fmt.Errorf("%w: %v", err, stderrLimitErr)
		}
		msg := strings.TrimSpace(string(errOut))
		if msg == "" {
			msg = strings.TrimSpace(string(data))
		}
		return data, fmt.Errorf("%w: %s", err, msg)
	}
	return data, nil
}

// runGitStdout 执行 Git 命令并返回 stdout 文本。
func runGitStdout(dir string, args ...string) (string, error) {
	out, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// runGitStdoutBytes 执行 Git 命令并限制 stdout 字节数。
func runGitStdoutBytes(dir string, limit int, args ...string) ([]byte, error) {
	if err := sanitizeGitArgs(args); err != nil {
		return nil, fmt.Errorf("git args rejected: %w", err)
	}
	ctx, cancel := commandContext(defaultTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv(dir)

	stdout := &limitedBuffer{limit: limit}
	stderr := &limitedBuffer{limit: maxGitTextOutputBytes}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	out, limitErr := stdout.bytes()
	if limitErr != nil {
		return out, limitErr
	}
	if err != nil {
		errOut, _ := stderr.bytes()
		return out, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(errOut)))
	}
	return out, nil
}

// ValidateRelPath 校验并清理仓库内相对路径。
func ValidateRelPath(p string, allowEmpty bool) (string, error) {
	return path_util.CleanRelPath(p, allowEmpty)
}

// ValidateCommitRef 校验 commit 引用格式，作为命令注入净化的入口白名单。
// ref 为空时视为由调用方使用默认值(HEAD 等)而放行；非空则必须匹配 commitRefRE，
// 否则拒绝，防止恶意输入进入 git 命令的 argv。
func ValidateCommitRef(ref string) error {
	if ref == "" {
		return nil
	}
	if !commitRefRE.MatchString(ref) {
		return fmt.Errorf("invalid commit ref: %s", ref)
	}
	return nil
}

// validateCommitRef 校验 commit 引用格式。
func validateCommitRef(ref string) error {
	return ValidateCommitRef(ref)
}

// ValidateTagName 校验 tag 名称格式。
func ValidateTagName(tagName string) error {
	if !tagNameRE.MatchString(tagName) {
		return fmt.Errorf("invalid tag name: %s", tagName)
	}
	return nil
}

// validateTreeish 校验 tree-ish 引用格式。
func validateTreeish(treeish string) error {
	if treeish == "" {
		return errors.New("treeish is required")
	}
	if ValidateCommitRef(treeish) == nil {
		return nil
	}
	if ValidateTagName(treeish) == nil {
		return nil
	}
	return fmt.Errorf("invalid treeish: %s", treeish)
}

// validateCommitRange 校验提交范围两端的引用格式。
func validateCommitRange(fromCommit, toCommit string) error {
	if err := validateCommitRef(fromCommit); err != nil {
		return fmt.Errorf("invalid fromCommit: %w", err)
	}
	if err := validateCommitRef(toCommit); err != nil {
		return fmt.Errorf("invalid toCommit: %w", err)
	}
	return nil
}

// cleanSubDir 清理可为空的仓库子目录路径。
func cleanSubDir(subDir string) (string, error) {
	return ValidateRelPath(subDir, true)
}

// scopedPath 将相对路径限定到指定子目录下。
func scopedPath(subDir, relPath string) (string, error) {
	cleanPath, err := ValidateRelPath(relPath, false)
	if err != nil {
		return "", err
	}
	if subDir == "" {
		return cleanPath, nil
	}
	return path.Join(subDir, cleanPath), nil
}

// initRepositoryLocked 在调用方持锁时初始化 Git 仓库。
func initRepositoryLocked(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create repo dir failed: %w", err)
	}
	if _, err := runGit(dir, "init"); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}
	if _, err := runGit(dir, "config", "user.name", "Skill Workspace"); err != nil {
		return fmt.Errorf("git config user.name failed: %w", err)
	}
	if _, err := runGit(dir, "config", "user.email", "skill@workspace.local"); err != nil {
		return fmt.Errorf("git config user.email failed: %w", err)
	}
	return nil
}

// commitAllInSubDirLocked 在调用方持锁时提交指定子目录全部变更。
func commitAllInSubDirLocked(dir, subDir, message string) (string, error) {
	if err := gitAddLocked(dir, nil, subDir); err != nil {
		return "", err
	}
	hasChanges, err := hasChangesInSubDirLocked(dir, subDir)
	if err != nil {
		return "", err
	}
	if !hasChanges {
		head, err := getHeadCommitLocked(dir)
		if err != nil {
			return "", fmt.Errorf("no changes and get HEAD failed: %w", err)
		}
		log.Printf("[git-util] CommitAll: no changes in %s, returning HEAD %s", dir, head)
		return head, nil
	}
	return gitCommitLocked(dir, message)
}

// getHeadCommitLocked 在调用方持锁时获取 HEAD 提交哈希。
func getHeadCommitLocked(dir string) (string, error) {
	out, err := runGitStdout(dir, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD failed: %w", err)
	}
	return out, nil
}

// hasHeadLocked 在调用方持锁时判断仓库是否已有 HEAD。
func hasHeadLocked(dir string) bool {
	_, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, "rev-parse", "--verify", "HEAD")
	return err == nil
}

// hasChangesInSubDirLocked 在调用方持锁时判断子目录是否有变更。
func hasChangesInSubDirLocked(dir, subDir string) (bool, error) {
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return false, fmt.Errorf("invalid subDir: %w", err)
	}
	args := []string{"status", "--porcelain"}
	if cleanSub != "" {
		args = append(args, "--", cleanSub+"/")
	}
	out, err := runGitStdout(dir, args...)
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return out != "", nil
}

// gitAddLocked 在调用方持锁时暂存路径。
func gitAddLocked(dir string, paths []string, subDir string) error {
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return fmt.Errorf("invalid subDir: %w", err)
	}
	if len(paths) == 0 {
		args := []string{"add", "-A"}
		if cleanSub != "" {
			args = []string{"add", "-A", "--", cleanSub + "/"}
		}
		if _, err := runGit(dir, args...); err != nil {
			return fmt.Errorf("git add failed: %w", err)
		}
		return nil
	}

	args := []string{"add", "--"}
	for _, p := range paths {
		scoped, err := scopedPath(cleanSub, p)
		if err != nil {
			return fmt.Errorf("invalid path %q: %w", p, err)
		}
		args = append(args, scoped)
	}
	if _, err := runGit(dir, args...); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	return nil
}

// gitResetLocked 在调用方持锁时取消暂存路径。
func gitResetLocked(dir string, paths []string, subDir string) error {
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return fmt.Errorf("invalid subDir: %w", err)
	}
	args := []string{"reset", "--"}
	if hasHeadLocked(dir) {
		args = []string{"reset", "HEAD", "--"}
	}
	if len(paths) == 0 {
		if cleanSub != "" {
			args = append(args, cleanSub+"/")
		}
	} else {
		for _, p := range paths {
			scoped, err := scopedPath(cleanSub, p)
			if err != nil {
				return fmt.Errorf("invalid path %q: %w", p, err)
			}
			args = append(args, scoped)
		}
	}
	if _, err := runGit(dir, args...); err != nil {
		return fmt.Errorf("git reset failed: %w", err)
	}
	return nil
}

// scopedPathspecs 将路径列表转换为限定子目录的 pathspec。
func scopedPathspecs(cleanSub string, paths []string) ([]string, error) {
	if len(paths) == 0 {
		if cleanSub == "" {
			return nil, nil
		}
		return []string{cleanSub + "/"}, nil
	}

	pathspecs := make([]string, 0, len(paths))
	for _, p := range paths {
		scoped, err := scopedPath(cleanSub, p)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", p, err)
		}
		pathspecs = append(pathspecs, scoped)
	}
	return pathspecs, nil
}

// gitDiscardWorkingTreeLocked 在调用方持锁时放弃工作区更改。
func gitDiscardWorkingTreeLocked(dir string, paths []string, subDir string) error {
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return fmt.Errorf("invalid subDir: %w", err)
	}
	pathspecs, err := scopedPathspecs(cleanSub, paths)
	if err != nil {
		return err
	}

	diffArgs := []string{"diff", "--name-only", "-z", "--"}
	diffArgs = append(diffArgs, pathspecs...)
	out, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, diffArgs...)
	if err != nil {
		return fmt.Errorf("git diff --name-only failed: %w", err)
	}

	changedPaths := make([]string, 0)
	for _, p := range strings.Split(string(out), "\x00") {
		if p != "" {
			changedPaths = append(changedPaths, p)
		}
	}
	if len(changedPaths) > 0 {
		args := []string{"restore", "--worktree", "--"}
		args = append(args, changedPaths...)
		if _, err := runGit(dir, args...); err != nil {
			return fmt.Errorf("git restore failed: %w", err)
		}
	}

	cleanArgs := []string{"clean", "-fd", "--"}
	cleanArgs = append(cleanArgs, pathspecs...)
	if _, err := runGit(dir, cleanArgs...); err != nil {
		return fmt.Errorf("git clean failed: %w", err)
	}
	return nil
}

// gitRestoreLocked 在调用方持锁时恢复指定子目录到 commit，同时覆盖暂存区和工作区。
func gitRestoreLocked(dir string, commit string, subDir string) error {
	if commit == "" {
		return errors.New("commit is required")
	}
	if err := validateCommitRef(commit); err != nil {
		return err
	}
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return fmt.Errorf("invalid subDir: %w", err)
	}
	pathspec := ""
	if cleanSub != "" {
		pathspec = cleanSub + "/"
	}

	args := []string{"restore", "--source=" + commit, "--staged", "--worktree", "--"}
	if pathspec != "" {
		args = append(args, pathspec)
	}
	if _, err := runGit(dir, args...); err != nil {
		return fmt.Errorf("git restore failed: %w", err)
	}

	// 清理未跟踪文件和目录，使工作区完全恢复到目标 commit 的状态。
	cleanArgs := []string{"clean", "-fd", "--"}
	if pathspec != "" {
		cleanArgs = append(cleanArgs, pathspec)
	}
	if _, err := runGit(dir, cleanArgs...); err != nil {
		return fmt.Errorf("git clean failed: %w", err)
	}

	return nil
}

// gitCommitLocked 在调用方持锁时提交已暂存变更。
func gitCommitLocked(dir, message string) (string, error) {
	if _, err := runGit(dir, "commit", "-m", message); err != nil {
		return "", fmt.Errorf("git commit failed: %w", err)
	}
	commitHash, err := getHeadCommitLocked(dir)
	if err != nil {
		return "", err
	}
	log.Printf("[git-util] GitCommit: committed %s in %s", commitHash, dir)
	return commitHash, nil
}

// InitRepository 初始化指定目录为 Git 仓库。
func InitRepository(dir string) error {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return initRepositoryLocked(dir)
}

// InitRepositoryLocked 在调用方持锁时初始化 Git 仓库。
func InitRepositoryLocked(dir string) error {
	return initRepositoryLocked(dir)
}

// CommitAllInSubDir 提交指定子目录的全部变更。
func CommitAllInSubDir(dir, subDir, message string) (string, error) {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return commitAllInSubDirLocked(dir, subDir, message)
}

// CommitAllInSubDirLocked 在调用方持锁时提交指定子目录全部变更。
func CommitAllInSubDirLocked(dir, subDir, message string) (string, error) {
	return commitAllInSubDirLocked(dir, subDir, message)
}

// GetHeadCommit 获取指定仓库的 HEAD 提交哈希。
func GetHeadCommit(dir string) (string, error) {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return getHeadCommitLocked(dir)
}

// GetDiff 获取指定提交范围和子目录的 diff。
func GetDiff(dir, fromCommit, toCommit string, subDir string) (string, error) {
	if toCommit == "" {
		toCommit = "HEAD"
	}
	if err := validateCommitRange(fromCommit, toCommit); err != nil {
		return "", err
	}
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return "", fmt.Errorf("invalid subDir: %w", err)
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	pathArg := ""
	if cleanSub != "" {
		pathArg = cleanSub + "/"
	}
	var args []string
	if fromCommit == "" {
		args = []string{"show", "--root", "--format=", toCommit}
		if pathArg != "" {
			args = append(args, "--", pathArg)
		}
	} else {
		args = []string{"diff", fromCommit, toCommit}
		if pathArg != "" {
			args = append(args, "--", pathArg)
		}
	}
	out, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, args...)
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(out), nil
}

type CommitInfo struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
}

// GetCommitLog 获取指定仓库的提交历史。
func GetCommitLog(dir string, count int) ([]CommitInfo, error) {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	if count <= 0 {
		count = 50
	}
	if !hasHeadLocked(dir) {
		return []CommitInfo{}, nil
	}
	out, err := runGitStdout(dir, "log", fmt.Sprintf("--format=%%H|%%s|%%ct"), "-n", strconv.Itoa(count))
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	lines := strings.Split(out, "\n")
	commits := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		timestamp, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			log.Printf("[git-util] parse commit timestamp failed: dir=%s hash=%s value=%s err=%v", dir, parts[0], parts[2], err)
			timestamp = 0
		}
		commits = append(commits, CommitInfo{
			Hash:    parts[0],
			Message: parts[1],
			Time:    timestamp,
		})
	}
	return commits, nil
}

// IsRepoInitialized 判断目录是否已初始化为 Git 仓库。
func IsRepoInitialized(dir string) bool {
	_, err := runGit(dir, "rev-parse", "--git-dir")
	return err == nil
}

// HasHead 判断指定仓库是否已有 HEAD 提交。
func HasHead(dir string) bool {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return hasHeadLocked(dir)
}

type FileChangeInfo struct {
	Path       string
	OldPath    string
	ChangeType string
}

type FileSnapshot struct {
	Content string
	Exists  bool
}

// GetChangedFiles 获取指定提交范围内的变更文件列表。
func GetChangedFiles(dir, fromCommit, toCommit string, subDir string) ([]FileChangeInfo, error) {
	if toCommit == "" {
		toCommit = "HEAD"
	}
	if err := validateCommitRange(fromCommit, toCommit); err != nil {
		return nil, err
	}
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return nil, fmt.Errorf("invalid subDir: %w", err)
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	pathArg := ""
	if cleanSub != "" {
		pathArg = cleanSub + "/"
	}
	var args []string
	if fromCommit == "" {
		args = []string{"show", "--root", "--name-status", "--format=", toCommit}
		if pathArg != "" {
			args = append(args, "--", pathArg)
		}
	} else {
		args = []string{"diff", "--name-status", fromCommit, toCommit}
		if pathArg != "" {
			args = append(args, "--", pathArg)
		}
	}
	out, err := runGit(dir, args...)
	if err != nil {
		return nil, fmt.Errorf("git diff --name-status failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	files := make([]FileChangeInfo, 0, len(lines))
	prefix := ""
	if cleanSub != "" {
		prefix = cleanSub + "/"
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		status := parts[0]
		changeType := "modified"
		switch {
		case strings.HasPrefix(status, "A"):
			changeType = "added"
		case strings.HasPrefix(status, "D"):
			changeType = "deleted"
		case strings.HasPrefix(status, "R"):
			changeType = "renamed"
		}
		var filePath, oldPath string
		if len(parts) == 3 {
			oldPath = strings.TrimPrefix(strings.TrimSpace(parts[1]), prefix)
			filePath = strings.TrimPrefix(strings.TrimSpace(parts[2]), prefix)
		} else {
			filePath = strings.TrimPrefix(strings.TrimSpace(parts[1]), prefix)
		}
		files = append(files, FileChangeInfo{
			Path:       filePath,
			OldPath:    oldPath,
			ChangeType: changeType,
		})
	}
	return files, nil
}

// GetFileContentAtCommit 读取指定提交中的文件内容。
func GetFileContentAtCommit(dir, commit, filePath string) (string, error) {
	snapshot, err := GetFileSnapshotAtCommit(dir, commit, filePath)
	if err != nil {
		return "", err
	}
	if !snapshot.Exists {
		return "", fmt.Errorf("git file not found: %s", filePath)
	}
	return snapshot.Content, nil
}

// GetFileSnapshotAtCommit 读取指定提交中的文件快照。
func GetFileSnapshotAtCommit(dir, commit, filePath string) (FileSnapshot, error) {
	if commit == "" {
		commit = "HEAD"
	}
	cleanPath, err := ValidateRelPath(filePath, false)
	if err != nil {
		return FileSnapshot{}, fmt.Errorf("invalid filePath: %w", err)
	}
	if err := validateCommitRef(commit); err != nil {
		return FileSnapshot{}, err
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	return getFileSnapshotAtTreeishLocked(dir, commit, cleanPath)
}

// GetFileSnapshotAtIndex 读取暂存区中的文件快照。
func GetFileSnapshotAtIndex(dir, filePath string) (FileSnapshot, error) {
	cleanPath, err := ValidateRelPath(filePath, false)
	if err != nil {
		return FileSnapshot{}, fmt.Errorf("invalid filePath: %w", err)
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	return getFileSnapshotAtSpecLocked(dir, ":"+cleanPath)
}

// getFileSnapshotAtTreeishLocked 在调用方持锁时读取 treeish 中的文件快照。
func getFileSnapshotAtTreeishLocked(dir, treeish, cleanPath string) (FileSnapshot, error) {
	if _, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, "rev-parse", "--verify", treeish+"^{tree}"); err != nil {
		if treeish == "HEAD" {
			return FileSnapshot{Exists: false}, nil
		}
		return FileSnapshot{}, fmt.Errorf("git rev-parse failed: %w", err)
	}
	return getFileSnapshotAtSpecLocked(dir, fmt.Sprintf("%s:%s", treeish, cleanPath))
}

// getFileSnapshotAtSpecLocked 在调用方持锁时读取 Git 对象规格对应的文件快照。
func getFileSnapshotAtSpecLocked(dir, spec string) (FileSnapshot, error) {
	out, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, "cat-file", "-t", spec)
	if err != nil {
		return FileSnapshot{Exists: false}, nil
	}
	if typ := strings.TrimSpace(string(out)); typ != "blob" {
		return FileSnapshot{}, fmt.Errorf("git object %s is %s, not blob", spec, typ)
	}
	content, err := runGitStdoutBytes(dir, maxGitFileOutputBytes, "show", spec)
	if err != nil {
		return FileSnapshot{}, fmt.Errorf("git show failed: %w", err)
	}
	return FileSnapshot{Content: string(content), Exists: true}, nil
}

// GetFileDiff 获取单个文件在提交范围内的 diff。
func GetFileDiff(dir, fromCommit, toCommit, filePath string) (string, error) {
	if toCommit == "" {
		toCommit = "HEAD"
	}
	if err := validateCommitRange(fromCommit, toCommit); err != nil {
		return "", err
	}
	cleanPath, err := ValidateRelPath(filePath, false)
	if err != nil {
		return "", fmt.Errorf("invalid filePath: %w", err)
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	var args []string
	if fromCommit == "" {
		args = []string{"show", "--root", "--format=", toCommit, "--", cleanPath}
	} else {
		args = []string{"diff", fromCommit, toCommit, "--", cleanPath}
	}
	out, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, args...)
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(out), nil
}

type GitStatusFile struct {
	Path       string
	ChangeType string
	Staged     bool
	OldPath    string
}

// GitStatus 获取指定仓库的文件状态。
func GitStatus(dir string, subDir string) ([]GitStatusFile, error) {
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return nil, fmt.Errorf("invalid subDir: %w", err)
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	args := []string{"status", "--porcelain", "-z", "-uall"}
	if cleanSub != "" {
		args = append(args, "--", cleanSub+"/")
	}
	out, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, args...)
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	data := string(out)
	if data == "" {
		return nil, nil
	}
	prefix := ""
	if cleanSub != "" {
		prefix = cleanSub + "/"
	}

	entries := strings.Split(data, "\x00")
	files := make([]GitStatusFile, 0, len(entries))
	for i := 0; i < len(entries); {
		entry := entries[i]
		if entry == "" {
			i++
			continue
		}
		if len(entry) < 3 {
			i++
			continue
		}
		xy := entry[:2]
		pathPart := strings.TrimPrefix(entry[3:], prefix)

		var oldPath string
		if isRenamedStatus(xy) && i+1 < len(entries) {
			i++
			oldPath = strings.TrimPrefix(entries[i], prefix)
		}

		files = append(files, statusFilesFromXY(xy, pathPart, oldPath)...)
		i++
	}
	return files, nil
}

// statusFilesFromXY 将 porcelain 状态位转换为状态文件列表。
func statusFilesFromXY(xy, pathPart, oldPath string) []GitStatusFile {
	x, y := xy[0], xy[1]
	if x == '?' && y == '?' {
		return []GitStatusFile{{Path: pathPart, ChangeType: "untracked", Staged: false}}
	}

	files := make([]GitStatusFile, 0, 2)
	if x != ' ' {
		files = append(files, GitStatusFile{
			Path:       pathPart,
			ChangeType: parseStatusByte(x),
			Staged:     true,
			OldPath:    oldPath,
		})
	}
	if y != ' ' {
		files = append(files, GitStatusFile{
			Path:       pathPart,
			ChangeType: parseStatusByte(y),
			Staged:     false,
			OldPath:    oldPath,
		})
	}
	return files
}

// isRenamedStatus 判断 porcelain 状态是否包含重命名。
func isRenamedStatus(xy string) bool {
	return len(xy) >= 2 && (xy[0] == 'R' || xy[1] == 'R')
}

// parseStatusByte 将 Git 状态字符转换为业务变更类型。
func parseStatusByte(status byte) string {
	switch status {
	case 'R':
		return "renamed"
	case 'D':
		return "deleted"
	case 'A':
		return "added"
	case '?':
		return "untracked"
	default:
		return "modified"
	}
}

// GitAdd 暂存指定仓库中的路径。
func GitAdd(dir string, paths []string, subDir string) error {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return gitAddLocked(dir, paths, subDir)
}

// GitAddLocked 在调用方持锁时暂存指定路径。
func GitAddLocked(dir string, paths []string, subDir string) error {
	return gitAddLocked(dir, paths, subDir)
}

// GitReset 取消暂存指定仓库中的路径。
func GitReset(dir string, paths []string, subDir string) error {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return gitResetLocked(dir, paths, subDir)
}

// GitResetLocked 在调用方持锁时取消暂存指定路径。
func GitResetLocked(dir string, paths []string, subDir string) error {
	return gitResetLocked(dir, paths, subDir)
}

// GitRestore 恢复指定仓库中的子目录到 commit，同时覆盖暂存区和工作区。
func GitRestore(dir string, commit string, subDir string) error {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return gitRestoreLocked(dir, commit, subDir)
}

// GitDiscardWorkingTree 放弃指定仓库中的工作区更改。
func GitDiscardWorkingTree(dir string, paths []string, subDir string) error {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return gitDiscardWorkingTreeLocked(dir, paths, subDir)
}

// GitCommit 提交指定仓库已暂存变更。
func GitCommit(dir, message string) (string, error) {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return gitCommitLocked(dir, message)
}

// GitCommitLocked 在调用方持锁时提交已暂存变更。
func GitCommitLocked(dir, message string) (string, error) {
	return gitCommitLocked(dir, message)
}

// GitDiffWorkingTree 获取工作区未暂存 diff。
func GitDiffWorkingTree(dir string, subDir string, filePath string) (string, error) {
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return "", fmt.Errorf("invalid subDir: %w", err)
	}
	pathArg := ""
	if filePath != "" {
		scoped, err := scopedPath(cleanSub, filePath)
		if err != nil {
			return "", fmt.Errorf("invalid filePath: %w", err)
		}
		pathArg = scoped
	} else if cleanSub != "" {
		pathArg = cleanSub + "/"
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	args := []string{"diff"}
	if pathArg != "" {
		args = append(args, "--", pathArg)
	}
	out, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, args...)
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(out), nil
}

// GitDiffStaged 获取暂存区 diff。
func GitDiffStaged(dir string, subDir string, filePath string) (string, error) {
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return "", fmt.Errorf("invalid subDir: %w", err)
	}
	pathArg := ""
	if filePath != "" {
		scoped, err := scopedPath(cleanSub, filePath)
		if err != nil {
			return "", fmt.Errorf("invalid filePath: %w", err)
		}
		pathArg = scoped
	} else if cleanSub != "" {
		pathArg = cleanSub + "/"
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	args := []string{"diff", "--cached"}
	if !hasHeadLocked(dir) {
		args = []string{"diff", "--cached", "--root"}
	}
	if pathArg != "" {
		args = append(args, "--", pathArg)
	}
	out, err := runGitStdoutBytes(dir, maxGitTextOutputBytes, args...)
	if err != nil {
		return "", fmt.Errorf("git diff --cached failed: %w", err)
	}
	return string(out), nil
}

// HasChangesInSubDir 判断指定子目录是否存在变更。
func HasChangesInSubDir(dir, subDir string) (bool, error) {
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()
	return hasChangesInSubDirLocked(dir, subDir)
}

// TagExists 判断指定仓库中 tag 是否存在。
func TagExists(dir, tagName string) (bool, error) {
	if err := ValidateTagName(tagName); err != nil {
		return false, err
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	out, err := runGitStdout(dir, "tag", "--list", tagName)
	if err != nil {
		return false, fmt.Errorf("git tag list failed: %w", err)
	}
	return out == tagName, nil
}

// CreateTag 在指定仓库当前 HEAD 上创建 tag。
func CreateTag(dir, tagName string) (string, error) {
	if err := ValidateTagName(tagName); err != nil {
		return "", err
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	headHash, err := getHeadCommitLocked(dir)
	if err != nil {
		return "", fmt.Errorf("get HEAD commit failed: %w", err)
	}
	if _, err := runGit(dir, "tag", tagName); err != nil {
		return "", fmt.Errorf("git tag %s failed: %w", tagName, err)
	}
	return headHash, nil
}

// DeleteTag 删除指定仓库中的 tag。
func DeleteTag(dir, tagName string) error {
	if err := ValidateTagName(tagName); err != nil {
		return err
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	if _, err := runGit(dir, "tag", "-d", tagName); err != nil {
		return fmt.Errorf("git tag -d %s failed: %w", tagName, err)
	}
	return nil
}

// ArchivePath 将指定 treeish 下的路径打包为 zip。
func ArchivePath(dir, treeish, subDir string) ([]byte, error) {
	if err := validateTreeish(treeish); err != nil {
		return nil, err
	}
	cleanSub, err := cleanSubDir(subDir)
	if err != nil {
		return nil, fmt.Errorf("invalid subDir: %w", err)
	}
	mu := getMu(dir)
	mu.Lock()
	defer mu.Unlock()

	target := treeish
	if cleanSub != "" {
		target = treeish + ":" + cleanSub
	}
	ctx, cancel := commandContext(defaultTimeout)
	defer cancel()
	out, err := runGitCombined(ctx, dir, maxGitArchiveSizeBytes, "archive", "--format=zip", target)
	if err != nil {
		return nil, fmt.Errorf("git archive %s failed: %w", target, err)
	}
	return out, nil
}
