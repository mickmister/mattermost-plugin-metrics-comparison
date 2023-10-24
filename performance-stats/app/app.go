package app

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-perf-stats-cli/prometheus"
)

type DBEntry struct {
	Method    string
	TotalTime float64
	Count     float64
	Average   float64
}

type APIEntry struct {
	Handler   string
	TotalTime float64
	Count     float64
	Average   float64
}

type App struct {
	client *prometheus.Client
}

func New(endpoint string) *App {
	client := prometheus.New(endpoint)
	return &App{
		client: client,
	}
}

func (a *App) GetDBMetrics(timeRange string) (map[string]*DBEntry, error) {
	data := map[string]*DBEntry{}
	totalTimeMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_db_store_time_sum[%s]) and increase(mattermost_db_store_time_count[%s]) > 0) by (method)", timeRange, timeRange))
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

	callsMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_db_store_time_count[%s]) > 0) by (method)", timeRange))
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
	averageMetrics, err := a.client.Query(fmt.Sprintf("(sum(increase(mattermost_db_store_time_count[2h]) > 0) by (method) - sum(increase(mattermost_db_store_time_count[2h] offset 30m) > 0) by (method)) / sum(increase(mattermost_db_store_time_count[2h] offset 30m) > 0) by (method) * 100"))
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

func (a *App) GetAPIMetrics(timeRange string) (map[string]*APIEntry, error) {
	data := map[string]*APIEntry{}
	totalTimeMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_api_time_sum[%s]) and increase(mattermost_api_time_count[%s]) > 0) by (handler)", timeRange, timeRange))
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

	callsMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_api_time_count[%s]) > 0) by (handler)", timeRange))
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
	averageMetrics, err := a.client.Query(fmt.Sprintf("(sum(increase(mattermost_api_time_sum[%s])) by (handler) / sum(increase(mattermost_api_time_count[%s]) > 0) by (handler))", timeRange, timeRange))
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

func (a *App) RunQueryWithMockData(query, firstOffset, secondOffset, length string, scaleBy, fname string) (map[string]*DBEntry, error) {
	g, err := getGrafanaResponseFromJSON(fname)
	if err != nil {
		return nil, err
	}

	data := map[string]*DBEntry{}

	for _, frame := range g.Results.A.Frames {
		totalTime := frame.Data.Values[1][0]
		method := frame.Schema.Fields[1].Labels.Method
		data[method] = &DBEntry{TotalTime: totalTime, Method: method}
	}

	for _, frame := range g.Results.B.Frames {
		calls := frame.Data.Values[1][0]
		method := frame.Schema.Fields[1].Labels.Method
		data[method].Count = calls
	}

	for _, frame := range g.Results.C.Frames {
		average := frame.Data.Values[1][0]
		method := frame.Schema.Fields[1].Labels.Method
		data[method].Average = average
	}

	return data, nil
}

func (a *App) RunQuery(query, firstOffset, secondOffset, length string, scaleBy string) (map[string]*DBEntry, error) {
	firstResultsMap := map[string]*DBEntry{}
	secondResultsMap := map[string]*DBEntry{}

	// firstQuery := replacePlaceholders(query, firstOffset, length)
	secondQuery := replacePlaceholders(query, secondOffset, length)

	timeRange := length
	queryResult, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_db_store_time_count[%s]) > 0) by (method)", timeRange))
	//
	// fmt.Println(firstQuery)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%+v\n", queryResult.Data.Result)

	for _, r := range queryResult.Data.Result {
		calls, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			return nil, err
		}
		firstResultsMap[r.Metric["method"]] = &DBEntry{TotalTime: calls, Method: r.Metric["method"]}
	}

	queryResult, err = a.client.Query(secondQuery)
	if err != nil {
		return nil, err
	}

	for _, r := range queryResult.Data.Result {
		calls, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			return nil, err
		}
		secondResultsMap[r.Metric["method"]] = &DBEntry{TotalTime: calls, Method: r.Metric["method"]}
	}

	finalResult := map[string]*DBEntry{}
	for handlerName, v1 := range firstResultsMap {
		v2 := secondResultsMap[handlerName]
		if v2 == nil {
			continue
		}

		if v1.TotalTime == 0 || v2.TotalTime == 0 {
			finalResult[handlerName] = &DBEntry{
				Method:    handlerName,
				TotalTime: 0,
			}
			continue
		}

		diff := v1.TotalTime - v2.TotalTime
		// percentDiff := diff
		percentDiff := (diff / v2.TotalTime) * 100

		finalResult[handlerName] = &DBEntry{
			Method:    handlerName,
			TotalTime: percentDiff,
		}
	}

	return finalResult, nil

	// totalTimeMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_api_time_sum[%s]) and increase(mattermost_api_time_count[%s]) > 0) by (handler)", timeRange, timeRange))
	// if err != nil {
	// 	return nil, err
	// }
	// for _, r := range totalTimeMetrics.Data.Result {
	// 	calls, err := strconv.ParseFloat(r.Value[1].(string), 64)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	data[r.Metric["handler"]] = &APIEntry{TotalTime: calls, Handler: r.Metric["handler"]}
	// }

	// callsMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_api_time_count[%s]) > 0) by (handler)", timeRange))
	// if err != nil {
	// 	return nil, err
	// }
	// for _, r := range callsMetrics.Data.Result {
	// 	entry := data[r.Metric["handler"]]
	// 	count, err := strconv.ParseFloat(r.Value[1].(string), 64)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	entry.Count = count
	// }
	// averageMetrics, err := a.client.Query(fmt.Sprintf("(sum(increase(mattermost_api_time_sum[%s])) by (handler) / sum(increase(mattermost_api_time_count[%s]) > 0) by (handler))", timeRange, timeRange))
	// if err != nil {
	// 	return nil, err
	// }
	// for _, r := range averageMetrics.Data.Result {
	// 	entry := data[r.Metric["handler"]]
	// 	average, err := strconv.ParseFloat(r.Value[1].(string), 64)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	entry.Average = average
	// }
	// return data, nil
}

func replacePlaceholders(query, offset, length string) string {
	// Replace placeholders in the query with actual values.
	query = strings.Replace(query, "{{.Length}}", length, -1)
	query = strings.Replace(query, "{{.Offset}}", offset, -1)
	return query
}

func ComputeReport(data1, data2 map[string]*DBEntry, criteria string, limit int) (biggestIncreases, biggestDecreases []*DBEntry) {
	data := map[string]*DBEntry{}

	for method, value1 := range data1 {
		value2, ok := data2[method]
		if !ok {
			// fmt.Println("missing value for method " + method)
			continue
		}

		total := (value2.TotalTime - value1.TotalTime) / value1.TotalTime
		total = math.Ceil(total*10000) / 100

		data[method] = &DBEntry{
			Method:    method,
			TotalTime: total,
			Count:     (value2.Count - value1.Count) / value1.Count,
			Average:   (value2.Average - value1.Average) / value2.Average,
		}
	}

	// if criteria != "total-time" && criteria != "average-time" && criteria != "count" {
	// 	fmt.Println("Invalid criteria")
	// 	return
	// }

	if criteria == "" {
		criteria = "total-time"
	}

	biggestIncreases = make([]*DBEntry, 0, len(data))
	for _, d := range data {
		biggestIncreases = append(biggestIncreases, d)
	}

	sort.Slice(biggestIncreases, func(i, j int) bool {
		// return dataList[i].TotalTime < dataList[j].TotalTime
		if criteria == "total-time" {
			return biggestIncreases[i].TotalTime > biggestIncreases[j].TotalTime
		} else if criteria == "average-time" {
			return biggestIncreases[i].Average > biggestIncreases[j].Average
		} else if criteria == "count" {
			return biggestIncreases[i].Count > biggestIncreases[j].Count
		}
		panic("unreachable code")
	})

	biggestDecreases = make([]*DBEntry, 0, len(data))
	for _, d := range data {
		biggestDecreases = append(biggestDecreases, d)
	}

	sort.Slice(biggestDecreases, func(i, j int) bool {
		// return dataList[i].TotalTime < dataList[j].TotalTime
		if criteria == "total-time" {
			return biggestDecreases[i].TotalTime > biggestDecreases[j].TotalTime
		} else if criteria == "average-time" {
			return biggestDecreases[i].Average > biggestDecreases[j].Average
		} else if criteria == "count" {
			return biggestDecreases[i].Count > biggestDecreases[j].Count
		}
		panic("unreachable code")
	})

	if limit > len(biggestDecreases) {
		limit = len(biggestDecreases)
	}

	return biggestIncreases[:limit], biggestDecreases[:limit]
}
