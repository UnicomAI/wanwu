package response

type ClientCumulative struct {
	Total int32 `json:"total"` // 累计客户端数量 [EN] Cumulative number of clients
}

type ClientStatistic struct {
	Overview ClientOverView `json:"overview"` // 统计面板 [EN] Statistics panel
	Trend    ClientTrend    `json:"trend"`    // 统计趋势 [EN] statistical trends
}

type ClientOverView struct {
	CumulativeClient StatisticOverviewItem `json:"cumulativeClient"` // 累计客户端 [EN] Cumulative clients
	AdditionClient   StatisticOverviewItem `json:"additionClient"`   // 新增客户端 [EN] Add new client
	ActiveClient     StatisticOverviewItem `json:"activeClient"`     // 日活客户端 [EN] Daily active client
	Browse           StatisticOverviewItem `json:"browse"`           // 浏览量 [EN] Views
}

type ClientTrend struct {
	Client StatisticChart `json:"client"` // 客户端 [EN] client
	Browse StatisticChart `json:"browse"` // 浏览量 [EN] Views
}

type StatisticOverviewItem struct {
	Value            float32 `json:"value"`            // 数量 [EN] quantity
	PeriodOverPeriod float32 `json:"periodOverPeriod"` // 环比上周期百分比 [EN] Period-on-month percentage
}

type StatisticChart struct {
	TableName string               `json:"tableName"` // 统计表名字 [EN] Statistics table name
	Lines     []StatisticChartLine `json:"lines"`     // 统计表中线段集合 [EN] Collection of line segments in statistical tables
}

type StatisticChartLine struct {
	LineName string                   `json:"lineName"` // 线段名字 [EN] Line segment name
	Items    []StatisticChartLineItem `json:"items"`    // 线段横纵坐标值 [EN] The horizontal and vertical coordinate values ​​of the line segment
}

type StatisticChartLineItem struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}
