package app

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

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
		calls, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			return nil, err
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
		count, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			return nil, err
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
		average, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			return nil, err
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
		calls, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			return nil, err
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
		count, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			return nil, err
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
		average, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			return nil, err
		}
		entry.Average = average
	}
	return data, nil
}

// sum(increase(mattermost_api_time_sum[%s]) and increase(mattermost_api_time_count[%s]) > 0) by (handler)

// func (a *App) RunQuery(query, firstOffset, secondOffset, length string, scaleBy string) (map[string]*DBEntry, error) {
// 	firstResultsMap := map[string]*DBEntry{}
// 	secondResultsMap := map[string]*DBEntry{}

// 	// firstQuery := replacePlaceholders(query, firstOffset, length)
// 	secondQuery := replacePlaceholders(query, secondOffset, length)

// 	timeRange := length
// 	queryResult, err := a.reportQueryClient.Query(fmt.Sprintf("sum(increase(mattermost_db_store_time_count[%s]) > 0) by (method)", timeRange))
// 	//
// 	// fmt.Println(firstQuery)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// fmt.Printf("%+v\n", queryResult.Data.Result)

// 	for _, r := range queryResult.Data.Result {
// 		calls, err := strconv.ParseFloat(r.Value[1].(string), 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		firstResultsMap[r.Metric["method"]] = &DBEntry{TotalTime: calls, Method: r.Metric["method"]}
// 	}

// 	queryResult, err = a.reportQueryClient.Query(secondQuery)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, r := range queryResult.Data.Result {
// 		calls, err := strconv.ParseFloat(r.Value[1].(string), 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		secondResultsMap[r.Metric["method"]] = &DBEntry{TotalTime: calls, Method: r.Metric["method"]}
// 	}

// 	finalResult := map[string]*DBEntry{}
// 	for handlerName, v1 := range firstResultsMap {
// 		v2 := secondResultsMap[handlerName]
// 		if v2 == nil {
// 			continue
// 		}

// 		if v1.TotalTime == 0 || v2.TotalTime == 0 {
// 			finalResult[handlerName] = &DBEntry{
// 				Method:    handlerName,
// 				TotalTime: 0,
// 			}
// 			continue
// 		}

// 		diff := v1.TotalTime - v2.TotalTime
// 		// percentDiff := diff
// 		percentDiff := (diff / v2.TotalTime) * 100

// 		finalResult[handlerName] = &DBEntry{
// 			Method:    handlerName,
// 			TotalTime: percentDiff,
// 		}
// 	}

// 	return finalResult, nil
// }

func replacePlaceholders(query, offset, length string) string {
	// Replace placeholders in the query with actual values.
	query = strings.Replace(query, "{{.Length}}", length, -1)
	query = strings.Replace(query, "{{.Offset}}", offset, -1)
	return query
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

	if sortBy == "" {
		sortBy = SortCategoryTotalTime
	}

	biggestIncreases := make([]*FullDBEntry, 0, len(data))
	for _, d := range data {
		biggestIncreases = append(biggestIncreases, d)
	}

	sort.Slice(biggestIncreases, func(i, j int) bool {
		if sortBy == SortCategoryTotalTime {
			return biggestIncreases[i].TotalTimeDifferencePercent > biggestIncreases[j].TotalTimeDifferencePercent
		} else if sortBy == SortCategoryAverageTime {
			return biggestIncreases[i].AverageDifferencePercent > biggestIncreases[j].AverageDifferencePercent
		} else if sortBy == SortCategoryCount {
			return biggestIncreases[i].CountDifferencePercent > biggestIncreases[j].CountDifferencePercent
		}
		panic("unreachable code")
	})

	biggestDecreases := make([]*FullDBEntry, 0, len(data))
	for _, d := range data {
		biggestDecreases = append(biggestDecreases, d)
	}

	sort.Slice(biggestDecreases, func(i, j int) bool {
		if sortBy == SortCategoryTotalTime {
			return biggestDecreases[i].TotalTimeDifferencePercent < biggestDecreases[j].TotalTimeDifferencePercent
		} else if sortBy == SortCategoryAverageTime {
			return biggestDecreases[i].AverageDifferencePercent < biggestDecreases[j].AverageDifferencePercent
		} else if sortBy == SortCategoryCount {
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
	criteria := runFlags.Sort
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

	if criteria == "" {
		criteria = SortCategoryTotalTime
	}

	biggestIncreases := make([]*FullAPIEntry, 0, len(data))
	for _, d := range data {
		biggestIncreases = append(biggestIncreases, d)
	}

	sort.Slice(biggestIncreases, func(i, j int) bool {
		if criteria == SortCategoryTotalTime {
			return biggestIncreases[i].TotalTimeDifferencePercent > biggestIncreases[j].TotalTimeDifferencePercent
		} else if criteria == SortCategoryAverageTime {
			return biggestIncreases[i].AverageDifferencePercent > biggestIncreases[j].AverageDifferencePercent
		} else if criteria == SortCategoryCount {
			return biggestIncreases[i].CountDifferencePercent > biggestIncreases[j].CountDifferencePercent
		}
		panic("unreachable code")
	})

	biggestDecreases := make([]*FullAPIEntry, 0, len(data))
	for _, d := range data {
		biggestDecreases = append(biggestDecreases, d)
	}

	sort.Slice(biggestDecreases, func(i, j int) bool {
		if criteria == SortCategoryTotalTime {
			return biggestDecreases[i].TotalTimeDifferencePercent < biggestDecreases[j].TotalTimeDifferencePercent
		} else if criteria == SortCategoryAverageTime {
			return biggestDecreases[i].AverageDifferencePercent < biggestDecreases[j].AverageDifferencePercent
		} else if criteria == SortCategoryCount {
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
