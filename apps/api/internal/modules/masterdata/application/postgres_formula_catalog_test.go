package application

import (
	"context"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
)

func TestPostgresFormulaCatalogRequiresDatabase(t *testing.T) {
	store := NewPostgresFormulaCatalog(nil, nil, PostgresFormulaCatalogConfig{})

	if _, err := store.List(context.Background(), domain.FormulaFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "formula-xff-v1"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if _, err := store.Create(context.Background(), CreateFormulaInput{}); err == nil {
		t.Fatal("Create() error = nil, want database required error")
	}
	if _, err := store.Activate(context.Background(), ActivateFormulaInput{}); err == nil {
		t.Fatal("Activate() error = nil, want database required error")
	}
	if _, err := store.CalculateRequirement(context.Background(), CalculateFormulaRequirementInput{}); err == nil {
		t.Fatal("CalculateRequirement() error = nil, want database required error")
	}
}
