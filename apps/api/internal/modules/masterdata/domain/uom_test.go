package domain

import (
	"errors"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewUOMValidatesDecimalScale(t *testing.T) {
	uom, err := NewUOM("kg", "Kilogram", "Kilogram", UOMGroupMass, decimal.QuantityScale, true, true, true, "")
	if err != nil {
		t.Fatalf("new uom: %v", err)
	}
	if uom.Code != "KG" || uom.DecimalScale != decimal.QuantityScale || !uom.AllowDecimal {
		t.Fatalf("uom = %+v, want normalized KG with quantity scale", uom)
	}

	_, err = NewUOM("pcs", "Piece", "Piece", UOMGroupCount, 6, false, false, true, "")
	if !errors.Is(err, ErrUOMInvalid) {
		t.Fatalf("error = %v, want invalid non-decimal scale", err)
	}
}

func TestNewUOMConversionValidatesScope(t *testing.T) {
	_, err := NewUOMConversion("global-with-item", "item-1", "KG", "G", "1000", UOMConversionGlobal, true)
	if !errors.Is(err, ErrUOMInvalid) {
		t.Fatalf("error = %v, want invalid global item scope", err)
	}

	_, err = NewUOMConversion("item-missing-item", "", "CARTON", "PCS", "48", UOMConversionItemSpecific, true)
	if !errors.Is(err, ErrUOMInvalid) {
		t.Fatalf("error = %v, want invalid item-specific scope", err)
	}
}
