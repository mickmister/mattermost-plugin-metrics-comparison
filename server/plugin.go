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
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-metrics-comparison/server/app"
	"github.com/mattermost/mattermost-plugin-metrics-comparison/server/prometheus"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	prometheusClient prometheus.PrometheusClient

	pluginapi *pluginapi.Client
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

	conf := p.getConfiguration()
	if conf.PrometheusURL == "" {
		return errors.New("please provide a config value for PrometheusURL")
	}

	p.pluginapi = pluginapi.NewClient(p.API, p.Driver)
	p.prometheusClient = prometheus.New(conf.PrometheusURL, p.pluginapi.Log)

	return nil
}

const fname1 = "_debug/grafana_sept3-9.json"
const fname2 = "_debug/grafana_sept10-16.json"

func getRunCommand() *model.AutocompleteData {
	var runCommand = model.NewAutocompleteData("run", "", "")

	defaults := app.GetDefaultReportFlags()
	runCommand.AddNamedTextArgument("query", "", "", defaults.Query, false)
	runCommand.AddNamedTextArgument("first", "", "", defaults.First, false)
	runCommand.AddNamedTextArgument("second", "", "", defaults.Second, false)
	runCommand.AddNamedTextArgument("length", "", "", defaults.Length, false)
	runCommand.AddNamedTextArgument("scaleBy", "", "", defaults.ScaleBy, false)
	runCommand.AddNamedTextArgument("sort", "", "", defaults.Sort, false)
	runCommand.AddNamedTextArgument("limit", "", "", fmt.Sprintf("%d", defaults.Limit), false)

	return runCommand
}

func parseRunCommandFlags(args []string) (app.RunReportFlags, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)

	defaults := app.GetDefaultReportFlags()
	query := fs.String("query", defaults.Query, "")
	first := fs.String("first", defaults.First, "")
	second := fs.String("second", defaults.Second, "")
	length := fs.String("length", defaults.Length, "")
	scaleBy := fs.String("scaleBy", defaults.ScaleBy, "")
	sort := fs.String("sort", defaults.Sort, "")
	limit := fs.Int("limit", defaults.Limit, "")

	err := fs.Parse(args)
	if err != nil {
		return defaults, err
	}

	return app.RunReportFlags{
		Query:   *query,
		First:   *first,
		Second:  *second,
		Length:  *length,
		ScaleBy: *scaleBy,
		Sort:    *sort,
		Limit:   *limit,
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

func (p *Plugin) executeRunCommand(flagValues app.RunReportFlags) (string, error) {
	a := app.New(p.prometheusClient)

	data1, err := a.GetDBMetrics(flagValues.First, flagValues.Length)
	if err != nil {
		err = errors.Wrap(err, "error getting API metrics")
		return err.Error(), err
	}

	data2, err := a.GetDBMetrics(flagValues.Second, flagValues.Length)
	if err != nil {
		err = errors.Wrap(err, "error running report")
		return err.Error(), err
	}

	report := app.ComputeReport(data1, data2, flagValues)
	return report.AsMarkdown(), nil
}

func (p *Plugin) executeRunCommandWithMockData(flagValues app.RunReportFlags) (string, error) {
	a := app.New(p.prometheusClient)

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		err = errors.Wrap(err, "error getting bundle assets folder path for test data")
		return err.Error(), err
	}

	fname := path.Join(bundlePath, "assets", fname1)

	data1, err := a.RunQueryWithMockData(flagValues.Query, flagValues.First, flagValues.Second, flagValues.Length, flagValues.ScaleBy, fname)
	if err != nil {
		err = errors.Wrap(err, "error running report")
		return err.Error(), err
	}

	fname = path.Join(bundlePath, "assets", fname2)

	data2, err := a.RunQueryWithMockData(flagValues.Query, flagValues.First, flagValues.Second, flagValues.Length, flagValues.ScaleBy, fname)
	if err != nil {
		err = errors.Wrap(err, "error running report")
		return err.Error(), err
	}

	report := app.ComputeReport(data1, data2, flagValues)
	return report.AsMarkdown(), nil
}

func getHelpText() string {
	return "TODO: Implement help text"
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
