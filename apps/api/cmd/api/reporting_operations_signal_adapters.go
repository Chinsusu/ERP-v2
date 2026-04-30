package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	qcdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/domain"
	reportingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/reporting/domain"
	returnsdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type operationsDailySignalSource interface {
	ListOperationsDailySignals(
		ctx context.Context,
		filters reportingdomain.ReportFilters,
	) ([]reportingdomain.OperationsDailySignal, error)
}

type prototypeOperationsDailySignalSource struct{}

type operationsDailyRuntimeSignalSource struct {
	receivings           warehouseDailyBoardReceivingLister
	inboundQC            warehouseDailyBoardInboundQCLister
	carrierManifests     operationsDailyCarrierManifestLister
	pickTasks            operationsDailyPickTaskLister
	returnReceipts       operationsDailyReturnReceiptLister
	stockCounts          operationsDailyStockCountLister
	subcontractOrders    warehouseDailyBoardSubcontractOrderLister
	subcontractTransfers warehouseDailyBoardSubcontractMaterialTransferLister
}

type operationsDailyCarrierManifestLister interface {
	Execute(context.Context, shippingdomain.CarrierManifestFilter) ([]shippingdomain.CarrierManifest, error)
}

type operationsDailyPickTaskLister interface {
	Execute(context.Context, shippingapp.PickTaskFilter) ([]shippingdomain.PickTask, error)
}

type operationsDailyReturnReceiptLister interface {
	Execute(context.Context, returnsdomain.ReturnReceiptFilter) ([]returnsdomain.ReturnReceipt, error)
}

type operationsDailyStockCountLister interface {
	Execute(context.Context) ([]inventorydomain.StockCountSession, error)
}

func (prototypeOperationsDailySignalSource) ListOperationsDailySignals(
	_ context.Context,
	_ reportingdomain.ReportFilters,
) ([]reportingdomain.OperationsDailySignal, error) {
	return prototypeOperationsDailySignals(), nil
}

func (s operationsDailyRuntimeSignalSource) ListOperationsDailySignals(
	ctx context.Context,
	filters reportingdomain.ReportFilters,
) ([]reportingdomain.OperationsDailySignal, error) {
	signals := make([]reportingdomain.OperationsDailySignal, 0)
	var err error

	if s.receivings != nil {
		signals, err = s.appendReceivingSignals(ctx, filters, signals)
		if err != nil {
			return nil, err
		}
	}
	if s.inboundQC != nil {
		signals, err = s.appendInboundQCSignals(ctx, filters, signals)
		if err != nil {
			return nil, err
		}
	}
	if s.pickTasks != nil {
		signals, err = s.appendPickTaskSignals(ctx, filters, signals)
		if err != nil {
			return nil, err
		}
	}
	if s.carrierManifests != nil {
		signals, err = s.appendCarrierManifestSignals(ctx, filters, signals)
		if err != nil {
			return nil, err
		}
	}
	if s.returnReceipts != nil {
		signals, err = s.appendReturnReceiptSignals(ctx, filters, signals)
		if err != nil {
			return nil, err
		}
	}
	if s.stockCounts != nil {
		signals, err = s.appendStockCountSignals(ctx, filters, signals)
		if err != nil {
			return nil, err
		}
	}
	if s.subcontractOrders != nil && s.subcontractTransfers != nil {
		signals, err = s.appendSubcontractTransferSignals(ctx, filters, signals)
		if err != nil {
			return nil, err
		}
	}

	return signals, nil
}

func (s operationsDailyRuntimeSignalSource) appendReceivingSignals(
	ctx context.Context,
	filters reportingdomain.ReportFilters,
	signals []reportingdomain.OperationsDailySignal,
) ([]reportingdomain.OperationsDailySignal, error) {
	rows, err := s.receivings.ListWarehouseReceivings(
		ctx,
		inventorydomain.NewWarehouseReceivingFilter(filters.WarehouseID, ""),
	)
	if err != nil {
		return nil, err
	}
	for _, receipt := range rows {
		status, severity := operationsDailyReceivingStatus(receipt.Status)
		signals = append(signals, reportingdomain.OperationsDailySignal{
			ID:            operationsDailySignalID("goods_receipt", receipt.ID),
			Area:          reportingdomain.OperationsDailyAreaInbound,
			SourceType:    "goods_receipt",
			SourceID:      receipt.ID,
			RefNo:         firstNonBlankOperationsDaily(receipt.ReceiptNo, receipt.ID),
			Title:         operationsDailyReceivingTitle(receipt.Status),
			WarehouseID:   receipt.WarehouseID,
			WarehouseCode: receipt.WarehouseCode,
			BusinessDate:  firstNonZeroOperationsDailyTime(receipt.UpdatedAt, receipt.CreatedAt),
			Status:        status,
			Severity:      severity,
			Quantity:      operationsDailyReceivingQuantity(receipt),
			UOMCode:       operationsDailyReceivingUOM(receipt),
			Owner:         firstNonBlankOperationsDaily(receipt.PostedBy, receipt.InspectReadyBy, receipt.SubmittedBy, receipt.CreatedBy),
		})
	}

	return signals, nil
}

func (s operationsDailyRuntimeSignalSource) appendInboundQCSignals(
	ctx context.Context,
	filters reportingdomain.ReportFilters,
	signals []reportingdomain.OperationsDailySignal,
) ([]reportingdomain.OperationsDailySignal, error) {
	rows, err := s.inboundQC.ListInboundQCInspections(
		ctx,
		qcapp.NewInboundQCInspectionFilter("", "", "", filters.WarehouseID),
	)
	if err != nil {
		return nil, err
	}
	for _, inspection := range rows {
		status, severity, exceptionCode := operationsDailyInboundQCStatus(inspection)
		signals = append(signals, reportingdomain.OperationsDailySignal{
			ID:            operationsDailySignalID("inbound_qc", inspection.ID),
			Area:          reportingdomain.OperationsDailyAreaQC,
			SourceType:    "inbound_qc",
			SourceID:      inspection.ID,
			RefNo:         strings.ToUpper(firstNonBlankOperationsDaily(inspection.ID, inspection.GoodsReceiptNo)),
			Title:         operationsDailyInboundQCTitle(inspection),
			WarehouseID:   inspection.WarehouseID,
			BusinessDate:  firstNonZeroOperationsDailyTime(inspection.DecidedAt, inspection.StartedAt, inspection.UpdatedAt, inspection.CreatedAt),
			Status:        status,
			Severity:      severity,
			Quantity:      operationsDailyQuantityString(inspection.Quantity),
			UOMCode:       inspection.UOMCode.String(),
			ExceptionCode: exceptionCode,
			Owner:         firstNonBlankOperationsDaily(inspection.DecidedBy, inspection.InspectorID, inspection.UpdatedBy, inspection.CreatedBy),
		})
	}

	return signals, nil
}

func (s operationsDailyRuntimeSignalSource) appendPickTaskSignals(
	ctx context.Context,
	filters reportingdomain.ReportFilters,
	signals []reportingdomain.OperationsDailySignal,
) ([]reportingdomain.OperationsDailySignal, error) {
	rows, err := s.pickTasks.Execute(ctx, shippingapp.PickTaskFilter{WarehouseID: filters.WarehouseID})
	if err != nil {
		return nil, err
	}
	for _, task := range rows {
		status, severity, exceptionCode := operationsDailyPickTaskStatus(task.Status)
		signals = append(signals, reportingdomain.OperationsDailySignal{
			ID:            operationsDailySignalID("pick_task", task.ID),
			Area:          reportingdomain.OperationsDailyAreaOutbound,
			SourceType:    "pick_task",
			SourceID:      task.ID,
			RefNo:         firstNonBlankOperationsDaily(task.PickTaskNo, task.ID),
			Title:         operationsDailyPickTaskTitle(task.Status),
			WarehouseID:   task.WarehouseID,
			WarehouseCode: task.WarehouseCode,
			BusinessDate:  firstNonZeroOperationsDailyTime(task.UpdatedAt, task.CreatedAt),
			Status:        status,
			Severity:      severity,
			Quantity:      operationsDailyPickTaskQuantity(task),
			UOMCode:       operationsDailyPickTaskUOM(task),
			ExceptionCode: exceptionCode,
			Owner:         firstNonBlankOperationsDaily(task.CompletedBy, task.StartedBy, task.AssignedTo),
		})
	}

	return signals, nil
}

func (s operationsDailyRuntimeSignalSource) appendCarrierManifestSignals(
	ctx context.Context,
	filters reportingdomain.ReportFilters,
	signals []reportingdomain.OperationsDailySignal,
) ([]reportingdomain.OperationsDailySignal, error) {
	rows, err := s.carrierManifests.Execute(
		ctx,
		shippingdomain.NewCarrierManifestFilter(filters.WarehouseID, "", "", ""),
	)
	if err != nil {
		return nil, err
	}
	for _, manifest := range rows {
		status, severity, exceptionCode := operationsDailyCarrierManifestStatus(manifest)
		signals = append(signals, reportingdomain.OperationsDailySignal{
			ID:            operationsDailySignalID("carrier_manifest", manifest.ID),
			Area:          reportingdomain.OperationsDailyAreaOutbound,
			SourceType:    "carrier_manifest",
			SourceID:      manifest.ID,
			RefNo:         firstNonBlankOperationsDaily(manifest.ID, manifest.HandoverBatch),
			Title:         operationsDailyCarrierManifestTitle(manifest),
			WarehouseID:   manifest.WarehouseID,
			WarehouseCode: manifest.WarehouseCode,
			BusinessDate:  operationsDailyDateFromString(manifest.Date, manifest.CreatedAt),
			Status:        status,
			Severity:      severity,
			Quantity:      operationsDailyCountQuantity(manifest.Summary().ExpectedCount),
			UOMCode:       "PCS",
			ExceptionCode: exceptionCode,
			Owner:         manifest.Owner,
		})
	}

	return signals, nil
}

func (s operationsDailyRuntimeSignalSource) appendReturnReceiptSignals(
	ctx context.Context,
	filters reportingdomain.ReportFilters,
	signals []reportingdomain.OperationsDailySignal,
) ([]reportingdomain.OperationsDailySignal, error) {
	rows, err := s.returnReceipts.Execute(
		ctx,
		returnsdomain.NewReturnReceiptFilter(filters.WarehouseID, ""),
	)
	if err != nil {
		return nil, err
	}
	for _, receipt := range rows {
		status, severity := operationsDailyReturnReceiptStatus(receipt.Status)
		signals = append(signals, reportingdomain.OperationsDailySignal{
			ID:            operationsDailySignalID("return_receipt", receipt.ID),
			Area:          reportingdomain.OperationsDailyAreaReturns,
			SourceType:    "return_receipt",
			SourceID:      receipt.ID,
			RefNo:         firstNonBlankOperationsDaily(receipt.ReceiptNo, receipt.ID),
			Title:         operationsDailyReturnReceiptTitle(receipt.Status),
			WarehouseID:   receipt.WarehouseID,
			WarehouseCode: receipt.WarehouseCode,
			BusinessDate:  firstNonZeroOperationsDailyTime(receipt.ReceivedAt, receipt.CreatedAt),
			Status:        status,
			Severity:      severity,
			Quantity:      operationsDailyReturnReceiptQuantity(receipt),
			UOMCode:       "PCS",
			Owner:         receipt.ReceivedBy,
		})
	}

	return signals, nil
}

func (s operationsDailyRuntimeSignalSource) appendStockCountSignals(
	ctx context.Context,
	filters reportingdomain.ReportFilters,
	signals []reportingdomain.OperationsDailySignal,
) ([]reportingdomain.OperationsDailySignal, error) {
	rows, err := s.stockCounts.Execute(ctx)
	if err != nil {
		return nil, err
	}
	for _, session := range rows {
		if filters.WarehouseID != "" && session.WarehouseID != filters.WarehouseID {
			continue
		}
		status, severity, exceptionCode := operationsDailyStockCountStatus(session)
		signals = append(signals, reportingdomain.OperationsDailySignal{
			ID:            operationsDailySignalID("stock_count", session.ID),
			Area:          reportingdomain.OperationsDailyAreaStock,
			SourceType:    "stock_count",
			SourceID:      session.ID,
			RefNo:         firstNonBlankOperationsDaily(session.CountNo, session.ID),
			Title:         operationsDailyStockCountTitle(session),
			WarehouseID:   session.WarehouseID,
			WarehouseCode: session.WarehouseCode,
			BusinessDate:  firstNonZeroOperationsDailyTime(session.UpdatedAt, session.CreatedAt),
			Status:        status,
			Severity:      severity,
			ExceptionCode: exceptionCode,
			Owner:         firstNonBlankOperationsDaily(session.SubmittedBy, session.CreatedBy),
		})
	}

	return signals, nil
}

func (s operationsDailyRuntimeSignalSource) appendSubcontractTransferSignals(
	ctx context.Context,
	filters reportingdomain.ReportFilters,
	signals []reportingdomain.OperationsDailySignal,
) ([]reportingdomain.OperationsDailySignal, error) {
	orders, err := s.subcontractOrders.ListSubcontractOrders(ctx, productionapp.SubcontractOrderFilter{})
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		transfers, err := s.subcontractTransfers.ListBySubcontractOrder(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		for _, transfer := range transfers {
			if filters.WarehouseID != "" && transfer.SourceWarehouseID != filters.WarehouseID {
				continue
			}
			signals = append(signals, reportingdomain.OperationsDailySignal{
				ID:            operationsDailySignalID("subcontract_order", transfer.ID),
				Area:          reportingdomain.OperationsDailyAreaSubcontract,
				SourceType:    "subcontract_order",
				SourceID:      order.ID,
				RefNo:         firstNonBlankOperationsDaily(transfer.TransferNo, order.OrderNo),
				Title:         operationsDailySubcontractTransferTitle(transfer),
				WarehouseID:   transfer.SourceWarehouseID,
				WarehouseCode: transfer.SourceWarehouseCode,
				BusinessDate:  firstNonZeroOperationsDailyTime(transfer.HandoverAt, transfer.CreatedAt),
				Status:        operationsDailySubcontractStatus(order.Status),
				Severity:      reportingdomain.OperationsDailySeverityNormal,
				Quantity:      operationsDailySubcontractTransferQuantity(transfer),
				UOMCode:       operationsDailySubcontractTransferUOM(transfer),
				Owner:         firstNonBlankOperationsDaily(transfer.HandoverBy, transfer.CreatedBy),
			})
		}
	}

	return signals, nil
}

func operationsDailyReceivingStatus(
	status inventorydomain.WarehouseReceivingStatus,
) (reportingdomain.OperationsDailyStatus, reportingdomain.OperationsDailySeverity) {
	switch inventorydomain.NormalizeWarehouseReceivingStatus(status) {
	case inventorydomain.WarehouseReceivingStatusPosted:
		return reportingdomain.OperationsDailyStatusCompleted, reportingdomain.OperationsDailySeverityNormal
	default:
		return reportingdomain.OperationsDailyStatusPending, reportingdomain.OperationsDailySeverityWarning
	}
}

func operationsDailyInboundQCStatus(
	inspection qcdomain.InboundQCInspection,
) (reportingdomain.OperationsDailyStatus, reportingdomain.OperationsDailySeverity, string) {
	switch qcdomain.NormalizeInboundQCInspectionStatus(inspection.Status) {
	case qcdomain.InboundQCInspectionStatusPending:
		return reportingdomain.OperationsDailyStatusPending, reportingdomain.OperationsDailySeverityWarning, ""
	case qcdomain.InboundQCInspectionStatusInProgress:
		return reportingdomain.OperationsDailyStatusInProgress, reportingdomain.OperationsDailySeverityNormal, ""
	case qcdomain.InboundQCInspectionStatusCompleted:
		switch qcdomain.NormalizeInboundQCResult(inspection.Result) {
		case qcdomain.InboundQCResultPass:
			return reportingdomain.OperationsDailyStatusCompleted, reportingdomain.OperationsDailySeverityNormal, ""
		case qcdomain.InboundQCResultFail:
			return reportingdomain.OperationsDailyStatusException, reportingdomain.OperationsDailySeverityDanger, "QC_FAIL"
		case qcdomain.InboundQCResultHold:
			return reportingdomain.OperationsDailyStatusBlocked, reportingdomain.OperationsDailySeverityWarning, "QC_HOLD"
		case qcdomain.InboundQCResultPartial:
			return reportingdomain.OperationsDailyStatusException, reportingdomain.OperationsDailySeverityWarning, "QC_PARTIAL"
		}
	}

	return reportingdomain.OperationsDailyStatusBlocked, reportingdomain.OperationsDailySeverityWarning, ""
}

func operationsDailyPickTaskStatus(
	status shippingdomain.PickTaskStatus,
) (reportingdomain.OperationsDailyStatus, reportingdomain.OperationsDailySeverity, string) {
	switch status {
	case shippingdomain.PickTaskStatusCreated, shippingdomain.PickTaskStatusAssigned:
		return reportingdomain.OperationsDailyStatusPending, reportingdomain.OperationsDailySeverityWarning, ""
	case shippingdomain.PickTaskStatusInProgress:
		return reportingdomain.OperationsDailyStatusInProgress, reportingdomain.OperationsDailySeverityNormal, ""
	case shippingdomain.PickTaskStatusCompleted:
		return reportingdomain.OperationsDailyStatusCompleted, reportingdomain.OperationsDailySeverityNormal, ""
	case shippingdomain.PickTaskStatusCancelled:
		return reportingdomain.OperationsDailyStatusBlocked, reportingdomain.OperationsDailySeverityWarning, "PICK_CANCELLED"
	default:
		return reportingdomain.OperationsDailyStatusException,
			reportingdomain.OperationsDailySeverityDanger,
			strings.ToUpper(strings.TrimSpace(string(status)))
	}
}

func operationsDailyCarrierManifestStatus(
	manifest shippingdomain.CarrierManifest,
) (reportingdomain.OperationsDailyStatus, reportingdomain.OperationsDailySeverity, string) {
	summary := manifest.Summary()
	switch manifest.Status {
	case shippingdomain.ManifestStatusCompleted, shippingdomain.ManifestStatusHandedOver:
		return reportingdomain.OperationsDailyStatusCompleted, reportingdomain.OperationsDailySeverityNormal, ""
	case shippingdomain.ManifestStatusScanning:
		if summary.MissingCount > 0 {
			return reportingdomain.OperationsDailyStatusBlocked, reportingdomain.OperationsDailySeverityDanger, "MISSING_HANDOVER_SCAN"
		}
		return reportingdomain.OperationsDailyStatusInProgress, reportingdomain.OperationsDailySeverityNormal, ""
	case shippingdomain.ManifestStatusException:
		return reportingdomain.OperationsDailyStatusException, reportingdomain.OperationsDailySeverityDanger, "HANDOVER_EXCEPTION"
	case shippingdomain.ManifestStatusCancelled:
		return reportingdomain.OperationsDailyStatusBlocked, reportingdomain.OperationsDailySeverityWarning, "HANDOVER_CANCELLED"
	default:
		return reportingdomain.OperationsDailyStatusPending, reportingdomain.OperationsDailySeverityWarning, ""
	}
}

func operationsDailyReturnReceiptStatus(
	status returnsdomain.ReturnReceiptStatus,
) (reportingdomain.OperationsDailyStatus, reportingdomain.OperationsDailySeverity) {
	switch returnsdomain.NormalizeReturnReceiptStatus(status) {
	case returnsdomain.ReturnStatusDispositioned:
		return reportingdomain.OperationsDailyStatusCompleted, reportingdomain.OperationsDailySeverityNormal
	case returnsdomain.ReturnStatusInspected:
		return reportingdomain.OperationsDailyStatusInProgress, reportingdomain.OperationsDailySeverityNormal
	default:
		return reportingdomain.OperationsDailyStatusPending, reportingdomain.OperationsDailySeverityWarning
	}
}

func operationsDailyStockCountStatus(
	session inventorydomain.StockCountSession,
) (reportingdomain.OperationsDailyStatus, reportingdomain.OperationsDailySeverity, string) {
	switch session.Status {
	case inventorydomain.StockCountStatusSubmitted:
		return reportingdomain.OperationsDailyStatusCompleted, reportingdomain.OperationsDailySeverityNormal, ""
	case inventorydomain.StockCountStatusVarianceReview:
		return reportingdomain.OperationsDailyStatusBlocked, reportingdomain.OperationsDailySeverityDanger, "VARIANCE_REVIEW"
	default:
		return reportingdomain.OperationsDailyStatusPending, reportingdomain.OperationsDailySeverityWarning, ""
	}
}

func operationsDailySubcontractStatus(
	status productiondomain.SubcontractOrderStatus,
) reportingdomain.OperationsDailyStatus {
	switch productiondomain.NormalizeSubcontractOrderStatus(status) {
	case productiondomain.SubcontractOrderStatusClosed:
		return reportingdomain.OperationsDailyStatusCompleted
	case productiondomain.SubcontractOrderStatusSampleRejected,
		productiondomain.SubcontractOrderStatusRejectedFactoryIssue:
		return reportingdomain.OperationsDailyStatusException
	default:
		return reportingdomain.OperationsDailyStatusInProgress
	}
}

func operationsDailyReceivingTitle(status inventorydomain.WarehouseReceivingStatus) string {
	switch inventorydomain.NormalizeWarehouseReceivingStatus(status) {
	case inventorydomain.WarehouseReceivingStatusDraft:
		return "Goods receipt draft awaiting submit"
	case inventorydomain.WarehouseReceivingStatusSubmitted:
		return "Goods receipt submitted awaiting receiving check"
	case inventorydomain.WarehouseReceivingStatusInspectReady:
		return "Goods receipt ready for inbound QC"
	case inventorydomain.WarehouseReceivingStatusPosted:
		return "Goods receipt posted"
	default:
		return "Goods receipt awaiting warehouse action"
	}
}

func operationsDailyInboundQCTitle(inspection qcdomain.InboundQCInspection) string {
	if inspection.Status == qcdomain.InboundQCInspectionStatusCompleted {
		return fmt.Sprintf("Inbound QC %s for %s", inspection.Result, firstNonBlankOperationsDaily(inspection.SKU, inspection.ID))
	}

	return fmt.Sprintf("Inbound QC %s for %s", inspection.Status, firstNonBlankOperationsDaily(inspection.SKU, inspection.ID))
}

func operationsDailyPickTaskTitle(status shippingdomain.PickTaskStatus) string {
	switch status {
	case shippingdomain.PickTaskStatusCreated, shippingdomain.PickTaskStatusAssigned:
		return "Pick task awaiting picker"
	case shippingdomain.PickTaskStatusInProgress:
		return "Pick task in progress"
	case shippingdomain.PickTaskStatusCompleted:
		return "Pick task completed"
	default:
		return "Pick task needs warehouse review"
	}
}

func operationsDailyCarrierManifestTitle(manifest shippingdomain.CarrierManifest) string {
	if manifest.Summary().MissingCount > 0 {
		return "Carrier handover missing scan"
	}
	switch manifest.Status {
	case shippingdomain.ManifestStatusCompleted, shippingdomain.ManifestStatusHandedOver:
		return "Carrier handover completed"
	case shippingdomain.ManifestStatusScanning:
		return "Carrier handover scanning"
	default:
		return "Carrier handover awaiting scan"
	}
}

func operationsDailyReturnReceiptTitle(status returnsdomain.ReturnReceiptStatus) string {
	switch returnsdomain.NormalizeReturnReceiptStatus(status) {
	case returnsdomain.ReturnStatusDispositioned:
		return "Return receipt dispositioned"
	case returnsdomain.ReturnStatusInspected:
		return "Return receipt inspected"
	default:
		return "Return receipt awaiting inspection"
	}
}

func operationsDailyStockCountTitle(session inventorydomain.StockCountSession) string {
	if session.Status == inventorydomain.StockCountStatusVarianceReview || session.HasVariance() {
		return "Cycle count variance needs review"
	}
	if session.Status == inventorydomain.StockCountStatusSubmitted {
		return "Cycle count submitted"
	}

	return "Cycle count open"
}

func operationsDailySubcontractTransferTitle(transfer productiondomain.SubcontractMaterialTransfer) string {
	if transfer.Status == productiondomain.SubcontractMaterialTransferStatusPartiallySent {
		return "Subcontract material issue partially sent"
	}

	return "Subcontract material issue sent to factory"
}

func operationsDailyReceivingQuantity(receipt inventorydomain.WarehouseReceiving) string {
	total := decimal.MustQuantity("0")
	for _, line := range receipt.Lines {
		next, err := decimal.AddQuantity(total, line.Quantity)
		if err != nil {
			return ""
		}
		total = next
	}

	return operationsDailyQuantityString(total)
}

func operationsDailyPickTaskQuantity(task shippingdomain.PickTask) string {
	total := decimal.MustQuantity("0")
	for _, line := range task.Lines {
		next, err := decimal.AddQuantity(total, line.QtyToPick)
		if err != nil {
			return ""
		}
		total = next
	}

	return operationsDailyQuantityString(total)
}

func operationsDailySubcontractTransferQuantity(transfer productiondomain.SubcontractMaterialTransfer) string {
	total := decimal.MustQuantity("0")
	for _, line := range transfer.Lines {
		next, err := decimal.AddQuantity(total, line.BaseIssueQty)
		if err != nil {
			return ""
		}
		total = next
	}

	return operationsDailyQuantityString(total)
}

func operationsDailyReturnReceiptQuantity(receipt returnsdomain.ReturnReceipt) string {
	total := 0
	for _, line := range receipt.Lines {
		total += line.Quantity
	}

	return operationsDailyCountQuantity(total)
}

func operationsDailyReceivingUOM(receipt inventorydomain.WarehouseReceiving) string {
	if len(receipt.Lines) == 0 {
		return ""
	}

	return receipt.Lines[0].BaseUOMCode.String()
}

func operationsDailyPickTaskUOM(task shippingdomain.PickTask) string {
	if len(task.Lines) == 0 {
		return ""
	}

	return task.Lines[0].BaseUOMCode.String()
}

func operationsDailySubcontractTransferUOM(transfer productiondomain.SubcontractMaterialTransfer) string {
	if len(transfer.Lines) == 0 {
		return ""
	}

	return transfer.Lines[0].BaseUOMCode.String()
}

func operationsDailyQuantityString(value decimal.Decimal) string {
	if strings.TrimSpace(value.String()) == "" || value.IsZero() {
		return ""
	}

	return value.String()
}

func operationsDailyCountQuantity(count int) string {
	if count <= 0 {
		return ""
	}

	return decimal.MustQuantity(fmt.Sprintf("%d", count)).String()
}

func operationsDailySignalID(sourceType string, sourceID string) string {
	return fmt.Sprintf("ops-%s-%s", strings.ReplaceAll(strings.TrimSpace(sourceType), "_", "-"), strings.TrimSpace(sourceID))
}

func operationsDailyDateFromString(value string, fallback time.Time) time.Time {
	parsed, err := time.ParseInLocation(
		reportingdomain.ReportDateLayout,
		strings.TrimSpace(value),
		reportingdomain.HoChiMinhLocation(),
	)
	if err == nil {
		return parsed
	}

	return fallback
}

func firstNonBlankOperationsDaily(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func firstNonZeroOperationsDailyTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value
		}
	}

	return time.Time{}
}
