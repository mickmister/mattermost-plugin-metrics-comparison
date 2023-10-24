package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

type Client struct {
	endpoint string
}

func New(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

func (c *Client) Query(query string) (*Response, error) {
	var response Response
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
	return &response, nil
}
