package app

type ReportType string

const (
	ReportTypeAPIHandlerComparison    ReportType = "api-comparison"
	ReportTypeDBStoreMethodComparison ReportType = "db-comparison"
	ReportTypeAPIHandlerSnapshot      ReportType = "api-snapshot"
	ReportTypeDBStoreMethodSnapshot   ReportType = "db-snapshot"
)

type SortCategory string

const (
	SortCategoryTotalTime   SortCategory = "total-time"
	SortCategoryCount       SortCategory = "count"
	SortCategoryAverageTime SortCategory = "average-time"
)

type RunReportFlags struct {
	Query      string
	ReportType ReportType
	First      string
	Second     string
	Length     string
	Sort       SortCategory
	ScaleBy    string
	Limit      int
}

const defaultReportType = ReportTypeAPIHandlerComparison
const defaultQuery = "sum(increase(mattermost_db_store_time_sum[{{.Length}}]) and increase(mattermost_db_store_time_count[{{.Offset}}]) > 0) by (method)"
const defaultFirst = "1m"
const defaultSecond = "1d"
const defaultLength = "1d"
const defaultScaleBy = "post"
const defaultSort = SortCategoryTotalTime
const defaultLimit = 20

func GetDefaultReportFlags() RunReportFlags {
	return RunReportFlags{
		ReportType: defaultReportType,
		Query:      defaultQuery,
		First:      defaultFirst,
		Second:     defaultSecond,
		Length:     defaultLength,
		ScaleBy:    defaultScaleBy,
		Sort:       defaultSort,
		Limit:      defaultLimit,
	}
}
