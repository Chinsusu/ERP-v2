import type {
  EndOfDayReconciliation,
  EndOfDayReconciliationQuery,
  EndOfDayReconciliationStatus,
  EndOfDayReconciliationSummary,
  ReconciliationLine,
  WarehouseDailyBoardData,
  WarehouseDailyBoardQuery,
  WarehouseDailyBoardSummary,
  WarehouseDailyTask,
  WarehouseDailyTaskStatus
} from "../types";

export const defaultWarehouseDailyBoardDate = "2026-04-26";

export const warehouseOptions = [
  { label: "All warehouses", value: "" },
  { label: "HCM", value: "wh-hcm" },
  { label: "HN", value: "wh-hn" }
] as const;

export const statusOptions: { label: string; value: "" | WarehouseDailyTaskStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Waiting", value: "waiting" },
  { label: "Picking", value: "picking" },
  { label: "Packed", value: "packed" },
  { label: "Handover", value: "handover" },
  { label: "Returns", value: "returns" },
  { label: "Mismatch", value: "mismatch" }
];

export const prototypeWarehouseDailyTasks: WarehouseDailyTask[] = [
  {
    id: "task-pick-260426-001",
    reference: "SO-260426-001",
    title: "Pick B2B serum order",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    status: "waiting",
    priority: "P1",
    owner: "Picking",
    dueAt: "2026-04-26T10:30:00Z",
    href: "/sales"
  },
  {
    id: "task-pick-260426-002",
    reference: "SO-260426-002",
    title: "Pick marketplace mixed order",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    status: "picking",
    priority: "P1",
    owner: "Picking",
    dueAt: "2026-04-26T11:00:00Z",
    href: "/sales"
  },
  {
    id: "task-pack-260426-003",
    reference: "PACK-260426-017",
    title: "Pack fragile glass serum set",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    status: "packed",
    priority: "P2",
    owner: "Packing",
    dueAt: "2026-04-26T12:15:00Z",
    href: "/shipping"
  },
  {
    id: "task-handover-260426-004",
    reference: "MAN-260426-GHN-01",
    title: "Scan GHN morning handover",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    status: "handover",
    priority: "P1",
    owner: "Handover",
    dueAt: "2026-04-26T13:00:00Z",
    href: "/shipping"
  },
  {
    id: "task-return-260426-005",
    reference: "RET-260426-009",
    title: "Inspect returned toner lot",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    status: "returns",
    priority: "P1",
    owner: "Returns",
    dueAt: "2026-04-26T15:00:00Z",
    href: "/returns"
  },
  {
    id: "task-mismatch-260426-006",
    reference: "VAR-260426-003",
    title: "Resolve LOT-2604A stock variance",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    status: "mismatch",
    priority: "P0",
    owner: "Lead",
    dueAt: "2026-04-26T09:15:00Z",
    href: "/inventory"
  },
  {
    id: "task-hn-pick-260426-007",
    reference: "SO-260426-HN-011",
    title: "Pick Hanoi dealer replenishment",
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    status: "waiting",
    priority: "P2",
    owner: "Picking",
    dueAt: "2026-04-26T10:45:00Z",
    href: "/sales"
  },
  {
    id: "task-hn-mismatch-260426-008",
    reference: "VAR-260426-HN-001",
    title: "Check bin balance before closing",
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    status: "mismatch",
    priority: "P0",
    owner: "Lead",
    dueAt: "2026-04-26T09:30:00Z",
    href: "/inventory"
  }
];

export const reconciliationStatusOptions: { label: string; value: "" | EndOfDayReconciliationStatus }[] = [
  { label: "All closing statuses", value: "" },
  { label: "Open", value: "open" },
  { label: "In Review", value: "in_review" },
  { label: "Closed", value: "closed" }
];

export const prototypeEndOfDayReconciliations: EndOfDayReconciliation[] = [
  createReconciliation({
    id: "rec-hcm-260426-day",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    date: "2026-04-26",
    shiftCode: "day",
    status: "in_review",
    owner: "Warehouse Lead",
    checklist: [
      { key: "shipments", label: "Shipments reconciled", complete: true, blocking: true },
      { key: "inbound", label: "Inbound and QC checked", complete: true, blocking: true },
      { key: "returns", label: "Returns triaged", complete: true, blocking: true },
      {
        key: "variance",
        label: "Stock variance reviewed",
        complete: false,
        blocking: true,
        note: "LOT-2604A short by 2"
      },
      {
        key: "pending_tasks",
        label: "P0 tasks cleared or noted",
        complete: false,
        blocking: true,
        note: "One mismatch pending lead sign-off"
      }
    ],
    lines: [
      {
        id: "line-hcm-serum-2604a",
        sku: "SERUM-30ML",
        batchNo: "LOT-2604A",
        binCode: "A-01",
        systemQuantity: 120,
        countedQuantity: 118,
        varianceQuantity: -2,
        reason: "cycle count variance",
        owner: "Warehouse Lead"
      },
      {
        id: "line-hcm-cream-2603b",
        sku: "CREAM-50G",
        batchNo: "LOT-2603B",
        binCode: "A-02",
        systemQuantity: 44,
        countedQuantity: 44,
        varianceQuantity: 0,
        owner: "Inventory"
      }
    ]
  }),
  createReconciliation({
    id: "rec-hn-260426-day",
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    date: "2026-04-26",
    shiftCode: "day",
    status: "open",
    owner: "HN Lead",
    checklist: [
      { key: "shipments", label: "Shipments reconciled", complete: true, blocking: true },
      { key: "returns", label: "Returns triaged", complete: true, blocking: true },
      { key: "variance", label: "Stock variance reviewed", complete: true, blocking: true },
      { key: "pending_tasks", label: "P0 tasks cleared or noted", complete: true, blocking: true }
    ],
    lines: [
      {
        id: "line-hn-toner-2604c",
        sku: "TONER-100ML",
        batchNo: "LOT-2604C",
        binCode: "HN-B-04",
        systemQuantity: 85,
        countedQuantity: 85,
        varianceQuantity: 0,
        owner: "HN Lead"
      }
    ]
  }),
  createReconciliation({
    id: "rec-hcm-260425-day",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    date: "2026-04-25",
    shiftCode: "day",
    status: "closed",
    owner: "Warehouse Lead",
    closedAt: "2026-04-25T17:42:00Z",
    closedBy: "user-warehouse-lead",
    checklist: [
      { key: "shipments", label: "Shipments reconciled", complete: true, blocking: true },
      { key: "returns", label: "Returns triaged", complete: true, blocking: true },
      { key: "variance", label: "Stock variance reviewed", complete: true, blocking: true },
      { key: "pending_tasks", label: "P0 tasks cleared or noted", complete: true, blocking: true }
    ],
    lines: [
      {
        id: "line-hcm-serum-2603z",
        sku: "SERUM-30ML",
        batchNo: "LOT-2603Z",
        binCode: "A-01",
        systemQuantity: 96,
        countedQuantity: 96,
        varianceQuantity: 0,
        owner: "Warehouse Lead"
      }
    ]
  })
];

export async function getWarehouseDailyBoard(query: WarehouseDailyBoardQuery = {}): Promise<WarehouseDailyBoardData> {
  const date = query.date ?? defaultWarehouseDailyBoardDate;
  const warehouseId = query.warehouseId ?? "";
  const baseTasks = prototypeWarehouseDailyTasks.filter((task) => {
    const taskDate = task.dueAt.slice(0, 10);
    if (taskDate !== date) {
      return false;
    }
    if (warehouseId && task.warehouseId !== warehouseId) {
      return false;
    }

    return true;
  });
  const filteredTasks = query.status ? baseTasks.filter((task) => task.status === query.status) : baseTasks;
  const tasks = sortWarehouseTasksByRisk(filteredTasks);
  const warehouseCode = warehouseId ? baseTasks[0]?.warehouseCode ?? warehouseId : "All";

  return {
    warehouseId: warehouseId || "all",
    warehouseCode,
    date,
    shiftStatus: hasPriorityZero(baseTasks) ? "open" : "closing",
    owner: warehouseId === "wh-hn" ? "HN Lead" : "Warehouse Lead",
    summary: summarizeWarehouseDailyBoard(baseTasks),
    tasks
  };
}

export async function getEndOfDayReconciliations(
  query: EndOfDayReconciliationQuery = {}
): Promise<EndOfDayReconciliation[]> {
  const date = query.date ?? defaultWarehouseDailyBoardDate;
  const warehouseId = query.warehouseId ?? "";

  return prototypeEndOfDayReconciliations.filter((reconciliation) => {
    if (date && reconciliation.date !== date) {
      return false;
    }
    if (warehouseId && reconciliation.warehouseId !== warehouseId) {
      return false;
    }
    if (query.status && reconciliation.status !== query.status) {
      return false;
    }

    return true;
  });
}

export async function closeEndOfDayReconciliation(
  reconciliationId: string,
  exceptionNote: string
): Promise<EndOfDayReconciliation> {
  const reconciliation = prototypeEndOfDayReconciliations.find((candidate) => candidate.id === reconciliationId);
  if (!reconciliation) {
    throw new Error("End-of-day reconciliation not found");
  }
  if (reconciliation.status === "closed") {
    throw new Error("End-of-day reconciliation is already closed");
  }
  if (!canCloseReconciliation(reconciliation, exceptionNote)) {
    throw new Error("Exception note is required before closing this shift");
  }

  return {
    ...reconciliation,
    status: "closed",
    closedAt: "2026-04-26T17:45:00Z",
    closedBy: "user-warehouse-lead",
    auditLogId: `audit-close-${reconciliation.id}`
  };
}

export function summarizeWarehouseDailyBoard(tasks: WarehouseDailyTask[]): WarehouseDailyBoardSummary {
  return {
    waiting: countByStatus(tasks, "waiting"),
    picking: countByStatus(tasks, "picking"),
    packed: countByStatus(tasks, "packed"),
    handover: countByStatus(tasks, "handover"),
    returns: countByStatus(tasks, "returns"),
    reconciliationMismatch: countByStatus(tasks, "mismatch"),
    overdue: tasks.filter((task) => task.priority === "P0").length
  };
}

export function sortWarehouseTasksByRisk(tasks: WarehouseDailyTask[]): WarehouseDailyTask[] {
  return [...tasks].sort((left, right) => {
    const riskDelta = taskRiskScore(left) - taskRiskScore(right);
    if (riskDelta !== 0) {
      return riskDelta;
    }

    return new Date(left.dueAt).getTime() - new Date(right.dueAt).getTime();
  });
}

export function summarizeReconciliationLines(lines: ReconciliationLine[]): EndOfDayReconciliationSummary {
  const systemQuantity = lines.reduce((total, line) => total + line.systemQuantity, 0);
  const countedQuantity = lines.reduce((total, line) => total + line.countedQuantity, 0);
  const varianceQuantity = lines.reduce((total, line) => total + line.varianceQuantity, 0);
  const varianceCount = lines.filter((line) => line.varianceQuantity !== 0).length;

  return {
    systemQuantity,
    countedQuantity,
    varianceQuantity,
    varianceCount,
    checklistTotal: 0,
    checklistCompleted: 0,
    readyToClose: varianceCount === 0
  };
}

export function warehouseTaskTone(status: WarehouseDailyTaskStatus): "normal" | "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "packed":
      return "success";
    case "handover":
      return "info";
    case "returns":
    case "picking":
      return "warning";
    case "mismatch":
      return "danger";
    case "waiting":
    default:
      return "normal";
  }
}

export function reconciliationStatusTone(
  status: EndOfDayReconciliationStatus
): "normal" | "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "closed":
      return "success";
    case "in_review":
      return "warning";
    case "open":
    default:
      return "info";
  }
}

function countByStatus(tasks: WarehouseDailyTask[], status: WarehouseDailyTaskStatus) {
  return tasks.filter((task) => task.status === status).length;
}

function hasPriorityZero(tasks: WarehouseDailyTask[]) {
  return tasks.some((task) => task.priority === "P0");
}

function taskRiskScore(task: WarehouseDailyTask) {
  if (task.status === "mismatch") {
    return 0;
  }
  if (task.priority === "P0") {
    return 1;
  }
  if (task.priority === "P1") {
    return 2;
  }

  return 3;
}

function canCloseReconciliation(reconciliation: EndOfDayReconciliation, exceptionNote: string) {
  const openBlockingItems = reconciliation.checklist.filter((item) => item.blocking && !item.complete);
  if (openBlockingItems.length === 0) {
    return true;
  }

  return exceptionNote.trim() !== "";
}

function createReconciliation(
  input: Omit<EndOfDayReconciliation, "summary">
): EndOfDayReconciliation {
  const lineSummary = summarizeReconciliationLines(input.lines);
  const checklistCompleted = input.checklist.filter((item) => item.complete).length;
  const openBlockingItems = input.checklist.filter((item) => item.blocking && !item.complete);

  return {
    ...input,
    summary: {
      ...lineSummary,
      checklistTotal: input.checklist.length,
      checklistCompleted,
      readyToClose: input.status !== "closed" && openBlockingItems.length === 0
    }
  };
}
