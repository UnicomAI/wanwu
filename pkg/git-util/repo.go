package git_util

import (
	"os"
	"path/filepath"
	"sync"
)

// Repo 封装对单个 git 仓库的所有操作。dir 在构造时绑定，方法签名中不再重复传递。
// 通过 Open(dir) 获取实例；调用方依赖 Repo 接口可在测试中注入 MockRepo。
//
// 方法分为两类：
//   - 普通方法：内部自行加锁，调用方无需关心并发。
//   - Locked 方法：不加锁，调用方须先持有 GetMutex().Lock()，
//     用于需要跨多步 git 操作保持原子性的场景。
type Repo interface {
	// IsInitialized 报告目录是否已初始化为 git 仓库。
	IsInitialized() bool
	HasHead() bool
	// GetMutex 返回与目录绑定的互斥锁。
	// 需要原子性保障的多步操作须先 Lock()，再调用对应的 Locked 方法，完成后 Unlock()。
	GetMutex() *sync.Mutex

	// --- 普通方法（内部加锁） ---

	GetHeadCommit() (string, error)
	GetDiff(fromCommit, toCommit, subDir string) (string, error)
	GetCommitLog(count int) ([]CommitInfo, error)
	GetChangedFiles(fromCommit, toCommit, subDir string) ([]FileChangeInfo, error)
	GetFileContentAtCommit(commit, filePath string) (string, error)
	GetFileSnapshotAtCommit(commit, filePath string) (FileSnapshot, error)
	GetFileSnapshotAtIndex(filePath string) (FileSnapshot, error)
	GetFileDiff(fromCommit, toCommit, filePath string) (string, error)
	Status(subDir string) ([]GitStatusFile, error)
	Add(paths []string, subDir string) error
	Reset(paths []string, subDir string) error
	Restore(commit string, subDir string) error
	DiscardWorkingTree(paths []string, subDir string) error
	Commit(message string) (string, error)
	DiffWorkingTree(subDir, filePath string) (string, error)
	DiffStaged(subDir, filePath string) (string, error)
	HasChangesInSubDir(subDir string) (bool, error)
	TagExists(tagName string) (bool, error)
	CreateTag(tagName string) (string, error)
	DeleteTag(tagName string) error
	ArchivePath(treeish, subDir string) ([]byte, error)
	ListTreeFiles(treeish, subDir string) ([]TreeEntry, error)
	GetTreeishCommitTimeMs(treeish string) (int64, error)
	ReadBlobAtTreeish(treeish, filePath string, maxBytes int64) (BlobInfo, error)

	// --- Locked 方法（调用方须持 GetMutex().Lock()） ---

	InitLocked() error
	CommitAllInSubDirLocked(subDir, message string) (string, error)
	AddLocked(paths []string, subDir string) error
}

// repo 是 Repo 的默认实现，通过 Open(dir) 创建。
type repo struct {
	dir string
}

// Open 返回与 dir 绑定的 Repo 实例。
// 同一 dir 的多次 Open 调用共享同一底层互斥锁（通过全局 sync.Map 保证）。
func Open(dir string) Repo {
	return &repo{dir: dir}
}

// IsInitialized 判断当前目录是否已是 Git 仓库。
func (r *repo) IsInitialized() bool { return IsRepoInitialized(r.dir) }

// HasHead 判断当前仓库是否已有 HEAD 提交。
func (r *repo) HasHead() bool { return HasHead(r.dir) }

// GetMutex 返回当前仓库目录对应的互斥锁。
func (r *repo) GetMutex() *sync.Mutex { return getMu(r.dir) }

// GetHeadCommit 获取当前 HEAD 提交哈希。
func (r *repo) GetHeadCommit() (string, error) { return GetHeadCommit(r.dir) }

// GetDiff 获取当前仓库指定提交范围的 diff。
func (r *repo) GetDiff(from, to, subDir string) (string, error) {
	return GetDiff(r.dir, from, to, subDir)
}

// GetCommitLog 获取当前仓库的提交历史。
func (r *repo) GetCommitLog(count int) ([]CommitInfo, error) { return GetCommitLog(r.dir, count) }

// GetChangedFiles 获取当前仓库指定提交范围的变更文件。
func (r *repo) GetChangedFiles(from, to, subDir string) ([]FileChangeInfo, error) {
	return GetChangedFiles(r.dir, from, to, subDir)
}

// GetFileContentAtCommit 读取指定提交中的文件内容。
func (r *repo) GetFileContentAtCommit(commit, filePath string) (string, error) {
	return GetFileContentAtCommit(r.dir, commit, filePath)
}

// GetFileSnapshotAtCommit 读取指定提交中的文件快照。
func (r *repo) GetFileSnapshotAtCommit(commit, filePath string) (FileSnapshot, error) {
	return GetFileSnapshotAtCommit(r.dir, commit, filePath)
}

// GetFileSnapshotAtIndex 读取暂存区中的文件快照。
func (r *repo) GetFileSnapshotAtIndex(filePath string) (FileSnapshot, error) {
	return GetFileSnapshotAtIndex(r.dir, filePath)
}

// GetFileDiff 获取单个文件在提交范围内的 diff。
func (r *repo) GetFileDiff(from, to, filePath string) (string, error) {
	return GetFileDiff(r.dir, from, to, filePath)
}

// Status 获取当前仓库的文件状态。
func (r *repo) Status(subDir string) ([]GitStatusFile, error) { return GitStatus(r.dir, subDir) }

// Add 暂存当前仓库中的路径。
func (r *repo) Add(paths []string, subDir string) error { return GitAdd(r.dir, paths, subDir) }

// Reset 取消暂存当前仓库中的路径。
func (r *repo) Reset(paths []string, subDir string) error { return GitReset(r.dir, paths, subDir) }

// Restore 恢复当前仓库中的子目录到指定 commit。
func (r *repo) Restore(commit string, subDir string) error { return GitRestore(r.dir, commit, subDir) }

// DiscardWorkingTree 放弃当前仓库工作区更改。
func (r *repo) DiscardWorkingTree(paths []string, subDir string) error {
	return GitDiscardWorkingTree(r.dir, paths, subDir)
}

// Commit 提交当前仓库已暂存变更。
func (r *repo) Commit(message string) (string, error) { return GitCommit(r.dir, message) }

// DiffWorkingTree 获取工作区未暂存 diff。
func (r *repo) DiffWorkingTree(subDir, filePath string) (string, error) {
	return GitDiffWorkingTree(r.dir, subDir, filePath)
}

// DiffStaged 获取暂存区 diff。
func (r *repo) DiffStaged(subDir, filePath string) (string, error) {
	return GitDiffStaged(r.dir, subDir, filePath)
}

// HasChangesInSubDir 判断子目录内是否存在变更。
func (r *repo) HasChangesInSubDir(subDir string) (bool, error) {
	return HasChangesInSubDir(r.dir, subDir)
}

// TagExists 判断指定 tag 是否存在。
func (r *repo) TagExists(tagName string) (bool, error) { return TagExists(r.dir, tagName) }

// CreateTag 在当前 HEAD 上创建 tag。
func (r *repo) CreateTag(tagName string) (string, error) { return CreateTag(r.dir, tagName) }

// DeleteTag 删除指定 tag。
func (r *repo) DeleteTag(tagName string) error { return DeleteTag(r.dir, tagName) }

// ArchivePath 打包指定 treeish 下的路径。
func (r *repo) ArchivePath(treeish, subDir string) ([]byte, error) {
	return ArchivePath(r.dir, treeish, subDir)
}

// ListTreeFiles 列出指定 treeish 下子目录中的全部 blob。
func (r *repo) ListTreeFiles(treeish, subDir string) ([]TreeEntry, error) {
	return ListTreeFiles(r.dir, treeish, subDir)
}

// GetTreeishCommitTimeMs 获取指定 treeish 对应提交的时间（毫秒）。
func (r *repo) GetTreeishCommitTimeMs(treeish string) (int64, error) {
	return GetTreeishCommitTimeMs(r.dir, treeish)
}

// ReadBlobAtTreeish 读取指定 treeish 下路径对应的对象。
func (r *repo) ReadBlobAtTreeish(treeish, filePath string, maxBytes int64) (BlobInfo, error) {
	return ReadBlobAtTreeish(r.dir, treeish, filePath, maxBytes)
}

// InitLocked 在调用方持锁时初始化仓库。
func (r *repo) InitLocked() error { return InitRepositoryLocked(r.dir) }

// CommitAllInSubDirLocked 在调用方持锁时提交子目录全部变更。
func (r *repo) CommitAllInSubDirLocked(subDir, message string) (string, error) {
	return CommitAllInSubDirLocked(r.dir, subDir, message)
}

// AddLocked 在调用方持锁时暂存路径。
func (r *repo) AddLocked(paths []string, subDir string) error {
	return GitAddLocked(r.dir, paths, subDir)
}

// HasGitMetadata 判断目录下是否存在 .git 文件/目录（不会进入子目录或解析仓库）。
func HasGitMetadata(dir string) (bool, error) {
	_, err := os.Lstat(filepath.Join(dir, ".git"))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
