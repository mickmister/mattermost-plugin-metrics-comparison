package prometheus

// totalTimeMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_api_time_sum[%s]) and increase(mattermost_api_time_count[%s]) > 0) by (handler)", timeRange, timeRange))
// callsMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_api_time_count[%s]) > 0) by (handler)", timeRange))
// averageMetrics, err := a.client.Query(fmt.Sprintf("(sum(increase(mattermost_api_time_sum[%s])) by (handler) / sum(increase(mattermost_api_time_count[%s]) > 0) by (handler))", timeRange, timeRange))

const queryTemplateAPIHandlerTotalTime = "sum(increase(mattermost_api_time_sum[{{.Length}}] {{.Offset}}) and increase(mattermost_api_time_count[{{.Length}}] {{.Offset}}) > 0) by (handler)"

func (c *Client) QueryAPIHandlerTotalTime(offset, length string) (*Response, error) {
	if offset != "" {
		offset = "offset " + offset
	}

	query := replacePlaceholders(queryTemplateAPIHandlerTotalTime, offset, length)

	return c.Query(query)
}

const queryTemplateAPIHandlerCount = "sum(increase(mattermost_api_time_count[{{.Length}}] {{.Offset}}) > 0) by (handler)"

func (c *Client) QueryAPIHandlerCount(offset, length string) (*Response, error) {
	if offset != "" {
		offset = "offset " + offset
	}

	query := replacePlaceholders(queryTemplateAPIHandlerCount, offset, length)

	return c.Query(query)
}

const queryTemplateAPIHandlerAverage = "(sum(increase(mattermost_api_time_sum[{{.Length}}] {{.Offset}})) by (handler) / sum(increase(mattermost_api_time_count[{{.Length}}] {{.Offset}}) > 0) by (handler))"

func (c *Client) QueryAPIHandlerAverage(offset, length string) (*Response, error) {
	if offset != "" {
		offset = "offset " + offset
	}

	query := replacePlaceholders(queryTemplateAPIHandlerAverage, offset, length)

	return c.Query(query)
}
