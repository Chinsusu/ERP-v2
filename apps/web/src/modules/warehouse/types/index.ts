export type WarehouseDailyTaskStatus = "waiting" | "picking" | "packed" | "handover" | "returns" | "mismatch";

export type WarehouseDailyTaskPriority = "P0" | "P1" | "P2";

export type WarehouseDailyShiftCode = "day" | "night";

export type WarehouseDailyTaskSource =
  | "order_queue"
  | "receiving"
  | "shipping"
  | "returns"
  | "stock_movement"
  | "reconciliation";

export type WarehouseDailyTask = {
  id: string;
  reference: string;
  title: string;
  warehouseId: string;
  warehouseCode: string;
  carrierCode?: string;
  shiftCode: WarehouseDailyShiftCode;
  status: WarehouseDailyTaskStatus;
  priority: WarehouseDailyTaskPriority;
  owner: string;
  dueAt: string;
  href: string;
  source: WarehouseDailyTaskSource;
  sourceField: string;
};

export type WarehouseDailyBoardQuery = {
  warehouseId?: string;
  date?: string;
  shiftCode?: WarehouseDailyShiftCode;
  carrierCode?: string;
  status?: WarehouseDailyTaskStatus;
};

export type WarehouseDailyBoardSummary = {
  waiting: number;
  picking: number;
  packed: number;
  handover: number;
  returns: number;
  reconciliationMismatch: number;
  overdue: number;
};

export type WarehouseFulfillmentMetrics = {
  warehouseId?: string;
  date?: string;
  shiftCode?: string;
  carrierCode?: string;
  totalOrders: number;
  newOrders: number;
  reservedOrders: number;
  pickingOrders: number;
  packedOrders: number;
  waitingHandoverOrders: number;
  missingOrders: number;
  handoverOrders: number;
  generatedAt: string;
};

export type WarehouseDailyBoardCounterSource = {
  counter: keyof WarehouseDailyBoardSummary;
  label: string;
  fields: string[];
};

export type WarehouseDailyBoardData = {
  warehouseId: string;
  warehouseCode: string;
  date: string;
  shiftCode: WarehouseDailyShiftCode;
  shiftStatus: "open" | "closing" | "closed";
  owner: string;
  summary: WarehouseDailyBoardSummary;
  fulfillment: WarehouseFulfillmentMetrics;
  sourceFields: WarehouseDailyBoardCounterSource[];
  tasks: WarehouseDailyTask[];
};

export type EndOfDayReconciliationStatus = "open" | "in_review" | "closed";

export type ReconciliationChecklistItem = {
  key: string;
  label: string;
  complete: boolean;
  blocking: boolean;
  note?: string;
};

export type ReconciliationLine = {
  id: string;
  sku: string;
  batchNo: string;
  binCode: string;
  systemQuantity: number;
  countedQuantity: number;
  varianceQuantity: number;
  reason?: string;
  owner: string;
};

export type EndOfDayReconciliationSummary = {
  systemQuantity: number;
  countedQuantity: number;
  varianceQuantity: number;
  varianceCount: number;
  checklistTotal: number;
  checklistCompleted: number;
  readyToClose: boolean;
};

export type EndOfDayReconciliationOperations = {
  orderCount: number;
  handoverOrderCount: number;
  returnOrderCount: number;
  stockMovementCount: number;
  stockCountSessionCount: number;
  pendingIssueCount: number;
};

export type EndOfDayReconciliation = {
  id: string;
  warehouseId: string;
  warehouseCode: string;
  date: string;
  shiftCode: string;
  status: EndOfDayReconciliationStatus;
  owner: string;
  closedAt?: string;
  closedBy?: string;
  auditLogId?: string;
  summary: EndOfDayReconciliationSummary;
  operations: EndOfDayReconciliationOperations;
  checklist: ReconciliationChecklistItem[];
  lines: ReconciliationLine[];
};

export type EndOfDayReconciliationQuery = {
  warehouseId?: string;
  date?: string;
  shiftCode?: WarehouseDailyShiftCode;
  status?: EndOfDayReconciliationStatus;
};
