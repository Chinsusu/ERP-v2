package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestCarrierManifestHandoverNegativeSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()

	cases := []struct {
		name             string
		code             string
		wantResult       string
		wantSeverity     string
		wantMessagePart  string
		wantLineTracking string
	}{
		{
			name:             "wrong carrier",
			code:             "VTP260426012",
			wantResult:       "MANIFEST_MISMATCH",
			wantSeverity:     "danger",
			wantMessagePart:  "carrier",
			wantLineTracking: "VTP260426012",
		},
		{
			name:             "unpacked shipment",
			code:             "GHN260426099",
			wantResult:       "INVALID_STATE",
			wantSeverity:     "danger",
			wantMessagePart:  "not packed",
			wantLineTracking: "GHN260426099",
		},
		{
			name:             "extra packed shipment",
			code:             "GHN260426004",
			wantResult:       "MANIFEST_MISMATCH",
			wantSeverity:     "danger",
			wantMessagePart:  "not expected",
			wantLineTracking: "GHN260426004",
		},
		{
			name:             "duplicate scan",
			code:             "GHN260426001",
			wantResult:       "DUPLICATE_SCAN",
			wantSeverity:     "warning",
			wantMessagePart:  "already scanned",
			wantLineTracking: "GHN260426001",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := shippingapp.NewPrototypeCarrierManifestStore()
			auditStore := audit.NewInMemoryLogStore()
			body := bytes.NewBufferString(fmt.Sprintf(
				`{"code":%q,"station_id":"dock-a","device_id":"scanner-negative","source":"qa_negative_smoke"}`,
				tc.code,
			))
			req := smokeRequestAsRole(
				httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/manifest-hcm-ghn-morning/scan", body),
				authConfig,
				auth.RoleWarehouseStaff,
			)
			req.SetPathValue("manifest_id", "manifest-hcm-ghn-morning")
			req.Header.Set(response.HeaderRequestID, "req-handover-negative-"+strings.ReplaceAll(tc.name, " ", "-"))
			rec := httptest.NewRecorder()

			verifyCarrierManifestScanHandler(shippingapp.NewVerifyCarrierManifestScan(store, auditStore)).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}
			payload := decodeSmokeSuccess[carrierManifestScanResponse](t, rec)
			if payload.Data.ResultCode != tc.wantResult ||
				payload.Data.Severity != tc.wantSeverity ||
				!strings.Contains(payload.Data.Message, tc.wantMessagePart) {
				t.Fatalf("scan payload = %+v, want %s/%s containing %q", payload.Data, tc.wantResult, tc.wantSeverity, tc.wantMessagePart)
			}
			if payload.Data.Line == nil || payload.Data.Line.TrackingNo != tc.wantLineTracking {
				t.Fatalf("line = %+v, want tracking %s", payload.Data.Line, tc.wantLineTracking)
			}
			if payload.Data.Manifest.Summary.MissingCount != 1 || payload.Data.AuditLogID == "" {
				t.Fatalf("manifest/audit = %+v/%q, want missing count unchanged and audit", payload.Data.Manifest.Summary, payload.Data.AuditLogID)
			}
			if payload.Data.ScanEvent.DeviceID != "scanner-negative" || payload.Data.ScanEvent.Source != "qa_negative_smoke" {
				t.Fatalf("scan event = %+v, want QA smoke device/source retained", payload.Data.ScanEvent)
			}

			events, err := store.ListScanEvents(context.Background(), "manifest-hcm-ghn-morning")
			if err != nil {
				t.Fatalf("list scan events: %v", err)
			}
			if len(events) != 1 || string(events[0].ResultCode) != tc.wantResult {
				t.Fatalf("scan events = %+v, want one %s event", events, tc.wantResult)
			}
		})
	}
}

func TestCarrierManifestHandoverMissingOrderBlocksConfirmSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	store := shippingapp.NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	req := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/manifest-hcm-ghn-morning/confirm-handover", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	req.SetPathValue("manifest_id", "manifest-hcm-ghn-morning")
	req.Header.Set(response.HeaderRequestID, "req-handover-missing-block")
	rec := httptest.NewRecorder()

	confirmCarrierManifestHandoverHandler(
		shippingapp.NewConfirmCarrierManifestHandover(store, auditStore, &recordingCarrierManifestSalesOrderHandover{}),
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	payload := decodeSmokeError(t, rec)
	if payload.Error.Code != response.ErrorCodeConflict ||
		payload.Error.Message != "Carrier manifest has missing orders" ||
		payload.Error.Details["missing_lines"] == nil {
		t.Fatalf("error payload = %+v, want missing order conflict details", payload.Error)
	}

	stored, err := store.Get(context.Background(), "manifest-hcm-ghn-morning")
	if err != nil {
		t.Fatalf("get manifest: %v", err)
	}
	if stored.Status != "scanning" || stored.Summary().MissingCount != 1 {
		t.Fatalf("manifest = %+v, want scanning with one missing line", stored)
	}
}
