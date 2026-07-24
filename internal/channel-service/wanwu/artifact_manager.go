package wanwu

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// ArtifactManager WGA write 产物清单管理器
// 维护 (channelID + userID + threadID) -> [产物文件名 basename] 的累积映射，持久化到 DB，重启不丢。
// 触发：handleWGASSEResponse 末尾把本次 run 的 write 产物（extractProducedFiles 提取）追加落库；
// 读取：sendWorkspaceFiles 在本次无强信号时，加载 thread 累积清单作为跨 run 回发信号。
type ArtifactManager struct {
	cli client.IClient
}

// NewArtifactManager 创建 write 产物清单管理器
func NewArtifactManager(cli client.IClient) *ArtifactManager {
	return &ArtifactManager{cli: cli}
}

// AppendArtifacts 把本次 run write 写过的文件名追加到 thread 累积清单（去重落库）。
// fileNames 应为 basename（与工作区文件名对齐）。threadID 为空或无文件名时静默跳过。
func (m *ArtifactManager) AppendArtifacts(ctx context.Context, channelID, userID, threadID string, fileNames []string) {
	if threadID == "" || len(fileNames) == 0 {
		return
	}
	if err := m.cli.AddWgaArtifacts(ctx, channelID, userID, threadID, fileNames); err != nil {
		log.Errorf("failed to persist wga artifacts channel=%s user=%s thread=%s: %v",
			channelID, userID, threadID, err)
	}
}

// ListArtifacts 读 thread 累积的 write 产物清单（basename 列表，按写入时间倒序）。
// 读取失败只记日志返回 nil，不影响本次回发（回退到现有 produced/mentioned/stem 路径）。
func (m *ArtifactManager) ListArtifacts(ctx context.Context, channelID, userID, threadID string) []string {
	if threadID == "" {
		return nil
	}
	arts, err := m.cli.ListWgaArtifacts(ctx, channelID, userID, threadID)
	if err != nil {
		log.Warnf("failed to load wga artifacts channel=%s thread=%s: %v", channelID, threadID, err)
		return nil
	}
	names := make([]string, 0, len(arts))
	for _, a := range arts {
		names = append(names, a.FileName)
	}
	return names
}
