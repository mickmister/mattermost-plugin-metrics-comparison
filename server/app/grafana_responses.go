package app

import (
	"encoding/json"
	"os"
)

func getGrafanaResponseFromJSON(fname string) (GrafanaDBStoreResponse, error) {
	var res GrafanaDBStoreResponse

	b, err := os.ReadFile(fname)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(b, &res)
	if err != nil {
		return res, err
	}

	return res, nil
}

type GrafanaDBStoreResponse struct {
	Results struct {
		A struct {
			Frames []struct {
				Data struct {
					Values [][]float64 `json:"values"`
				} `json:"data"`
				Schema struct {
					Fields []struct {
						Config struct {
							DisplayNameFromDs string `json:"displayNameFromDS,omitempty"`
							Interval          int    `json:"interval,omitempty"`
						} `json:"config"`
						Labels *struct {
							Method string `json:"method"`
						} `json:"labels,omitempty"`
						Name     string `json:"name"`
						Type     string `json:"type"`
						TypeInfo struct {
							Frame string `json:"frame"`
						} `json:"typeInfo"`
					} `json:"fields"`
					Meta struct {
						Custom struct {
							ResultType string `json:"resultType"`
						} `json:"custom"`
						ExecutedQueryString string `json:"executedQueryString"`
						Type                string `json:"type"`
						TypeVersion         []int  `json:"typeVersion"`
					} `json:"meta"`
					RefID string `json:"refId"`
				} `json:"schema"`
			} `json:"frames"`
			Status int `json:"status"`
		} `json:"A"`
		B struct {
			Frames []struct {
				Data struct {
					Values [][]float64 `json:"values"`
				} `json:"data"`
				Schema struct {
					Fields []struct {
						Config struct {
							DisplayNameFromDs string `json:"displayNameFromDS,omitempty"`
							Interval          int    `json:"interval,omitempty"`
						} `json:"config"`
						Labels *struct {
							Method string `json:"method"`
						} `json:"labels,omitempty"`
						Name     string `json:"name"`
						Type     string `json:"type"`
						TypeInfo struct {
							Frame string `json:"frame"`
						} `json:"typeInfo"`
					} `json:"fields"`
					Meta struct {
						Custom struct {
							ResultType string `json:"resultType"`
						} `json:"custom"`
						ExecutedQueryString string `json:"executedQueryString"`
						Type                string `json:"type"`
						TypeVersion         []int  `json:"typeVersion"`
					} `json:"meta"`
					RefID string `json:"refId"`
				} `json:"schema"`
			} `json:"frames"`
			Status int `json:"status"`
		} `json:"B"`
		C struct {
			Frames []struct {
				Data struct {
					Values [][]float64 `json:"values"`
				} `json:"data"`
				Schema struct {
					Fields []struct {
						Config struct {
							DisplayNameFromDs string `json:"displayNameFromDS,omitempty"`
							Interval          int    `json:"interval,omitempty"`
						} `json:"config"`
						Labels *struct {
							Method string `json:"method"`
						} `json:"labels,omitempty"`
						Name     string `json:"name"`
						Type     string `json:"type"`
						TypeInfo struct {
							Frame string `json:"frame"`
						} `json:"typeInfo"`
					} `json:"fields"`
					Meta struct {
						Custom struct {
							ResultType string `json:"resultType"`
						} `json:"custom"`
						ExecutedQueryString string `json:"executedQueryString"`
						Type                string `json:"type"`
						TypeVersion         []int  `json:"typeVersion"`
					} `json:"meta"`
					RefID string `json:"refId"`
				} `json:"schema"`
			} `json:"frames"`
			Status int `json:"status"`
		} `json:"C"`
	} `json:"results"`
}
