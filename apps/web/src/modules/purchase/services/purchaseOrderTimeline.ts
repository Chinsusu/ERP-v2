import type { StatusTone } from "../../../shared/design-system/components";
import { decimalScales, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type { PurchaseOrder, PurchaseOrderLine, PurchaseOrderStatus } from "../types";

export type PurchaseOrderTimelineStatus = "complete" | "current" | "pending" | "blocked";

export type PurchaseOrderTimelineItem = {
  id: string;
  label: string;
  description: string;
  status: PurchaseOrderTimelineStatus;
  tone: StatusTone;
  occurredAt?: string;
  action?: {
    label: string;
    href: string;
    disabled?: boolean;
  };
};

const quantityScale = decimalScales.quantity;

export function buildPurchaseOrderTimeline(order: PurchaseOrder): PurchaseOrderTimelineItem[] {
  const terminalStatus = terminalPurchaseOrderStatus(order.status);
  const items: PurchaseOrderTimelineItem[] = [
    {
      id: "created",
      label: "Tạo PO",
      description: `${order.poNo} đã được tạo cho ${order.supplierName}.`,
      status: "complete",
      tone: "success",
      occurredAt: order.createdAt
    },
    {
      id: "submitted",
      label: "Gửi duyệt",
      description: "PO chờ người phụ trách mua hàng hoặc quản lý duyệt.",
      status: submittedStatus(order),
      tone: toneForTimelineStatus(submittedStatus(order)),
      occurredAt: order.submittedAt
    },
    {
      id: "approved",
      label: "Duyệt PO",
      description: "Sau khi duyệt, PO có thể dùng để theo dõi giao hàng và nhập kho.",
      status: approvedStatus(order, terminalStatus),
      tone: toneForTimelineStatus(approvedStatus(order, terminalStatus)),
      occurredAt: order.approvedAt
    },
    {
      id: "receiving",
      label: "Nhập hàng",
      description: receivingDescription(order),
      status: receivingStatus(order, terminalStatus),
      tone: toneForTimelineStatus(receivingStatus(order, terminalStatus)),
      occurredAt: order.receivedAt ?? order.partiallyReceivedAt,
      action: {
        label: "Mở nhập hàng",
        href: purchaseOrderReceivingHref(order),
        disabled: !["approved", "partially_received", "received", "closed"].includes(order.status)
      }
    },
    {
      id: "payable",
      label: "Công nợ NCC",
      description: payableDescription(order),
      status: payableStatus(order, terminalStatus),
      tone: toneForTimelineStatus(payableStatus(order, terminalStatus)),
      action: {
        label: "Mở AP",
        href: purchaseOrderPayableHref(order),
        disabled: !["partially_received", "received", "closed"].includes(order.status)
      }
    },
    {
      id: "closed",
      label: "Đóng PO",
      description: "PO được đóng khi đã xử lý xong mua hàng, nhập kho và đối soát.",
      status: closedStatus(order, terminalStatus),
      tone: toneForTimelineStatus(closedStatus(order, terminalStatus)),
      occurredAt: order.closedAt
    }
  ];

  if (order.status === "cancelled") {
    items.push({
      id: "cancelled",
      label: "Hủy PO",
      description: order.cancelReason ?? "PO đã bị hủy.",
      status: "complete",
      tone: "danger",
      occurredAt: order.cancelledAt
    });
  }

  if (order.status === "rejected") {
    items.push({
      id: "rejected",
      label: "Từ chối PO",
      description: order.rejectReason ?? "PO đã bị từ chối.",
      status: "complete",
      tone: "danger",
      occurredAt: order.rejectedAt
    });
  }

  return items;
}

export function remainingPurchaseLineQuantity(line: PurchaseOrderLine) {
  const remaining = toScaledQuantity(line.orderedQty) - toScaledQuantity(line.receivedQty);
  return fromScaledQuantity(remaining > BigInt(0) ? remaining : BigInt(0));
}

export function purchaseOrderSourcePlanNo(order: PurchaseOrder) {
  const match = order.note?.match(/\bPP-\d{6}-\d+\b/i);
  return match ? match[0].toUpperCase() : undefined;
}

export function purchaseOrderReceivingHref(order: PurchaseOrder) {
  const params = new URLSearchParams();
  params.set("po_id", order.id);
  if (order.warehouseId) {
    params.set("warehouse_id", order.warehouseId);
  }

  return `/receiving?${params.toString()}#receiving-draft`;
}

export function purchaseOrderPayableHref(order: PurchaseOrder) {
  const params = new URLSearchParams();
  params.set("ap_q", order.poNo || order.id);

  return `/finance?${params.toString()}#supplier-payables`;
}

function submittedStatus(order: PurchaseOrder): PurchaseOrderTimelineStatus {
  if (order.submittedAt || !["draft", "cancelled"].includes(order.status)) {
    return "complete";
  }
  if (order.status === "cancelled") {
    return "blocked";
  }

  return "current";
}

function approvedStatus(order: PurchaseOrder, terminalStatus: PurchaseOrderStatus | undefined): PurchaseOrderTimelineStatus {
  if (order.approvedAt || ["approved", "partially_received", "received", "closed"].includes(order.status)) {
    return "complete";
  }
  if (terminalStatus) {
    return "blocked";
  }
  if (order.status === "submitted") {
    return "current";
  }

  return "pending";
}

function receivingStatus(order: PurchaseOrder, terminalStatus: PurchaseOrderStatus | undefined): PurchaseOrderTimelineStatus {
  if (terminalStatus) {
    return "blocked";
  }
  if (["received", "closed"].includes(order.status)) {
    return "complete";
  }
  if (["approved", "partially_received"].includes(order.status)) {
    return "current";
  }

  return "pending";
}

function closedStatus(order: PurchaseOrder, terminalStatus: PurchaseOrderStatus | undefined): PurchaseOrderTimelineStatus {
  if (terminalStatus) {
    return "blocked";
  }
  if (order.status === "closed") {
    return "complete";
  }
  if (order.status === "received") {
    return "current";
  }

  return "pending";
}

function payableStatus(order: PurchaseOrder, terminalStatus: PurchaseOrderStatus | undefined): PurchaseOrderTimelineStatus {
  if (terminalStatus) {
    return "blocked";
  }
  if (order.status === "closed") {
    return "complete";
  }
  if (["partially_received", "received"].includes(order.status)) {
    return "current";
  }

  return "pending";
}

function terminalPurchaseOrderStatus(status: PurchaseOrderStatus) {
  return ["cancelled", "rejected"].includes(status) ? status : undefined;
}

function receivingDescription(order: PurchaseOrder) {
  const lineCount = order.lineCount ?? order.lines.length;
  const receivedLineCount = order.receivedLineCount ?? order.lines.filter((line) => line.receivedQty !== "0.000000").length;

  if (order.status === "received" || order.status === "closed") {
    return `${receivedLineCount}/${lineCount} dòng đã nhận đủ theo PO.`;
  }
  if (order.status === "partially_received") {
    return `${receivedLineCount}/${lineCount} dòng đã nhập một phần.`;
  }

  return "Theo dõi giao hàng, phiếu nhập kho và số lượng còn thiếu.";
}

function payableDescription(order: PurchaseOrder) {
  if (order.status === "closed") {
    return "PO đã đối soát và sẵn sàng theo dõi thanh toán NCC.";
  }
  if (["partially_received", "received"].includes(order.status)) {
    return "Các phiếu nhập đã hạch toán và đạt QC sẽ tạo AP cho NCC.";
  }

  return "AP chỉ mở sau khi có phiếu nhập theo PO được hạch toán.";
}

function toneForTimelineStatus(status: PurchaseOrderTimelineStatus): StatusTone {
  switch (status) {
    case "complete":
      return "success";
    case "current":
      return "info";
    case "blocked":
      return "danger";
    case "pending":
    default:
      return "normal";
  }
}

function toScaledQuantity(value: string) {
  return BigInt(normalizeDecimalInput(value, quantityScale).replace(".", ""));
}

function fromScaledQuantity(value: bigint) {
  const digits = value.toString().padStart(quantityScale + 1, "0");
  return `${digits.slice(0, -quantityScale)}.${digits.slice(-quantityScale)}`;
}
