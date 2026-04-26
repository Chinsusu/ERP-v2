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
