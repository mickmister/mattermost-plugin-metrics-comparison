package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Response struct {
	Status string
	Data   struct {
		ResultType string
		Result     []struct {
			Metric map[string]string
			Value  []interface{}
		}
	}
}

type PrometheusClient interface {
	Query(query string) (*Response, error)
}

type ReportQueryClient struct {
	client PrometheusClient
}

func NewReportClient(client PrometheusClient) *ReportQueryClient {
	return &ReportQueryClient{
		client,
	}
}

type prometheusClient struct {
	endpoint string
}

func NewPrometheusClient(endpoint string) PrometheusClient {
	return &prometheusClient{
		endpoint,
	}
}

func (c *prometheusClient) Query(query string) (*Response, error) {
	var response Response

	// ptof("Running PromQL query: " + query)

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", c.endpoint, url.QueryEscape(query)))
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Data.Result) == 0 {
		return nil, errors.Errorf("received no data from prometheus query `%s`", query)
	}

	// ptofj(&response)

	return &response, nil
}

func replacePlaceholders(query, offset, length string) string {
	query = strings.ReplaceAll(query, "{{.Length}}", length)
	query = strings.ReplaceAll(query, "{{.Offset}}", offset)
	return query
}
