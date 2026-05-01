package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	qcdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testInboundQCOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresInboundQCInspectionStorePersistsDecisionLifecycle(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := seedInboundQCInspectionSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	store := NewPostgresInboundQCInspectionStore(db, PostgresInboundQCInspectionStoreConfig{
		DefaultOrgID: testInboundQCOrgID,
	})
	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())

	cases := []struct {
		name           string
		result         qcdomain.InboundQCResult
		passedQuantity decimal.Decimal
		failedQuantity decimal.Decimal
		holdQuantity   decimal.Decimal
		reason         string
	}{
		{
			name:           "pass",
			result:         qcdomain.InboundQCResultPass,
			passedQuantity: decimal.MustQuantity("24"),
			failedQuantity: decimal.MustQuantity("0"),
			holdQuantity:   decimal.MustQuantity("0"),
		},
		{
			name:           "fail",
			result:         qcdomain.InboundQCResultFail,
			passedQuantity: decimal.MustQuantity("0"),
			failedQuantity: decimal.MustQuantity("24"),
			holdQuantity:   decimal.MustQuantity("0"),
			reason:         "damaged packaging",
		},
		{
			name:           "hold",
			result:         qcdomain.InboundQCResultHold,
			passedQuantity: decimal.MustQuantity("0"),
			failedQuantity: decimal.MustQuantity("0"),
			holdQuantity:   decimal.MustQuantity("24"),
			reason:         "pending COA",
		},
		{
			name:           "partial",
			result:         qcdomain.InboundQCResultPartial,
			passedQuantity: decimal.MustQuantity("16"),
			failedQuantity: decimal.MustQuantity("3"),
			holdQuantity:   decimal.MustQuantity("5"),
			reason:         "sample hold split",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			created := newPostgresInboundQCInspectionFixture(t, suffix, tc.name)
			if err := store.Save(ctx, created); err != nil {
				t.Fatalf("save pending inspection: %v", err)
			}

			started, err := created.Start("user-qa", created.CreatedAt.Add(10*time.Minute))
			if err != nil {
				t.Fatalf("start inspection: %v", err)
			}
			if err := store.Save(ctx, started); err != nil {
				t.Fatalf("save started inspection: %v", err)
			}
			completed, err := started.RecordDecision(qcdomain.InboundQCDecisionInput{
				Result:         tc.result,
				PassedQuantity: tc.passedQuantity,
				FailedQuantity: tc.failedQuantity,
				HoldQuantity:   tc.holdQuantity,
				Checklist:      postgresInboundQCChecklist(tc.result),
				Reason:         tc.reason,
				ActorID:        "user-qa",
				ChangedAt:      created.CreatedAt.Add(20 * time.Minute),
			})
			if err != nil {
				t.Fatalf("record decision: %v", err)
			}
			if err := store.Save(ctx, completed); err != nil {
				t.Fatalf("save completed inspection: %v", err)
			}

			loaded, err := store.Get(ctx, created.ID)
			if err != nil {
				t.Fatalf("get persisted inspection: %v", err)
			}
			if loaded.Status != qcdomain.InboundQCInspectionStatusCompleted ||
				loaded.Result != tc.result ||
				loaded.PassedQuantity.String() != tc.passedQuantity.String() ||
				loaded.FailedQuantity.String() != tc.failedQuantity.String() ||
				loaded.HoldQuantity.String() != tc.holdQuantity.String() ||
				len(loaded.Checklist) != 3 {
				t.Fatalf("loaded inspection = %+v, want completed %s decision with checklist", loaded, tc.result)
			}
		})
	}

	rows, err := store.List(ctx, NewInboundQCInspectionFilter(
		qcdomain.InboundQCInspectionStatusCompleted,
		"",
		"",
		"wh-hcm-fg",
	))
	if err != nil {
		t.Fatalf("list completed inspections: %v", err)
	}
	if !containsInboundQCInspection(rows, "iqc-s10-04-02-"+suffix+"-partial") {
		t.Fatalf("completed inspection list missing partial decision")
	}
}

func newPostgresInboundQCInspectionFixture(
	t *testing.T,
	suffix string,
	resultName string,
) qcdomain.InboundQCInspection {
	t.Helper()

	createdAt := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	inspection, err := qcdomain.NewInboundQCInspection(qcdomain.NewInboundQCInspectionInput{
		ID:                  "iqc-s10-04-02-" + suffix + "-" + resultName,
		OrgID:               "org-my-pham",
		GoodsReceiptID:      "grn-s10-04-02-" + suffix,
		GoodsReceiptNo:      "GRN-S10-04-02-" + suffix,
		GoodsReceiptLineID:  "grn-line-s10-04-02-" + suffix + "-" + resultName,
		PurchaseOrderID:     "po-s10-04-02-" + suffix,
		PurchaseOrderLineID: "po-line-s10-04-02-" + suffix,
		ItemID:              "item-serum-30ml",
		SKU:                 "SERUM-30ML",
		ItemName:            "Vitamin C Serum",
		BatchID:             "batch-s10-04-02-" + suffix,
		BatchNo:             "LOT-S10-04-02",
		LotNo:               "LOT-S10-04-02",
		ExpiryDate:          time.Date(2027, 5, 1, 0, 0, 0, 0, time.UTC),
		WarehouseID:         "wh-hcm-fg",
		LocationID:          "loc-hcm-fg-qc-01",
		Quantity:            decimal.MustQuantity("24"),
		UOMCode:             "EA",
		InspectorID:         "user-qa",
		Checklist:           postgresInboundQCChecklist(qcdomain.InboundQCChecklistStatusPending),
		CreatedAt:           createdAt,
		CreatedBy:           "user-warehouse",
	})
	if err != nil {
		t.Fatalf("new inspection fixture: %v", err)
	}

	return inspection
}

func postgresInboundQCChecklist(
	status any,
) []qcdomain.NewInboundQCChecklistItemInput {
	checkStatus := qcdomain.InboundQCChecklistStatus(fmt.Sprint(status))
	if checkStatus == "" || checkStatus == qcdomain.InboundQCChecklistStatusPending {
		return []qcdomain.NewInboundQCChecklistItemInput{
			{ID: "check-packaging", Code: "PACKAGING", Label: "Packaging condition", Required: true},
			{ID: "check-lot-expiry", Code: "LOT_EXPIRY", Label: "Lot and expiry match delivery", Required: true},
			{ID: "check-sample", Code: "SAMPLE", Label: "Sample retained", Required: false},
		}
	}
	if checkStatus == qcdomain.InboundQCChecklistStatusFail {
		return []qcdomain.NewInboundQCChecklistItemInput{
			{ID: "check-packaging", Code: "PACKAGING", Label: "Packaging condition", Required: true, Status: qcdomain.InboundQCChecklistStatusFail},
			{ID: "check-lot-expiry", Code: "LOT_EXPIRY", Label: "Lot and expiry match delivery", Required: true, Status: qcdomain.InboundQCChecklistStatusPass},
			{ID: "check-sample", Code: "SAMPLE", Label: "Sample retained", Required: false, Status: qcdomain.InboundQCChecklistStatusNotApplicable},
		}
	}

	return []qcdomain.NewInboundQCChecklistItemInput{
		{ID: "check-packaging", Code: "PACKAGING", Label: "Packaging condition", Required: true, Status: qcdomain.InboundQCChecklistStatusPass},
		{ID: "check-lot-expiry", Code: "LOT_EXPIRY", Label: "Lot and expiry match delivery", Required: true, Status: qcdomain.InboundQCChecklistStatusPass},
		{ID: "check-sample", Code: "SAMPLE", Label: "Sample retained", Required: false, Status: qcdomain.InboundQCChecklistStatusNotApplicable},
	}
}

func containsInboundQCInspection(rows []qcdomain.InboundQCInspection, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}

	return false
}

func seedInboundQCInspectionSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testInboundQCOrgID,
	)

	return err
}
