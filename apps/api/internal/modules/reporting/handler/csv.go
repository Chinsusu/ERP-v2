package handler

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrInvalidCSVExport = errors.New("csv export is invalid")

type CSVExport struct {
	Filename string
	Headers  []string
	Rows     [][]string
}

func WriteCSV(w http.ResponseWriter, r *http.Request, export CSVExport) error {
	if len(export.Headers) == 0 {
		return ErrInvalidCSVExport
	}
	for _, row := range export.Rows {
		if len(row) != len(export.Headers) {
			return ErrInvalidCSVExport
		}
	}

	requestID := response.RequestID(r)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+sanitizeCSVFilename(export.Filename)+`"`)
	w.Header().Set(response.HeaderRequestID, requestID)
	w.WriteHeader(http.StatusOK)

	writer := csv.NewWriter(w)
	if err := writer.Write(export.Headers); err != nil {
		return err
	}
	if err := writer.WriteAll(export.Rows); err != nil {
		return err
	}
	writer.Flush()

	return writer.Error()
}

func sanitizeCSVFilename(value string) string {
	name := strings.TrimSpace(value)
	if name == "" {
		return "report.csv"
	}
	replacer := strings.NewReplacer(
		`"`, "_",
		"\\", "_",
		"/", "_",
		"\r", "_",
		"\n", "_",
	)
	name = replacer.Replace(name)
	if !strings.HasSuffix(strings.ToLower(name), ".csv") {
		name += ".csv"
	}

	return name
}
