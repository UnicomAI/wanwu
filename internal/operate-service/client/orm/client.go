package orm

import (
	"log"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"gorm.io/gorm"
)

type SystemCustomKey string
type SystemCustomMode string

const (
	SystemCustomTabKey   SystemCustomKey = "system_custom_tab"
	SystemCustomLoginKey SystemCustomKey = "system_custom_login"
	SystemCustomHomeKey  SystemCustomKey = "system_custom_home"
)
const (
	SystemCustomModeLight SystemCustomMode = "light"
	SystemCustomModeDark  SystemCustomMode = "dark"
)

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) (*Client, error) {
	// auto migrate
	if err := db.AutoMigrate(
		model.SystemCustom{},
		model.ClientRecord{},
		model.ClientDailyStats{},
	); err != nil {
		return nil, err
	}
	// init corn
	if err := CronInit(db); err != nil {
		log.Fatalf("init corn failed, err: %v", err)
	}
	return &Client{
		db: db,
	}, nil
}

func toErrStatus(key string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: key,
		Args:    args,
	}
}

type SystemCustom struct {
	Login LoginConfig `json:"login"` // 登录页配置 [EN] Login page configuration
	Tab   TabConfig   `json:"tab"`   // 标签页配置 [EN] Tab configuration
	Home  HomeConfig  `json:"home"`  // 首页配置 [EN] Home page configuration
}

type LoginConfig struct {
	LoginBgPath string `json:"loginBgPath"` // 登录页背景图路径 [EN] Login page background image path
	LogoPath    string `json:"logoPath"`    // 登录页logo路径 [EN] Login page logo path
	WelcomeText string `json:"welcomeText"` // 登录页欢迎词 [EN] Login page welcome message
	ButtonColor string `json:"buttonColor"` // 登录按钮颜色 [EN] Login button color
}

type TabConfig struct {
	LogoPath string `json:"logoPath"` // 标签页logo路径 [EN] Tab logo path
	Title    string `json:"title"`    // 标签页标题 [EN] Tab title
}

type HomeConfig struct {
	LogoPath string `json:"logoPath"` // 平台logo路径 [EN] Platform logo path
	Name     string `json:"name"`     // 平台名称 [EN] Platform name
	BgColor  string `json:"bgColor"`  // 平台背景颜色 [EN] Platform background color
}

type ClientStatistic struct {
	Overview ClientOverView `json:"overview"` // 统计面板 [EN] Statistics panel
	Trend    ClientTrend    `json:"trend"`    // 统计趋势 [EN] statistical trends
}

type ClientOverView struct {
	Cumulative ClientOverviewItem `json:"cumulative"` // 累计客户端 [EN] Cumulative clients
	New        ClientOverviewItem `json:"new"`        // 新增客户端 [EN] Add new client
	Active     ClientOverviewItem `json:"active"`     // 日活客户端 [EN] Daily active client
}

type ClientOverviewItem struct {
	Value            float32 `json:"value"`            // 数量 [EN] quantity
	PeriodOverPeriod float32 `json:"periodOverPeriod"` // 环比上周期百分比 [EN] Period-on-month percentage
}

type ClientTrend struct {
	Client StatisticChart `json:"client"`
}

type StatisticChart struct {
	Name  string               `json:"name"`  // 统计表名字 [EN] Statistics table name
	Lines []StatisticChartLine `json:"lines"` // 统计表中线段集合 [EN] Collection of line segments in statistical tables
}

type StatisticChartLine struct {
	Name  string                   `json:"name"`  // 线段名字 [EN] Line segment name
	Items []StatisticChartLineItem `json:"items"` // 线段横纵坐标值 [EN] The horizontal and vertical coordinate values ​​of the line segment
}

type StatisticChartLineItem struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}
