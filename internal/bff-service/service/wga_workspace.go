// Package service 提供 WGA（General Agent）工作空间管理的公共能力。
//
// # 设计目标
//
// 本文件从 BFF 层抽取工作空间相关的公共操作，为不同业务场景提供可复用的能力：
//   - General Agent：当前主要使用方，使用 versioned 或 overwrite 模式
//   - 其他业务：可通过传入自定义 Store 使用相同能力，支持不同的 baseDir/mode/threadID
//
// # 存储模式支持
//
// wga-persistent 包提供两种存储模式，本文件所有方法均支持：
//   - ModeVersioned：分轮存储，每次执行创建独立 run 目录，保留历史
//   - ModeOverwrite：覆盖模式，每次执行覆盖同一目录，不保留历史
//
// 后续扩展新模式时，只需在 wga-persistent 包中实现新的 SessionPersistent 接口，
// 本文件中的业务方法无需修改即可复用。
//
// # 跨 Thread 场景支持
//
// Store 在创建时绑定一个 threadID（决定文件存储位置），但业务上下文可能需要不同的 threadID：
//   - 场景示例：多个对话共享同一个 workspace，但需要发送不同 threadID 的事件
//   - 解决方案：Store 用于访问文件，业务 threadID 通过参数传入（如 BuildWgaWorkspaceEvent 的 cfg.ThreadID）
//
// # 方法分类
//
// 1. 工厂方法：NewGeneralAgentWorkspaceStore（General Agent 专用），其他业务应补充对应的业务工厂方法
// 2. 目录获取方法：接收 Store 参数，封装错误处理
// 3. 工作空间信息统计：纯函数，不依赖 Store
// 4. 文件树构建：纯函数，不依赖 Store
// 5. AG-UI 事件构建：BuildWgaWorkspaceEvent，支持跨 thread 场景
// 6. 高层业务方法：接收 Store 参数，支持不同模式复用
// 7. 工具方法：纯函数，不依赖 Store
package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	path_util "github.com/UnicomAI/wanwu/pkg/path-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_persistent "github.com/UnicomAI/wanwu/pkg/wga-persistent"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
)

// ============================================================
// 核心工厂方法
// ============================================================

// NewGeneralAgentWorkspaceStore 从配置创建 General Agent 专用的持久化存储 Store。
//
// 此方法为 General Agent 场景专用，从统一配置读取 baseDir 和 mode。
// 其他业务应补充对应的业务工厂方法，参考此方法封装配置读取和错误处理：
//
//	func NewXxxWorkspaceStore(threadID string) (*wga_persistent.Store, error) {
//	    // 从业务配置读取 baseDir、mode
//	    return wga_persistent.NewStore(mode, baseDir, threadID)
//	}
//
// 参数：
//   - threadID: 绑定的会话 ID，决定工作空间存储位置
//
// 返回：
//   - Store: 可用于所有工作空间操作
//   - error: 配置未启用或创建失败时返回错误
func NewGeneralAgentWorkspaceStore(threadID string) (*wga_persistent.Store, error) {
	cfg := config.WgaCfg()
	if !cfg.Persistent.Enabled {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "persistent not enabled")
	}

	mode := wga_persistent.ModeVersioned
	if cfg.Persistent.Mode == string(wga_persistent.ModeOverwrite) {
		mode = wga_persistent.ModeOverwrite
	}

	return wga_persistent.NewStore(mode, cfg.Persistent.BaseDir, threadID)
}

// NewGeneralAgentSkillWorkspaceStore creates the overwrite workspace used by Skill Chat.
//
// The store is bound to customSkillID instead of the WGA business threadID so multiple
// chat threads can continue editing the same skill workspace.
func NewGeneralAgentSkillWorkspaceStore(customSkillID string) (*wga_persistent.Store, error) {
	cfg := config.WgaCfg()
	if !cfg.Persistent.Enabled {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "persistent not enabled")
	}

	return wga_persistent.NewStore(wga_persistent.ModeOverwrite, generalAgentSkillWorkspaceBaseDir(cfg), customSkillID)
}

func generalAgentSkillWorkspaceBaseDir(cfg *config.WgaConfig) string {
	if cfg.Persistent.SkillBaseDir != "" {
		return cfg.Persistent.SkillBaseDir
	}
	baseDir := cfg.Persistent.BaseDir
	if baseDir == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(baseDir), "overwrite")
}

// ============================================================
// 目录获取方法
//
// 这些方法接收 Store 参数，封装底层调用的错误处理，返回标准化的错误信息。
// 支持所有存储模式（versioned/overwrite），支持跨 thread 场景（Store 绑定的
// threadID 决定文件访问位置，与业务上下文的 threadID 解耦）。
// ============================================================

// GetWgaWorkspaceRunDir 获取指定 run 的工作空间目录
// 封装 GetRunDir 调用和错误处理，返回标准化的错误信息
// 返回值：目录路径，如果不存在则返回空字符串和错误
func GetWgaWorkspaceRunDir(store *wga_persistent.Store, runID string) (string, error) {
	ok, info, err := store.GetRunDir(runID)
	if err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if !ok {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, "run directory not found")
	}
	return info.Dir, nil
}

// GetWgaWorkspaceLastRunDir 获取最新 run 的工作空间目录
// 用于恢复上次执行的工作空间状态
// 返回值：目录路径，如果不存在则返回空字符串（无错误）
func GetWgaWorkspaceLastRunDir(store *wga_persistent.Store) (string, error) {
	ok, info, err := store.GetLastRunDir()
	if err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if !ok {
		return "", nil // 不存在不返回错误
	}
	return info.Dir, nil
}

// GetWgaWorkspaceThreadDir 获取 thread 级别的工作空间目录
// 用于删除整个会话的工作空间
func GetWgaWorkspaceThreadDir(store *wga_persistent.Store) string {
	return store.GetThreadDir().Dir
}

// ============================================================
// 工作空间信息统计
//
// 纯函数，不依赖 Store。可直接用于任意目录的统计计算。
// 支持后续扩展新模式的业务复用，无需修改。
// ============================================================

// WgaWorkspaceInfo 工作空间统计信息
type WgaWorkspaceInfo struct {
	Dir       string // 目录路径
	FileCount int    // 文件数量
	TotalSize int64  // 总大小（字节）
}

// GetWgaWorkspaceInfo 获取工作空间统计信息
// 计算目录下的文件数量和总大小
// 如果目录不存在或为空，返回 FileCount=0, TotalSize=0
func GetWgaWorkspaceInfo(dir string) (*WgaWorkspaceInfo, error) {
	totalSize, fileCount, err := getWgaWorkspaceInfoInternal(dir)
	if err != nil {
		return nil, err
	}
	return &WgaWorkspaceInfo{
		Dir:       dir,
		FileCount: fileCount,
		TotalSize: totalSize,
	}, nil
}

// getWgaWorkspaceInfoInternal 内部方法：递归计算目录大小和文件数
func getWgaWorkspaceInfoInternal(currentDir string) (int64, int, error) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return 0, 0, err
	}

	var totalSize int64
	var fileCount int

	for _, entry := range entries {
		// 跳过隐藏文件
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		currentPath := filepath.Join(currentDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			log.Warnf("[wga] failed to get file info: %s: %v", currentPath, err)
			continue
		}

		if entry.IsDir() {
			dirSize, dirFileCount, err := getWgaWorkspaceInfoInternal(currentPath)
			if err != nil {
				log.Warnf("[wga] failed to get workspace info for dir: %s: %v", currentPath, err)
				continue
			}
			totalSize += dirSize
			fileCount += dirFileCount
		} else {
			totalSize += info.Size()
			fileCount++
		}
	}

	return totalSize, fileCount, nil
}

// ============================================================
// 文件树构建
//
// 纯函数，不依赖 Store。用于 info 接口和 download 目录打包场景。
// ============================================================

// BuildWgaWorkspaceFileTree 构建工作空间文件树
func BuildWgaWorkspaceFileTree(dir string) ([]*response.GeneralAgentFileNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []*response.GeneralAgentFileNode
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			log.Warnf("[wga] failed to get file info: %s: %v", entry.Name(), err)
			continue
		}

		fileNode := &response.GeneralAgentFileNode{
			Name: entry.Name(),
		}

		if entry.IsDir() {
			fileNode.Type = "directory"
			children, err := BuildWgaWorkspaceFileTree(filepath.Join(dir, entry.Name()))
			if err != nil {
				log.Warnf("[wga] failed to build file tree for dir: %s: %v", entry.Name(), err)
			}
			fileNode.Children = children
		} else {
			fileNode.Type = "file"
			fileNode.Size = info.Size()
			fullPath := filepath.Join(dir, entry.Name())
			if data, err := os.ReadFile(fullPath); err == nil {
				fileNode.MimeType = http.DetectContentType(data)
			}
			if fileNode.MimeType == "" {
				fileNode.MimeType = "application/octet-stream"
			}
		}

		files = append(files, fileNode)
	}

	return files, nil
}

// CalculateWgaFileTreeTotalSize 计算文件树总大小
func CalculateWgaFileTreeTotalSize(files []*response.GeneralAgentFileNode) int64 {
	var total int64
	for _, f := range files {
		if f.Type == "directory" {
			total += CalculateWgaFileTreeTotalSize(f.Children)
		} else {
			total += f.Size
		}
	}
	return total
}

// ============================================================
// AG-UI 事件构建
//
// BuildWgaWorkspaceEvent 构建工作空间活动事件，用于 RUN_FINISHED 后通知前端更新。
//
// 模式行为差异：
//   - ModeVersioned：检查工作空间是否有变化，无变化则不发送事件（避免无效通知）
//   - ModeOverwrite：直接发送事件，不检查变化（每次执行都是新内容）
//
// 跨 Thread 场景支持：
//   - Store.ThreadID：决定访问哪个 workspace 的文件
//   - cfg.ThreadID：事件内容中的 threadID，可以是业务上下文的对话 ID
//   - 示例：多个对话共享同一个 workspace，每个对话收到自己的事件
//
// 后续扩展新模式时，可根据 store.GetThreadDir().Mode 添加不同的行为逻辑。
// ============================================================

// WgaWorkspaceEventConfig 构建工作空间事件的配置
type WgaWorkspaceEventConfig struct {
	ThreadID           string
	RunID              string
	StepID             string
	LastWorkspaceSize  int64 // 上次统计的大小（用于对比，仅 versioned 模式有效）
	LastWorkspaceCount int   // 上次统计的数量（用于对比，仅 versioned 模式有效）
}

// BuildWgaWorkspaceEvent 构建 workspace 活动事件
// versioned 模式：如果工作空间为空或与上次相同，返回 nil
// overwrite 模式：直接发送 event（不检查变化）
// 用于 RUN_FINISHED 后注入 ACTIVITY_SNAPSHOT 事件
func BuildWgaWorkspaceEvent(store *wga_persistent.Store, cfg *WgaWorkspaceEventConfig) (aguievents.Event, error) {
	ok, info, err := store.GetRunDir(cfg.RunID)
	if err != nil {
		log.Warnf("[wga] failed to get run dir %s: %v", cfg.RunID, err)
		return nil, nil
	}
	if !ok {
		log.Warnf("[wga] run dir %s not found", cfg.RunID)
		return nil, nil
	}

	statInfo, err := os.Stat(info.Dir)
	if err != nil {
		log.Warnf("[wga] failed to stat dir %s: %v", info.Dir, err)
		return nil, nil
	}
	if !statInfo.IsDir() {
		log.Warnf("[wga] path %s is not a directory", info.Dir)
		return nil, nil
	}

	wsInfo, err := GetWgaWorkspaceInfo(info.Dir)
	if err != nil {
		log.Warnf("[wga] failed to get workspace info for %s: %v", info.Dir, err)
		return nil, nil
	}
	if wsInfo.FileCount == 0 {
		return nil, nil
	}

	// 获取 store 的模式
	threadDir := store.GetThreadDir()

	// versioned 模式：检查是否有变化
	if threadDir.Mode == wga_persistent.ModeVersioned {
		if wsInfo.TotalSize == cfg.LastWorkspaceSize && wsInfo.FileCount == cfg.LastWorkspaceCount {
			// 工作空间内容未变化，不发送事件
			return nil, nil
		}
	}
	// overwrite 模式：直接发送事件，不检查变化

	stepID := cfg.StepID
	if stepID == "" {
		stepID = aguievents.GenerateStepID()
	}

	return aguievents.NewActivitySnapshotEvent(
		stepID,
		ag_ui_util.ActivityTypeWorkspace,
		&ag_ui_util.WorkspaceActivityContent{
			RunID:     cfg.RunID,
			ThreadID:  cfg.ThreadID,
			FileCount: wsInfo.FileCount,
			TotalSize: wsInfo.TotalSize,
			Timestamp: time.Now().UnixMilli(),
		},
	), nil
}

// ============================================================
// 高层业务方法
//
// 这些方法接收 Store 参数，封装完整的业务逻辑。
// 支持不同模式复用：调用方传入不同 mode 的 Store，方法内部自动适配。
// 支持跨 thread 场景：Store 决定文件位置，与业务 threadID 解耦。
// ============================================================

// WgaWorkspaceDownloadResult 下载结果
type WgaWorkspaceDownloadResult struct {
	FileName    string // 文件名或目录名
	Data        []byte // 文件内容或 zip 压缩包
	ContentType string // MIME 类型
	IsDir       bool   // 是否是目录
}

// DownloadWgaWorkspace 下载工作空间文件或目录
// store: 持久化存储，由调用方创建（支持不同模式）
// runID: 执行 ID
// path: 相对路径，如果为空则下载整个 run 目录
func DownloadWgaWorkspace(store *wga_persistent.Store, runID, path string) (*WgaWorkspaceDownloadResult, error) {
	workDir, err := GetWgaWorkspaceRunDir(store, runID)
	if err != nil {
		return nil, err
	}

	targetPath := workDir
	if path != "" {
		resolved, _, err := path_util.JoinWithinBase(workDir, path, true)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("invalid path: %v", err))
		}
		if err := path_util.EnsureNoSymlinkInPath(workDir, resolved, true); err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("invalid path: %v", err))
		}
		targetPath = resolved
	}

	fi, err := os.Stat(targetPath)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("path not found: %v", err))
	}

	if fi.IsDir() {
		zipName := fmt.Sprintf("workspace_%s_%s.zip", runID, filepath.Base(path))
		zipData, err := util.ZipDir(targetPath + "/.")
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to create zip: %v", err))
		}
		return &WgaWorkspaceDownloadResult{
			FileName:    zipName,
			Data:        zipData,
			ContentType: "application/zip",
			IsDir:       true,
		}, nil
	}

	fileName := filepath.Base(path)
	fileData, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to read file: %v", err))
	}

	return &WgaWorkspaceDownloadResult{
		FileName:    fileName,
		Data:        fileData,
		ContentType: http.DetectContentType(fileData),
		IsDir:       false,
	}, nil
}

// WgaWorkspacePreviewResult 预览结果
type WgaWorkspacePreviewResult struct {
	Data        []byte
	ContentType string
	FileName    string
}

// PreviewWgaWorkspace 预览工作空间文件
// store: 持久化存储，由调用方创建（支持不同模式）
// runID: 执行 ID
// path: 文件相对路径（必须为文件，不支持目录）
func PreviewWgaWorkspace(store *wga_persistent.Store, runID, path string) (*WgaWorkspacePreviewResult, error) {
	workDir, err := GetWgaWorkspaceRunDir(store, runID)
	if err != nil {
		return nil, err
	}

	targetPath, _, err := path_util.JoinWithinBase(workDir, path, false)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("invalid path: %v", err))
	}
	if err := path_util.EnsureNoSymlinkInPath(workDir, targetPath, true); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("invalid path: %v", err))
	}

	fi, err := os.Stat(targetPath)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("path not found: %v", err))
	}
	if fi.IsDir() {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "path is a directory, not a file")
	}

	fileName := filepath.Base(path)
	fileData, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to read file: %v", err))
	}

	return &WgaWorkspacePreviewResult{
		Data:        fileData,
		ContentType: http.DetectContentType(fileData),
		FileName:    fileName,
	}, nil
}

// WgaWorkspaceTreeResult 工作空间文件树结果
type WgaWorkspaceTreeResult struct {
	Files     []*response.GeneralAgentFileNode
	FileCount int
	TotalSize int64
}

// GetWgaWorkspaceTree 获取工作空间文件树
func GetWgaWorkspaceTree(store *wga_persistent.Store, runID string) (*WgaWorkspaceTreeResult, error) {
	workDir, err := GetWgaWorkspaceRunDir(store, runID)
	if err != nil {
		return nil, err
	}

	files, err := BuildWgaWorkspaceFileTree(workDir)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to read directory: %v", err))
	}

	return &WgaWorkspaceTreeResult{
		Files:     files,
		FileCount: len(files),
		TotalSize: CalculateWgaFileTreeTotalSize(files),
	}, nil
}

// CleanupWgaWorkspace 清理整个 thread 的工作空间
func CleanupWgaWorkspace(store *wga_persistent.Store) error {
	threadDir := GetWgaWorkspaceThreadDir(store)
	if threadDir == "" {
		return nil
	}

	return util.DeleteDir(threadDir)
}

// WgaWorkspaceRunInfo 单次执行的目录信息
type WgaWorkspaceRunInfo struct {
	RunID     string
	Dir       string
	Timestamp int64
	FileCount int
	TotalSize int64
}

// ListWgaWorkspaceRuns 列出所有 run 的工作空间信息
//
// versioned 模式返回所有历史 run（按时间戳降序），overwrite 模式最多返回一个。
func ListWgaWorkspaceRuns(store *wga_persistent.Store) ([]*WgaWorkspaceRunInfo, error) {
	runDirs, err := store.ListRunDirs()
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}

	var result []*WgaWorkspaceRunInfo
	for _, info := range runDirs {
		wsInfo, err := GetWgaWorkspaceInfo(info.Dir)
		if err != nil {
			log.Warnf("[wga] failed to get workspace info for run %s: %v", info.RunID, err)
			continue
		}

		result = append(result, &WgaWorkspaceRunInfo{
			RunID:     info.RunID,
			Dir:       info.Dir,
			Timestamp: info.Timestamp,
			FileCount: wsInfo.FileCount,
			TotalSize: wsInfo.TotalSize,
		})
	}

	return result, nil
}

// WgaWorkspaceDirs 工作空间输入输出目录
type WgaWorkspaceDirs struct {
	InputDir  string // 输入目录（用于挂载到沙箱）
	OutputDir string // 输出目录（沙箱写入的目录）
}

// PrepareWgaWorkspaceDirs 为执行准备输入输出目录
//
// versioned 模式下 withCopyLastOutput=true 时从上次输出复制，overwrite 模式忽略此参数。
func PrepareWgaWorkspaceDirs(store *wga_persistent.Store, runID string, withCopyLastOutput bool) (*WgaWorkspaceDirs, error) {
	_, info, err := store.GetRunDir(runID, wga_persistent.WithMkdir(withCopyLastOutput))
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}

	return &WgaWorkspaceDirs{
		InputDir:  filepath.Clean(info.Dir) + "/.",
		OutputDir: info.Dir,
	}, nil
}

// ============================================================
// 工具方法
//
// 纯函数，不依赖 Store，可直接用于任意场景。
// ============================================================

// DownloadWgaWorkspaceURLs 下载 URL 文件到指定目录
func DownloadWgaWorkspaceURLs(ctx context.Context, urls map[string]string, dir string) error {
	if len(urls) == 0 {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir %s failed: %w", dir, err)
	}
	for fileName, urlStr := range urls {
		// Sanitize fileName to prevent path traversal (e.g. ../../etc/cron.d/evil)
		cleanName, err := path_util.CleanRelPath(fileName, false)
		if err != nil {
			log.Warnf("[DownloadWgaWorkspaceURLs] skip unsafe fileName %q: %v", fileName, err)
			continue
		}
		log.Infof("[DownloadWgaWorkspaceURLs] downloading URL %s to %s", urlStr, dir)
		resp, err := trace_util.NewResty(ctx).R().SetContext(ctx).Get(urlStr)
		if err != nil {
			log.Errorf("[DownloadWgaWorkspaceURLs] download URL %s failed: %v", urlStr, err)
			continue
		}
		bodyBytes := resp.Body()
		filePath := filepath.Join(dir, cleanName)
		if err := os.WriteFile(filePath, bodyBytes, 0644); err != nil {
			log.Errorf("[DownloadWgaWorkspaceURLs] save file %s failed: %v", filePath, err)
			continue
		}
		log.Infof("[DownloadWgaWorkspaceURLs] downloaded URL %s to %s, size: %d bytes", urlStr, filePath, len(bodyBytes))
	}
	return nil
}
