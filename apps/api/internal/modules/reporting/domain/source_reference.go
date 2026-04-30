package domain

import (
	"errors"
	"strings"
)

var ErrInvalidReportSourceReference = errors.New("report source reference is invalid")

type ReportSourceReference struct {
	EntityType  string
	ID          string
	Label       string
	Href        string
	Unavailable bool
}

type ReportSourceReferenceInput struct {
	EntityType  string
	ID          string
	Label       string
	Href        string
	Unavailable bool
}

func NewReportSourceReference(input ReportSourceReferenceInput) (ReportSourceReference, error) {
	reference := ReportSourceReference{
		EntityType:  strings.TrimSpace(input.EntityType),
		ID:          strings.TrimSpace(input.ID),
		Label:       strings.TrimSpace(input.Label),
		Href:        strings.TrimSpace(input.Href),
		Unavailable: input.Unavailable,
	}
	if reference.EntityType == "" || reference.ID == "" || reference.Label == "" {
		return ReportSourceReference{}, ErrInvalidReportSourceReference
	}
	if !reference.Unavailable && reference.Href == "" {
		return ReportSourceReference{}, ErrInvalidReportSourceReference
	}
	if reference.Unavailable {
		reference.Href = ""
	}

	return reference, nil
}
