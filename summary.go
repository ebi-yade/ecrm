package ecrm

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

type summary struct {
	Repo             string `json:"repository"`
	ExpiredImages    int64  `json:"expired_images"`
	TotalImages      int64  `json:"total_images"`
	ExpiredImageSize int64  `json:"expired_image_size"`
	TotalImageSize   int64  `json:"total_image_size"`
}

func (s *summary) row() []string {
	return []string{
		s.Repo,
		fmt.Sprintf("%d (%s)", s.TotalImages, humanize.Bytes(uint64(s.TotalImageSize))),
		fmt.Sprintf("%d (%s)", -s.ExpiredImages, humanize.Bytes(uint64(s.ExpiredImageSize))),
		fmt.Sprintf("%d (%s)", s.TotalImages-s.ExpiredImages, humanize.Bytes(uint64(s.TotalImageSize-s.ExpiredImageSize))),
	}
}

func newOutputFormatFrom(s string) (outputFormat, error) {
	switch s {
	case "table":
		return formatTable, nil
	case "json":
		return formatJSON, nil
	default:
		return outputFormat(0), fmt.Errorf("invalid format name: %s", s)
	}
}

type outputFormat int

func (f outputFormat) String() string {
	switch f {
	case formatTable:
		return "table"
	case formatJSON:
		return "json"
	default:
		return "unknown"
	}
}

const (
	formatTable outputFormat = iota + 1
	formatJSON
)

type summaries []*summary

func (s *summaries) print(w io.Writer, noColor bool, format outputFormat) error {
	switch format {
	case formatTable:
		return s.printTable(w, noColor)
	case formatJSON:
		return s.printJSON(w)
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

func (s *summaries) printJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func (s *summaries) printTable(w io.Writer, noColor bool) error {
	t := tablewriter.NewWriter(w)
	t.SetHeader(s.header())
	t.SetBorder(false)
	for _, s := range *s {
		row := s.row()
		colors := make([]tablewriter.Colors, len(row))
		if strings.HasPrefix(row[2], "0 ") {
			row[2] = ""
		} else {
			colors[2] = tablewriter.Colors{tablewriter.FgBlueColor}
		}
		if strings.HasPrefix(row[3], "0 ") {
			colors[3] = tablewriter.Colors{tablewriter.FgYellowColor}
		}
		if noColor {
			t.Append(row)
		} else {
			t.Rich(row, colors)
		}
	}
	t.Render()
	return nil
}

func (s *summaries) header() []string {
	return []string{
		"repository",
		"total",
		"expired",
		"keep",
	}
}
