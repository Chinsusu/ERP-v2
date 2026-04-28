package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
)

func TestPrototypeCarrierCatalogListsActiveCarriers(t *testing.T) {
	service := NewListCarriers(NewPrototypeCarrierCatalog())

	rows, err := service.Execute(context.Background(), domain.NewCarrierFilter("", domain.CarrierStatusActive))
	if err != nil {
		t.Fatalf("list carriers: %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3 active carriers", len(rows))
	}
	if rows[0].Code != "GHN" || rows[1].Code != "NJV" || rows[2].Code != "VTP" {
		t.Fatalf("rows = %+v, want sorted active carriers", rows)
	}
}

func TestPrototypeCarrierCatalogGetsCarrierByCode(t *testing.T) {
	service := NewGetCarrier(NewPrototypeCarrierCatalog())

	carrier, err := service.Execute(context.Background(), "ghn")
	if err != nil {
		t.Fatalf("get carrier: %v", err)
	}
	if carrier.Code != "GHN" ||
		carrier.Name != "GHN Express" ||
		carrier.HandoverZone != "handover-a" ||
		carrier.Status != domain.CarrierStatusActive ||
		carrier.SLAProfile == "" {
		t.Fatalf("carrier = %+v, want GHN master data", carrier)
	}
}

func TestPrototypeCarrierCatalogReturnsNotFound(t *testing.T) {
	service := NewGetCarrier(NewPrototypeCarrierCatalog())

	_, err := service.Execute(context.Background(), "unknown")
	if !errors.Is(err, ErrCarrierNotFound) {
		t.Fatalf("err = %v, want carrier not found", err)
	}
}
