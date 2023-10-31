package main

import (
	"flag"
	"fmt"
	"net/http"
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

	reportQueryClient *prometheus.ReportQueryClient

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

	promClient := prometheus.NewPrometheusClient(conf.PrometheusURL)
	p.reportQueryClient = prometheus.NewReportClient(promClient)

	return nil
}

const fname1 = "_debug/grafana_sept3-9.json"
const fname2 = "_debug/grafana_sept10-16.json"

func getRunCommand() *model.AutocompleteData {
	var runCommand = model.NewAutocompleteData("run", "", "")

	defaults := app.GetDefaultReportFlags()
	runCommand.AddNamedStaticListArgument("report", "", false, []model.AutocompleteListItem{
		{
			Item: string(app.ReportTypeAPIHandlerComparison),
		},
		{
			Item: string(app.ReportTypeAPIHandlerComparison),
		},
	})
	runCommand.AddNamedTextArgument("query", "", "", defaults.Query, false)
	runCommand.AddNamedTextArgument("first", "", "", defaults.First, false)
	runCommand.AddNamedTextArgument("second", "", "", defaults.Second, false)
	runCommand.AddNamedTextArgument("length", "", "", defaults.Length, false)
	runCommand.AddNamedTextArgument("scaleBy", "", "", defaults.ScaleBy, false)
	runCommand.AddNamedStaticListArgument("sort", "", false, []model.AutocompleteListItem{
		{
			Item: string(app.SortCategoryTotalTime),
		},
		{
			Item: string(app.SortCategoryCount),
		},
		{
			Item: string(app.SortCategoryAverageTime),
		},
	})
	runCommand.AddNamedTextArgument("limit", "", "", fmt.Sprintf("%d", defaults.Limit), false)
	runCommand.AddNamedTextArgument("public", "", "", "false", false)

	return runCommand
}

func parseRunCommandFlags(args []string) (app.RunReportFlags, bool, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)

	defaults := app.GetDefaultReportFlags()
	reportType := fs.String("report", string(defaults.ReportType), "")
	query := fs.String("query", defaults.Query, "")
	first := fs.String("first", defaults.First, "")
	second := fs.String("second", defaults.Second, "")
	length := fs.String("length", defaults.Length, "")
	scaleBy := fs.String("scaleBy", defaults.ScaleBy, "")
	sort := fs.String("sort", string(defaults.Sort), "")
	limit := fs.Int("limit", defaults.Limit, "")
	public := fs.String("public", "false", "")

	err := fs.Parse(args)
	if err != nil {
		return defaults, false, err
	}

	isPublicPost := *public == "true"

	return app.RunReportFlags{
		ReportType: app.ReportType(*reportType),
		Query:      *query,
		First:      *first,
		Second:     *second,
		Length:     *length,
		ScaleBy:    *scaleBy,
		Sort:       app.SortCategory(*sort),
		Limit:      *limit,
	}, isPublicPost, nil
}

func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	parts := strings.Fields(args.Command)
	if len(parts) < 1 {
		return p.ephemeralResponse("Please provide a command." + getHelpText()), nil
	}
	if parts[0] != "/"+trigger {
		return p.ephemeralResponse("Invalid command. Please use " + trigger + getHelpText()), nil
	}

	if len(parts) < 2 {
		return p.ephemeralResponse("Invalid command. Please include a subcommand." + getHelpText()), nil
	}

	subcommand := parts[1]
	if subcommand != "run" {
		return p.ephemeralResponse("Invalid subcommand" + subcommand + "." + getHelpText()), nil
	}

	commandFlags := parts[2:]

	flagValues, isPublic, err := parseRunCommandFlags(commandFlags)
	if err != nil {
		return p.ephemeralResponse(errors.Wrap(err, "error parsing command flags").Error()), nil
	}

	reportText, err := p.executeRunCommand(flagValues)
	if err != nil {
		errMsg := errors.Wrap(err, "failed to execute run command").Error()
		p.API.LogError(errMsg)
		return p.ephemeralResponse(errMsg), nil
	}

	resText := "`" + args.Command + "`\n\n" + reportText

	if isPublic {
		return p.publicResponse(args.ChannelId, args.UserId, resText), nil
	}

	return p.ephemeralResponse(resText), nil
}

func (p *Plugin) publicResponse(channelId, userId, text string) *model.CommandResponse {
	post := &model.Post{
		ChannelId: channelId,
		UserId:    userId,
		Message:   text,
	}
	p.pluginapi.Post.CreatePost(post)
	return &model.CommandResponse{}
}

func (p *Plugin) ephemeralResponse(text string) *model.CommandResponse {
	return &model.CommandResponse{Text: text}
}

func (p *Plugin) executeRunCommand(flagValues app.RunReportFlags) (string, error) {
	a := app.New(p.reportQueryClient)

	if flagValues.ReportType == app.ReportTypeDBStoreMethodComparison {
		report, err := a.RunDBComparisonReport(flagValues)
		if err != nil {
			err = errors.Wrap(err, "error running db comparison report")
			return err.Error(), err
		}

		return report.AsMarkdown(), nil
	}

	if flagValues.ReportType == app.ReportTypeAPIHandlerComparison {
		report, err := a.RunAPIComparisonReport(flagValues)
		if err != nil {
			err = errors.Wrap(err, "error running db comparison report")
			return err.Error(), err
		}

		return report.AsMarkdown(), nil
	}

	err := errors.Errorf("unknown run subcommand %s", flagValues.ReportType)
	return err.Error(), err
}

func getHelpText() string {
	return "TODO: Implement help text"
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
