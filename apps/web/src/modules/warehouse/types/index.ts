export type WarehouseDailyTaskStatus = "waiting" | "picking" | "packed" | "handover" | "returns" | "mismatch";

export type WarehouseDailyTaskPriority = "P0" | "P1" | "P2";

export type WarehouseDailyTask = {
  id: string;
  reference: string;
  title: string;
  warehouseId: string;
  warehouseCode: string;
  status: WarehouseDailyTaskStatus;
  priority: WarehouseDailyTaskPriority;
  owner: string;
  dueAt: string;
  href: string;
};

export type WarehouseDailyBoardQuery = {
  warehouseId?: string;
  date?: string;
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

export type WarehouseDailyBoardData = {
  warehouseId: string;
  warehouseCode: string;
  date: string;
  shiftStatus: "open" | "closing" | "closed";
  owner: string;
  summary: WarehouseDailyBoardSummary;
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
  checklist: ReconciliationChecklistItem[];
  lines: ReconciliationLine[];
};

export type EndOfDayReconciliationQuery = {
  warehouseId?: string;
  date?: string;
  status?: EndOfDayReconciliationStatus;
};
