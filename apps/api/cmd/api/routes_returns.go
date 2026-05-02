package main

type returnsRouteHandlers struct {
	returnMasterData  routeHandler
	returnScan        routeHandler
	returnReceipts    routeHandler
	returnInspection  routeHandler
	returnDisposition routeHandler
	returnAttachment  routeHandler
}

func registerReturnsRoutes(routes routeGroup, handlers returnsRouteHandlers) {
	routes.token("/api/v1/return-reasons", handlers.returnMasterData)
	routes.token("/api/v1/returns/scan", handlers.returnScan)
	routes.token("/api/v1/returns/receipts", handlers.returnReceipts)
	routes.token("/api/v1/returns/{return_receipt_id}/inspect", handlers.returnInspection)
	routes.token("/api/v1/returns/{return_receipt_id}/disposition", handlers.returnDisposition)
	routes.token("/api/v1/returns/{return_receipt_id}/attachments", handlers.returnAttachment)
}
