"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import {
  DataTable,
  EmptyState,
  ErrorState,
  LoadingState,
  StatusChip,
  type DataTableColumn,
  type StatusTone
} from "@/shared/design-system/components";
import { t } from "@/shared/i18n";
import {
  approvePurchaseOrder,
  cancelPurchaseOrder,
  closePurchaseOrder,
  formatPurchaseDate,
  formatPurchaseMoney,
  formatPurchaseQuantity,
  getPurchaseOrder,
  purchaseOrderStatusTone,
  submitPurchaseOrder
} from "../services/purchaseOrderService";
import {
  buildPurchaseOrderTimeline,
  purchaseOrderSourcePlanNo,
  remainingPurchaseLineQuantity,
  type PurchaseOrderTimelineItem
} from "../services/purchaseOrderTimeline";
import type { PurchaseOrder, PurchaseOrderLine, PurchaseOrderStatus } from "../types";

type PurchaseOrderDetailPrototypeProps = {
  poId: string;
};

export function PurchaseOrderDetailPrototype({ poId }: PurchaseOrderDetailPrototypeProps) {
  const [order, setOrder] = useState<PurchaseOrder>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [busyAction, setBusyAction] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string }>();

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    setFeedback(undefined);
    getPurchaseOrder(poId)
      .then((nextOrder) => {
        if (active) {
          setOrder(nextOrder);
        }
      })
      .catch((loadError) => {
        if (active) {
          setError(errorText(loadError));
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [poId]);

  const sourcePlanNo = useMemo(() => (order ? purchaseOrderSourcePlanNo(order) : undefined), [order]);
  const timeline = useMemo(() => (order ? buildPurchaseOrderTimeline(order) : []), [order]);

  async function runAction(action: "submit" | "approve" | "cancel" | "close") {
    if (!order || busyAction) {
      return;
    }

    setBusyAction(action);
    setFeedback(undefined);
    try {
      const reason = action === "cancel" ? "Hủy từ trang chi tiết PO" : undefined;
      const result =
        action === "submit"
          ? await submitPurchaseOrder(order.id, order.version)
          : action === "approve"
            ? await approvePurchaseOrder(order.id, order.version)
            : action === "cancel"
              ? await cancelPurchaseOrder(order.id, reason ?? "", order.version)
              : await closePurchaseOrder(order.id, order.version);
      setOrder(result.purchaseOrder);
      setFeedback({
        tone: action === "cancel" ? "warning" : "success",
        message: `${result.purchaseOrder.poNo} / ${purchaseOrderStatusLabel(result.currentStatus)}`
      });
    } catch (actionError) {
      setFeedback({
        tone: "danger",
        message: actionError instanceof Error ? actionError.message : purchaseCopy("feedback.actionFailed")
      });
    } finally {
      setBusyAction("");
    }
  }

  if (loading) {
    return <LoadingState title="Đang tải đơn mua hàng" />;
  }

  if (error || !order) {
    return (
      <ErrorState
        title="Không tải được đơn mua hàng"
        description={error ?? "Không tìm thấy đơn mua hàng."}
        action={
          <Link className="erp-button erp-button--secondary" href="/purchase">
            Quay lại mua hàng
          </Link>
        }
      />
    );
  }

  return (
    <main className="erp-masterdata-page erp-purchase-detail-page">
      <header className="erp-page-header">
        <div>
          <span className="erp-production-step-label">Đơn mua hàng</span>
          <h1 className="erp-page-title">{order.poNo}</h1>
          <p className="erp-page-description">
            {order.supplierName} / {order.warehouseCode ?? "-"} / {formatPurchaseDate(order.expectedDate)}
          </p>
        </div>
        <div className="erp-page-actions">
          {sourcePlanNo ? (
            <Link className="erp-button erp-button--secondary" href={`/production/plans/${sourcePlanNo.toLowerCase()}`}>
              Mở kế hoạch
            </Link>
          ) : null}
          <Link className="erp-button erp-button--secondary" href="/purchase">
            Quay lại danh sách
          </Link>
        </div>
      </header>

      {feedback ? (
        <p className={`erp-purchase-feedback erp-purchase-feedback--${feedback.tone}`} role="status">
          <span>{feedback.message}</span>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => setFeedback(undefined)}>
            Tắt
          </button>
        </p>
      ) : null}

      <section className="erp-production-selected-plan-card" aria-label="Tóm tắt đơn mua hàng">
        <div className="erp-production-selected-plan-main">
          <span className="erp-production-step-label">Trạng thái PO</span>
          <h2>{purchaseOrderStatusLabel(order.status)}</h2>
          <p>{order.note ?? "Không có ghi chú nguồn."}</p>
          <div className="erp-production-selected-plan-badges">
            <StatusChip tone={purchaseOrderStatusTone(order.status)}>{purchaseOrderStatusLabel(order.status)}</StatusChip>
            <StatusChip tone="info">{order.lineCount ?? order.lines.length} dòng</StatusChip>
            <StatusChip tone={order.receivedLineCount && order.receivedLineCount > 0 ? "success" : "warning"}>
              {order.receivedLineCount ?? 0} dòng đã nhận
            </StatusChip>
          </div>
        </div>
        <dl className="erp-production-selected-plan-meta">
          <div>
            <dt>Nhà cung cấp</dt>
            <dd>{order.supplierName}</dd>
          </div>
          <div>
            <dt>Kho nhận</dt>
            <dd>{order.warehouseCode ?? "-"}</dd>
          </div>
          <div>
            <dt>Tổng tiền</dt>
            <dd>{formatPurchaseMoney(order.totalAmount, order.currencyCode)}</dd>
          </div>
          <div>
            <dt>Phiên bản</dt>
            <dd>{order.version}</dd>
          </div>
        </dl>
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Timeline PO</h2>
            <p className="erp-page-description">Theo dõi PO từ tạo nháp, duyệt, nhập hàng đến đóng hoặc hủy.</p>
          </div>
        </header>
        <ol className="erp-document-timeline" aria-label="Timeline đơn mua hàng">
          {timeline.map((item) => (
            <TimelineItem key={item.id} item={item} />
          ))}
        </ol>
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Dòng hàng PO</h2>
            <p className="erp-page-description">Số lượng mua, đã nhận và còn lại theo từng SKU.</p>
          </div>
        </header>
        <DataTable
          columns={lineColumns}
          rows={order.lines}
          getRowKey={(line) => line.id}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="PO chưa có dòng hàng" />}
        />
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Thao tác PO</h2>
            <p className="erp-page-description">Thao tác chỉ mở khi đúng trạng thái hiện tại của PO.</p>
          </div>
        </header>
        <div className="erp-purchase-actions erp-purchase-detail-actions">
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={order.status !== "draft" || Boolean(busyAction)}
            onClick={() => runAction("submit")}
          >
            {purchaseCopy("actions.submit")}
          </button>
          <button
            className="erp-button erp-button--primary"
            type="button"
            disabled={order.status !== "submitted" || Boolean(busyAction)}
            onClick={() => runAction("approve")}
          >
            {purchaseCopy("actions.approve")}
          </button>
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={!["approved", "partially_received", "received"].includes(order.status) || Boolean(busyAction)}
            onClick={() => runAction("close")}
          >
            {purchaseCopy("actions.close")}
          </button>
          <button
            className="erp-button erp-button--danger"
            type="button"
            disabled={!["draft", "submitted", "approved"].includes(order.status) || Boolean(busyAction)}
            onClick={() => runAction("cancel")}
          >
            {purchaseCopy("actions.cancel")}
          </button>
        </div>
      </section>
    </main>
  );
}

const lineColumns: DataTableColumn<PurchaseOrderLine>[] = [
  {
    key: "sku",
    header: purchaseCopy("line.columns.sku"),
    render: (line) => (
      <span className="erp-purchase-order-cell">
        <strong>{line.skuCode}</strong>
        <small>{line.itemName}</small>
      </span>
    ),
    width: "260px"
  },
  {
    key: "ordered",
    header: purchaseCopy("line.columns.ordered"),
    render: (line) => formatPurchaseQuantity(line.orderedQty, line.uomCode),
    align: "right",
    width: "150px"
  },
  {
    key: "received",
    header: purchaseCopy("line.columns.received"),
    render: (line) => formatPurchaseQuantity(line.receivedQty, line.uomCode),
    align: "right",
    width: "150px"
  },
  {
    key: "remaining",
    header: "Còn lại",
    render: (line) => formatPurchaseQuantity(remainingPurchaseLineQuantity(line), line.uomCode),
    align: "right",
    width: "150px"
  },
  {
    key: "unitPrice",
    header: purchaseCopy("line.columns.unitPrice"),
    render: (line) => formatPurchaseMoney(line.unitPrice, line.currencyCode),
    align: "right",
    width: "150px"
  },
  {
    key: "amount",
    header: purchaseCopy("line.columns.amount"),
    render: (line) => formatPurchaseMoney(line.lineAmount, line.currencyCode),
    align: "right",
    width: "150px"
  }
];

function TimelineItem({ item }: { item: PurchaseOrderTimelineItem }) {
  return (
    <li className="erp-document-timeline-step" data-status={item.status}>
      <span className="erp-document-timeline-marker" aria-hidden="true" />
      <div className="erp-document-timeline-content">
        <div className="erp-document-timeline-heading">
          <strong>{item.label}</strong>
          <StatusChip tone={item.tone}>{timelineStatusLabel(item.status)}</StatusChip>
        </div>
        <p>{item.description}</p>
        {item.action ? (
          item.action.disabled ? (
            <button className="erp-button erp-button--secondary erp-button--compact erp-document-timeline-action" type="button" disabled>
              {item.action.label}
            </button>
          ) : (
            <Link className="erp-button erp-button--secondary erp-button--compact erp-document-timeline-action" href={item.action.href}>
              {item.action.label}
            </Link>
          )
        ) : null}
        {item.occurredAt ? <small>{formatPurchaseDate(item.occurredAt)}</small> : null}
      </div>
    </li>
  );
}

function timelineStatusLabel(status: PurchaseOrderTimelineItem["status"]) {
  switch (status) {
    case "complete":
      return "Đã xong";
    case "current":
      return "Đang xử lý";
    case "blocked":
      return "Dừng";
    case "pending":
    default:
      return "Chờ";
  }
}

function purchaseOrderStatusLabel(status: PurchaseOrderStatus) {
  return purchaseCopy(`status.${status}`);
}

function purchaseCopy(key: string, values?: Record<string, string | number>) {
  return t(`purchase.${key}`, { values });
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Yêu cầu thất bại";
}
