package domain

import (
	"errors"
	"testing"
)

func TestNewWarehouseNormalizesFields(t *testing.T) {
	warehouse, err := NewWarehouse(NewWarehouseInput{
		ID:              "wh-test",
		Code:            " wh-hcm-fg ",
		Name:            "Finished Goods HCM",
		Type:            WarehouseTypeFinishedGood,
		SiteCode:        " hcm ",
		AllowSaleIssue:  true,
		AllowProdIssue:  false,
		AllowQuarantine: false,
		Status:          WarehouseStatusActive,
	})
	if err != nil {
		t.Fatalf("new warehouse: %v", err)
	}

	if warehouse.Code != "WH-HCM-FG" || warehouse.SiteCode != "HCM" {
		t.Fatalf("codes = %q/%q, want normalized uppercase", warehouse.Code, warehouse.SiteCode)
	}
	if warehouse.Type != WarehouseTypeFinishedGood || warehouse.Status != WarehouseStatusActive {
		t.Fatalf("type/status = %q/%q, want finished_good active", warehouse.Type, warehouse.Status)
	}
}

func TestNewWarehouseRejectsInvalidType(t *testing.T) {
	_, err := NewWarehouse(NewWarehouseInput{
		ID:       "wh-test",
		Code:     "WH-HCM-X",
		Name:     "Invalid Warehouse",
		Type:     WarehouseType("unknown"),
		SiteCode: "HCM",
		Status:   WarehouseStatusActive,
	})
	if !errors.Is(err, ErrWarehouseInvalidType) {
		t.Fatalf("error = %v, want invalid warehouse type", err)
	}
}

func TestNewLocationNormalizesAndFiltersWarehouseFields(t *testing.T) {
	location, err := NewLocation(NewLocationInput{
		ID:            "loc-test",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: " wh-hcm-fg ",
		Code:          " fg-a01 ",
		Name:          "Aisle A01",
		Type:          LocationTypePick,
		ZoneCode:      " pick ",
		AllowPick:     true,
		AllowStore:    true,
		Status:        LocationStatusActive,
	})
	if err != nil {
		t.Fatalf("new location: %v", err)
	}

	if location.Code != "FG-A01" || location.ZoneCode != "PICK" || location.WarehouseCode != "WH-HCM-FG" {
		t.Fatalf("location codes = %+v, want normalized uppercase", location)
	}

	filter := NewLocationFilter("pick", "wh-hcm-fg", LocationStatusActive, LocationTypePick, 1, 20)
	if !filter.Matches(location) {
		t.Fatal("filter did not match warehouse/status/type/search")
	}
}

func TestNewLocationRejectsInvalidType(t *testing.T) {
	_, err := NewLocation(NewLocationInput{
		ID:            "loc-test",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		Code:          "FG-A01",
		Name:          "Aisle A01",
		Type:          LocationType("unknown"),
		Status:        LocationStatusActive,
	})
	if !errors.Is(err, ErrLocationInvalidType) {
		t.Fatalf("error = %v, want invalid location type", err)
	}
}
