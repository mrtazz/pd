package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/alecthomas/kong"
	"github.com/olekukonko/tablewriter"

	"github.com/davecgh/go-spew/spew"
)

const (
	version    = "0.1.0"
	pdEnvToken = "PD_TOKEN"
	// csv headers
	incidentNumberHeader = "incident_number"
	descriptionHeader    = "description"
	createdAtHeader      = "created_on"
	resolvedAtHeader     = "resolved_on"

	pdBaseURL            = "https://foo.pagerduty.com"
	timestampLayout      = "2006-01-02T15:04:05-07:00"
	pdAPITimestampLayout = "2006-01-02T15:04:05Z"
	flagTimestampLayout  = "2006-01-02"

	mdTimestampLayout = "2006-01-02 15:04 MST"

	mdTemplate = `
| Incident | Description | Created | Resolved | Duration |
|   ---    |     ---     |  ---    |  ---     | ---      |
{{- range . }}
| [{{ .Incident }}]({{ .IncidentLink }}) | {{ .Description }} | {{ .CreatedAt }} | {{ .ResolvedAt }} | {{ .Duration }} |
{{- end }}
`
)

var (
	flags struct {
		Run struct {
			CSV string `required:"" help:"PagerDuty incidents .csv export to use"`
		} `cmd:"" help:"run the agent"`
		Import struct {
			TeamID string `required:"" help:"PagerDuty team ID to get incidents for"`
			Since  string `required:"" help:"time range to get incidents for" default:"168h"`
		} `cmd:"" help:"import incidents from pagerduty API"`
		Oncall struct {
			UserID string `required:"" help:"PagerDuty user ID to get incidents for"`
			Since  string `required:"" help:"time range to get incidents for, e.g. 2006-01-02"`
		} `cmd:"" help:"retrieve on-call schedules for user from the pagerduty API"`
		Version struct {
		} `cmd:"" help:"print version and exit."`
	}
)

type incident struct {
	incidentNumber string
	description    string
	createdAt      time.Time
	resolvedAt     time.Time
}

type oncall struct {
	user   string
	start  time.Time
	end    time.Time
	policy string
}

func timeFormatInLocal(t time.Time) string {
	utcTime := t.UTC().Format(mdTimestampLayout)
	loc, _ := time.LoadLocation("Europe/Berlin")
	localTime := t.In(loc).Format("15:04 MST")

	return fmt.Sprintf("%s (%s)", utcTime, localTime)
}

func (i incident) CreatedAt() string {
	return timeFormatInLocal(i.createdAt)
}
func (i incident) ResolvedAt() string {
	return timeFormatInLocal(i.resolvedAt)
}
func (i incident) Description() string {
	return i.description
}
func (i incident) Incident() string {
	return i.incidentNumber
}
func (i incident) IncidentLink() string {
	return fmt.Sprintf("%s/incidents/%s", pdBaseURL, i.incidentNumber)
}

func (i incident) Duration() string {
	return fmt.Sprintf("%s", i.resolvedAt.Sub(i.createdAt))
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func getIncidentsFromPD(teamID string, since time.Duration, token string) ([]incident, error) {

	ctx := context.Background()

	ret := make([]incident, 0, 25)
	client := pagerduty.NewClient(token)
	team, err := client.GetTeamWithContext(ctx, teamID)
	if err != nil {
		return ret, err
	}

	fmt.Printf("Getting incidents for team '%s', with ID '%s':", team.Name, teamID)
	var opts pagerduty.ListIncidentsOptions
	opts.TeamIDs = []string{teamID}
	opts.Statuses = []string{"triggered", "acknowledged", "resolved"}
	opts.Since = (time.Now().Add(-since)).Format(pdAPITimestampLayout)
	resp, err := client.ListIncidentsWithContext(ctx, opts)
	if err != nil {
		return ret, err
	}

	for _, inc := range resp.Incidents {
		thisIncident := incident{
			incidentNumber: fmt.Sprintf("%d", inc.IncidentNumber),
			description:    inc.Description,
		}
		if thisIncident.createdAt, err = time.Parse(pdAPITimestampLayout, inc.CreatedAt); err != nil {
			log.Println("error parsing created at: " + err.Error())
			continue
		}
		if thisIncident.resolvedAt, err = time.Parse(pdAPITimestampLayout, inc.LastStatusChangeAt); err != nil {
			log.Println("error parsing updated at: " + err.Error())
			continue
		}
		ret = append(ret, thisIncident)
	}

	return ret, nil
}

func parseRecords(input [][]string) []incident {
	ret := make([]incident, 0, len(input))
	var incidentNumberIdx int
	var descriptionIdx int
	var createdAtIdx int
	var resolvedAtIdx int

	for idx, headerField := range input[0] {
		switch headerField {
		case incidentNumberHeader:
			incidentNumberIdx = idx
		case descriptionHeader:
			descriptionIdx = idx
		case createdAtHeader:
			createdAtIdx = idx
		case resolvedAtHeader:
			resolvedAtIdx = idx
		}
	}

	for _, incidentLine := range input[1:len(input)] {
		thisIncident := incident{
			incidentNumber: incidentLine[incidentNumberIdx],
			description:    incidentLine[descriptionIdx],
		}
		var err error
		if thisIncident.createdAt, err = time.Parse(timestampLayout, incidentLine[createdAtIdx]); err != nil {
			log.Println("error parsing created at: " + err.Error())
			continue
		}
		if thisIncident.resolvedAt, err = time.Parse(timestampLayout, incidentLine[resolvedAtIdx]); err != nil {
			log.Println("error parsing created at: " + err.Error())
			continue
		}
		ret = append(ret, thisIncident)
	}

	return ret
}

func renderMarkdown(i []incident) (string, error) {
	tmpl, err := template.New("markdown").Parse(mdTemplate)
	if err != nil {
		return "", fmt.Errorf("template error: %w", err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, i)
	if err != nil {
		return "", fmt.Errorf("template error: %w", err)
	}
	return b.String(), nil
}

func main() {
	ctx := kong.Parse(&flags)
	switch ctx.Command() {
	case "run":
		run()
	case "import":
		importFromPD()
	case "oncall":
		getOnCallTimes()
	case "version":
		fmt.Printf(version)
		return
	default:
		log.Fatal("Unknown command: " + ctx.Command())
	}
}

func run() {
	path := flags.Run.CSV
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
	}
	records := readCsvFile(path)
	incidents := parseRecords(records)
	tpl, err := renderMarkdown(incidents)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(tpl)
}

func getOnCallTimesFromPD(userID string, since time.Time, token string) ([]oncall, error) {
	ctx := context.Background()

	ret := make([]oncall, 0, 25)
	client := pagerduty.NewClient(token)
	user, err := client.GetUserWithContext(ctx, userID, pagerduty.GetUserOptions{})
	if err != nil {
		return ret, err
	}

	fmt.Printf("Getting on-call times for user '%s', with ID '%s' since %s:\n", user.Name, userID, since)
	var opts pagerduty.ListOnCallOptions
	opts.UserIDs = []string{userID}
	opts.Since = since.Format(pdAPITimestampLayout)
	opts.ScheduleIDs = []string{}
	resp, err := client.ListOnCallsWithContext(ctx, opts)
	if err != nil {
		return ret, err
	}

	for _, oc := range resp.OnCalls {
		spew.Dump(oc)
		thisOC := oncall{
			user:   user.Name,
			policy: oc.Schedule.Summary,
		}
		if thisOC.start, err = time.Parse(pdAPITimestampLayout, oc.Start); err != nil {
			log.Println("error parsing oncall start time: " + err.Error())
			continue
		}
		if thisOC.end, err = time.Parse(pdAPITimestampLayout, oc.End); err != nil {
			log.Println("error parsing oncall end time: " + err.Error())
			continue
		}
		ret = append(ret, thisOC)
	}

	return ret, nil
}

func getOnCallTimes() {
	token := os.Getenv(pdEnvToken)

	since, err := time.Parse(flagTimestampLayout, flags.Oncall.Since)
	if err != nil {
		log.Println(err)
		return
	}
	oncalls, err := getOnCallTimesFromPD(flags.Oncall.UserID, since, token)
	if err != nil {
		log.Println(err)
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Start", "End", "Policy"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	for _, v := range oncalls {
		table.Append([]string{
			v.start.Format(flagTimestampLayout),
			v.end.Format(flagTimestampLayout),
			v.policy,
		})
	}
	table.Render()
}

func importFromPD() {
	token := os.Getenv(pdEnvToken)

	duration, err := time.ParseDuration(flags.Import.Since)
	if err != nil {
		log.Println(err)
		return
	}
	incidents, err := getIncidentsFromPD(flags.Import.TeamID, duration, token)
	if err != nil {
		log.Println(err)
		return
	}
	tpl, err := renderMarkdown(incidents)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(tpl)
}
