package response

type ClientCumulative struct {
	Total int32 `json:"total"` // Cumulative number of clients
}

type ClientStatistic struct {
	Overview ClientOverView `json:"overview"` // Statistics panel
	Trend    ClientTrend    `json:"trend"`    // statistical trends
}

type ClientOverView struct {
	CumulativeClient StatisticOverviewItem `json:"cumulativeClient"` // Cumulative clients
	AdditionClient   StatisticOverviewItem `json:"additionClient"`   // Add new client
	ActiveClient     StatisticOverviewItem `json:"activeClient"`     // Daily active client
	Browse           StatisticOverviewItem `json:"browse"`           // Views
}

type ClientTrend struct {
	Client StatisticChart `json:"client"` // client
	Browse StatisticChart `json:"browse"` // Views
}

type StatisticOverviewItem struct {
	Value            float32 `json:"value"`            // quantity
	PeriodOverPeriod float32 `json:"periodOverPeriod"` // Period-on-month percentage
}

type StatisticChart struct {
	TableName string               `json:"tableName"` // Statistics table name
	Lines     []StatisticChartLine `json:"lines"`     // Collection of line segments in statistical tables
}

type StatisticChartLine struct {
	LineName string                   `json:"lineName"` // Line segment name
	Items    []StatisticChartLineItem `json:"items"`    // The horizontal and vertical coordinate values ​​of the line segment
}

type StatisticChartLineItem struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}
