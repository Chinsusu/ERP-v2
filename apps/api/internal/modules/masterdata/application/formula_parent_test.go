package application

import (
	"errors"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
)

func TestCreateFormulaInputForParentUsesFinishedProductMasterData(t *testing.T) {
	parent := domain.Item{
		ID:       "item-grn",
		SKUCode:  "GRN",
		Name:     "Dau goi Retro Nano 350ml",
		Type:     domain.ItemTypeFinishedGood,
		Status:   domain.ItemStatusActive,
		UOMBase:  "BOTTLE",
		ItemCode: "ITEM-GRN",
	}

	input, err := createFormulaInputForParent(CreateFormulaInput{
		FormulaCode:      "FORMULA-GRN-v1",
		FinishedItemID:   "free-text-should-be-replaced",
		FinishedSKU:      "FREE",
		FinishedItemName: "Free text",
		FinishedItemType: "raw_material",
		FormulaVersion:   "v1",
	}, parent)
	if err != nil {
		t.Fatalf("createFormulaInputForParent() error = %v", err)
	}

	if input.FinishedItemID != "item-grn" ||
		input.FinishedSKU != "GRN" ||
		input.FinishedItemName != "Dau goi Retro Nano 350ml" ||
		input.FinishedItemType != string(domain.ItemTypeFinishedGood) {
		t.Fatalf("input parent fields = %+v, want fields copied from product master", input)
	}
}

func TestCreateFormulaInputForParentRejectsInvalidParents(t *testing.T) {
	tests := []struct {
		name string
		item domain.Item
		want error
	}{
		{
			name: "missing parent",
			item: domain.Item{},
			want: ErrFormulaParentItemNotFound,
		},
		{
			name: "inactive parent",
			item: domain.Item{ID: "item-inactive", SKUCode: "INACTIVE", Name: "Inactive", Type: domain.ItemTypeFinishedGood, Status: domain.ItemStatusInactive},
			want: ErrFormulaParentItemInactive,
		},
		{
			name: "raw material parent",
			item: domain.Item{ID: "item-raw", SKUCode: "ACT_BAICAPIL", Name: "BAICAPIL", Type: domain.ItemTypeRawMaterial, Status: domain.ItemStatusActive},
			want: domain.ErrFormulaInvalidFinishedItemType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createFormulaInputForParent(CreateFormulaInput{}, tt.item)
			if !errors.Is(err, tt.want) {
				t.Fatalf("createFormulaInputForParent() error = %v, want %v", err, tt.want)
			}
		})
	}
}
