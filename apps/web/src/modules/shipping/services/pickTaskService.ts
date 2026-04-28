import { ApiError, apiGetRaw, apiPost } from "../../../shared/api/client";
import type { PickTask, PickTaskExceptionCode, PickTaskLine, PickTaskQuery, PickTaskStatus } from "../types";

const defaultAccessToken = "local-dev-access-token";

type PickTaskApi = {
  id: string;
  org_id: string;
  pick_task_no: string;
  sales_order_id: string;
  order_no: string;
  warehouse_id: string;
  warehouse_code: string;
  status: PickTaskStatus;
  assigned_to?: string;
  assigned_at?: string;
  started_at?: string;
  started_by?: string;
  completed_at?: string;
  completed_by?: string;
  audit_log_id?: string;
  lines: PickTaskLineApi[];
  created_at: string;
  updated_at: string;
};

type PickTaskLineApi = {
  id: string;
  line_no: number;
  sales_order_line_id: string;
  stock_reservation_id: string;
  item_id: string;
  sku_code: string;
  batch_id: string;
  batch_no: string;
  warehouse_id: string;
  bin_id: string;
  bin_code: string;
  qty_to_pick: string;
  qty_picked: string;
  base_uom_code: string;
  status: PickTaskLine["status"];
  picked_at?: string;
  picked_by?: string;
  created_at: string;
  updated_at: string;
};

export const pickTaskWarehouseOptions = [
  { label: "All warehouses", value: "" },
  { label: "Finished Goods HCM", value: "wh-hcm-fg", code: "WH-HCM-FG" }
] as const;

export const pickTaskStatusOptions: { label: string; value: "" | PickTaskStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Created", value: "created" },
  { label: "Assigned", value: "assigned" },
  { label: "In progress", value: "in_progress" },
  { label: "Completed", value: "completed" },
  { label: "Missing stock", value: "missing_stock" },
  { label: "Wrong SKU", value: "wrong_sku" },
  { label: "Wrong batch", value: "wrong_batch" },
  { label: "Wrong location", value: "wrong_location" },
  { label: "Cancelled", value: "cancelled" }
];

export const pickTaskExceptionOptions: { label: string; value: PickTaskExceptionCode }[] = [
  { label: "Missing stock", value: "missing_stock" },
  { label: "Wrong SKU", value: "wrong_sku" },
  { label: "Wrong batch", value: "wrong_batch" },
  { label: "Wrong location", value: "wrong_location" },
  { label: "Cancel task", value: "cancelled" }
];

let prototypePickTasks = createPrototypePickTasks();

export async function getPickTasks(query: PickTaskQuery = {}): Promise<PickTask[]> {
  try {
    const tasks = await apiGetRaw<PickTaskApi[]>(`/pick-tasks${queryString(toApiQuery(query))}`, {
      accessToken: defaultAccessToken
    });

    return tasks.map(fromApiPickTask);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypePickTasks(query);
  }
}

export async function startPickTask(id: string): Promise<PickTask> {
  try {
    return fromApiPickTask(
      await apiPost<PickTaskApi, Record<string, never>>(
        `/pick-tasks/${encodeURIComponent(id)}/start`,
        {},
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypePickTask(id, "start");
  }
}

export async function confirmPickTaskLine(id: string, lineId: string, pickedQty: string): Promise<PickTask> {
  try {
    return fromApiPickTask(
      await apiPost<PickTaskApi, { line_id: string; picked_qty: string }>(
        `/pick-tasks/${encodeURIComponent(id)}/confirm-line`,
        { line_id: lineId, picked_qty: pickedQty },
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return confirmPrototypePickTaskLine(id, lineId, pickedQty);
  }
}

export async function completePickTask(id: string): Promise<PickTask> {
  try {
    return fromApiPickTask(
      await apiPost<PickTaskApi, Record<string, never>>(
        `/pick-tasks/${encodeURIComponent(id)}/complete`,
        {},
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypePickTask(id, "complete");
  }
}

export async function reportPickTaskException(
  id: string,
  exceptionCode: PickTaskExceptionCode,
  investigation = ""
): Promise<PickTask> {
  try {
    return fromApiPickTask(
      await apiPost<PickTaskApi, { exception_code: PickTaskExceptionCode; investigation: string }>(
        `/pick-tasks/${encodeURIComponent(id)}/exception`,
        { exception_code: exceptionCode, investigation },
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypePickTask(id, exceptionCode);
  }
}

export function pickTaskStatusTone(status: PickTaskStatus): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "completed":
      return "success";
    case "assigned":
    case "in_progress":
      return "info";
    case "created":
      return "warning";
    case "missing_stock":
    case "wrong_sku":
    case "wrong_batch":
    case "wrong_location":
    case "cancelled":
      return "danger";
    default:
      return "normal";
  }
}

export function pickTaskLineStatusTone(
  status: PickTaskLine["status"]
): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "picked":
      return "success";
    case "pending":
      return "warning";
    case "missing_stock":
    case "wrong_sku":
    case "wrong_batch":
    case "wrong_location":
    case "cancelled":
      return "danger";
    default:
      return "normal";
  }
}

export function resetPrototypePickTasksForTest() {
  prototypePickTasks = createPrototypePickTasks();
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function fromApiPickTask(task: PickTaskApi): PickTask {
  return {
    id: task.id,
    orgId: task.org_id,
    pickTaskNo: task.pick_task_no,
    salesOrderId: task.sales_order_id,
    orderNo: task.order_no,
    warehouseId: task.warehouse_id,
    warehouseCode: task.warehouse_code,
    status: task.status,
    assignedTo: task.assigned_to,
    assignedAt: task.assigned_at,
    startedAt: task.started_at,
    startedBy: task.started_by,
    completedAt: task.completed_at,
    completedBy: task.completed_by,
    auditLogId: task.audit_log_id,
    lines: task.lines.map(fromApiPickTaskLine),
    createdAt: task.created_at,
    updatedAt: task.updated_at
  };
}

function fromApiPickTaskLine(line: PickTaskLineApi): PickTaskLine {
  return {
    id: line.id,
    lineNo: line.line_no,
    salesOrderLineId: line.sales_order_line_id,
    stockReservationId: line.stock_reservation_id,
    itemId: line.item_id,
    skuCode: line.sku_code,
    batchId: line.batch_id,
    batchNo: line.batch_no,
    warehouseId: line.warehouse_id,
    binId: line.bin_id,
    binCode: line.bin_code,
    qtyToPick: line.qty_to_pick,
    qtyPicked: line.qty_picked,
    baseUOMCode: line.base_uom_code,
    status: line.status,
    pickedAt: line.picked_at,
    pickedBy: line.picked_by,
    createdAt: line.created_at,
    updatedAt: line.updated_at
  };
}

function toApiQuery(query: PickTaskQuery) {
  return {
    warehouse_id: query.warehouseId,
    status: query.status,
    assigned_to: query.assignedTo
  };
}

function queryString(query: Record<string, string | undefined>) {
  const params = new URLSearchParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value) {
      params.set(key, value);
    }
  });

  const value = params.toString();
  return value ? `?${value}` : "";
}

function filterPrototypePickTasks(query: PickTaskQuery) {
  return prototypePickTasks
    .filter((task) => {
      if (query.warehouseId && task.warehouseId !== query.warehouseId) {
        return false;
      }
      if (query.status && task.status !== query.status) {
        return false;
      }
      if (query.assignedTo && task.assignedTo !== query.assignedTo) {
        return false;
      }

      return true;
    })
    .map(clonePickTask);
}

function transitionPrototypePickTask(id: string, action: "start" | "complete" | PickTaskExceptionCode) {
  const task = getPrototypePickTask(id);
  if (action === "start") {
    if (task.status === "created") {
      task.status = "assigned";
      task.assignedTo = "user-picker";
      task.assignedAt = "2026-04-28T09:32:00Z";
    }
    if (task.status !== "assigned" && task.status !== "in_progress") {
      throw new Error("Pick task status transition is invalid");
    }
    task.status = "in_progress";
    task.startedAt = task.startedAt || "2026-04-28T09:33:00Z";
    task.startedBy = task.startedBy || "user-picker";
    task.updatedAt = "2026-04-28T09:33:00Z";
    task.auditLogId = "audit-pick-task-started-prototype";
    return clonePickTask(task);
  }
  if (action === "complete") {
    if (task.status === "completed") {
      return clonePickTask(task);
    }
    if (task.status !== "in_progress" || task.lines.some((line) => line.status !== "picked")) {
      throw new Error("Pick task status transition is invalid");
    }
    task.status = "completed";
    task.completedAt = "2026-04-28T09:40:00Z";
    task.completedBy = "user-picker";
    task.updatedAt = "2026-04-28T09:40:00Z";
    task.auditLogId = "audit-pick-task-completed-prototype";
    return clonePickTask(task);
  }
  if (task.status === "completed" || isExceptionStatus(task.status)) {
    throw new Error("Pick task status transition is invalid");
  }

  task.status = action;
  task.updatedAt = "2026-04-28T09:36:00Z";
  task.auditLogId = "audit-pick-task-exception-prototype";
  return clonePickTask(task);
}

function confirmPrototypePickTaskLine(id: string, lineId: string, pickedQty: string) {
  const task = getPrototypePickTask(id);
  if (task.status !== "in_progress") {
    throw new Error("Pick task status transition is invalid");
  }
  const line = task.lines.find((candidate) => candidate.id === lineId);
  if (!line) {
    throw new Error("Pick task line was not found");
  }
  if (normalizeDecimal(pickedQty) !== normalizeDecimal(line.qtyToPick)) {
    throw new Error("Pick task quantity is invalid");
  }
  line.qtyPicked = line.qtyToPick;
  line.status = "picked";
  line.pickedAt = "2026-04-28T09:38:00Z";
  line.pickedBy = "user-picker";
  line.updatedAt = "2026-04-28T09:38:00Z";
  task.updatedAt = line.updatedAt;
  task.auditLogId = "audit-pick-task-line-confirmed-prototype";

  return clonePickTask(task);
}

function getPrototypePickTask(id: string) {
  const task = prototypePickTasks.find((candidate) => candidate.id === id);
  if (!task) {
    throw new Error("Pick task not found");
  }

  return task;
}

function createPrototypePickTasks(): PickTask[] {
  return [
    {
      id: "pick-so-260428-0001",
      orgId: "org-my-pham",
      pickTaskNo: "PICK-SO-260428-0001",
      salesOrderId: "so-260428-0001",
      orderNo: "SO-260428-0001",
      warehouseId: "wh-hcm-fg",
      warehouseCode: "WH-HCM-FG",
      status: "created",
      lines: [
        {
          id: "pick-so-260428-0001-line-01",
          lineNo: 1,
          salesOrderLineId: "so-260428-0001-line-01",
          stockReservationId: "rsv-so-260428-0001-line-01",
          itemId: "item-serum-30ml",
          skuCode: "SERUM-30ML",
          batchId: "batch-serum-2604a",
          batchNo: "LOT-2604A",
          warehouseId: "wh-hcm-fg",
          binId: "bin-hcm-pick-a01",
          binCode: "PICK-A-01",
          qtyToPick: "3.000000",
          qtyPicked: "0.000000",
          baseUOMCode: "EA",
          status: "pending",
          createdAt: "2026-04-28T09:30:00Z",
          updatedAt: "2026-04-28T09:30:00Z"
        }
      ],
      createdAt: "2026-04-28T09:30:00Z",
      updatedAt: "2026-04-28T09:30:00Z"
    },
    {
      id: "pick-so-260428-0002",
      orgId: "org-my-pham",
      pickTaskNo: "PICK-SO-260428-0002",
      salesOrderId: "so-260428-0002",
      orderNo: "SO-260428-0002",
      warehouseId: "wh-hcm-fg",
      warehouseCode: "WH-HCM-FG",
      status: "in_progress",
      assignedTo: "user-picker",
      assignedAt: "2026-04-28T09:10:00Z",
      startedAt: "2026-04-28T09:12:00Z",
      startedBy: "user-picker",
      lines: [
        {
          id: "pick-so-260428-0002-line-01",
          lineNo: 1,
          salesOrderLineId: "so-260428-0002-line-01",
          stockReservationId: "rsv-so-260428-0002-line-01",
          itemId: "item-cream-50g",
          skuCode: "CREAM-50G",
          batchId: "batch-cream-2604b",
          batchNo: "LOT-2604B",
          warehouseId: "wh-hcm-fg",
          binId: "bin-hcm-pick-b02",
          binCode: "PICK-B-02",
          qtyToPick: "5.000000",
          qtyPicked: "0.000000",
          baseUOMCode: "EA",
          status: "pending",
          createdAt: "2026-04-28T09:08:00Z",
          updatedAt: "2026-04-28T09:12:00Z"
        }
      ],
      createdAt: "2026-04-28T09:08:00Z",
      updatedAt: "2026-04-28T09:12:00Z"
    }
  ];
}

function isExceptionStatus(status: PickTaskStatus) {
  return ["missing_stock", "wrong_sku", "wrong_batch", "wrong_location", "cancelled"].includes(status);
}

function clonePickTask(task: PickTask): PickTask {
  return {
    ...task,
    lines: task.lines.map((line) => ({ ...line }))
  };
}

function normalizeDecimal(value: string) {
  const [whole, fraction = ""] = value.trim().split(".");
  return `${whole}.${fraction.padEnd(6, "0").slice(0, 6)}`;
}
