package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type inventoryRouteHandlers struct {
	stockMovement          routeHandler
	availableStock         routeHandler
	stockAdjustments       routeHandler
	stockAdjustmentSubmit  routeHandler
	stockAdjustmentApprove routeHandler
	stockAdjustmentReject  routeHandler
	stockAdjustmentPost    routeHandler
	stockTransfers         routeHandler
	stockTransferSubmit    routeHandler
	stockTransferApprove   routeHandler
	stockTransferPost      routeHandler
	warehouseIssues        routeHandler
	warehouseIssueSubmit   routeHandler
	warehouseIssueApprove  routeHandler
	warehouseIssuePost     routeHandler
	stockCounts            routeHandler
	stockCountSubmit       routeHandler
	batches                routeHandler
	batchQCTransitions     routeHandler
	batchDetail            routeHandler
}

func registerInventoryRoutes(routes routeGroup, handlers inventoryRouteHandlers) {
	routes.permission("/api/v1/inventory/stock-movements", auth.PermissionRecordCreate, handlers.stockMovement)
	routes.permission("/api/v1/inventory/available-stock", auth.PermissionInventoryView, handlers.availableStock)
	routes.token("/api/v1/stock-adjustments", handlers.stockAdjustments)
	routes.token("/api/v1/stock-adjustments/{stock_adjustment_id}/submit", handlers.stockAdjustmentSubmit)
	routes.token("/api/v1/stock-adjustments/{stock_adjustment_id}/approve", handlers.stockAdjustmentApprove)
	routes.token("/api/v1/stock-adjustments/{stock_adjustment_id}/reject", handlers.stockAdjustmentReject)
	routes.token("/api/v1/stock-adjustments/{stock_adjustment_id}/post", handlers.stockAdjustmentPost)
	routes.token("/api/v1/stock-transfers", handlers.stockTransfers)
	routes.token("/api/v1/stock-transfers/{stock_transfer_id}/submit", handlers.stockTransferSubmit)
	routes.token("/api/v1/stock-transfers/{stock_transfer_id}/approve", handlers.stockTransferApprove)
	routes.token("/api/v1/stock-transfers/{stock_transfer_id}/post", handlers.stockTransferPost)
	routes.token("/api/v1/warehouse-issues", handlers.warehouseIssues)
	routes.token("/api/v1/warehouse-issues/{warehouse_issue_id}/submit", handlers.warehouseIssueSubmit)
	routes.token("/api/v1/warehouse-issues/{warehouse_issue_id}/approve", handlers.warehouseIssueApprove)
	routes.token("/api/v1/warehouse-issues/{warehouse_issue_id}/post", handlers.warehouseIssuePost)
	routes.token("/api/v1/stock-counts", handlers.stockCounts)
	routes.token("/api/v1/stock-counts/{stock_count_id}/submit", handlers.stockCountSubmit)
	routes.permission("/api/v1/inventory/batches", auth.PermissionInventoryView, handlers.batches)
	routes.token("/api/v1/inventory/batches/{batch_id}/qc-transitions", handlers.batchQCTransitions)
	routes.permission("/api/v1/inventory/batches/{batch_id}", auth.PermissionInventoryView, handlers.batchDetail)
}
