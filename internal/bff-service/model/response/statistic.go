package response

type StatisticUserName struct {
	UserID   string `json:"userId"`
	UserName string `json:"username"`
}

type ClientStatistic struct {
	Overview ClientOverView `json:"overview"` // 统计面板
	Trend    ClientTrend    `json:"trend"`    // 统计趋势
}

type ClientOverView struct {
	CumulativeClient StatisticOverviewItem `json:"cumulativeClient"` // 累计客户端
	AdditionClient   StatisticOverviewItem `json:"additionClient"`   // 新增客户端
	ActiveClient     StatisticOverviewItem `json:"activeClient"`     // 日活客户端
	Browse           StatisticOverviewItem `json:"browse"`           // 浏览量
}

type ClientTrend struct {
	Client StatisticChart `json:"client"` // 客户端
	Browse StatisticChart `json:"browse"` // 浏览量
}

type StatisticOverviewItem struct {
	Value            float32 `json:"value"`            // 数量
	PeriodOverPeriod float32 `json:"periodOverPeriod"` // 环比上周期百分比
}

type StatisticChart struct {
	TableName string               `json:"tableName"` // 统计表名字
	Lines     []StatisticChartLine `json:"lines"`     // 统计表中线段集合
}

type StatisticChartLine struct {
	LineName string                   `json:"lineName"` // 线段名字
	Items    []StatisticChartLineItem `json:"items"`    // 线段横纵坐标值
}

type StatisticChartLineItem struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}
