package app

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/mattermost/mattermost-plugin-metrics-comparison/server/prometheus"
	"github.com/pkg/errors"
)

type DBEntry struct {
	Method    string
	TotalTime float64
	Count     float64
	Average   float64
}

type FullDBEntry struct {
	Method                     string
	Timeframe1                 *DBEntry
	Timeframe2                 *DBEntry
	TotalTimeDifferencePercent float64
	CountDifferencePercent     float64
	AverageDifferencePercent   float64
}

type APIEntry struct {
	Handler   string
	TotalTime float64
	Count     float64
	Average   float64
}

type FullAPIEntry struct {
	Handler                    string
	Timeframe1                 *APIEntry
	Timeframe2                 *APIEntry
	TotalTimeDifferencePercent float64
	CountDifferencePercent     float64
	AverageDifferencePercent   float64
}

type App struct {
	reportQueryClient *prometheus.ReportQueryClient
}

type DBStoreReport struct {
	BiggestIncreases []*FullDBEntry
	BiggestDecreases []*FullDBEntry
	RunFlags         RunReportFlags
}

type APIHandlerReport struct {
	BiggestIncreases []*FullAPIEntry
	BiggestDecreases []*FullAPIEntry
	RunFlags         RunReportFlags
}

func New(client *prometheus.ReportQueryClient) *App {
	return &App{
		reportQueryClient: client,
	}
}

func (a *App) GetDBMetrics(offset, length string) (map[string]*DBEntry, error) {
	data := map[string]*DBEntry{}
	totalTimeMetrics, err := a.reportQueryClient.QueryDBStoreTotalTime(offset, length)
	// totalTimeMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_db_store_time_sum[%s]) and increase(mattermost_db_store_time_count[%s]) > 0) by (method)", timeRange, timeRange))
	if err != nil {
		return nil, err
	}
	for _, r := range totalTimeMetrics.Data.Result {
		calls, err2 := strconv.ParseFloat(r.Value[1].(string), 64)
		if err2 != nil {
			return nil, err2
		}
		data[r.Metric["method"]] = &DBEntry{TotalTime: calls, Method: r.Metric["method"]}
	}

	callsMetrics, err := a.reportQueryClient.QueryDBStoreCount(offset, length)
	// callsMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_db_store_time_count[%s]) > 0) by (method)", timeRange))
	if err != nil {
		return nil, err
	}
	for _, r := range callsMetrics.Data.Result {
		entry := data[r.Metric["method"]]
		count, err2 := strconv.ParseFloat(r.Value[1].(string), 64)
		if err2 != nil {
			return nil, err2
		}
		entry.Count = count
	}
	averageMetrics, err := a.reportQueryClient.QueryDBStoreAverage(offset, length)
	// averageMetrics, err := a.client.Query(fmt.Sprintf("(sum(increase(mattermost_db_store_time_count[2h]) > 0) by (method) - sum(increase(mattermost_db_store_time_count[2h] offset 30m) > 0) by (method)) / sum(increase(mattermost_db_store_time_count[2h] offset 30m) > 0) by (method) * 100"))
	if err != nil {
		return nil, err
	}
	for _, r := range averageMetrics.Data.Result {
		entry := data[r.Metric["method"]]
		average, err2 := strconv.ParseFloat(r.Value[1].(string), 64)
		if err2 != nil {
			return nil, err2
		}
		if average == 0 {
			fmt.Println("zero")
		}
		entry.Average = average
	}
	return data, nil
}

func (a *App) GetAPIMetrics(offset, length string) (map[string]*APIEntry, error) {
	data := map[string]*APIEntry{}
	totalTimeMetrics, err := a.reportQueryClient.QueryAPIHandlerTotalTime(offset, length)
	// totalTimeMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_api_time_sum[%s]) and increase(mattermost_api_time_count[%s]) > 0) by (handler)", timeRange, timeRange))
	if err != nil {
		return nil, err
	}
	for _, r := range totalTimeMetrics.Data.Result {
		calls, err2 := strconv.ParseFloat(r.Value[1].(string), 64)
		if err2 != nil {
			return nil, err2
		}
		data[r.Metric["handler"]] = &APIEntry{TotalTime: calls, Handler: r.Metric["handler"]}
	}

	callsMetrics, err := a.reportQueryClient.QueryAPIHandlerCount(offset, length)
	// callsMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_api_time_count[%s]) > 0) by (handler)", timeRange))
	if err != nil {
		return nil, err
	}
	for _, r := range callsMetrics.Data.Result {
		entry := data[r.Metric["handler"]]
		count, err2 := strconv.ParseFloat(r.Value[1].(string), 64)
		if err2 != nil {
			return nil, err2
		}
		entry.Count = count
	}
	averageMetrics, err := a.reportQueryClient.QueryAPIHandlerAverage(offset, length)
	// averageMetrics, err := a.client.Query(fmt.Sprintf("(sum(increase(mattermost_api_time_sum[%s])) by (handler) / sum(increase(mattermost_api_time_count[%s]) > 0) by (handler))", timeRange, timeRange))
	if err != nil {
		return nil, err
	}
	for _, r := range averageMetrics.Data.Result {
		entry := data[r.Metric["handler"]]
		average, err2 := strconv.ParseFloat(r.Value[1].(string), 64)
		if err2 != nil {
			return nil, err2
		}
		entry.Average = average
	}
	return data, nil
}

func (a *App) RunDBComparisonReport(runFlags RunReportFlags) (*DBStoreReport, error) {
	data1, err := a.GetDBMetrics(runFlags.First, runFlags.Length)
	if err != nil {
		return nil, errors.Wrap(err, "error getting DB metrics for first time frame")
	}

	data2, err := a.GetDBMetrics(runFlags.Second, runFlags.Length)
	if err != nil {
		return nil, errors.Wrap(err, "error getting DB metrics for second time frame")
	}

	data := map[string]*FullDBEntry{}
	sortBy := runFlags.Sort
	limit := runFlags.Limit

	for method, value1 := range data1 {
		value2, ok := data2[method]
		if !ok {
			// fmt.Println("missing value for method " + method)
			continue
		}

		data[method] = &FullDBEntry{
			Method:                     method,
			Timeframe1:                 value1,
			Timeframe2:                 value2,
			TotalTimeDifferencePercent: numToPercent((value2.TotalTime - value1.TotalTime) / value1.TotalTime),
			CountDifferencePercent:     numToPercent((value2.Count - value1.Count) / value1.Count),
			AverageDifferencePercent:   numToPercent((value2.Average - value1.Average) / value2.Average),
		}
	}

	biggestIncreases := make([]*FullDBEntry, 0, len(data))
	for _, d := range data {
		biggestIncreases = append(biggestIncreases, d)
	}

	sort.Slice(biggestIncreases, func(i, j int) bool {
		switch sortBy {
		case SortCategoryTotalTime:
			return biggestIncreases[i].TotalTimeDifferencePercent > biggestIncreases[j].TotalTimeDifferencePercent
		case SortCategoryAverageTime:
			return biggestIncreases[i].AverageDifferencePercent > biggestIncreases[j].AverageDifferencePercent
		case SortCategoryCount:
			return biggestIncreases[i].CountDifferencePercent > biggestIncreases[j].CountDifferencePercent
		}
		panic("unreachable code")
	})

	biggestDecreases := make([]*FullDBEntry, 0, len(data))
	for _, d := range data {
		biggestDecreases = append(biggestDecreases, d)
	}

	sort.Slice(biggestDecreases, func(i, j int) bool {
		switch sortBy {
		case SortCategoryTotalTime:
			return biggestDecreases[i].TotalTimeDifferencePercent < biggestDecreases[j].TotalTimeDifferencePercent
		case SortCategoryAverageTime:
			return biggestDecreases[i].AverageDifferencePercent < biggestDecreases[j].AverageDifferencePercent
		case SortCategoryCount:
			return biggestDecreases[i].CountDifferencePercent < biggestDecreases[j].CountDifferencePercent
		}
		panic("unreachable code")
	})

	if limit > len(biggestDecreases) {
		limit = len(biggestDecreases)
	}

	return &DBStoreReport{
		BiggestIncreases: biggestIncreases[:limit],
		BiggestDecreases: biggestDecreases[:limit],
		RunFlags:         runFlags,
	}, nil
}

func (a *App) RunAPIComparisonReport(runFlags RunReportFlags) (*APIHandlerReport, error) {
	data1, err := a.GetAPIMetrics(runFlags.First, runFlags.Length)
	if err != nil {
		return nil, errors.Wrap(err, "error getting DB metrics for first time frame")
	}

	data2, err := a.GetAPIMetrics(runFlags.Second, runFlags.Length)
	if err != nil {
		return nil, errors.Wrap(err, "error getting DB metrics for second time frame")
	}

	data := map[string]*FullAPIEntry{}
	sortBy := runFlags.Sort
	limit := runFlags.Limit

	for method, value1 := range data1 {
		value2, ok := data2[method]
		if !ok {
			// fmt.Println("missing value for method " + method)
			continue
		}

		data[method] = &FullAPIEntry{
			Handler:                    method,
			Timeframe1:                 value1,
			Timeframe2:                 value2,
			TotalTimeDifferencePercent: numToPercent((value2.TotalTime - value1.TotalTime) / value1.TotalTime),
			CountDifferencePercent:     numToPercent((value2.Count - value1.Count) / value1.Count),
			AverageDifferencePercent:   numToPercent((value2.Average - value1.Average) / value2.Average),
		}
	}

	biggestIncreases := make([]*FullAPIEntry, 0, len(data))
	for _, d := range data {
		biggestIncreases = append(biggestIncreases, d)
	}

	sort.Slice(biggestIncreases, func(i, j int) bool {
		switch sortBy {
		case SortCategoryTotalTime:
			return biggestIncreases[i].TotalTimeDifferencePercent > biggestIncreases[j].TotalTimeDifferencePercent
		case SortCategoryAverageTime:
			return biggestIncreases[i].AverageDifferencePercent > biggestIncreases[j].AverageDifferencePercent
		case SortCategoryCount:
			return biggestIncreases[i].CountDifferencePercent > biggestIncreases[j].CountDifferencePercent
		}
		panic("unreachable code")
	})

	biggestDecreases := make([]*FullAPIEntry, 0, len(data))
	for _, d := range data {
		biggestDecreases = append(biggestDecreases, d)
	}

	sort.Slice(biggestDecreases, func(i, j int) bool {
		switch sortBy {
		case SortCategoryTotalTime:
			return biggestDecreases[i].TotalTimeDifferencePercent < biggestDecreases[j].TotalTimeDifferencePercent
		case SortCategoryAverageTime:
			return biggestDecreases[i].AverageDifferencePercent < biggestDecreases[j].AverageDifferencePercent
		case SortCategoryCount:
			return biggestDecreases[i].CountDifferencePercent < biggestDecreases[j].CountDifferencePercent
		}
		panic("unreachable code")
	})

	if limit > len(biggestDecreases) {
		limit = len(biggestDecreases)
	}

	return &APIHandlerReport{
		BiggestIncreases: biggestIncreases[:limit],
		BiggestDecreases: biggestDecreases[:limit],
		RunFlags:         runFlags,
	}, nil
}

func numToPercent(num float64) float64 {
	return math.Ceil(num*10000) / 100
}
