package main

import (
	"flag"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"

	cli "github.com/mattermost/mattermost-perf-stats-cli/app"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

const trigger = "performance-report"

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func (p *Plugin) OnActivate() error {
	err := p.API.RegisterCommand(&model.Command{
		Trigger:      trigger,
		AutoComplete: true,
		AutocompleteData: &model.AutocompleteData{
			Trigger:     trigger,
			SubCommands: []*model.AutocompleteData{getRunCommand()},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

const defaultQuery = "sum(increase(mattermost_db_store_time_sum[{{.Length}}]) and increase(mattermost_db_store_time_count[{{.Offset}}]) > 0) by (method)"
const defaultFirst = "3d"
const defaultSecond = "5d"
const defaultLength = "1d"
const defaultScaleBy = "post"
const defaultSort = "total-time"
const defaultLimit = 20

const fname1 = "_debug/grafana_sept3-9.json"
const fname2 = "_debug/grafana_sept10-16.json"

type runCommandFlags struct {
	query   string
	first   string
	second  string
	length  string
	Sort    string
	scaleBy string
	limit   int
}

func getRunCommand() *model.AutocompleteData {
	var runCommand = model.NewAutocompleteData("run", "", "")

	runCommand.AddNamedTextArgument("query", "", "", defaultQuery, false)
	runCommand.AddNamedTextArgument("first", "", "", defaultFirst, false)
	runCommand.AddNamedTextArgument("second", "", "", defaultSecond, false)
	runCommand.AddNamedTextArgument("length", "", "", defaultLength, false)
	runCommand.AddNamedTextArgument("scaleBy", "", "", defaultScaleBy, false)
	runCommand.AddNamedTextArgument("sort", "", "", defaultSort, false)
	runCommand.AddNamedTextArgument("limit", "", "", fmt.Sprintf("%d", defaultLimit), false)

	return runCommand
}

func parseRunCommandFlags(args []string) (*runCommandFlags, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)

	query := fs.String("query", defaultQuery, "")
	first := fs.String("first", defaultFirst, "")
	second := fs.String("second", defaultSecond, "")
	length := fs.String("length", defaultLength, "")
	scaleBy := fs.String("scaleBy", defaultScaleBy, "")
	limit := fs.Int("limit", defaultLimit, "")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	return &runCommandFlags{
		query:   *query,
		first:   *first,
		second:  *second,
		length:  *length,
		scaleBy: *scaleBy,
		limit:   *limit,
	}, nil
}

func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	parts := strings.Fields(args.Command)
	if len(parts) < 1 {
		return &model.CommandResponse{Text: "Please provide a command." + getHelpText()}, nil
	}
	if parts[0] != "/"+trigger {
		return &model.CommandResponse{Text: "Invalid command. Please use " + trigger + getHelpText()}, nil
	}

	if len(parts) < 2 {
		return &model.CommandResponse{Text: "Invalid command. Please include a subcommand." + getHelpText()}, nil
	}

	subcommand := parts[1]
	if subcommand != "run" {
		return &model.CommandResponse{Text: "Invalid subcommand" + subcommand + "." + getHelpText()}, nil
	}

	commandFlags := parts[2:]
	if len(commandFlags) < 1 {
		return &model.CommandResponse{Text: "No flags received." + getHelpText()}, nil
	}

	flagValues, err := parseRunCommandFlags(commandFlags)
	if err != nil {
		return &model.CommandResponse{Text: errors.Wrap(err, "error parsing command flags").Error()}, nil
	}

	resText, err := p.executeRunCommand(flagValues)
	if err != nil {
		p.API.LogError("failed to execute run command", "err", err)
	}

	return &model.CommandResponse{Text: resText}, nil
}

func (p *Plugin) executeRunCommand(flagValues *runCommandFlags) (string, error) {
	a := cli.New("")

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		err = errors.Wrap(err, "error getting bundle assets folder path for test data")
		return err.Error(), err
	}

	fname := path.Join(bundlePath, "assets", fname1)

	data1, err := a.RunQueryWithMockData(flagValues.query, flagValues.first, flagValues.second, flagValues.length, flagValues.scaleBy, fname)
	if err != nil {
		err = errors.Wrap(err, "error running report")
		return err.Error(), err
	}

	fname = path.Join(bundlePath, "assets", fname2)

	// is this supposed to compute one time frame, or both?
	data2, err := a.RunQueryWithMockData(flagValues.query, flagValues.first, flagValues.second, flagValues.length, flagValues.scaleBy, fname)
	if err != nil {
		err = errors.Wrap(err, "error running report")
		return err.Error(), err
	}

	biggestIncreases, biggestDecreases := cli.ComputeReport(data1, data2, "total-time", flagValues.limit)

	resText := "## Biggest Increases\n\n"
	resText += generateMarkdownTable(biggestIncreases)

	resText += "\n\n## Biggest Decreases\n\n"
	resText += generateMarkdownTable(biggestDecreases)

	return resText, nil
}

func generateMarkdownTable(entries []*cli.DBEntry) string {
	resText := "| Method | Total Time | Count | Average |\n"
	resText += "| -------- | -------- | -------- | -------- |\n"
	for _, v := range entries {
		resText += fmt.Sprintf("|%s | %v | %v | %v |\n", v.Method, v.TotalTime, v.Count, v.Average)
	}

	return resText
}

func getHelpText() string {
	return "TODO: Implement help text"
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
