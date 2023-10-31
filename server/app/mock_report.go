package app

func (a *App) RunQueryWithMockData(fname string) (map[string]*DBEntry, error) {
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
