package application

import (
	"context"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
)

func TestListReturnMasterDataReturnsSeededCatalogs(t *testing.T) {
	service := NewListReturnMasterData()

	data, err := service.Execute(context.Background())
	if err != nil {
		t.Fatalf("list return master data: %v", err)
	}

	if len(data.Reasons) == 0 {
		t.Fatal("return reasons are empty")
	}
	if len(data.Conditions) != 5 {
		t.Fatalf("return conditions = %d, want 5", len(data.Conditions))
	}
	if len(data.Dispositions) != 3 {
		t.Fatalf("return dispositions = %d, want 3", len(data.Dispositions))
	}
	if data.Dispositions[0].Code != domain.ReturnDispositionReusable {
		t.Fatalf("first disposition = %q, want reusable", data.Dispositions[0].Code)
	}
}
