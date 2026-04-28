import { apiGet } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import { getGoodsReceipts } from "../../receiving/services/warehouseReceivingService";
import type { GoodsReceipt, GoodsReceiptStockMovement } from "../../receiving/types";
import { getReturnReceipts } from "../../returns/services/returnReceivingService";
import type { ReturnReceipt } from "../../returns/types";
import { getCarrierManifests } from "../../shipping/services/carrierManifestService";
import type { CarrierManifest } from "../../shipping/types";
import type {
  EndOfDayReconciliation,
  EndOfDayReconciliationQuery,
  EndOfDayReconciliationStatus,
  EndOfDayReconciliationSummary,
  ReconciliationLine,
  WarehouseDailyBoardCounterSource,
  WarehouseDailyBoardData,
  WarehouseDailyBoardQuery,
  WarehouseDailyBoardSummary,
  WarehouseFulfillmentMetrics,
  WarehouseDailyShiftCode,
  WarehouseDailyTask,
  WarehouseDailyTaskPriority,
  WarehouseDailyTaskStatus
} from "../types";

type WarehouseFulfillmentMetricsApi = components["schemas"]["WarehouseFulfillmentMetrics"];
type WarehouseFulfillmentMetricsApiQuery = operations["getWarehouseDailyBoardFulfillmentMetrics"]["parameters"]["query"];

const defaultAccessToken = "local-dev-access-token";
export const defaultWarehouseDailyBoardDate = "2026-04-26";
export const defaultWarehouseDailyBoardShiftCode: WarehouseDailyShiftCode = "day";

export const warehouseShiftOptions: { label: string; value: WarehouseDailyShiftCode }[] = [
  { label: "Day", value: "day" },
  { label: "Night", value: "night" }
];

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

export type WarehouseFulfillmentDrillDownKey =
  | "new"
  | "reserved"
  | "picking"
  | "packed"
  | "waiting_handover"
  | "missing"
  | "handover";

export const warehouseDailyBoardCounterSources: WarehouseDailyBoardCounterSource[] = [
  {
    counter: "waiting",
    label: "New work",
    fields: ["order_queue.status=waiting", "goods_receipts.status=draft|submitted|inspect_ready"]
  },
  {
    counter: "picking",
    label: "Picking",
    fields: ["order_queue.status=picking"]
  },
  {
    counter: "packed",
    label: "Packed",
    fields: ["order_queue.status=packed"]
  },
  {
    counter: "handover",
    label: "Handover",
    fields: ["carrier_manifests.status=ready|scanning|exception", "carrier_manifest_lines.scanned"]
  },
  {
    counter: "returns",
    label: "Returns",
    fields: ["return_receipts.status=pending_inspection", "return_receipts.disposition"]
  },
  {
    counter: "reconciliationMismatch",
    label: "Stock variance",
    fields: ["reconciliation_lines.variance_quantity", "stock_movements.stock_status=qc_hold"]
  },
  {
    counter: "overdue",
    label: "P0 exceptions",
    fields: ["warehouse_daily_tasks.priority=P0"]
  }
];

export const prototypeWarehouseDailyTasks: WarehouseDailyTask[] = [
  {
    id: "task-pick-260426-001",
    reference: "SO-260426-001",
    title: "Pick B2B serum order",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    shiftCode: "day",
    status: "waiting",
    priority: "P1",
    owner: "Picking",
    dueAt: "2026-04-26T10:30:00Z",
    href: "/sales",
    source: "order_queue",
    sourceField: "order_queue.status"
  },
  {
    id: "task-pick-260426-002",
    reference: "SO-260426-002",
    title: "Pick marketplace mixed order",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    shiftCode: "day",
    status: "picking",
    priority: "P1",
    owner: "Picking",
    dueAt: "2026-04-26T11:00:00Z",
    href: "/sales",
    source: "order_queue",
    sourceField: "order_queue.status"
  },
  {
    id: "task-pack-260426-003",
    reference: "PACK-260426-017",
    title: "Pack fragile glass serum set",
    warehouseId: "wh-hcm",
    warehouseCode: "HCM",
    shiftCode: "day",
    status: "packed",
    priority: "P2",
    owner: "Packing",
    dueAt: "2026-04-26T12:15:00Z",
    href: "/shipping",
    source: "order_queue",
    sourceField: "order_queue.status"
  },
  {
    id: "task-hn-pick-260426-007",
    reference: "SO-260426-HN-011",
    title: "Pick Hanoi dealer replenishment",
    warehouseId: "wh-hn",
    warehouseCode: "HN",
    shiftCode: "day",
    status: "waiting",
    priority: "P2",
    owner: "Picking",
    dueAt: "2026-04-26T10:45:00Z",
    href: "/sales",
    source: "order_queue",
    sourceField: "order_queue.status"
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

export type WarehouseDailyBoardSources = {
  orderTasks?: WarehouseDailyTask[];
  goodsReceipts?: GoodsReceipt[];
  carrierManifests?: CarrierManifest[];
  returnReceipts?: ReturnReceipt[];
  reconciliations?: EndOfDayReconciliation[];
  fulfillmentMetrics?: WarehouseFulfillmentMetrics;
};

export async function getWarehouseDailyBoard(query: WarehouseDailyBoardQuery = {}): Promise<WarehouseDailyBoardData> {
  const date = query.date ?? defaultWarehouseDailyBoardDate;
  const warehouseId = query.warehouseId ?? "";
  const shiftCode = query.shiftCode ?? defaultWarehouseDailyBoardShiftCode;
  const carrierCode = normalizeCarrierCode(query.carrierCode);
  const [goodsReceipts, carrierManifests, returnReceipts, fulfillmentMetrics] = await Promise.all([
    getGoodsReceipts(),
    getCarrierManifests({ warehouseId: warehouseId || undefined, date, carrierCode: carrierCode || undefined }),
    getReturnReceipts({ warehouseId: warehouseId || undefined }),
    getWarehouseFulfillmentMetrics({
      ...query,
      date,
      shiftCode,
      carrierCode: carrierCode || undefined
    })
  ]);

  return composeWarehouseDailyBoard(
    {
      ...query,
      date,
      shiftCode,
      carrierCode: carrierCode || undefined
    },
    {
      orderTasks: prototypeWarehouseDailyTasks,
      goodsReceipts,
      carrierManifests,
      returnReceipts,
      reconciliations: prototypeEndOfDayReconciliations,
      fulfillmentMetrics
    }
  );
}

export async function getWarehouseFulfillmentMetrics(
  query: WarehouseDailyBoardQuery = {}
): Promise<WarehouseFulfillmentMetrics | undefined> {
  try {
    const metrics = await apiGet("/warehouse/daily-board/fulfillment-metrics", {
      accessToken: defaultAccessToken,
      query: toFulfillmentMetricsApiQuery(query)
    });

    return fromFulfillmentMetricsApi(metrics);
  } catch {
    return undefined;
  }
}

export function buildWarehouseFulfillmentDrillDownHref(
  key: WarehouseFulfillmentDrillDownKey,
  query: WarehouseDailyBoardQuery = {}
) {
  switch (key) {
    case "reserved":
      return salesOrderDrillDownHref(query, "reserved");
    case "missing":
      return carrierManifestDrillDownHref(query, "exception");
    case "handover":
      return carrierManifestDrillDownHref(query, "handed_over");
    case "waiting_handover":
      return warehouseTaskDrillDownHref(query, "handover");
    case "picking":
      return warehouseTaskDrillDownHref(query, "picking");
    case "packed":
      return warehouseTaskDrillDownHref(query, "packed");
    case "new":
    default:
      return warehouseTaskDrillDownHref(query, "waiting");
  }
}

export function composeWarehouseDailyBoard(
  query: WarehouseDailyBoardQuery = {},
  sources: WarehouseDailyBoardSources = {}
): WarehouseDailyBoardData {
  const date = query.date ?? defaultWarehouseDailyBoardDate;
  const warehouseId = query.warehouseId ?? "";
  const shiftCode = query.shiftCode ?? defaultWarehouseDailyBoardShiftCode;
  const carrierCode = normalizeCarrierCode(query.carrierCode);
  const sourceTasks = [
    ...(sources.orderTasks ?? []),
    ...receivingTasks(sources.goodsReceipts ?? []),
    ...stockMovementTasks(sources.goodsReceipts ?? []),
    ...shippingTasks(sources.carrierManifests ?? []),
    ...returnTasks(sources.returnReceipts ?? []),
    ...reconciliationTasks(sources.reconciliations ?? [])
  ];
  const baseTasks = sourceTasks.filter((task) => {
    if (task.dueAt.slice(0, 10) !== date) {
      return false;
    }
    if (task.shiftCode !== shiftCode) {
      return false;
    }
    if (warehouseId && !matchesWarehouseScope(task.warehouseId, warehouseId)) {
      return false;
    }
    if (carrierCode && task.carrierCode !== carrierCode) {
      return false;
    }

    return true;
  });
  const filteredTasks = query.status ? baseTasks.filter((task) => task.status === query.status) : baseTasks;
  const tasks = sortWarehouseTasksByRisk(filteredTasks);
  const warehouseCode = warehouseCodeForBoard(warehouseId, baseTasks);

  return {
    warehouseId: warehouseId || "all",
    warehouseCode,
    date,
    shiftCode,
    shiftStatus: hasPriorityZero(baseTasks) ? "open" : "closing",
    owner: warehouseId === "wh-hn" ? "HN Lead" : "Warehouse Lead",
    summary: summarizeWarehouseDailyBoard(baseTasks),
    fulfillment:
      sources.fulfillmentMetrics ??
      summarizeWarehouseFulfillmentMetrics(
        {
          warehouseId: warehouseId || undefined,
          date,
          shiftCode,
          carrierCode: carrierCode || undefined
        },
        baseTasks
      ),
    sourceFields: warehouseDailyBoardCounterSources,
    tasks
  };
}

export async function getEndOfDayReconciliations(
  query: EndOfDayReconciliationQuery = {}
): Promise<EndOfDayReconciliation[]> {
  const date = query.date ?? defaultWarehouseDailyBoardDate;
  const warehouseId = query.warehouseId ?? "";
  const shiftCode = query.shiftCode ?? defaultWarehouseDailyBoardShiftCode;

  return prototypeEndOfDayReconciliations.filter((reconciliation) => {
    if (date && reconciliation.date !== date) {
      return false;
    }
    if (shiftCode && reconciliation.shiftCode !== shiftCode) {
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

function receivingTasks(receipts: GoodsReceipt[]): WarehouseDailyTask[] {
  return receipts
    .filter((receipt) => receipt.status !== "posted")
    .map((receipt) => ({
      id: `task-receiving-${receipt.id}`,
      reference: receipt.receiptNo,
      title: receivingTaskTitle(receipt),
      warehouseId: receipt.warehouseId,
      warehouseCode: receipt.warehouseCode,
      shiftCode: defaultWarehouseDailyBoardShiftCode,
      status: "waiting",
      priority: receivingPriority(receipt),
      owner: receipt.status === "inspect_ready" ? "QC" : "Receiving",
      dueAt: receipt.inspectReadyAt ?? receipt.submittedAt ?? receipt.updatedAt,
      href: "/receiving",
      source: "receiving",
      sourceField: "goods_receipts.status"
    }));
}

function stockMovementTasks(receipts: GoodsReceipt[]): WarehouseDailyTask[] {
  return receipts.flatMap((receipt) =>
    (receipt.stockMovements ?? [])
      .filter((movement) => movement.stockStatus === "qc_hold")
      .map((movement) => stockMovementTask(receipt, movement))
  );
}

function stockMovementTask(receipt: GoodsReceipt, movement: GoodsReceiptStockMovement): WarehouseDailyTask {
  return {
    id: `task-stock-movement-${movement.movementNo.toLowerCase()}`,
    reference: movement.movementNo,
    title: `Review QC-hold movement for ${receipt.receiptNo}`,
    warehouseId: movement.warehouseId,
    warehouseCode: receipt.warehouseCode,
    shiftCode: defaultWarehouseDailyBoardShiftCode,
    status: "mismatch",
    priority: "P0",
    owner: "Inventory",
    dueAt: receipt.postedAt ?? receipt.updatedAt,
    href: "/inventory",
    source: "stock_movement",
    sourceField: "stock_movements.stock_status"
  };
}

function shippingTasks(manifests: CarrierManifest[]): WarehouseDailyTask[] {
  return manifests
    .filter((manifest) => manifest.status !== "completed")
    .map((manifest) => ({
      id: `task-shipping-${manifest.id}`,
      reference: manifest.id.toUpperCase(),
      title: shippingTaskTitle(manifest),
      warehouseId: manifest.warehouseId,
      warehouseCode: manifest.warehouseCode,
      carrierCode: manifest.carrierCode,
      shiftCode: defaultWarehouseDailyBoardShiftCode,
      status: "handover",
      priority: shippingPriority(manifest),
      owner: manifest.owner,
      dueAt: manifest.createdAt,
      href: "/shipping",
      source: "shipping",
      sourceField: "carrier_manifests.status,carrier_manifest_lines.scanned"
    }));
}

function returnTasks(receipts: ReturnReceipt[]): WarehouseDailyTask[] {
  return receipts
    .filter((receipt) => receipt.status === "pending_inspection")
    .map((receipt) => ({
      id: `task-return-${receipt.id}`,
      reference: receipt.returnCode ?? receipt.receiptNo,
      title: returnTaskTitle(receipt),
      warehouseId: receipt.warehouseId,
      warehouseCode: receipt.warehouseCode,
      shiftCode: defaultWarehouseDailyBoardShiftCode,
      status: "returns",
      priority: receipt.unknownCase ? "P0" : "P1",
      owner: "Returns",
      dueAt: receipt.receivedAt,
      href: "/returns",
      source: "returns",
      sourceField: "return_receipts.status,return_receipts.disposition"
    }));
}

function reconciliationTasks(reconciliations: EndOfDayReconciliation[]): WarehouseDailyTask[] {
  return reconciliations.flatMap((reconciliation) =>
    reconciliation.lines
      .filter((line) => line.varianceQuantity !== 0)
      .map((line) => ({
        id: `task-reconciliation-${line.id}`,
        reference: `VAR-${reconciliation.date.replaceAll("-", "")}-${line.sku}`,
        title: `Resolve ${line.batchNo} stock variance`,
        warehouseId: reconciliation.warehouseId,
        warehouseCode: reconciliation.warehouseCode,
        shiftCode: reconciliation.shiftCode as WarehouseDailyShiftCode,
        status: "mismatch" as const,
        priority: "P0" as const,
        owner: line.owner,
        dueAt: `${reconciliation.date}T09:15:00Z`,
        href: "/inventory",
        source: "reconciliation" as const,
        sourceField: "reconciliation_lines.variance_quantity"
      }))
  );
}

function receivingTaskTitle(receipt: GoodsReceipt) {
  switch (receipt.status) {
    case "inspect_ready":
      return `Post inspected receipt ${receipt.referenceDocId}`;
    case "submitted":
      return `Inspect submitted receipt ${receipt.referenceDocId}`;
    case "draft":
    default:
      return `Complete receiving draft ${receipt.referenceDocId}`;
  }
}

function receivingPriority(receipt: GoodsReceipt): WarehouseDailyTaskPriority {
  return receipt.status === "inspect_ready" ? "P1" : "P2";
}

function shippingTaskTitle(manifest: CarrierManifest) {
  if (manifest.status === "exception") {
    return `${manifest.carrierCode} handover exception`;
  }
  if (manifest.summary.missingCount > 0) {
    return `${manifest.carrierCode} handover has ${manifest.summary.missingCount} missing scan`;
  }

  return `${manifest.carrierCode} handover ready`;
}

function shippingPriority(manifest: CarrierManifest): WarehouseDailyTaskPriority {
  if (manifest.status === "exception") {
    return "P0";
  }
  if (manifest.summary.missingCount > 0) {
    return "P1";
  }

  return "P2";
}

function returnTaskTitle(receipt: ReturnReceipt) {
  if (receipt.unknownCase) {
    return `Investigate unknown return ${receipt.scanCode}`;
  }

  return `Inspect return ${receipt.originalOrderNo ?? receipt.receiptNo}`;
}

function toFulfillmentMetricsApiQuery(query: WarehouseDailyBoardQuery): WarehouseFulfillmentMetricsApiQuery {
  return {
    warehouse_id: query.warehouseId,
    date: query.date,
    shift_code: query.shiftCode,
    carrier_code: normalizeCarrierCode(query.carrierCode) || undefined
  };
}

function fromFulfillmentMetricsApi(metrics: WarehouseFulfillmentMetricsApi): WarehouseFulfillmentMetrics {
  return {
    warehouseId: metrics.warehouse_id,
    date: metrics.date,
    shiftCode: metrics.shift_code,
    carrierCode: metrics.carrier_code,
    totalOrders: metrics.total_orders,
    newOrders: metrics.new_orders,
    reservedOrders: metrics.reserved_orders,
    pickingOrders: metrics.picking_orders,
    packedOrders: metrics.packed_orders,
    waitingHandoverOrders: metrics.waiting_handover_orders,
    missingOrders: metrics.missing_orders,
    handoverOrders: metrics.handover_orders,
    generatedAt: metrics.generated_at
  };
}

export function summarizeWarehouseFulfillmentMetrics(
  query: WarehouseDailyBoardQuery,
  tasks: WarehouseDailyTask[]
): WarehouseFulfillmentMetrics {
  const date = query.date ?? defaultWarehouseDailyBoardDate;
  const orderTasks = tasks.filter((task) => task.source === "order_queue");
  const handoverTasks = tasks.filter((task) => task.source === "shipping");
  const newOrders = countOrderTasksByStatus(orderTasks, "waiting");
  const pickingOrders = countOrderTasksByStatus(orderTasks, "picking");
  const packedOrders = countOrderTasksByStatus(orderTasks, "packed");
  const waitingHandoverOrders = handoverTasks.length;
  const missingOrders = handoverTasks.filter((task) => task.priority !== "P2").length;

  return {
    warehouseId: query.warehouseId,
    date,
    shiftCode: query.shiftCode ?? defaultWarehouseDailyBoardShiftCode,
    carrierCode: normalizeCarrierCode(query.carrierCode) || undefined,
    totalOrders: newOrders + pickingOrders + packedOrders + waitingHandoverOrders,
    newOrders,
    reservedOrders: 0,
    pickingOrders,
    packedOrders,
    waitingHandoverOrders,
    missingOrders,
    handoverOrders: 0,
    generatedAt: `${date}T00:00:00Z`
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

function countOrderTasksByStatus(tasks: WarehouseDailyTask[], status: WarehouseDailyTaskStatus) {
  return tasks.filter((task) => task.status === status).length;
}

function hasPriorityZero(tasks: WarehouseDailyTask[]) {
  return tasks.some((task) => task.priority === "P0");
}

function normalizeCarrierCode(carrierCode?: string) {
  return carrierCode?.trim().toUpperCase() ?? "";
}

function warehouseTaskDrillDownHref(query: WarehouseDailyBoardQuery, queue: WarehouseDailyTaskStatus) {
  return drillDownHref(
    "/warehouse",
    {
      warehouse_id: query.warehouseId,
      date: query.date,
      shift_code: query.shiftCode,
      carrier_code: normalizeCarrierCode(query.carrierCode) || undefined,
      queue
    },
    "task-board"
  );
}

function salesOrderDrillDownHref(query: WarehouseDailyBoardQuery, status: string) {
  return drillDownHref(
    "/sales",
    {
      warehouse_id: salesWarehouseIdForDrillDown(query.warehouseId),
      date: query.date,
      status
    },
    "sales-list"
  );
}

function carrierManifestDrillDownHref(query: WarehouseDailyBoardQuery, status: string) {
  return drillDownHref(
    "/shipping",
    {
      warehouse_id: query.warehouseId,
      date: query.date,
      carrier_code: normalizeCarrierCode(query.carrierCode) || undefined,
      status
    },
    "carrier-manifest-list"
  );
}

function drillDownHref(path: string, query: Record<string, string | undefined>, hash: string) {
  const params = new URLSearchParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value) {
      params.set(key, value);
    }
  });
  const queryString = params.toString();

  return `${path}${queryString ? `?${queryString}` : ""}#${hash}`;
}

function salesWarehouseIdForDrillDown(warehouseId?: string) {
  if (warehouseId === "wh-hcm") {
    return "wh-hcm-fg";
  }

  return warehouseId;
}

function matchesWarehouseScope(taskWarehouseId: string, queryWarehouseId: string) {
  if (taskWarehouseId === queryWarehouseId) {
    return true;
  }

  return taskWarehouseId.startsWith(`${queryWarehouseId}-`);
}

function warehouseCodeForBoard(warehouseId: string, tasks: WarehouseDailyTask[]) {
  if (!warehouseId) {
    return "All";
  }
  if (warehouseId === "wh-hcm") {
    return "HCM";
  }
  if (warehouseId === "wh-hn") {
    return "HN";
  }

  return tasks[0]?.warehouseCode ?? warehouseId;
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
