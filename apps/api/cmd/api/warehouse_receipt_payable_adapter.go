package main

import (
	"context"
	"fmt"
	"time"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type warehouseReceiptSupplierPayableAdapter struct {
	service financeapp.SupplierPayableService
}

func (a warehouseReceiptSupplierPayableAdapter) CreateWarehouseReceiptPayable(
	ctx context.Context,
	input inventoryapp.CreateWarehouseReceiptPayableInput,
) (inventoryapp.WarehouseReceiptPayableCreationResult, error) {
	receipt := input.Receipt
	order := input.PurchaseOrder
	lines := make([]financeapp.SupplierPayableLineInput, 0, len(input.Lines))
	for _, line := range input.Lines {
		lines = append(lines, financeapp.SupplierPayableLineInput{
			ID:          fmt.Sprintf("ap-%s-%s", receipt.ID, line.ReceiptLineID),
			Description: fmt.Sprintf("Accepted %s from %s", line.SKU, receipt.ReceiptNo),
			SourceDocument: financeapp.SourceDocumentInput{
				Type: "purchase_order",
				ID:   order.ID,
				No:   order.PONo,
			},
			Amount: line.Amount.String(),
		})
	}
	result, err := a.service.CreateSupplierPayable(ctx, financeapp.CreateSupplierPayableInput{
		ID:           fmt.Sprintf("ap-%s", receipt.ID),
		PayableNo:    fmt.Sprintf("AP-%s", receipt.ReceiptNo),
		SupplierID:   order.SupplierID,
		SupplierCode: order.SupplierCode,
		SupplierName: order.SupplierName,
		Status:       "open",
		SourceDocument: financeapp.SourceDocumentInput{
			Type: "warehouse_receipt",
			ID:   receipt.ID,
			No:   receipt.ReceiptNo,
		},
		Lines:        lines,
		TotalAmount:  input.TotalAmount.String(),
		CurrencyCode: decimal.CurrencyVND.String(),
		DueDate:      warehouseReceiptPayableDueDate(order.ExpectedDate),
		ActorID:      input.ActorID,
		RequestID:    input.RequestID,
	})
	if err != nil {
		return inventoryapp.WarehouseReceiptPayableCreationResult{}, err
	}

	return inventoryapp.WarehouseReceiptPayableCreationResult{
		PayableID:  result.SupplierPayable.ID,
		PayableNo:  result.SupplierPayable.PayableNo,
		AuditLogID: result.AuditLogID,
	}, nil
}

func warehouseReceiptPayableDueDate(expectedDate string) string {
	parsed, err := time.Parse(time.DateOnly, expectedDate)
	if err != nil {
		return expectedDate
	}

	return parsed.AddDate(0, 0, 7).Format(time.DateOnly)
}
