package domain

import (
	"errors"
	"testing"
)

func TestNewReportSourceReferenceNormalizesAvailableReference(t *testing.T) {
	reference, err := NewReportSourceReference(ReportSourceReferenceInput{
		EntityType: " goods_receipt ",
		ID:         " gr-1 ",
		Label:      " GR-001 ",
		Href:       " /receiving?source_id=gr-1 ",
	})
	if err != nil {
		t.Fatalf("NewReportSourceReference returned error: %v", err)
	}

	if reference.EntityType != "goods_receipt" ||
		reference.ID != "gr-1" ||
		reference.Label != "GR-001" ||
		reference.Href != "/receiving?source_id=gr-1" ||
		reference.Unavailable {
		t.Fatalf("reference = %+v, want normalized available reference", reference)
	}
}

func TestNewReportSourceReferenceAllowsUnavailableReferenceWithoutHref(t *testing.T) {
	reference, err := NewReportSourceReference(ReportSourceReferenceInput{
		EntityType:  "external_note",
		ID:          "note-1",
		Label:       "NOTE-001",
		Unavailable: true,
	})
	if err != nil {
		t.Fatalf("NewReportSourceReference returned error: %v", err)
	}

	if !reference.Unavailable || reference.Href != "" {
		t.Fatalf("reference = %+v, want unavailable reference with empty href", reference)
	}
}

func TestNewReportSourceReferenceRejectsInvalidReferences(t *testing.T) {
	tests := []ReportSourceReferenceInput{
		{ID: "gr-1", Label: "GR-001", Href: "/receiving?source_id=gr-1"},
		{EntityType: "goods_receipt", Label: "GR-001", Href: "/receiving?source_id=gr-1"},
		{EntityType: "goods_receipt", ID: "gr-1", Href: "/receiving?source_id=gr-1"},
		{EntityType: "goods_receipt", ID: "gr-1", Label: "GR-001"},
	}

	for _, input := range tests {
		_, err := NewReportSourceReference(input)
		if !errors.Is(err, ErrInvalidReportSourceReference) {
			t.Fatalf("input = %+v error = %v, want ErrInvalidReportSourceReference", input, err)
		}
	}
}
