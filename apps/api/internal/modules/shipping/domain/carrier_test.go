package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewCarrierNormalizesFieldsAndDefaults(t *testing.T) {
	carrier, err := NewCarrier(NewCarrierInput{
		Code:         " ghn ",
		Name:         "GHN Express",
		HandoverZone: "handover-a",
		CreatedAt:    time.Date(2026, 4, 28, 8, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new carrier: %v", err)
	}

	if carrier.ID != "carrier-ghn" ||
		carrier.Code != "GHN" ||
		carrier.Status != CarrierStatusActive ||
		carrier.SLAProfile != "standard" ||
		!carrier.IsActive() {
		t.Fatalf("carrier = %+v, want normalized active carrier with default SLA profile", carrier)
	}
}

func TestNewCarrierRejectsInvalidStatus(t *testing.T) {
	_, err := NewCarrier(NewCarrierInput{
		Code:         "GHN",
		Name:         "GHN Express",
		HandoverZone: "handover-a",
		Status:       CarrierStatus("paused"),
		CreatedAt:    time.Date(2026, 4, 28, 8, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrCarrierInvalidStatus) {
		t.Fatalf("err = %v, want invalid status", err)
	}
}

func TestNewCarrierRequiresMasterFields(t *testing.T) {
	_, err := NewCarrier(NewCarrierInput{
		Code:      "GHN",
		Name:      "GHN Express",
		CreatedAt: time.Date(2026, 4, 28, 8, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrCarrierRequiredField) {
		t.Fatalf("err = %v, want required field", err)
	}
}
