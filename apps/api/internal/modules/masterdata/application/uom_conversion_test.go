package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestUOMCatalogConvertsGlobalMassAndVolumeToBaseUOM(t *testing.T) {
	catalog := NewPrototypeUOMCatalog()

	mass, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		SKU:         "VITC-POWDER",
		Quantity:    decimal.MustQuantity("25"),
		FromUOMCode: "KG",
		BaseUOMCode: "G",
	})
	if err != nil {
		t.Fatalf("convert mass: %v", err)
	}
	if mass.BaseQuantity != "25000.000000" || mass.ConversionFactor != "1000.000000" || mass.ConversionType != domain.UOMConversionGlobal {
		t.Fatalf("mass result = %+v, want 25000 G by factor 1000", mass)
	}

	volume, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		SKU:         "TONER-LIQUID",
		Quantity:    decimal.MustQuantity("1.5"),
		FromUOMCode: "L",
		BaseUOMCode: "ML",
	})
	if err != nil {
		t.Fatalf("convert volume: %v", err)
	}
	if volume.BaseQuantity != "1500.000000" || volume.ConversionFactor != "1000.000000" {
		t.Fatalf("volume result = %+v, want 1500 ML by factor 1000", volume)
	}
}

func TestUOMCatalogConvertsFractionalQuantitiesToBaseUOM(t *testing.T) {
	catalog := NewPrototypeUOMCatalog()

	mass, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		SKU:         "VITC-POWDER",
		Quantity:    decimal.MustQuantity("0.125"),
		FromUOMCode: "KG",
		BaseUOMCode: "G",
	})
	if err != nil {
		t.Fatalf("convert fractional mass: %v", err)
	}
	if mass.Quantity != "0.125000" ||
		mass.SourceUOMCode != "KG" ||
		mass.BaseQuantity != "125.000000" ||
		mass.BaseUOMCode != "G" ||
		mass.ConversionFactor != "1000.000000" {
		t.Fatalf("mass result = %+v, want 0.125 KG converted to 125.000000 G", mass)
	}

	volume, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		SKU:         "TONER-LIQUID",
		Quantity:    decimal.MustQuantity("0.001"),
		FromUOMCode: "L",
		BaseUOMCode: "ML",
	})
	if err != nil {
		t.Fatalf("convert fractional volume: %v", err)
	}
	if volume.BaseQuantity != "1.000000" ||
		volume.BaseUOMCode != "ML" ||
		volume.ConversionFactor != "1000.000000" {
		t.Fatalf("volume result = %+v, want 0.001 L converted to 1.000000 ML", volume)
	}
}

func TestUOMCatalogConvertsItemSpecificPackUOM(t *testing.T) {
	catalog := NewPrototypeUOMCatalog()

	result, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		ItemID:      "item-serum-30ml",
		SKU:         "SERUM-30ML",
		Quantity:    decimal.MustQuantity("2"),
		FromUOMCode: "CARTON",
		BaseUOMCode: "PCS",
	})
	if err != nil {
		t.Fatalf("convert carton: %v", err)
	}
	if result.BaseQuantity != "96.000000" || result.ConversionFactor != "48.000000" || result.ConversionType != domain.UOMConversionItemSpecific {
		t.Fatalf("result = %+v, want 96 PCS by item-specific factor 48", result)
	}
}

func TestUOMCatalogPassesThroughBaseUOM(t *testing.T) {
	catalog := NewPrototypeUOMCatalog()

	result, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		ItemID:      "item-serum-30ml",
		SKU:         "SERUM-30ML",
		Quantity:    decimal.MustQuantity("7"),
		FromUOMCode: "PCS",
		BaseUOMCode: "PCS",
	})
	if err != nil {
		t.Fatalf("base passthrough: %v", err)
	}
	if !result.IsBasePassthrough || result.BaseQuantity != "7.000000" || result.ConversionFactor != "1.000000" {
		t.Fatalf("result = %+v, want direct passthrough", result)
	}
}

func TestUOMCatalogAcceptsOperationalPackagingUnits(t *testing.T) {
	catalog := NewPrototypeUOMCatalog()

	for _, code := range []string{"BAG", "PACK", "ROLL", "CM"} {
		result, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
			SKU:         "PACKAGING-SHEET-IMPORT",
			Quantity:    decimal.MustQuantity("1"),
			FromUOMCode: code,
			BaseUOMCode: code,
		})
		if err != nil {
			t.Fatalf("base passthrough for %s: %v", code, err)
		}
		if !result.IsBasePassthrough || result.BaseUOMCode.String() != code {
			t.Fatalf("result for %s = %+v, want direct passthrough", code, result)
		}
	}
}

func TestUOMCatalogReturnsStandardDetailsForMissingConversion(t *testing.T) {
	catalog := NewPrototypeUOMCatalog()

	_, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		ItemID:      "item-cream-50g",
		SKU:         "CREAM-50G",
		Quantity:    decimal.MustQuantity("2"),
		FromUOMCode: "CARTON",
		BaseUOMCode: "PCS",
	})
	if !errors.Is(err, domain.ErrUOMConversionMissing) {
		t.Fatalf("error = %v, want missing conversion", err)
	}
	var conversionErr domain.UOMConversionError
	if !errors.As(err, &conversionErr) {
		t.Fatalf("error type = %T, want UOMConversionError", err)
	}
	details := conversionErr.Details()
	if details["sku_code"] != "CREAM-50G" || details["from_uom_code"] != "CARTON" || details["base_uom_code"] != "PCS" {
		t.Fatalf("details = %+v, want sku/from/to uom", details)
	}
}

func TestUOMCatalogRejectsInvalidAndInactiveConversion(t *testing.T) {
	catalog := NewPrototypeUOMCatalog()
	_, err := catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		SKU:         "BAD-UOM",
		Quantity:    decimal.MustQuantity("1"),
		FromUOMCode: "UNKNOWN",
		BaseUOMCode: "PCS",
	})
	if !errors.Is(err, domain.ErrUOMInvalid) {
		t.Fatalf("error = %v, want invalid uom", err)
	}

	inactive, err := domain.NewUOMConversion("item-serum-carton-pcs", "item-serum-30ml", "CARTON", "PCS", "48", domain.UOMConversionItemSpecific, false)
	if err != nil {
		t.Fatalf("new inactive conversion: %v", err)
	}
	catalog.UpsertConversion(inactive)
	_, err = catalog.ConvertToBase(context.Background(), ConvertToBaseInput{
		ItemID:      "item-serum-30ml",
		SKU:         "SERUM-30ML",
		Quantity:    decimal.MustQuantity("2"),
		FromUOMCode: "CARTON",
		BaseUOMCode: "PCS",
	})
	if !errors.Is(err, domain.ErrUOMConversionInactive) {
		t.Fatalf("error = %v, want inactive conversion", err)
	}
}
