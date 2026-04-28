import { ApiError, apiGetRaw, apiPost } from "../../../shared/api/client";
import type {
  ConfirmPackTaskLineInput,
  PackTask,
  PackTaskExceptionCode,
  PackTaskLine,
  PackTaskQuery,
  PackTaskStatus
} from "../types";

const defaultAccessToken = "local-dev-access-token";

type PackTaskApi = {
  id: string;
  org_id: string;
  pack_task_no: string;
  sales_order_id: string;
  sales_order_status?: string;
  order_no: string;
  pick_task_id: string;
  pick_task_no: string;
  warehouse_id: string;
  warehouse_code: string;
  status: PackTaskStatus;
  assigned_to?: string;
  assigned_at?: string;
  started_at?: string;
  started_by?: string;
  packed_at?: string;
  packed_by?: string;
  audit_log_id?: string;
  lines: PackTaskLineApi[];
  created_at: string;
  updated_at: string;
};

type PackTaskLineApi = {
  id: string;
  line_no: number;
  pick_task_line_id: string;
  sales_order_line_id: string;
  item_id: string;
  sku_code: string;
  batch_id: string;
  batch_no: string;
  warehouse_id: string;
  qty_to_pack: string;
  qty_packed: string;
  base_uom_code: string;
  status: PackTaskLine["status"];
  packed_at?: string;
  packed_by?: string;
  created_at: string;
  updated_at: string;
};

export const packTaskWarehouseOptions = [
  { label: "All warehouses", value: "" },
  { label: "Finished Goods HCM", value: "wh-hcm-fg", code: "WH-HCM-FG" }
] as const;

export const packTaskStatusOptions: { label: string; value: "" | PackTaskStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Created", value: "created" },
  { label: "In progress", value: "in_progress" },
  { label: "Packed", value: "packed" },
  { label: "Pack exception", value: "pack_exception" },
  { label: "Cancelled", value: "cancelled" }
];

export const packTaskExceptionOptions: { label: string; value: PackTaskExceptionCode }[] = [
  { label: "Pack exception", value: "pack_exception" }
];

let prototypePackTasks = createPrototypePackTasks();

export async function getPackTasks(query: PackTaskQuery = {}): Promise<PackTask[]> {
  try {
    const tasks = await apiGetRaw<PackTaskApi[]>(`/pack-tasks${queryString(toApiQuery(query))}`, {
      accessToken: defaultAccessToken
    });

    return tasks.map(fromApiPackTask);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypePackTasks(query);
  }
}

export async function startPackTask(id: string): Promise<PackTask> {
  try {
    return fromApiPackTask(
      await apiPost<PackTaskApi, Record<string, never>>(
        `/pack-tasks/${encodeURIComponent(id)}/start`,
        {},
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypePackTask(id, "start");
  }
}

export async function confirmPackTask(id: string, lines: ConfirmPackTaskLineInput[]): Promise<PackTask> {
  try {
    return fromApiPackTask(
      await apiPost<PackTaskApi, { lines: { line_id: string; packed_qty: string }[] }>(
        `/pack-tasks/${encodeURIComponent(id)}/confirm`,
        {
          lines: lines.map((line) => ({
            line_id: line.lineId,
            packed_qty: line.packedQty
          }))
        },
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return confirmPrototypePackTask(id, lines);
  }
}

export async function reportPackTaskException(
  id: string,
  exceptionCode: PackTaskExceptionCode,
  investigation = "",
  lineId = ""
): Promise<PackTask> {
  try {
    return fromApiPackTask(
      await apiPost<PackTaskApi, { line_id?: string; exception_code: PackTaskExceptionCode; investigation: string }>(
        `/pack-tasks/${encodeURIComponent(id)}/exception`,
        { line_id: lineId || undefined, exception_code: exceptionCode, investigation },
        { accessToken: defaultAccessToken }
      )
    );
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypePackTask(id, exceptionCode, lineId);
  }
}

export function packTaskStatusTone(status: PackTaskStatus): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "packed":
      return "success";
    case "in_progress":
      return "info";
    case "created":
      return "warning";
    case "pack_exception":
    case "cancelled":
      return "danger";
    default:
      return "normal";
  }
}

export function packTaskLineStatusTone(
  status: PackTaskLine["status"]
): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "packed":
      return "success";
    case "pending":
      return "warning";
    case "pack_exception":
    case "cancelled":
      return "danger";
    default:
      return "normal";
  }
}

export function resetPrototypePackTasksForTest() {
  prototypePackTasks = createPrototypePackTasks();
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function fromApiPackTask(task: PackTaskApi): PackTask {
  return {
    id: task.id,
    orgId: task.org_id,
    packTaskNo: task.pack_task_no,
    salesOrderId: task.sales_order_id,
    salesOrderStatus: task.sales_order_status,
    orderNo: task.order_no,
    pickTaskId: task.pick_task_id,
    pickTaskNo: task.pick_task_no,
    warehouseId: task.warehouse_id,
    warehouseCode: task.warehouse_code,
    status: task.status,
    assignedTo: task.assigned_to,
    assignedAt: task.assigned_at,
    startedAt: task.started_at,
    startedBy: task.started_by,
    packedAt: task.packed_at,
    packedBy: task.packed_by,
    auditLogId: task.audit_log_id,
    lines: task.lines.map(fromApiPackTaskLine),
    createdAt: task.created_at,
    updatedAt: task.updated_at
  };
}

function fromApiPackTaskLine(line: PackTaskLineApi): PackTaskLine {
  return {
    id: line.id,
    lineNo: line.line_no,
    pickTaskLineId: line.pick_task_line_id,
    salesOrderLineId: line.sales_order_line_id,
    itemId: line.item_id,
    skuCode: line.sku_code,
    batchId: line.batch_id,
    batchNo: line.batch_no,
    warehouseId: line.warehouse_id,
    qtyToPack: line.qty_to_pack,
    qtyPacked: line.qty_packed,
    baseUOMCode: line.base_uom_code,
    status: line.status,
    packedAt: line.packed_at,
    packedBy: line.packed_by,
    createdAt: line.created_at,
    updatedAt: line.updated_at
  };
}

function toApiQuery(query: PackTaskQuery) {
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

function filterPrototypePackTasks(query: PackTaskQuery) {
  return prototypePackTasks
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
    .map(clonePackTask);
}

function transitionPrototypePackTask(id: string, action: "start" | PackTaskExceptionCode, lineId = "") {
  const task = getPrototypePackTask(id);
  if (action === "start") {
    if (task.status !== "created" && task.status !== "in_progress") {
      throw new Error("Pack task status transition is invalid");
    }
    task.status = "in_progress";
    task.startedAt = task.startedAt || "2026-04-28T10:47:00Z";
    task.startedBy = task.startedBy || "user-packer";
    task.updatedAt = "2026-04-28T10:47:00Z";
    task.auditLogId = "audit-pack-task-started-prototype";
    return clonePackTask(task);
  }
  if (task.status === "packed" || task.status === "cancelled") {
    throw new Error("Pack task status transition is invalid");
  }
  const exceptionLine = lineId ? task.lines.find((candidate) => candidate.id === lineId) : undefined;
  if (lineId && (!exceptionLine || exceptionLine.status !== "pending")) {
    throw new Error("Pack task status transition is invalid");
  }
  task.status = "pack_exception";
  task.updatedAt = "2026-04-28T10:50:00Z";
  task.auditLogId = "audit-pack-task-exception-prototype";
  if (exceptionLine) {
    exceptionLine.status = "pack_exception";
    exceptionLine.updatedAt = task.updatedAt;
  }

  return clonePackTask(task);
}

function confirmPrototypePackTask(id: string, lines: ConfirmPackTaskLineInput[]) {
  const task = getPrototypePackTask(id);
  if (task.status === "packed") {
    return clonePackTask(task);
  }
  if (task.status !== "in_progress") {
    throw new Error("Pack task status transition is invalid");
  }
  const byLineId = new Map(lines.map((line) => [line.lineId, line.packedQty]));
  task.lines.forEach((line) => {
    const packedQty = byLineId.get(line.id);
    if (!packedQty || normalizeDecimal(packedQty) !== normalizeDecimal(line.qtyToPack)) {
      throw new Error("Pack task quantity is invalid");
    }
    line.qtyPacked = line.qtyToPack;
    line.status = "packed";
    line.packedAt = "2026-04-28T10:55:00Z";
    line.packedBy = "user-packer";
    line.updatedAt = "2026-04-28T10:55:00Z";
  });
  task.status = "packed";
  task.salesOrderStatus = "packed";
  task.packedAt = "2026-04-28T10:55:00Z";
  task.packedBy = "user-packer";
  task.updatedAt = "2026-04-28T10:55:00Z";
  task.auditLogId = "audit-pack-task-confirmed-prototype";

  return clonePackTask(task);
}

function getPrototypePackTask(id: string) {
  const task = prototypePackTasks.find((candidate) => candidate.id === id);
  if (!task) {
    throw new Error("Pack task not found");
  }

  return task;
}

function createPrototypePackTasks(): PackTask[] {
  return [
    {
      id: "pack-so-260428-0003",
      orgId: "org-my-pham",
      packTaskNo: "PACK-SO-260428-0003",
      salesOrderId: "so-260428-0003",
      salesOrderStatus: "packing",
      orderNo: "SO-260428-0003",
      pickTaskId: "pick-so-260428-0003",
      pickTaskNo: "PICK-SO-260428-0003",
      warehouseId: "wh-hcm-fg",
      warehouseCode: "WH-HCM-FG",
      status: "created",
      lines: [
        {
          id: "pack-so-260428-0003-line-01",
          lineNo: 1,
          pickTaskLineId: "pick-so-260428-0003-line-01",
          salesOrderLineId: "so-260428-0003-line-01",
          itemId: "item-serum-30ml",
          skuCode: "SERUM-30ML",
          batchId: "batch-serum-2604a",
          batchNo: "LOT-2604A",
          warehouseId: "wh-hcm-fg",
          qtyToPack: "3.000000",
          qtyPacked: "0.000000",
          baseUOMCode: "EA",
          status: "pending",
          createdAt: "2026-04-28T10:45:00Z",
          updatedAt: "2026-04-28T10:45:00Z"
        }
      ],
      createdAt: "2026-04-28T10:45:00Z",
      updatedAt: "2026-04-28T10:45:00Z"
    },
    {
      id: "pack-so-260428-0004",
      orgId: "org-my-pham",
      packTaskNo: "PACK-SO-260428-0004",
      salesOrderId: "so-260428-0004",
      salesOrderStatus: "packing",
      orderNo: "SO-260428-0004",
      pickTaskId: "pick-so-260428-0004",
      pickTaskNo: "PICK-SO-260428-0004",
      warehouseId: "wh-hcm-fg",
      warehouseCode: "WH-HCM-FG",
      status: "in_progress",
      assignedTo: "user-packer",
      assignedAt: "2026-04-28T10:20:00Z",
      startedAt: "2026-04-28T10:22:00Z",
      startedBy: "user-packer",
      lines: [
        {
          id: "pack-so-260428-0004-line-01",
          lineNo: 1,
          pickTaskLineId: "pick-so-260428-0004-line-01",
          salesOrderLineId: "so-260428-0004-line-01",
          itemId: "item-cream-50g",
          skuCode: "CREAM-50G",
          batchId: "batch-cream-2604b",
          batchNo: "LOT-2604B",
          warehouseId: "wh-hcm-fg",
          qtyToPack: "5.000000",
          qtyPacked: "0.000000",
          baseUOMCode: "EA",
          status: "pending",
          createdAt: "2026-04-28T10:18:00Z",
          updatedAt: "2026-04-28T10:22:00Z"
        }
      ],
      createdAt: "2026-04-28T10:18:00Z",
      updatedAt: "2026-04-28T10:22:00Z"
    }
  ];
}

function clonePackTask(task: PackTask): PackTask {
  return {
    ...task,
    lines: task.lines.map((line) => ({ ...line }))
  };
}

function normalizeDecimal(value: string) {
  const [whole, fraction = ""] = value.trim().split(".");
  return `${whole}.${fraction.padEnd(6, "0").slice(0, 6)}`;
}
