package main

import (
	"context"
	"fmt"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
)

type subcontractSupplierPayableAdapter struct {
	service financeapp.SupplierPayableService
}

func (a subcontractSupplierPayableAdapter) CreateSubcontractPayable(
	ctx context.Context,
	input productionapp.CreateSubcontractPayableInput,
) (productionapp.SubcontractPayableCreationResult, error) {
	order := input.SubcontractOrder
	milestone := input.Milestone
	result, err := a.service.CreateSupplierPayable(ctx, financeapp.CreateSupplierPayableInput{
		ID:           fmt.Sprintf("ap-%s", milestone.ID),
		PayableNo:    fmt.Sprintf("AP-%s", milestone.MilestoneNo),
		SupplierID:   order.FactoryID,
		SupplierCode: order.FactoryCode,
		SupplierName: order.FactoryName,
		Status:       "open",
		SourceDocument: financeapp.SourceDocumentInput{
			Type: "subcontract_payment_milestone",
			ID:   milestone.ID,
			No:   milestone.MilestoneNo,
		},
		Lines: []financeapp.SupplierPayableLineInput{
			{
				ID:          fmt.Sprintf("ap-%s-line-1", milestone.ID),
				Description: fmt.Sprintf("Final subcontract payment for %s", order.OrderNo),
				SourceDocument: financeapp.SourceDocumentInput{
					Type: "subcontract_order",
					ID:   order.ID,
					No:   order.OrderNo,
				},
				Amount: milestone.Amount.String(),
			},
		},
		TotalAmount:  milestone.Amount.String(),
		CurrencyCode: milestone.CurrencyCode.String(),
		ActorID:      input.ActorID,
		RequestID:    input.RequestID,
	})
	if err != nil {
		return productionapp.SubcontractPayableCreationResult{}, err
	}

	return productionapp.SubcontractPayableCreationResult{
		PayableID:  result.SupplierPayable.ID,
		PayableNo:  result.SupplierPayable.PayableNo,
		AuditLogID: result.AuditLogID,
	}, nil
}
