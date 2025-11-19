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
	Login LoginConfig `json:"login"` // Login page configuration
	Tab   TabConfig   `json:"tab"`   // Tab configuration
	Home  HomeConfig  `json:"home"`  // Home page configuration
}

type LoginConfig struct {
	LoginBgPath string `json:"loginBgPath"` // Login page background image path
	LogoPath    string `json:"logoPath"`    // Login page logo path
	WelcomeText string `json:"welcomeText"` // Login page welcome message
	ButtonColor string `json:"buttonColor"` // Login button color
}

type TabConfig struct {
	LogoPath string `json:"logoPath"` // Tab logo path
	Title    string `json:"title"`    // Tab title
}

type HomeConfig struct {
	LogoPath string `json:"logoPath"` // Platform logo path
	Name     string `json:"name"`     // Platform name
	BgColor  string `json:"bgColor"`  // Platform background color
}

type ClientStatistic struct {
	Overview ClientOverView `json:"overview"` // Statistics panel
	Trend    ClientTrend    `json:"trend"`    // statistical trends
}

type ClientOverView struct {
	Cumulative ClientOverviewItem `json:"cumulative"` // Cumulative clients
	New        ClientOverviewItem `json:"new"`        // Add new client
	Active     ClientOverviewItem `json:"active"`     // Daily active client
}

type ClientOverviewItem struct {
	Value            float32 `json:"value"`            // quantity
	PeriodOverPeriod float32 `json:"periodOverPeriod"` // Period-on-month percentage
}

type ClientTrend struct {
	Client StatisticChart `json:"client"`
}

type StatisticChart struct {
	Name  string               `json:"name"`  // Statistics table name
	Lines []StatisticChartLine `json:"lines"` // Collection of line segments in statistical tables
}

type StatisticChartLine struct {
	Name  string                   `json:"name"`  // Line segment name
	Items []StatisticChartLineItem `json:"items"` // The horizontal and vertical coordinate values ​​of the line segment
}

type StatisticChartLineItem struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}
