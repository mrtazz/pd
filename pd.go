package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrtazz/pd/pkg/formatter"
	"github.com/mrtazz/pd/pkg/pagerduty"

	"github.com/alecthomas/kong"
)

const (
	pdEnvToken = "PD_TOKEN"

	flagTimestampLayout        = "2006-01-02"
	oncallShiftTimestampLayout = "2006-01-02T15:04"
)

var (
	version = "0.1.1"
	flags   struct {
		Incidents struct {
			TeamID   string `help:"PagerDuty team ID to get incidents for"`
			BaseURL  string `help:"Base URL for pagerduty (e.g. https://company.pagerduty.com"`
			Since    string `help:"time range start to get incidents for" default:"2006-01-02"`
			Until    string `help:"time range end to get incidents for" default:"2006-02-02"`
			CSV      string `help:"PagerDuty incidents .csv export to use"`
			Format   string `help:"output format for the incident list (markdown or csv)" default:"markdown"`
			Priority string `help:"which priority to list incidents for (empty means all priorities)" default:""`
		} `cmd:"" help:"get incidents from pagerduty API or csv"`
		Oncall struct {
			UserID string `required:"" help:"PagerDuty user ID to get incidents for"`
			Since  string `required:"" help:"time range to get incidents for, e.g. 2006-01-02"`
		} `cmd:"" help:"retrieve on-call schedules for user from the pagerduty API"`
		Version struct {
		} `cmd:"" help:"print version and exit."`
	}
)

func main() {
	ctx := kong.Parse(&flags)
	switch ctx.Command() {
	case "incidents":
		if flags.Incidents.CSV != "" {
			incidentsFromCSV()
		}
		if flags.Incidents.TeamID != "" {
			incidentsFromAPI()
		}
	case "oncall":
		getOnCallTimes()
	case "version":
		fmt.Printf(version)
		return
	default:
		log.Fatal("Unknown command: " + ctx.Command())
	}
}

func incidentsFromCSV() {
	path := flags.Incidents.CSV
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
	}

	incidents, _ := pagerduty.ParseIncidentsCSV(path)

	printIncidentsAsMarkdownTable(incidents)

}

func getOnCallTimes() {
	token := os.Getenv(pdEnvToken)

	client := pagerduty.New(token)

	since, err := time.Parse(flagTimestampLayout, flags.Oncall.Since)
	if err != nil {
		log.Println(err)
		return
	}
	oncalls, err := client.GetOncallShiftsForUser(flags.Oncall.UserID, since)
	if err != nil {
		log.Println(err)
		return
	}
	shiftsForTable := make([][]string, 0, len(oncalls))

	for _, o := range oncalls {
		shiftsForTable = append(shiftsForTable,
			[]string{
				o.Start.Format(oncallShiftTimestampLayout),
				o.End.Format(oncallShiftTimestampLayout),
				o.Schedule},
		)
	}
	header := []string{"Start", "End", "Schedule"}
	text, _ := formatter.RenderMarkdownTable(header, shiftsForTable)

	fmt.Println(text)
}

func incidentsFromAPI() {
	token := os.Getenv(pdEnvToken)
	client := pagerduty.New(token)

	if flags.Incidents.BaseURL != "" {
		client = client.WithBaseURL(flags.Incidents.BaseURL)
	}

	since, err := time.Parse(flagTimestampLayout, flags.Incidents.Since)
	if err != nil {
		log.Println(err)
		return
	}
	until, err := time.Parse(flagTimestampLayout, flags.Incidents.Until)
	if err != nil {
		log.Println(err)
		return
	}
	if until.Before(since) {
		// if we have an invalid time we default to a timespan of 30 days
		until = since.Add(30 * 24 * time.Hour)
	}
	incidents, err := client.GetIncidentsForTeam(flags.Incidents.TeamID, since, until, flags.Incidents.Priority)
	if err != nil {
		log.Println(err)
		return
	}
	switch flags.Incidents.Format {
	case "markdown":
		printIncidentsAsMarkdownTable(incidents)
	case "csv":
		printIncidentsAsCSV(incidents)
	default:
		log.Printf("unknown output format: '%s'\n", flags.Incidents.Format)

	}
}

func printIncidentsAsMarkdownTable(incidents []pagerduty.Incident) {
	incidentsForTable := make([][]string, 0, len(incidents))

	for _, i := range incidents {
		incidentsForTable = append(incidentsForTable,
			[]string{fmt.Sprintf("[%d](%s)", i.Number, i.URL()),
				i.Priority,
				i.Description,
				formatter.FormatTimeWithUTCAndLocal(i.CreatedAt),
				formatter.FormatTimeWithUTCAndLocal(i.ResolvedAt), i.Duration()},
		)
	}

	header := []string{"Incident", "Priority", "Description", "Created", "Last Update", "Duration"}

	text, _ := formatter.RenderMarkdownTable(header, incidentsForTable)

	fmt.Println(text)
}

func printIncidentsAsCSV(incidents []pagerduty.Incident) {
	incidentsForTable := make([][]string, 0, len(incidents))

	for _, i := range incidents {
		incidentsForTable = append(incidentsForTable,
			[]string{fmt.Sprintf("[%d](%s)", i.Number, i.URL()),
				i.Priority,
				i.Description,
				formatter.FormatTimeWithUTCAndLocal(i.CreatedAt),
				formatter.FormatTimeWithUTCAndLocal(i.ResolvedAt), i.Duration()},
		)
	}

	header := []string{"Incident", "Priority", "Description", "Created", "Last Update", "Duration"}

	text, _ := formatter.RenderCSV(header, incidentsForTable)

	fmt.Println(text)
}
