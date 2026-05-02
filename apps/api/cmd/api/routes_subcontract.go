package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type subcontractRouteHandlers struct {
	subcontractOrders                 routeHandler
	subcontractOrderDetail            routeHandler
	subcontractOrderSubmit            routeHandler
	subcontractOrderApprove           routeHandler
	subcontractOrderConfirmFactory    routeHandler
	subcontractOrderRecordDeposit     routeHandler
	subcontractOrderIssueMaterials    routeHandler
	subcontractOrderStartProduction   routeHandler
	subcontractOrderReceiveGoods      routeHandler
	subcontractOrderReportDefect      routeHandler
	subcontractOrderAccept            routeHandler
	subcontractOrderPartialAccept     routeHandler
	subcontractOrderFinalPaymentReady routeHandler
	subcontractOrderSubmitSample      routeHandler
	subcontractOrderApproveSample     routeHandler
	subcontractOrderRejectSample      routeHandler
	subcontractOrderCancel            routeHandler
	subcontractOrderClose             routeHandler
}

func registerSubcontractRoutes(routes routeGroup, handlers subcontractRouteHandlers) {
	routes.permission("/api/v1/subcontract-orders", auth.PermissionSubcontractView, handlers.subcontractOrders)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}", auth.PermissionSubcontractView, handlers.subcontractOrderDetail)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/submit", auth.PermissionRecordCreate, handlers.subcontractOrderSubmit)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/approve", auth.PermissionRecordCreate, handlers.subcontractOrderApprove)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/confirm-factory", auth.PermissionRecordCreate, handlers.subcontractOrderConfirmFactory)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/record-deposit", auth.PermissionRecordCreate, handlers.subcontractOrderRecordDeposit)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/issue-materials", auth.PermissionRecordCreate, handlers.subcontractOrderIssueMaterials)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/start-mass-production", auth.PermissionRecordCreate, handlers.subcontractOrderStartProduction)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/receive-finished-goods", auth.PermissionRecordCreate, handlers.subcontractOrderReceiveGoods)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/report-factory-defect", auth.PermissionRecordCreate, handlers.subcontractOrderReportDefect)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/accept", auth.PermissionRecordCreate, handlers.subcontractOrderAccept)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/partial-accept", auth.PermissionRecordCreate, handlers.subcontractOrderPartialAccept)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/mark-final-payment-ready", auth.PermissionRecordCreate, handlers.subcontractOrderFinalPaymentReady)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/submit-sample", auth.PermissionRecordCreate, handlers.subcontractOrderSubmitSample)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/approve-sample", auth.PermissionRecordCreate, handlers.subcontractOrderApproveSample)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/reject-sample", auth.PermissionRecordCreate, handlers.subcontractOrderRejectSample)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/cancel", auth.PermissionRecordCreate, handlers.subcontractOrderCancel)
	routes.permission("/api/v1/subcontract-orders/{subcontract_order_id}/close", auth.PermissionRecordCreate, handlers.subcontractOrderClose)
}
