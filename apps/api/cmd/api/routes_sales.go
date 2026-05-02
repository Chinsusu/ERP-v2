package main

type salesRouteHandlers struct {
	salesOrders       routeHandler
	salesOrderDetail  routeHandler
	salesOrderConfirm routeHandler
	salesOrderCancel  routeHandler
}

func registerSalesRoutes(routes routeGroup, handlers salesRouteHandlers) {
	routes.token("/api/v1/sales-orders", handlers.salesOrders)
	routes.token("/api/v1/sales-orders/{sales_order_id}", handlers.salesOrderDetail)
	routes.token("/api/v1/sales-orders/{sales_order_id}/confirm", handlers.salesOrderConfirm)
	routes.token("/api/v1/sales-orders/{sales_order_id}/cancel", handlers.salesOrderCancel)
}
