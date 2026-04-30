package main

import (
	"context"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	reportingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/reporting/domain"
	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
)

func TestOperationsDailyRuntimeSignalSourceUsesDomainStores(t *testing.T) {
	receivingService := inventoryapp.NewWarehouseReceivingService(
		inventoryapp.NewPrototypeWarehouseReceivingStore(),
		nil,
		nil,
		nil,
		nil,
	)
	source := operationsDailyRuntimeSignalSource{
		receivings:       receivingService,
		carrierManifests: shippingapp.NewListCarrierManifests(shippingapp.NewPrototypeCarrierManifestStore()),
		pickTasks:        shippingapp.NewListPickTasks(shippingapp.NewPrototypePickTaskStore(mustPrototypePickTask())),
		returnReceipts:   returnsapp.NewListReturnReceipts(returnsapp.NewPrototypeReturnReceiptStore()),
	}
	filters, err := reportingdomain.NewReportFilters(reportingdomain.ReportFilterInput{
		FromDate: "2026-04-26",
		ToDate:   "2026-04-28",
	})
	if err != nil {
		t.Fatalf("new report filters: %v", err)
	}

	signals, err := source.ListOperationsDailySignals(context.Background(), filters)
	if err != nil {
		t.Fatalf("list operations daily signals: %v", err)
	}
	report, err := reportingdomain.NewOperationsDailyReport(
		filters,
		signals,
		reportingdomain.OperationsDailyOptions{},
	)
	if err != nil {
		t.Fatalf("new operations daily report: %v", err)
	}

	wantSources := map[string]string{
		"goods_receipt":    "grn-hcm-260427-draft",
		"carrier_manifest": "manifest-hcm-ghn-morning",
		"pick_task":        "pick-so-260428-0001",
		"return_receipt":   "rr-260426-0001",
	}
	for sourceType, sourceID := range wantSources {
		row, ok := findOperationsDailyRow(report.Rows, sourceType, sourceID)
		if !ok {
			t.Fatalf("missing %s %s in report rows: %+v", sourceType, sourceID, report.Rows)
		}
		if row.SourceReference.EntityType != sourceType ||
			row.SourceReference.ID != sourceID ||
			row.SourceReference.Unavailable {
			t.Fatalf("source reference for %s = %+v, want linked source reference", sourceType, row.SourceReference)
		}
	}

	manifest, ok := findOperationsDailyRow(report.Rows, "carrier_manifest", "manifest-hcm-ghn-morning")
	if !ok {
		t.Fatalf("missing carrier manifest row")
	}
	if manifest.Status != "blocked" ||
		manifest.Severity != "danger" ||
		manifest.ExceptionCode != "MISSING_HANDOVER_SCAN" {
		t.Fatalf("manifest row = %+v, want blocked missing scan signal", manifest)
	}
}

func findOperationsDailyRow(
	rows []reportingdomain.OperationsDailyRow,
	sourceType string,
	sourceID string,
) (reportingdomain.OperationsDailyRow, bool) {
	for _, row := range rows {
		if row.SourceType == sourceType && row.SourceID == sourceID {
			return row, true
		}
	}

	return reportingdomain.OperationsDailyRow{}, false
}
