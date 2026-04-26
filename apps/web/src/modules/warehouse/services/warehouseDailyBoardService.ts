import type {
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
  const tasks = query.status ? baseTasks.filter((task) => task.status === query.status) : baseTasks;
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

function countByStatus(tasks: WarehouseDailyTask[], status: WarehouseDailyTaskStatus) {
  return tasks.filter((task) => task.status === status).length;
}

function hasPriorityZero(tasks: WarehouseDailyTask[]) {
  return tasks.some((task) => task.priority === "P0");
}
