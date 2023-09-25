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
	version    = "0.1.0"
	pdEnvToken = "PD_TOKEN"

	pdBaseURL           = "https://foo.pagerduty.com"
	flagTimestampLayout = "2006-01-02"
)

var (
	flags struct {
		Incidents struct {
			TeamID string `help:"PagerDuty team ID to get incidents for"`
			Since  string `help:"time range to get incidents for" default:"168h"`
			CSV    string `help:"PagerDuty incidents .csv export to use"`
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
				o.Start.Format(flagTimestampLayout),
				o.End.Format(flagTimestampLayout),
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

	since, err := time.Parse(flagTimestampLayout, flags.Incidents.Since)
	if err != nil {
		log.Println(err)
		return
	}
	incidents, err := client.GetIncidentsForTeam(flags.Incidents.TeamID, since)
	if err != nil {
		log.Println(err)
		return
	}
	printIncidentsAsMarkdownTable(incidents)
}

func printIncidentsAsMarkdownTable(incidents []pagerduty.Incident) {
	incidentsForTable := make([][]string, 0, len(incidents))

	for _, i := range incidents {
		incidentsForTable = append(incidentsForTable,
			[]string{fmt.Sprintf("%d", i.Number), i.URL(), i.Description,
				formatter.FormatTimeWithUTCAndLocal(i.CreatedAt),
				formatter.FormatTimeWithUTCAndLocal(i.ResolvedAt), i.Duration()},
		)
	}

	header := []string{"Incident", "Description", "Created", "Resolved", "Duration"}

	text, _ := formatter.RenderMarkdownTable(header, incidentsForTable)

	fmt.Println(text)
}
