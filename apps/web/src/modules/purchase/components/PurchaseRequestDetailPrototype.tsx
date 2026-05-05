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
import {
  approvePurchaseRequest,
  convertPurchaseRequestToPurchaseOrder,
  formatPurchaseRequestDate,
  formatPurchaseRequestQuantity,
  getPurchaseRequest,
  purchaseRequestStatusLabel,
  purchaseRequestStatusTone,
  submitPurchaseRequest
} from "../services/purchaseRequestService";
import { purchaseSupplierOptions, purchaseWarehouseOptions } from "../services/purchaseOrderService";
import type { PurchaseOrder, PurchaseRequest, PurchaseRequestLine } from "../types";

type PurchaseRequestDetailPrototypeProps = {
  requestId: string;
};

type PurchaseRequestTimelineItem = {
  id: string;
  label: string;
  description: string;
  status: "complete" | "current" | "pending" | "blocked";
  tone: StatusTone;
  occurredAt?: string;
  action?: {
    label: string;
    href?: string;
    disabled?: boolean;
  };
};

export function PurchaseRequestDetailPrototype({ requestId }: PurchaseRequestDetailPrototypeProps) {
  const [request, setRequest] = useState<PurchaseRequest>();
  const [createdOrder, setCreatedOrder] = useState<PurchaseOrder>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [busyAction, setBusyAction] = useState("");
  const [feedback, setFeedback] = useState<{ tone: StatusTone; message: string }>();
  const [supplierId, setSupplierId] = useState(purchaseSupplierOptions[0]?.value ?? "");
  const [warehouseId, setWarehouseId] = useState(purchaseWarehouseOptions[0]?.value ?? "");
  const [expectedDate, setExpectedDate] = useState(defaultDateOffset(3));

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(undefined);
    setFeedback(undefined);

    getPurchaseRequest(requestId)
      .then((nextRequest) => {
        if (active) {
          setRequest(nextRequest);
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
  }, [requestId]);

  const timeline = useMemo(() => (request ? buildPurchaseRequestTimeline(request) : []), [request]);

  async function runAction(action: "submit" | "approve" | "convert") {
    if (!request || busyAction) {
      return;
    }
    setBusyAction(action);
    setFeedback(undefined);
    try {
      if (action === "submit") {
        const result = await submitPurchaseRequest(request.id);
        setRequest(result.purchaseRequest);
        setFeedback({ tone: "success", message: `${result.purchaseRequest.requestNo} / ${purchaseRequestStatusLabel(result.currentStatus)}` });
      } else if (action === "approve") {
        const result = await approvePurchaseRequest(request.id);
        setRequest(result.purchaseRequest);
        setFeedback({ tone: "success", message: `${result.purchaseRequest.requestNo} / ${purchaseRequestStatusLabel(result.currentStatus)}` });
      } else {
        const result = await convertPurchaseRequestToPurchaseOrder(request.id, {
          supplierId,
          warehouseId,
          expectedDate,
          currencyCode: "VND",
          unitPrice: "0"
        });
        setRequest(result.purchaseRequest);
        setCreatedOrder(result.purchaseOrder);
        setFeedback({ tone: "success", message: `Đã tạo ${result.purchaseOrder.poNo} từ ${result.purchaseRequest.requestNo}` });
      }
    } catch (actionError) {
      setFeedback({ tone: "danger", message: errorText(actionError) });
    } finally {
      setBusyAction("");
    }
  }

  if (loading) {
    return <LoadingState title="Đang tải đề nghị mua" />;
  }

  if (error || !request) {
    return (
      <ErrorState
        title="Không tải được đề nghị mua"
        description={error ?? "Không tìm thấy đề nghị mua."}
        action={
          <Link className="erp-button erp-button--secondary" href="/purchase">
            Quay lại mua hàng
          </Link>
        }
      />
    );
  }

  const convertedPOId = request.convertedPurchaseOrderId ?? createdOrder?.id;
  const convertedPONo = request.convertedPurchaseOrderNo ?? createdOrder?.poNo;
  const canConvert = request.status === "approved" && supplierId !== "" && warehouseId !== "" && expectedDate !== "";

  return (
    <main className="erp-masterdata-page erp-purchase-detail-page">
      <header className="erp-page-header">
        <div>
          <span className="erp-production-step-label">Đề nghị mua</span>
          <h1 className="erp-page-title">{request.requestNo}</h1>
          <p className="erp-page-description">Nguồn: kế hoạch sản xuất {request.sourceProductionPlanNo}</p>
        </div>
        <div className="erp-page-actions">
          <Link className="erp-button erp-button--secondary" href={`/production/plans/${encodeURIComponent(request.sourceProductionPlanId)}`}>
            Mở kế hoạch
          </Link>
          <Link className="erp-button erp-button--secondary" href="/purchase">
            Quay lại mua hàng
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

      <section className="erp-production-selected-plan-card" aria-label="Tóm tắt đề nghị mua">
        <div className="erp-production-selected-plan-main">
          <span className="erp-production-step-label">Trạng thái đề nghị</span>
          <h2>{purchaseRequestStatusLabel(request.status)}</h2>
          <p>
            {request.lines.length} dòng vật tư cần mua từ {request.sourceProductionPlanNo}
          </p>
          <div className="erp-production-selected-plan-badges">
            <StatusChip tone={purchaseRequestStatusTone(request.status)}>{purchaseRequestStatusLabel(request.status)}</StatusChip>
            <StatusChip tone="info">{request.lines.length} dòng</StatusChip>
            {convertedPONo ? <StatusChip tone="success">{convertedPONo}</StatusChip> : null}
          </div>
        </div>
        <dl className="erp-production-selected-plan-meta">
          <div>
            <dt>Kế hoạch</dt>
            <dd>{request.sourceProductionPlanNo}</dd>
          </div>
          <div>
            <dt>Ngày tạo</dt>
            <dd>{formatPurchaseRequestDate(request.createdAt)}</dd>
          </div>
          <div>
            <dt>PO đã tạo</dt>
            <dd>{convertedPONo ?? "-"}</dd>
          </div>
        </dl>
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Timeline đề nghị mua</h2>
            <p className="erp-page-description">Theo dõi từ tạo nháp, gửi duyệt, duyệt đến chuyển thành PO.</p>
          </div>
        </header>
        <ol className="erp-document-timeline" aria-label="Timeline đề nghị mua">
          {timeline.map((item) => (
            <TimelineItem key={item.id} item={item} />
          ))}
        </ol>
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Dòng vật tư đề nghị mua</h2>
            <p className="erp-page-description">Các dòng này được sinh từ nhu cầu vật tư thiếu của kế hoạch sản xuất.</p>
          </div>
        </header>
        <DataTable
          columns={requestLineColumns}
          rows={request.lines}
          getRowKey={(line) => line.id}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="Đề nghị mua chưa có dòng vật tư" />}
        />
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Thao tác đề nghị mua</h2>
            <p className="erp-page-description">PO chỉ được tạo sau khi đề nghị mua đã được duyệt.</p>
          </div>
        </header>
        <div className="erp-purchase-actions erp-purchase-detail-actions">
          <button
            className="erp-button erp-button--secondary"
            type="button"
            disabled={request.status !== "draft" || Boolean(busyAction)}
            onClick={() => runAction("submit")}
          >
            {busyAction === "submit" ? "Đang gửi" : "Gửi duyệt"}
          </button>
          <button
            className="erp-button erp-button--primary"
            type="button"
            disabled={request.status !== "submitted" || Boolean(busyAction)}
            onClick={() => runAction("approve")}
          >
            {busyAction === "approve" ? "Đang duyệt" : "Duyệt đề nghị"}
          </button>
          {convertedPOId ? (
            <Link className="erp-button erp-button--primary" href={`/purchase/orders/${encodeURIComponent(convertedPOId)}`}>
              Mở PO {convertedPONo}
            </Link>
          ) : null}
        </div>
        <div className="erp-production-next-action-fields">
          <label className="erp-field">
            <span>Nhà cung cấp</span>
            <select className="erp-input" value={supplierId} onChange={(event) => setSupplierId(event.currentTarget.value)}>
              {purchaseSupplierOptions.map((supplier) => (
                <option key={supplier.value} value={supplier.value}>
                  {supplier.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>Kho nhận</span>
            <select className="erp-input" value={warehouseId} onChange={(event) => setWarehouseId(event.currentTarget.value)}>
              {purchaseWarehouseOptions.map((warehouse) => (
                <option key={warehouse.value} value={warehouse.value}>
                  {warehouse.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>Ngày dự kiến</span>
            <input className="erp-input" type="date" value={expectedDate} onChange={(event) => setExpectedDate(event.currentTarget.value)} />
          </label>
          <button
            className="erp-button erp-button--primary"
            type="button"
            disabled={!canConvert || Boolean(busyAction)}
            onClick={() => runAction("convert")}
          >
            {busyAction === "convert" ? "Đang tạo PO" : "Tạo PO từ đề nghị"}
          </button>
        </div>
      </section>
    </main>
  );
}

const requestLineColumns: DataTableColumn<PurchaseRequestLine>[] = [
  {
    key: "sku",
    header: "Vật tư",
    render: (line) => (
      <span className="erp-purchase-order-cell">
        <strong>{line.sku}</strong>
        <small>{line.itemName}</small>
      </span>
    ),
    width: "280px"
  },
  {
    key: "qty",
    header: "Số lượng đề nghị",
    render: (line) => formatPurchaseRequestQuantity(line.requestedQty, line.uomCode),
    align: "right",
    width: "170px"
  },
  {
    key: "source",
    header: "Dòng kế hoạch",
    render: (line) => line.sourceProductionPlanLineId,
    width: "220px"
  },
  {
    key: "note",
    header: "Ghi chú",
    render: (line) => line.note ?? "-",
    width: "240px"
  }
];

function TimelineItem({ item }: { item: PurchaseRequestTimelineItem }) {
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
            <Link className="erp-button erp-button--secondary erp-button--compact erp-document-timeline-action" href={item.action.href ?? "#"}>
              {item.action.label}
            </Link>
          )
        ) : null}
        {item.occurredAt ? <small>{formatPurchaseRequestDate(item.occurredAt)}</small> : null}
      </div>
    </li>
  );
}

function buildPurchaseRequestTimeline(request: PurchaseRequest): PurchaseRequestTimelineItem[] {
  const submitted = ["submitted", "approved", "converted_to_po"].includes(request.status);
  const approved = ["approved", "converted_to_po"].includes(request.status);
  const converted = request.status === "converted_to_po";

  return [
    {
      id: "draft",
      label: "Tạo đề nghị mua",
      description: `${request.requestNo} được sinh từ kế hoạch ${request.sourceProductionPlanNo}.`,
      status: "complete",
      tone: "success",
      occurredAt: request.createdAt
    },
    {
      id: "submit",
      label: "Gửi duyệt",
      description: "Mua hàng kiểm tra vật tư thiếu trước khi trình duyệt.",
      status: submitted ? "complete" : request.status === "draft" ? "current" : "pending",
      tone: submitted ? "success" : "warning",
      occurredAt: request.submittedAt
    },
    {
      id: "approve",
      label: "Duyệt đề nghị",
      description: "Quản lý xác nhận nhu cầu mua trước khi tạo PO.",
      status: approved ? "complete" : request.status === "submitted" ? "current" : "pending",
      tone: approved ? "success" : request.status === "submitted" ? "info" : "normal",
      occurredAt: request.approvedAt
    },
    {
      id: "convert",
      label: "Tạo PO",
      description: converted
        ? `Đã tạo ${request.convertedPurchaseOrderNo}.`
        : "Tạo PO từ đề nghị đã duyệt, giữ link ngược về kế hoạch sản xuất.",
      status: converted ? "complete" : request.status === "approved" ? "current" : "pending",
      tone: converted ? "success" : request.status === "approved" ? "warning" : "normal",
      occurredAt: request.convertedAt,
      action: request.convertedPurchaseOrderId
        ? {
            label: "Mở PO",
            href: `/purchase/orders/${encodeURIComponent(request.convertedPurchaseOrderId)}`
          }
        : undefined
    }
  ];
}

function timelineStatusLabel(status: PurchaseRequestTimelineItem["status"]) {
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

function defaultDateOffset(days: number) {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date.toISOString().slice(0, 10);
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Yêu cầu thất bại";
}
