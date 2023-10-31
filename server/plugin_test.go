package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-plugin-metrics-comparison/server/prometheus"
	mock_prometheus "github.com/mattermost/mattermost-plugin-metrics-comparison/server/prometheus/mock_client"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHTTP(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	plugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()
	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)
	bodyString := string(bodyBytes)

	assert.Equal("Hello, world!", bodyString)
}

func TestExecuteCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prom := mock_prometheus.NewMockPrometheusClient(ctrl)
	api := &plugintest.API{}

	reportClient := prometheus.NewReportClient(prom)

	p := Plugin{reportQueryClient: reportClient}
	p.API = api

	args := &model.CommandArgs{
		Command: "/performance-report run --first 1d --second 3d --length 1d",
	}

	mockResponses := []struct {
		query    string
		response *prometheus.Response
	}{
		{
			"sum(increase(mattermost_api_time_sum[1d] offset 1d) and increase(mattermost_api_time_count[1d] offset 1d) > 0) by (handler)",
			makeFakeResponseForAPIHandler([]mockAPIHandlerQueryMetricData{{
				handlerName: "GetUser",
				value:       "3.2",
			}}),
		},
		{
			"sum(increase(mattermost_api_time_count[1d] offset 1d) > 0) by (handler)",
			makeFakeResponseForAPIHandler([]mockAPIHandlerQueryMetricData{{
				handlerName: "GetUser",
				value:       "3",
			}}),
		},
		{
			"(sum(increase(mattermost_api_time_sum[1d] offset 1d)) by (handler) / sum(increase(mattermost_api_time_count[1d] offset 1d) > 0) by (handler))",
			makeFakeResponseForAPIHandler([]mockAPIHandlerQueryMetricData{{
				handlerName: "GetUser",
				value:       "1",
			}}),
		},
		{
			"sum(increase(mattermost_api_time_sum[1d] offset 3d) and increase(mattermost_api_time_count[1d] offset 3d) > 0) by (handler)",
			makeFakeResponseForAPIHandler([]mockAPIHandlerQueryMetricData{{
				handlerName: "GetUser",
				value:       "5.2",
			}}),
		},
		{
			"sum(increase(mattermost_api_time_count[1d] offset 3d) > 0) by (handler)",
			makeFakeResponseForAPIHandler([]mockAPIHandlerQueryMetricData{{
				handlerName: "GetUser",
				value:       "5",
			}}),
		},
		{
			"(sum(increase(mattermost_api_time_sum[1d] offset 3d)) by (handler) / sum(increase(mattermost_api_time_count[1d] offset 3d) > 0) by (handler))",
			makeFakeResponseForAPIHandler([]mockAPIHandlerQueryMetricData{{
				handlerName: "GetUser",
				value:       "1",
			}}),
		},
	}

	for _, mockData := range mockResponses {
		prom.EXPECT().Query(mockData.query).Return(mockData.response, nil).AnyTimes()
	}

	// prom.EXPECT().Query("sum(increase(mattermost_api_time_sum[1d] offset 1d) and increase(mattermost_api_time_count[1d] offset 1d) > 0) by (handler)").Return(, nil)

	res, appErr := p.ExecuteCommand(nil, args)
	require.Nil(t, appErr)

	require.NotNil(t, res)

	expected := "`/performance-report run --first 1d --second 3d --length 1d`\n\n## Percentage Differences of API Handlers\n\nParameters:\n- First time frame start: 1d\n- Second time frame start: 3d\n- Time frame length: 1d\n\n### Biggest Increases\n\n| Handler | Total Diff | Total 1 | Total 2 | Count Diff | Count 1 | Count 2 | Average Diff | Average 1 | Average 2 |\n| -------- | -------- | -------- | -------- |\n|GetUser | 62.5% | 3.200 | 5.200 | 66.67% | 3 | 5 | 0% | 1.00000 | 1.00000 |\n\n\n### Biggest Decreases\n\n| Handler | Total Diff | Total 1 | Total 2 | Count Diff | Count 1 | Count 2 | Average Diff | Average 1 | Average 2 |\n| -------- | -------- | -------- | -------- |\n|GetUser | 62.5% | 3.200 | 5.200 | 66.67% | 3 | 5 | 0% | 1.00000 | 1.00000 |\n"
	require.Equal(t, expected, res.Text)
}

type mockAPIHandlerQueryMetricData struct {
	handlerName string
	value       string
}

func makeFakeResponseForAPIHandler(data []mockAPIHandlerQueryMetricData) *prometheus.Response {
	values := []struct {
		Metric map[string]string
		Value  []interface{}
	}{}

	for _, entry := range data {
		values = append(values, struct {
			Metric map[string]string
			Value  []interface{}
		}{
			Metric: map[string]string{
				"handler": entry.handlerName,
			},
			Value: []interface{}{
				"",
				entry.value,
			},
		})
	}

	return &prometheus.Response{
		Status: "success",
		Data: struct {
			ResultType string
			Result     []struct {
				Metric map[string]string
				Value  []interface{}
			}
		}{
			Result: values,
		},
	}
}
