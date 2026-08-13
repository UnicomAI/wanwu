package orm

import (
	"context"
	"errors"
	"log"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/iam"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"github.com/UnicomAI/wanwu/internal/operate-service/config"
	"gorm.io/gorm"
)

type SystemCustomKey string
type SystemCustomMode string

const (
	SystemCustomTabKey          SystemCustomKey = "system_custom_tab"
	SystemCustomLoginKey        SystemCustomKey = "system_custom_login"
	SystemCustomHomeKey         SystemCustomKey = "system_custom_home"
	SystemCustomGeneralAgentKey SystemCustomKey = "system_custom_general_agent"
)
const (
	SystemCustomModeLight SystemCustomMode = "light"
	SystemCustomModeDark  SystemCustomMode = "dark"
)

type Client struct {
	db *gorm.DB
	// iam 消息中心的受众校验与 joinedAt 查询数据源
	iam iam.IClient
	// notice 消息中心配置（名单上限）
	notice config.NoticeConfig
}

func NewClient(db *gorm.DB, iamCli iam.IClient, noticeCfg config.NoticeConfig) (*Client, error) {
	// auto migrate
	if err := db.AutoMigrate(
		model.SystemCustom{},
		model.ClientRecord{},
		model.ClientDailyRecord{},
		// --- 消息中心五表 ---
		model.NoticeMessage{},
		model.MessageOrg{},
		model.MessageAudience{},
		model.UserStatus{},
		model.ReadWatermark{},
	); err != nil {
		return nil, err
	}
	c := &Client{
		db:     db,
		iam:    iamCli,
		notice: noticeCfg.Fill(),
	}
	// init corn
	if err := CronInit(db); err != nil {
		log.Fatalf("init corn failed, err: %v", err)
	}
	return c, nil
}

// transaction 事务封装：业务逻辑返回 *err_code.Status，非 nil 时回滚。
// 与 iam/rag/model/mcp/assistant 各服务 orm 包内的同名封装保持一致。
func (c *Client) transaction(ctx context.Context, fc func(tx *gorm.DB) *err_code.Status) *err_code.Status {
	var status *err_code.Status
	_ = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if status = fc(tx); status != nil {
			return errors.New(status.String())
		}
		return nil
	})
	return status
}

func toErrStatus(key string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: key,
		Args:    args,
	}
}

type SystemCustom struct {
	Login        LoginConfig        `json:"login"`        // 登录页配置
	Tab          TabConfig          `json:"tab"`          // 标签页配置
	Home         HomeConfig         `json:"home"`         // 首页配置
	GeneralAgent GeneralAgentConfig `json:"generalAgent"` // 通用智能体配置
}

type LoginConfig struct {
	LoginBgPath string `json:"loginBgPath"` // 登录页背景图路径
	LogoPath    string `json:"logoPath"`    // 登录页logo路径
	WelcomeText string `json:"welcomeText"` // 登录页欢迎词
	ButtonColor string `json:"buttonColor"` // 登录按钮颜色
}

type TabConfig struct {
	LogoPath string `json:"logoPath"` // 标签页logo路径
	Title    string `json:"title"`    // 标签页标题
}

type HomeConfig struct {
	LogoPath string `json:"logoPath"` // 平台logo路径
	Name     string `json:"name"`     // 平台名称
	BgColor  string `json:"bgColor"`  // 平台背景颜色
}

type GeneralAgentConfig struct {
	GeneralAgentIconPath string `json:"generalAgentIconPath"` // 通用智能体图标路径
	GeneralAgentWelcome  string `json:"generalAgentWelcome"`  // 通用智能体欢迎语
	GeneralAgentMenuName string `json:"generalAgentMenuName"` // 通用智能体菜单名称
}

type ClientStatistic struct {
	Overview ClientOverView `json:"overview"` // 统计面板
	Trend    ClientTrend    `json:"trend"`    // 统计趋势
}

type ClientOverView struct {
	Cumulative ClientOverviewItem `json:"cumulative"` // 累计客户端
	New        ClientOverviewItem `json:"new"`        // 新增客户端
	Active     ClientOverviewItem `json:"active"`     // 日活客户端
}

type ClientOverviewItem struct {
	Value            float32 `json:"value"`            // 数量
	PeriodOverPeriod float32 `json:"periodOverPeriod"` // 环比上周期百分比
}

type ClientTrend struct {
	Client StatisticChart `json:"client"`
}

type StatisticChart struct {
	Name  string               `json:"name"`  // 统计表名字
	Lines []StatisticChartLine `json:"lines"` // 统计表中线段集合
}

type StatisticChartLine struct {
	Name  string                   `json:"name"`  // 线段名字
	Items []StatisticChartLineItem `json:"items"` // 线段横纵坐标值
}

type StatisticChartLineItem struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}
