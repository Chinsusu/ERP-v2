package application

import (
	"reflect"
	"strings"
	"testing"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
)

func TestBuildPostgresProductionPlanListQueryAppliesFiltersOnce(t *testing.T) {
	query, args := buildPostgresProductionPlanListQuery("00000000-0000-4000-8000-000000000001", ProductionPlanFilter{
		Search:       "xff",
		OutputItemID: "item-xff-150",
		Statuses: []productiondomain.ProductionPlanStatus{
			productiondomain.ProductionPlanStatusDraft,
		},
	})

	if strings.Count(query, "WHERE") != 1 {
		t.Fatalf("query = %q, want exactly one WHERE clause", query)
	}
	if !strings.Contains(query, "org_id = $1::uuid") ||
		!strings.Contains(query, "output_item_ref = $2") ||
		!strings.Contains(query, "lower(plan_no) LIKE $3") ||
		!strings.Contains(query, "status IN ($4)") {
		t.Fatalf("query = %q, want org, output item, search and status filters", query)
	}

	wantArgs := []any{
		"00000000-0000-4000-8000-000000000001",
		"item-xff-150",
		"%xff%",
		"draft",
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", args, wantArgs)
	}
}
