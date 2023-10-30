package prometheus

// totalTimeMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_db_store_time_sum[%s]) and increase(mattermost_db_store_time_count[%s]) > 0) by (method)", timeRange, timeRange))
// callsMetrics, err := a.client.Query(fmt.Sprintf("sum(increase(mattermost_db_store_time_count[%s]) > 0) by (method)", timeRange))
// averageMetrics, err := a.client.Query(fmt.Sprintf("(sum(increase(mattermost_db_store_time_count[2h]) > 0) by (method) - sum(increase(mattermost_db_store_time_count[2h] offset 30m) > 0) by (method)) / sum(increase(mattermost_db_store_time_count[2h] offset 30m) > 0) by (method) * 100"))

const queryTemplateDBStoreTotalTime = "sum(increase(mattermost_db_store_time_sum[{{.Length}}] {{.Offset}}) and increase(mattermost_db_store_time_count[{{.Length}}] {{.Offset}}) > 0) by (method)"
const queryTemplateDBStoreCount = "sum(increase(mattermost_db_store_time_count[{{.Length}}] {{.Offset}}) > 0) by (method)"
const queryTemplateDBStoreAverage = "(sum(increase(mattermost_db_store_time_sum[{{.Length}}] {{.Offset}})) by (method) / sum(increase(mattermost_db_store_time_count[{{.Length}}] {{.Offset}}) > 0) by (method))"

func (c *Client) QueryDBStoreTotalTime(offset, length string) (*Response, error) {
	if offset != "" {
		offset = "offset " + offset
	}

	query := replacePlaceholders(queryTemplateDBStoreTotalTime, offset, length)

	return c.Query(query)
}

func (c *Client) QueryDBStoreCount(offset, length string) (*Response, error) {
	if offset != "" {
		offset = "offset " + offset
	}

	query := replacePlaceholders(queryTemplateDBStoreCount, offset, length)

	return c.Query(query)
}

func (c *Client) QueryDBStoreAverage(offset, length string) (*Response, error) {
	if offset != "" {
		offset = "offset " + offset
	}

	query := replacePlaceholders(queryTemplateDBStoreAverage, offset, length)

	return c.Query(query)
}
