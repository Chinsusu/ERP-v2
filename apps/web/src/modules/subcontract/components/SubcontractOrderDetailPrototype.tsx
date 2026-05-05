"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import {
  DataTable,
  EmptyState,
  ErrorState,
  LoadingState,
  StatusChip,
  type DataTableColumn
} from "@/shared/design-system/components";
import { formatProductionPlanQuantity } from "@/modules/production-planning/services/productionPlanService";
import {
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  getSubcontractOrder,
  subcontractOrderStatusTone
} from "../services/subcontractOrderService";
import {
  buildSubcontractOrderTimeline,
  productionFactoryOrderSourcePlanHref,
  subcontractOperationsHref,
  type SubcontractOrderTimelineItem
} from "../services/subcontractOrderTimeline";
import type { SubcontractFinalPaymentStatus, SubcontractOrder, SubcontractOrderMaterialLine } from "../types";

type SubcontractOrderDetailPrototypeProps = {
  orderId: string;
};

export function SubcontractOrderDetailPrototype({ orderId }: SubcontractOrderDetailPrototypeProps) {
  const [order, setOrder] = useState<SubcontractOrder>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    getSubcontractOrder(orderId)
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
  }, [orderId]);

  const timeline = useMemo(() => (order ? buildSubcontractOrderTimeline(order) : []), [order]);
  const sourcePlanHref = useMemo(() => (order ? productionFactoryOrderSourcePlanHref(order) : undefined), [order]);

  if (loading) {
    return <LoadingState title="Đang tải lệnh nhà máy" />;
  }

  if (error || !order) {
    return (
      <ErrorState
        title="Không tải được lệnh nhà máy"
        description={error ?? "Không tìm thấy lệnh nhà máy."}
        action={
          <Link className="erp-button erp-button--secondary" href="/production">
            Quay lại sản xuất
          </Link>
        }
      />
    );
  }

  return (
    <main className="erp-masterdata-page erp-purchase-detail-page">
      <header className="erp-page-header">
        <div>
          <span className="erp-production-step-label">Lệnh nhà máy / Gia công ngoài</span>
          <h1 className="erp-page-title">{order.orderNo}</h1>
          <p className="erp-page-description">
            {order.factoryName} / {order.sku} / {formatProductionPlanQuantity(String(order.quantity), order.uomCode ?? "PCS")}
          </p>
        </div>
        <div className="erp-page-actions">
          {sourcePlanHref ? (
            <Link className="erp-button erp-button--secondary" href={sourcePlanHref}>
              Mở kế hoạch
            </Link>
          ) : null}
          <Link className="erp-button erp-button--secondary" href={subcontractOperationsHref(order)}>
            Mở xử lý lệnh
          </Link>
          <Link className="erp-button erp-button--secondary" href="/production">
            Quay lại sản xuất
          </Link>
        </div>
      </header>

      <section className="erp-production-selected-plan-card" aria-label="Tóm tắt lệnh nhà máy">
        <div className="erp-production-selected-plan-main">
          <span className="erp-production-step-label">Trạng thái lệnh</span>
          <h2>{formatSubcontractOrderStatus(order.status)}</h2>
          <p>
            Lệnh gửi nhà máy ngoài từ kế hoạch {order.sourceProductionPlanNo ?? "-"}; thành phẩm chỉ vào tồn khả dụng sau khi nhận về và QC đạt.
          </p>
          <div className="erp-production-selected-plan-badges">
            <StatusChip tone={subcontractOrderStatusTone(order.status)}>{formatSubcontractOrderStatus(order.status)}</StatusChip>
            <StatusChip tone={order.sampleRequired ? "warning" : "normal"}>
              {order.sampleRequired ? "Cần duyệt mẫu" : "Không yêu cầu mẫu"}
            </StatusChip>
            <StatusChip tone={closeoutTone(order)}>{closeoutLabel(order)}</StatusChip>
          </div>
        </div>
        <dl className="erp-production-selected-plan-meta">
          <div>
            <dt>Nhà máy</dt>
            <dd>{order.factoryName}</dd>
          </div>
          <div>
            <dt>Ngày nhận dự kiến</dt>
            <dd>{formatDate(order.expectedDeliveryDate)}</dd>
          </div>
          <div>
            <dt>Kế hoạch nguồn</dt>
            <dd>{order.sourceProductionPlanNo ?? "-"}</dd>
          </div>
          <div>
            <dt>Đã nhận / Đạt QC</dt>
            <dd>
              {formatProductionPlanQuantity(order.receivedQty ?? "0", order.uomCode ?? "PCS")} /{" "}
              {formatProductionPlanQuantity(order.acceptedQty ?? "0", order.uomCode ?? "PCS")}
            </dd>
          </div>
          <div>
            <dt>Đặt cọc</dt>
            <dd>{formatSubcontractDepositStatus(order.depositStatus)}</dd>
          </div>
          <div>
            <dt>Thanh toán cuối</dt>
            <dd>{formatFinalPaymentStatus(order.finalPaymentStatus)}</dd>
          </div>
        </dl>
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Timeline lệnh nhà máy</h2>
            <p className="erp-page-description">
              Theo dõi lệnh từ kế hoạch sản xuất, xác nhận nhà máy, xuất vật tư, nhận thành phẩm, QC đến đóng lệnh.
            </p>
          </div>
        </header>
        <ol className="erp-document-timeline" aria-label="Timeline lệnh nhà máy">
          {timeline.map((item) => (
            <TimelineItem item={item} key={item.id} />
          ))}
        </ol>
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Vật tư xuất cho nhà máy</h2>
            <p className="erp-page-description">Các dòng nguyên liệu/bao bì cần bàn giao cho nhà máy theo lệnh này.</p>
          </div>
          <Link className="erp-button erp-button--secondary" href={subcontractOperationsHref(order, "subcontract-transfer")}>
            Mở xuất vật tư
          </Link>
        </header>
        <DataTable
          columns={materialLineColumns}
          rows={order.materialLines}
          getRowKey={(line) => line.id}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="Lệnh này chưa có dòng vật tư" />}
        />
      </section>
    </main>
  );
}

const materialLineColumns: DataTableColumn<SubcontractOrderMaterialLine>[] = [
  {
    key: "sku",
    header: "Vật tư",
    render: (line) => (
      <span className="erp-masterdata-product-cell">
        <strong>{line.skuCode}</strong>
        <small>{line.itemName}</small>
      </span>
    ),
    width: "260px"
  },
  {
    key: "planned",
    header: "Cần xuất",
    render: (line) => formatProductionPlanQuantity(line.plannedQty, line.uomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "issued",
    header: "Đã xuất",
    render: (line) => formatProductionPlanQuantity(line.issuedQty, line.uomCode),
    align: "right",
    width: "140px"
  },
  {
    key: "qc",
    header: "Kiểm soát",
    render: (line) => (line.lotTraceRequired ? "Lô, QC" : "Không kiểm lô"),
    width: "140px"
  },
  {
    key: "note",
    header: "Ghi chú",
    render: (line) => line.note ?? "-",
    width: "220px"
  }
];

function TimelineItem({ item }: { item: SubcontractOrderTimelineItem }) {
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
        {item.occurredAt ? <small>{formatDate(item.occurredAt)}</small> : null}
      </div>
    </li>
  );
}

function timelineStatusLabel(status: SubcontractOrderTimelineItem["status"]) {
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

function closeoutLabel(order: SubcontractOrder) {
  if (order.status === "closed") {
    return "Đã đóng";
  }
  if (["accepted", "final_payment_ready"].includes(order.status)) {
    return "QC đạt";
  }
  if (order.status === "rejected_with_factory_issue") {
    return "Claim nhà máy";
  }
  if (["finished_goods_received", "qc_in_progress"].includes(order.status)) {
    return "Chờ QC";
  }

  return "Đang xử lý";
}

function closeoutTone(order: SubcontractOrder) {
  if (order.status === "closed" || ["accepted", "final_payment_ready"].includes(order.status)) {
    return "success" as const;
  }
  if (order.status === "rejected_with_factory_issue") {
    return "danger" as const;
  }
  if (["finished_goods_received", "qc_in_progress"].includes(order.status)) {
    return "warning" as const;
  }

  return "info" as const;
}

function formatFinalPaymentStatus(status: SubcontractFinalPaymentStatus) {
  switch (status) {
    case "released":
      return "Đã thanh toán";
    case "hold":
      return "Đang giữ";
    case "pending":
    default:
      return "Chờ xử lý";
  }
}

function formatDate(value?: string) {
  if (!value) {
    return "-";
  }

  return new Intl.DateTimeFormat("vi-VN", { day: "2-digit", month: "2-digit", year: "numeric" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Request failed";
}
