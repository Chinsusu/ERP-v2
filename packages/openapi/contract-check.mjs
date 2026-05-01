import { readFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const repoRoot = process.cwd();
const openapiPath = path.join(repoRoot, "packages/openapi/openapi.yaml");
const apiMainPath = path.join(repoRoot, "apps/api/cmd/api/main.go");

const [openapi, apiMain] = await Promise.all([
  readFile(openapiPath, "utf8"),
  readFile(apiMainPath, "utf8")
]);

const sprint4Routes = [
  {
    path: "/purchase-orders",
    apiRoute: "/api/v1/purchase-orders",
    operationIds: ["listPurchaseOrders", "createPurchaseOrder"]
  },
  {
    path: "/purchase-orders/{purchase_order_id}",
    apiRoute: "/api/v1/purchase-orders/{purchase_order_id}",
    operationIds: ["getPurchaseOrder", "updatePurchaseOrder"]
  },
  {
    path: "/purchase-orders/{purchase_order_id}/submit",
    apiRoute: "/api/v1/purchase-orders/{purchase_order_id}/submit",
    operationIds: ["submitPurchaseOrder"]
  },
  {
    path: "/purchase-orders/{purchase_order_id}/approve",
    apiRoute: "/api/v1/purchase-orders/{purchase_order_id}/approve",
    operationIds: ["approvePurchaseOrder"]
  },
  {
    path: "/purchase-orders/{purchase_order_id}/cancel",
    apiRoute: "/api/v1/purchase-orders/{purchase_order_id}/cancel",
    operationIds: ["cancelPurchaseOrder"]
  },
  {
    path: "/purchase-orders/{purchase_order_id}/close",
    apiRoute: "/api/v1/purchase-orders/{purchase_order_id}/close",
    operationIds: ["closePurchaseOrder"]
  },
  {
    path: "/goods-receipts",
    apiRoute: "/api/v1/goods-receipts",
    operationIds: ["listGoodsReceipts", "createGoodsReceipt"]
  },
  {
    path: "/goods-receipts/{receipt_id}",
    apiRoute: "/api/v1/goods-receipts/{receipt_id}",
    operationIds: ["getGoodsReceipt"]
  },
  {
    path: "/goods-receipts/{receipt_id}/submit",
    apiRoute: "/api/v1/goods-receipts/{receipt_id}/submit",
    operationIds: ["submitGoodsReceipt"]
  },
  {
    path: "/goods-receipts/{receipt_id}/inspect-ready",
    apiRoute: "/api/v1/goods-receipts/{receipt_id}/inspect-ready",
    operationIds: ["markGoodsReceiptInspectReady"]
  },
  {
    path: "/goods-receipts/{receipt_id}/post",
    apiRoute: "/api/v1/goods-receipts/{receipt_id}/post",
    operationIds: ["postGoodsReceipt"]
  },
  {
    path: "/inbound-qc-inspections",
    apiRoute: "/api/v1/inbound-qc-inspections",
    operationIds: ["listInboundQCInspections", "createInboundQCInspection"]
  },
  {
    path: "/inbound-qc-inspections/{inspection_id}",
    apiRoute: "/api/v1/inbound-qc-inspections/{inspection_id}",
    operationIds: ["getInboundQCInspection"]
  },
  {
    path: "/inbound-qc-inspections/{inspection_id}/start",
    apiRoute: "/api/v1/inbound-qc-inspections/{inspection_id}/start",
    operationIds: ["startInboundQCInspection"]
  },
  {
    path: "/inbound-qc-inspections/{inspection_id}/pass",
    apiRoute: "/api/v1/inbound-qc-inspections/{inspection_id}/pass",
    operationIds: ["passInboundQCInspection"]
  },
  {
    path: "/inbound-qc-inspections/{inspection_id}/fail",
    apiRoute: "/api/v1/inbound-qc-inspections/{inspection_id}/fail",
    operationIds: ["failInboundQCInspection"]
  },
  {
    path: "/inbound-qc-inspections/{inspection_id}/partial",
    apiRoute: "/api/v1/inbound-qc-inspections/{inspection_id}/partial",
    operationIds: ["partialInboundQCInspection"]
  },
  {
    path: "/inbound-qc-inspections/{inspection_id}/hold",
    apiRoute: "/api/v1/inbound-qc-inspections/{inspection_id}/hold",
    operationIds: ["holdInboundQCInspection"]
  },
  {
    path: "/supplier-rejections",
    apiRoute: "/api/v1/supplier-rejections",
    operationIds: ["listSupplierRejections", "createSupplierRejection"]
  },
  {
    path: "/supplier-rejections/{supplier_rejection_id}",
    apiRoute: "/api/v1/supplier-rejections/{supplier_rejection_id}",
    operationIds: ["getSupplierRejection"]
  },
  {
    path: "/supplier-rejections/{supplier_rejection_id}/submit",
    apiRoute: "/api/v1/supplier-rejections/{supplier_rejection_id}/submit",
    operationIds: ["submitSupplierRejection"]
  },
  {
    path: "/supplier-rejections/{supplier_rejection_id}/confirm",
    apiRoute: "/api/v1/supplier-rejections/{supplier_rejection_id}/confirm",
    operationIds: ["confirmSupplierRejection"]
  },
  {
    path: "/warehouse/daily-board/inbound-metrics",
    apiRoute: "/api/v1/warehouse/daily-board/inbound-metrics",
    operationIds: ["getWarehouseDailyBoardInboundMetrics"]
  }
];

const sprint5Routes = [
  {
    path: "/subcontract-orders",
    apiRoute: "/api/v1/subcontract-orders",
    operationIds: ["listSubcontractOrders", "createSubcontractOrder"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}",
    operationIds: ["getSubcontractOrder", "updateSubcontractOrder"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/submit",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/submit",
    operationIds: ["submitSubcontractOrder"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/approve",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/approve",
    operationIds: ["approveSubcontractOrder"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/confirm-factory",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/confirm-factory",
    operationIds: ["confirmFactorySubcontractOrder"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/record-deposit",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/record-deposit",
    operationIds: ["recordSubcontractDeposit"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/issue-materials",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/issue-materials",
    operationIds: ["issueSubcontractMaterials"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/start-mass-production",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/start-mass-production",
    operationIds: ["startMassProductionSubcontractOrder"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/receive-finished-goods",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/receive-finished-goods",
    operationIds: ["receiveSubcontractFinishedGoods"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/report-factory-defect",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/report-factory-defect",
    operationIds: ["reportSubcontractFactoryDefect"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/accept",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/accept",
    operationIds: ["acceptSubcontractFinishedGoods"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/partial-accept",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/partial-accept",
    operationIds: ["partialAcceptSubcontractFinishedGoods"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/mark-final-payment-ready",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/mark-final-payment-ready",
    operationIds: ["markSubcontractFinalPaymentReady"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/submit-sample",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/submit-sample",
    operationIds: ["submitSubcontractSample"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/approve-sample",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/approve-sample",
    operationIds: ["approveSubcontractSample"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/reject-sample",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/reject-sample",
    operationIds: ["rejectSubcontractSample"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/cancel",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/cancel",
    operationIds: ["cancelSubcontractOrder"]
  },
  {
    path: "/subcontract-orders/{subcontract_order_id}/close",
    apiRoute: "/api/v1/subcontract-orders/{subcontract_order_id}/close",
    operationIds: ["closeSubcontractOrder"]
  },
  {
    path: "/warehouse/daily-board/subcontract-metrics",
    apiRoute: "/api/v1/warehouse/daily-board/subcontract-metrics",
    operationIds: ["getWarehouseDailyBoardSubcontractMetrics"]
  }
];

const sprint6Routes = [
  {
    path: "/customer-receivables",
    apiRoute: "/api/v1/customer-receivables",
    operationIds: ["listCustomerReceivables", "createCustomerReceivable"]
  },
  {
    path: "/customer-receivables/{customer_receivable_id}",
    apiRoute: "/api/v1/customer-receivables/{customer_receivable_id}",
    operationIds: ["getCustomerReceivable"]
  },
  {
    path: "/customer-receivables/{customer_receivable_id}/record-receipt",
    apiRoute: "/api/v1/customer-receivables/{customer_receivable_id}/record-receipt",
    operationIds: ["recordCustomerReceivableReceipt"]
  },
  {
    path: "/customer-receivables/{customer_receivable_id}/mark-disputed",
    apiRoute: "/api/v1/customer-receivables/{customer_receivable_id}/mark-disputed",
    operationIds: ["markCustomerReceivableDisputed"]
  },
  {
    path: "/customer-receivables/{customer_receivable_id}/void",
    apiRoute: "/api/v1/customer-receivables/{customer_receivable_id}/void",
    operationIds: ["voidCustomerReceivable"]
  },
  {
    path: "/supplier-payables",
    apiRoute: "/api/v1/supplier-payables",
    operationIds: ["listSupplierPayables", "createSupplierPayable"]
  },
  {
    path: "/supplier-payables/{supplier_payable_id}",
    apiRoute: "/api/v1/supplier-payables/{supplier_payable_id}",
    operationIds: ["getSupplierPayable"]
  },
  {
    path: "/supplier-payables/{supplier_payable_id}/request-payment",
    apiRoute: "/api/v1/supplier-payables/{supplier_payable_id}/request-payment",
    operationIds: ["requestSupplierPayablePayment"]
  },
  {
    path: "/supplier-payables/{supplier_payable_id}/approve-payment",
    apiRoute: "/api/v1/supplier-payables/{supplier_payable_id}/approve-payment",
    operationIds: ["approveSupplierPayablePayment"]
  },
  {
    path: "/supplier-payables/{supplier_payable_id}/reject-payment",
    apiRoute: "/api/v1/supplier-payables/{supplier_payable_id}/reject-payment",
    operationIds: ["rejectSupplierPayablePayment"]
  },
  {
    path: "/supplier-payables/{supplier_payable_id}/record-payment",
    apiRoute: "/api/v1/supplier-payables/{supplier_payable_id}/record-payment",
    operationIds: ["recordSupplierPayablePayment"]
  },
  {
    path: "/supplier-payables/{supplier_payable_id}/void",
    apiRoute: "/api/v1/supplier-payables/{supplier_payable_id}/void",
    operationIds: ["voidSupplierPayable"]
  },
  {
    path: "/cash-transactions",
    apiRoute: "/api/v1/cash-transactions",
    operationIds: ["listCashTransactions", "createCashTransaction"]
  },
  {
    path: "/cash-transactions/{cash_transaction_id}",
    apiRoute: "/api/v1/cash-transactions/{cash_transaction_id}",
    operationIds: ["getCashTransaction"]
  },
  {
    path: "/finance/dashboard",
    apiRoute: "/api/v1/finance/dashboard",
    operationIds: ["getFinanceDashboard"]
  },
  {
    path: "/cod-remittances",
    apiRoute: "/api/v1/cod-remittances",
    operationIds: ["listCODRemittances", "createCODRemittance"]
  },
  {
    path: "/cod-remittances/{cod_remittance_id}",
    apiRoute: "/api/v1/cod-remittances/{cod_remittance_id}",
    operationIds: ["getCODRemittance"]
  },
  {
    path: "/cod-remittances/{cod_remittance_id}/match",
    apiRoute: "/api/v1/cod-remittances/{cod_remittance_id}/match",
    operationIds: ["matchCODRemittance"]
  },
  {
    path: "/cod-remittances/{cod_remittance_id}/record-discrepancy",
    apiRoute: "/api/v1/cod-remittances/{cod_remittance_id}/record-discrepancy",
    operationIds: ["recordCODRemittanceDiscrepancy"]
  },
  {
    path: "/cod-remittances/{cod_remittance_id}/submit",
    apiRoute: "/api/v1/cod-remittances/{cod_remittance_id}/submit",
    operationIds: ["submitCODRemittance"]
  },
  {
    path: "/cod-remittances/{cod_remittance_id}/approve",
    apiRoute: "/api/v1/cod-remittances/{cod_remittance_id}/approve",
    operationIds: ["approveCODRemittance"]
  },
  {
    path: "/cod-remittances/{cod_remittance_id}/close",
    apiRoute: "/api/v1/cod-remittances/{cod_remittance_id}/close",
    operationIds: ["closeCODRemittance"]
  }
];

const sprint7Routes = [
  {
    path: "/reports/inventory-snapshot",
    apiRoute: "/api/v1/reports/inventory-snapshot",
    operationIds: ["getInventorySnapshotReport"]
  },
  {
    path: "/reports/inventory-snapshot/export.csv",
    apiRoute: "/api/v1/reports/inventory-snapshot/export.csv",
    operationIds: ["exportInventorySnapshotReportCSV"]
  },
  {
    path: "/reports/operations-daily",
    apiRoute: "/api/v1/reports/operations-daily",
    operationIds: ["getOperationsDailyReport"]
  },
  {
    path: "/reports/operations-daily/export.csv",
    apiRoute: "/api/v1/reports/operations-daily/export.csv",
    operationIds: ["exportOperationsDailyReportCSV"]
  },
  {
    path: "/reports/finance-summary",
    apiRoute: "/api/v1/reports/finance-summary",
    operationIds: ["getFinanceSummaryReport"]
  },
  {
    path: "/reports/finance-summary/export.csv",
    apiRoute: "/api/v1/reports/finance-summary/export.csv",
    operationIds: ["exportFinanceSummaryReportCSV"]
  }
];

const requiredSuccessSchemas = [
  "PurchaseOrderListSuccessResponse",
  "PurchaseOrderSuccessResponse",
  "PurchaseOrderActionResultSuccessResponse",
  "GoodsReceiptListSuccessResponse",
  "GoodsReceiptSuccessResponse",
  "InboundQCInspectionListSuccessResponse",
  "InboundQCInspectionSuccessResponse",
  "InboundQCActionResultSuccessResponse",
  "SupplierRejectionListSuccessResponse",
  "SupplierRejectionSuccessResponse",
  "SupplierRejectionActionResultSuccessResponse",
  "WarehouseInboundMetricsSuccessResponse",
  "SubcontractOrderListSuccessResponse",
  "SubcontractOrderSuccessResponse",
  "SubcontractOrderActionResultSuccessResponse",
  "SubcontractPaymentMilestoneResultSuccessResponse",
  "IssueSubcontractMaterialsSuccessResponse",
  "ReceiveSubcontractFinishedGoodsSuccessResponse",
  "ReportSubcontractFactoryDefectSuccessResponse",
  "AcceptSubcontractFinishedGoodsSuccessResponse",
  "PartialAcceptSubcontractFinishedGoodsSuccessResponse",
  "SubcontractSampleApprovalResultSuccessResponse",
  "WarehouseSubcontractMetricsSuccessResponse",
  "CustomerReceivableListSuccessResponse",
  "CustomerReceivableSuccessResponse",
  "CustomerReceivableActionSuccessResponse",
  "SupplierPayableListSuccessResponse",
  "SupplierPayableSuccessResponse",
  "SupplierPayableActionSuccessResponse",
  "CashTransactionListSuccessResponse",
  "CashTransactionSuccessResponse",
  "FinanceDashboardSuccessResponse",
  "CODRemittanceListSuccessResponse",
  "CODRemittanceSuccessResponse",
  "CODRemittanceActionSuccessResponse",
  "InventorySnapshotReportSuccessResponse",
  "OperationsDailyReportSuccessResponse",
  "FinanceSummaryReportSuccessResponse"
];

const failures = [];

const routes = [...sprint4Routes, ...sprint5Routes, ...sprint6Routes, ...sprint7Routes];

for (const route of routes) {
  requireContains(openapi, `  ${route.path}:`, `OpenAPI path missing: ${route.path}`);
  requireContains(apiMain, `"${route.apiRoute}"`, `API route registration missing: ${route.apiRoute}`);
  for (const operationId of route.operationIds) {
    requireContains(openapi, `operationId: ${operationId}`, `OpenAPI operationId missing: ${operationId}`);
  }
}

for (const schemaName of requiredSuccessSchemas) {
  const block = schemaBlock(openapi, schemaName);
  if (!block) {
    failures.push(`OpenAPI success response schema missing: ${schemaName}`);
    continue;
  }
  requireContains(block, "allOf:", `${schemaName} must use the standard success envelope`);
  requireContains(block, '$ref: "#/components/schemas/SuccessResponse"', `${schemaName} must reference SuccessResponse`);
  requireContains(block, "data:", `${schemaName} must expose response data`);
}

const colonActionPattern =
  /^  \/(purchase-orders|goods-receipts|inbound-qc-inspections|supplier-rejections|subcontract-orders|customer-receivables|supplier-payables|cash-transactions|finance\/dashboard|cod-remittances|reports\/(inventory-snapshot|operations-daily|finance-summary)|warehouse\/daily-board\/(inbound|subcontract)-metrics)[^\n]*:[A-Za-z0-9_-]+:/m;
if (colonActionPattern.test(openapi)) {
  failures.push("Tracked OpenAPI paths must use slash action style, not colon action style.");
}

if (failures.length > 0) {
  console.error("OpenAPI route/envelope contract check failed:");
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exit(1);
}

console.log(`OpenAPI route/envelope contract check passed: ${routes.length} routes and ${requiredSuccessSchemas.length} envelopes.`);

function requireContains(haystack, needle, message) {
  if (!haystack.includes(needle)) {
    failures.push(message);
  }
}

function schemaBlock(source, schemaName) {
  const pattern = new RegExp(`^    ${schemaName}:\\n(?<body>(?:      .+\\n|        .+\\n|          .+\\n|            .+\\n|              .+\\n|                .+\\n|\\s*\\n)*)`, "m");
  const match = source.match(pattern);
  return match?.groups?.body ? `${schemaName}:\n${match.groups.body}` : "";
}
