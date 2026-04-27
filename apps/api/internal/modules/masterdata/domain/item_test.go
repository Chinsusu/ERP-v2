package domain

import (
	"errors"
	"testing"
)

func TestNewItemNormalizesCosmeticMasterDataFields(t *testing.T) {
	item, err := NewItem(NewItemInput{
		ID:               "item-test",
		ItemCode:         " item-serum-test ",
		SKUCode:          " serum-test-30ml ",
		Name:             "Serum Test 30ml",
		Type:             ItemTypeFinishedGood,
		Group:            "serum",
		BrandCode:        "myh",
		UOMBase:          "ea",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    365,
		QCRequired:       true,
		Status:           ItemStatusActive,
		StandardCost:     "12500",
		IsSellable:       true,
		IsProducible:     true,
		SpecVersion:      "SPEC-TEST",
	})
	if err != nil {
		t.Fatalf("new item: %v", err)
	}

	if item.ItemCode != "ITEM-SERUM-TEST" || item.SKUCode != "SERUM-TEST-30ML" {
		t.Fatalf("codes = %q/%q, want normalized uppercase", item.ItemCode, item.SKUCode)
	}
	if item.UOMBase != "EA" || item.UOMPurchase != "EA" || item.UOMIssue != "EA" {
		t.Fatalf("uom = %q/%q/%q, want EA defaults", item.UOMBase, item.UOMPurchase, item.UOMIssue)
	}
	if item.BrandCode != "MYH" {
		t.Fatalf("brand = %q, want MYH", item.BrandCode)
	}
}

func TestNewItemRequiresShelfLifeWhenExpiryControlled(t *testing.T) {
	_, err := NewItem(NewItemInput{
		ID:               "item-test",
		ItemCode:         "ITEM-SERUM-TEST",
		SKUCode:          "SERUM-TEST-30ML",
		Name:             "Serum Test 30ml",
		Type:             ItemTypeFinishedGood,
		UOMBase:          "EA",
		ExpiryControlled: true,
		Status:           ItemStatusActive,
	})
	if !errors.Is(err, ErrItemInvalidShelfLife) {
		t.Fatalf("error = %v, want invalid shelf life", err)
	}
}

func TestItemFilterMatchesCosmeticFields(t *testing.T) {
	item, err := NewItem(NewItemInput{
		ID:           "item-test",
		ItemCode:     "ITEM-SERUM-TEST",
		SKUCode:      "SERUM-TEST-30ML",
		Name:         "Serum Test 30ml",
		Type:         ItemTypeFinishedGood,
		Group:        "serum",
		BrandCode:    "MYH",
		UOMBase:      "EA",
		Status:       ItemStatusActive,
		StandardCost: "12500",
		IsSellable:   true,
		IsProducible: true,
	})
	if err != nil {
		t.Fatalf("new item: %v", err)
	}

	filter := NewItemFilter("myh", ItemStatusActive, ItemTypeFinishedGood, 1, 20)
	if !filter.Matches(item) {
		t.Fatal("filter did not match brand/status/type")
	}
}
