package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresSubcontractFactoryClaimStorePersistsClaim(t *testing.T) {
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

	if err := seedSubcontractOrderSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	openedAt := time.Date(2026, 5, 2, 15, 0, 0, 0, time.UTC)
	order := buildQCSubcontractOrderForFactoryClaimStore(t, suffix, openedAt)
	orderStore := NewPostgresSubcontractOrderStore(
		db,
		PostgresSubcontractOrderStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed qc subcontract order: %v", err)
	}

	builder := NewSubcontractFactoryClaimService()
	builder.clock = func() time.Time { return openedAt }
	buildResult, err := builder.BuildClaim(ctx, BuildSubcontractFactoryClaimInput{
		ID:              "sfc-s16-06-01-" + suffix,
		ClaimNo:         "SFC-S16-06-01-" + suffix,
		Order:           order,
		ReceiptID:       "sfgr-s16-06-01-" + suffix,
		ReceiptNo:       "SFGR-S16-06-01-" + suffix,
		ReasonCode:      "packaging_damaged",
		Reason:          "Outer cartons crushed and bottle caps scratched",
		Severity:        "p1",
		AffectedQty:     "12",
		UOMCode:         "PCS",
		BaseAffectedQty: "12",
		BaseUOMCode:     "PCS",
		OwnerID:         "factory-owner",
		OpenedBy:        "qa-user",
		OpenedAt:        openedAt,
		ActorID:         "qa-user",
		Evidence: []BuildSubcontractFactoryClaimEvidenceInput{{
			ID:           "sfc-evidence-s16-06-01-" + suffix,
			EvidenceType: "qc_photo",
			ObjectKey:    "subcontract/sfc-s16-06-01/damaged-cartons.jpg",
		}},
	})
	if err != nil {
		t.Fatalf("build factory claim: %v", err)
	}

	store := NewPostgresSubcontractFactoryClaimStore(
		db,
		PostgresSubcontractFactoryClaimStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	if err := store.Save(ctx, buildResult.Claim); err != nil {
		t.Fatalf("save factory claim: %v", err)
	}

	loaded, err := store.Get(ctx, buildResult.Claim.ID)
	if err != nil {
		t.Fatalf("get factory claim: %v", err)
	}
	if loaded.ID != buildResult.Claim.ID ||
		loaded.Status != productiondomain.SubcontractFactoryClaimStatusOpen ||
		loaded.SubcontractOrderID != order.ID ||
		loaded.FactoryName != order.FactoryName ||
		loaded.AffectedQty.String() != "12.000000" ||
		loaded.BaseAffectedQty.String() != "12.000000" ||
		loaded.ReasonCode != "PACKAGING_DAMAGED" ||
		loaded.Severity != "P1" ||
		!loaded.DueAt.Equal(openedAt.AddDate(0, 0, order.ClaimWindowDays)) ||
		len(loaded.Evidence) != 1 ||
		loaded.Evidence[0].ObjectKey != "subcontract/sfc-s16-06-01/damaged-cartons.jpg" {
		t.Fatalf("loaded claim = %+v, want persisted open factory claim", loaded)
	}
	claims, err := store.ListBySubcontractOrder(ctx, order.OrderNo)
	if err != nil {
		t.Fatalf("list factory claims: %v", err)
	}
	if len(claims) != 1 || claims[0].ID != buildResult.Claim.ID {
		t.Fatalf("claims = %+v, want persisted claim by order", claims)
	}

	acknowledged, err := loaded.Acknowledge("factory-owner", openedAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("acknowledge factory claim: %v", err)
	}
	if err := store.Save(ctx, acknowledged); err != nil {
		t.Fatalf("save acknowledged factory claim: %v", err)
	}
	resolved, err := acknowledged.Resolve("qa-lead", "factory accepted replacement batch", openedAt.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("resolve factory claim: %v", err)
	}
	if err := store.Save(ctx, resolved); err != nil {
		t.Fatalf("save resolved factory claim: %v", err)
	}
	loaded, err = store.Get(ctx, resolved.ClaimNo)
	if err != nil {
		t.Fatalf("get resolved factory claim: %v", err)
	}
	if loaded.Status != productiondomain.SubcontractFactoryClaimStatusResolved ||
		loaded.AcknowledgedBy != "factory-owner" ||
		loaded.ResolvedBy != "qa-lead" ||
		loaded.ResolutionNote != "factory accepted replacement batch" ||
		loaded.Version != 3 {
		t.Fatalf("resolved claim = %+v, want persisted lifecycle actor refs", loaded)
	}
}

func TestPostgresSubcontractFactoryClaimStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSubcontractFactoryClaimStore(nil, PostgresSubcontractFactoryClaimStoreConfig{})

	if err := store.Save(context.Background(), productiondomain.SubcontractFactoryClaim{}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "sfc-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if _, err := store.ListBySubcontractOrder(context.Background(), "sco-missing"); err == nil {
		t.Fatal("ListBySubcontractOrder() error = nil, want database required error")
	}
}

func buildQCSubcontractOrderForFactoryClaimStore(
	t *testing.T,
	suffix string,
	changedAt time.Time,
) productiondomain.SubcontractOrder {
	t.Helper()

	order := buildMassProductionSubcontractOrderForFinishedGoodsReceiptStore(t, suffix)
	received, err := order.MarkFinishedGoodsReceived("warehouse-user", changedAt.Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("mark finished goods received: %v", err)
	}
	qc, err := received.StartQC("qa-user", changedAt.Add(-time.Hour))
	if err != nil {
		t.Fatalf("start qc: %v", err)
	}

	return qc
}
