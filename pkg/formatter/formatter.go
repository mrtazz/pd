package formatter

import (
	"bytes"
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
)

const (
	mdTimestampLayout = "2006-01-02 15:04 MST"
)

var (
	// Location defines the local timezone
	Location = "Europe/Berlin"
)

// RenderMarkdownTable renders the given data as a markdown table
func RenderMarkdownTable(header []string, data [][]string) (string, error) {
	var buff bytes.Buffer
	table := tablewriter.NewWriter(&buff)
	table.SetHeader(header)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	table.AppendBulk(data)
	table.Render()

	return buff.String(), nil
}

// FormatTimeWithUTCAndLocal formats a given time as UTC and local time
func FormatTimeWithUTCAndLocal(t time.Time) string {
	utcTime := t.UTC().Format(mdTimestampLayout)
	loc, _ := time.LoadLocation(Location)
	localTime := t.In(loc).Format("15:04 MST")

	return fmt.Sprintf("%s (%s)", utcTime, localTime)
}
