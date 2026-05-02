package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type purchaseRouteHandlers struct {
	purchaseOrders       routeHandler
	purchaseOrderDetail  routeHandler
	purchaseOrderSubmit  routeHandler
	purchaseOrderApprove routeHandler
	purchaseOrderCancel  routeHandler
	purchaseOrderClose   routeHandler
}

func registerPurchaseRoutes(routes routeGroup, handlers purchaseRouteHandlers) {
	routes.permission("/api/v1/purchase-orders", auth.PermissionPurchaseView, handlers.purchaseOrders)
	routes.permission("/api/v1/purchase-orders/{purchase_order_id}", auth.PermissionPurchaseView, handlers.purchaseOrderDetail)
	routes.permission("/api/v1/purchase-orders/{purchase_order_id}/submit", auth.PermissionRecordCreate, handlers.purchaseOrderSubmit)
	routes.permission("/api/v1/purchase-orders/{purchase_order_id}/approve", auth.PermissionRecordCreate, handlers.purchaseOrderApprove)
	routes.permission("/api/v1/purchase-orders/{purchase_order_id}/cancel", auth.PermissionRecordCreate, handlers.purchaseOrderCancel)
	routes.permission("/api/v1/purchase-orders/{purchase_order_id}/close", auth.PermissionRecordCreate, handlers.purchaseOrderClose)
}
