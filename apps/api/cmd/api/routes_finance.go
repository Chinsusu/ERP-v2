package main

type financeRouteHandlers struct {
	customerReceivables             routeHandler
	customerReceivableDetail        routeHandler
	customerReceivableRecordReceipt routeHandler
	customerReceivableMarkDisputed  routeHandler
	customerReceivableVoid          routeHandler
	supplierPayables                routeHandler
	supplierPayableDetail           routeHandler
	supplierPayableRequestPayment   routeHandler
	supplierPayableApprovePayment   routeHandler
	supplierPayableRejectPayment    routeHandler
	supplierPayableRecordPayment    routeHandler
	supplierPayableVoid             routeHandler
	cashTransactions                routeHandler
	cashTransactionDetail           routeHandler
	financeDashboard                routeHandler
	codRemittances                  routeHandler
	codRemittanceDetail             routeHandler
	codRemittanceMatch              routeHandler
	codRemittanceDiscrepancy        routeHandler
	codRemittanceSubmit             routeHandler
	codRemittanceApprove            routeHandler
	codRemittanceClose              routeHandler
}

func registerFinanceRoutes(routes routeGroup, handlers financeRouteHandlers) {
	routes.token("/api/v1/customer-receivables", handlers.customerReceivables)
	routes.token("/api/v1/customer-receivables/{customer_receivable_id}", handlers.customerReceivableDetail)
	routes.token("/api/v1/customer-receivables/{customer_receivable_id}/record-receipt", handlers.customerReceivableRecordReceipt)
	routes.token("/api/v1/customer-receivables/{customer_receivable_id}/mark-disputed", handlers.customerReceivableMarkDisputed)
	routes.token("/api/v1/customer-receivables/{customer_receivable_id}/void", handlers.customerReceivableVoid)
	routes.token("/api/v1/supplier-payables", handlers.supplierPayables)
	routes.token("/api/v1/supplier-payables/{supplier_payable_id}", handlers.supplierPayableDetail)
	routes.token("/api/v1/supplier-payables/{supplier_payable_id}/request-payment", handlers.supplierPayableRequestPayment)
	routes.token("/api/v1/supplier-payables/{supplier_payable_id}/approve-payment", handlers.supplierPayableApprovePayment)
	routes.token("/api/v1/supplier-payables/{supplier_payable_id}/reject-payment", handlers.supplierPayableRejectPayment)
	routes.token("/api/v1/supplier-payables/{supplier_payable_id}/record-payment", handlers.supplierPayableRecordPayment)
	routes.token("/api/v1/supplier-payables/{supplier_payable_id}/void", handlers.supplierPayableVoid)
	routes.token("/api/v1/cash-transactions", handlers.cashTransactions)
	routes.token("/api/v1/cash-transactions/{cash_transaction_id}", handlers.cashTransactionDetail)
	routes.token("/api/v1/finance/dashboard", handlers.financeDashboard)
	routes.token("/api/v1/cod-remittances", handlers.codRemittances)
	routes.token("/api/v1/cod-remittances/{cod_remittance_id}", handlers.codRemittanceDetail)
	routes.token("/api/v1/cod-remittances/{cod_remittance_id}/match", handlers.codRemittanceMatch)
	routes.token("/api/v1/cod-remittances/{cod_remittance_id}/record-discrepancy", handlers.codRemittanceDiscrepancy)
	routes.token("/api/v1/cod-remittances/{cod_remittance_id}/submit", handlers.codRemittanceSubmit)
	routes.token("/api/v1/cod-remittances/{cod_remittance_id}/approve", handlers.codRemittanceApprove)
	routes.token("/api/v1/cod-remittances/{cod_remittance_id}/close", handlers.codRemittanceClose)
}
