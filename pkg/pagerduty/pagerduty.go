package pagerduty

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pdApi "github.com/PagerDuty/go-pagerduty"
)

const (
	// timestamp layouts
	pdAPITimestampLayout = "2006-01-02T15:04:05Z"
	mdTimestampLayout    = "2006-01-02 15:04 MST"
	csvTimestampLayout   = "2006-01-02T15:04:05-07:00"
	dayTimestampLayout   = "2006-01-02"

	// csv headers
	incidentNumberHeader = "incident_number"
	descriptionHeader    = "description"
	createdAtHeader      = "created_on"
	resolvedAtHeader     = "resolved_on"
)

var (
	// BaseURL of the pagerduty account
	BaseURL = "https://pagerduty.com"
)

// Incident represents a PD incident
type Incident struct {
	Number      int
	Description string
	CreatedAt   time.Time
	ResolvedAt  time.Time
}

// URL returns the incident URL
func (i Incident) URL() string {
	return fmt.Sprintf("%s/incidents/%d", BaseURL, i.Number)
}

// Duration returns the format duration of the incident
func (i Incident) Duration() string {
	return fmt.Sprintf("%s", i.ResolvedAt.Sub(i.CreatedAt))
}

// OncallShift represents an on-call shift
type OncallShift struct {
	User     string
	Start    time.Time
	End      time.Time
	Schedule string
}

// Client is the interface to interact with the PD API
type Client interface {
	GetIncidentsForTeam(teamID string, since time.Time) ([]Incident, error)
	GetOncallShiftsForUser(userID string, since time.Time) ([]OncallShift, error)
}

type defaultClient struct {
	api *pdApi.Client
}

// New returns a pagerduty client
func New(token string) Client {
	return defaultClient{
		api: pdApi.NewClient(token),
	}
}

func (d defaultClient) GetIncidentsForTeam(teamID string, since time.Time) ([]Incident, error) {

	ctx := context.Background()

	ret := make([]Incident, 0, 25)

	var opts pdApi.ListIncidentsOptions
	opts.TeamIDs = []string{teamID}
	opts.Statuses = []string{"triggered", "acknowledged", "resolved"}
	opts.Since = since.Format(pdAPITimestampLayout)
	resp, err := d.api.ListIncidentsWithContext(ctx, opts)
	if err != nil {
		return ret, err
	}

	for _, inc := range resp.Incidents {
		thisIncident := Incident{
			Number:      int(inc.IncidentNumber),
			Description: inc.Description,
		}
		if thisIncident.CreatedAt, err = time.Parse(pdAPITimestampLayout, inc.CreatedAt); err != nil {
			log.Println("error parsing created at: " + err.Error())
			continue
		}
		if thisIncident.ResolvedAt, err = time.Parse(pdAPITimestampLayout, inc.LastStatusChangeAt); err != nil {
			log.Println("error parsing updated at: " + err.Error())
			continue
		}
		ret = append(ret, thisIncident)
	}

	return ret, nil
}

func (d defaultClient) GetOncallShiftsForUser(userID string, since time.Time) ([]OncallShift, error) {
	ctx := context.Background()

	ret := make([]OncallShift, 0, 25)
	user, err := d.api.GetUserWithContext(ctx, userID, pdApi.GetUserOptions{})
	if err != nil {
		return ret, err
	}

	fmt.Printf("Getting on-call times for user '%s', with ID '%s' since %s:\n", user.Name, userID, since)
	var opts pdApi.ListOnCallOptions
	opts.UserIDs = []string{userID}
	opts.Since = since.Format(pdAPITimestampLayout)
	opts.ScheduleIDs = []string{}
	resp, err := d.api.ListOnCallsWithContext(ctx, opts)
	if err != nil {
		return ret, err
	}

	for _, oc := range resp.OnCalls {
		thisOC := OncallShift{
			User:     user.Name,
			Schedule: oc.Schedule.Summary,
		}
		if thisOC.Start, err = time.Parse(pdAPITimestampLayout, oc.Start); err != nil {
			log.Println("error parsing oncall start time: " + err.Error())
			continue
		}
		if thisOC.End, err = time.Parse(pdAPITimestampLayout, oc.End); err != nil {
			log.Println("error parsing oncall end time: " + err.Error())
			continue
		}
		ret = append(ret, thisOC)
	}

	return ret, nil
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

func parseRecords(input [][]string) []Incident {
	ret := make([]Incident, 0, len(input))
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
		thisIncident := Incident{
			Description: incidentLine[descriptionIdx],
		}
		var err error
		if thisIncident.Number, err = strconv.Atoi(incidentLine[incidentNumberIdx]); err != nil {
			log.Println("error parsing incident number: " + err.Error())
			continue
		}
		if thisIncident.CreatedAt, err = time.Parse(csvTimestampLayout, incidentLine[createdAtIdx]); err != nil {
			log.Println("error parsing created at: " + err.Error())
			continue
		}
		if thisIncident.ResolvedAt, err = time.Parse(csvTimestampLayout, incidentLine[resolvedAtIdx]); err != nil {
			log.Println("error parsing created at: " + err.Error())
			continue
		}
		ret = append(ret, thisIncident)
	}

	return ret
}

// ParseIncidentsCSV returns incidents parsed from a given path to a csv
// export
func ParseIncidentsCSV(path string) ([]Incident, error) {
	records := readCsvFile(path)
	incidents := parseRecords(records)

	return incidents, nil
}
