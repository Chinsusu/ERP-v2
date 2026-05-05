import type { StatusTone } from "../../../shared/design-system/components";
import type { GoodsReceipt, GoodsReceiptStatus } from "../../receiving/types";
import type { PurchaseOrder } from "../types";

export type PurchaseOrderReceiptRow = {
  id: string;
  receiptNo: string;
  status: GoodsReceiptStatus;
  statusLabel: string;
  statusTone: StatusTone;
  lineCount: number;
  qcSummary: string;
  createdAt: string;
  postedAt?: string;
  href: string;
};

export function buildPurchaseOrderReceiptRows(order: PurchaseOrder, receipts: GoodsReceipt[]): PurchaseOrderReceiptRow[] {
  return receipts
    .filter((receipt) => receipt.referenceDocType === "purchase_order" && receipt.referenceDocId === order.id)
    .map((receipt) => ({
      id: receipt.id,
      receiptNo: receipt.receiptNo,
      status: receipt.status,
      statusLabel: purchaseOrderReceiptStatusLabel(receipt.status),
      statusTone: purchaseOrderReceiptStatusTone(receipt.status),
      lineCount: receipt.lines.length,
      qcSummary: purchaseOrderReceiptQcSummary(receipt),
      createdAt: receipt.createdAt,
      postedAt: receipt.postedAt,
      href: purchaseOrderReceiptListHref(order, receipt.status)
    }));
}

export function purchaseOrderReceiptListHref(order: PurchaseOrder, status?: GoodsReceiptStatus) {
  const params = new URLSearchParams();
  params.set("po_id", order.id);
  if (order.warehouseId) {
    params.set("warehouse_id", order.warehouseId);
  }
  if (status) {
    params.set("status", status);
  }

  return `/receiving?${params.toString()}#receiving-list`;
}

export function purchaseOrderReceiptStatusLabel(status: GoodsReceiptStatus) {
  switch (status) {
    case "posted":
      return "Đã hạch toán";
    case "inspect_ready":
      return "Sẵn sàng QC";
    case "submitted":
      return "Đã gửi";
    case "draft":
    default:
      return "Nháp";
  }
}

export function purchaseOrderReceiptStatusTone(status: GoodsReceiptStatus): StatusTone {
  switch (status) {
    case "posted":
      return "success";
    case "inspect_ready":
      return "warning";
    case "submitted":
      return "info";
    case "draft":
    default:
      return "normal";
  }
}

function purchaseOrderReceiptQcSummary(receipt: GoodsReceipt) {
  const pass = receipt.lines.filter((line) => line.qcStatus === "pass").length;
  const hold = receipt.lines.filter((line) => line.qcStatus === "hold").length;
  const fail = receipt.lines.filter((line) => line.qcStatus === "fail").length;
  const quarantine = receipt.lines.filter((line) => line.qcStatus === "quarantine" || line.qcStatus === "retest_required").length;
  const parts = [
    pass > 0 ? `Đạt ${pass}` : "",
    hold > 0 ? `Giữ ${hold}` : "",
    fail > 0 ? `Lỗi ${fail}` : "",
    quarantine > 0 ? `Cách ly ${quarantine}` : ""
  ].filter(Boolean);

  return parts.length > 0 ? parts.join(" / ") : "-";
}
