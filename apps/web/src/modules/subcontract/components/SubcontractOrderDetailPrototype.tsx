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
  createSubcontractFactoryDispatch,
  formatSubcontractFactoryDispatchStatus,
  formatSubcontractDepositStatus,
  formatSubcontractOrderStatus,
  getSubcontractFactoryDispatches,
  getSubcontractOrder,
  markSubcontractFactoryDispatchReady,
  markSubcontractFactoryDispatchSent,
  recordSubcontractFactoryDispatchResponse,
  subcontractFactoryDispatchStatusTone,
  subcontractOrderStatusTone
} from "../services/subcontractOrderService";
import {
  buildSubcontractOrderTimeline,
  productionFactoryOrderSourcePlanHref,
  subcontractOperationsHref,
  type SubcontractOrderTimelineItem
} from "../services/subcontractOrderTimeline";
import {
  buildSubcontractFactoryExecutionTracker,
  type FactoryExecutionWorkItem
} from "../services/subcontractFactoryExecutionTracker";
import type {
  SubcontractFactoryDispatch,
  SubcontractFinalPaymentStatus,
  SubcontractOrder,
  SubcontractOrderMaterialLine
} from "../types";

type SubcontractOrderDetailPrototypeProps = {
  orderId: string;
};

export function SubcontractOrderDetailPrototype({ orderId }: SubcontractOrderDetailPrototypeProps) {
  const [order, setOrder] = useState<SubcontractOrder>();
  const [dispatches, setDispatches] = useState<SubcontractFactoryDispatch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [dispatchError, setDispatchError] = useState<string | undefined>();
  const [dispatchMessage, setDispatchMessage] = useState<string | undefined>();
  const [dispatchBusy, setDispatchBusy] = useState(false);
  const [responseNote, setResponseNote] = useState("");

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    Promise.all([getSubcontractOrder(orderId), getSubcontractFactoryDispatches(orderId)])
      .then(([nextOrder, nextDispatches]) => {
        if (active) {
          setOrder(nextOrder);
          setDispatches(nextDispatches);
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

  const latestDispatch = dispatches[0];
  const timeline = useMemo(
    () => (order ? buildSubcontractOrderTimeline(order, { dispatchStatus: latestDispatch?.status }) : []),
    [latestDispatch?.status, order]
  );
  const executionTracker = useMemo(
    () => (order ? buildSubcontractFactoryExecutionTracker(order, { dispatchStatus: latestDispatch?.status }) : undefined),
    [latestDispatch?.status, order]
  );
  const sourcePlanHref = useMemo(() => (order ? productionFactoryOrderSourcePlanHref(order) : undefined), [order]);

  const reloadFactoryDispatches = async (nextOrder?: SubcontractOrder) => {
    const targetOrder = nextOrder ?? order;
    if (!targetOrder) {
      return;
    }
    const nextDispatches = await getSubcontractFactoryDispatches(targetOrder.id);
    setDispatches(nextDispatches);
  };

  const runDispatchAction = async (
    action: () => Promise<{ order: SubcontractOrder; dispatch: SubcontractFactoryDispatch }>,
    message: string
  ) => {
    setDispatchBusy(true);
    setDispatchError(undefined);
    setDispatchMessage(undefined);
    try {
      const result = await action();
      setOrder(result.order);
      await reloadFactoryDispatches(result.order);
      setDispatchMessage(message);
      if (result.dispatch.status === "confirmed") {
        setResponseNote("");
      }
    } catch (cause) {
      setDispatchError(errorText(cause));
    } finally {
      setDispatchBusy(false);
    }
  };

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

      {executionTracker ? (
        <section className="erp-masterdata-list-card" id="factory-execution-tracker">
          <header className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Theo dõi thực thi nhà máy</h2>
              <p className="erp-page-description">
                Một danh sách công việc cho lệnh này, từ gửi nhà máy, đặt cọc, bàn giao vật tư, duyệt mẫu, sản xuất, nhận hàng đến QC và thanh toán.
              </p>
            </div>
            <StatusChip tone={executionTracker.currentGate.tone}>{factoryExecutionStatusLabel(executionTracker.currentGate.status)}</StatusChip>
          </header>
          <div className="erp-production-selected-plan-main">
            <span className="erp-production-step-label">Việc cần xử lý</span>
            <h3>{executionTracker.currentGate.title}</h3>
            <p>{executionTracker.currentGate.description}</p>
            <div className="erp-production-selected-plan-badges">
              <StatusChip tone={executionTracker.currentGate.tone}>{factoryExecutionStatusLabel(executionTracker.currentGate.status)}</StatusChip>
              <StatusChip tone="normal">{executionTracker.currentGate.metric}</StatusChip>
            </div>
            {executionTracker.currentGate.action ? (
              executionTracker.currentGate.action.disabled ? (
                <button className="erp-button erp-button--secondary erp-button--compact" type="button" disabled>
                  {executionTracker.currentGate.action.label}
                </button>
              ) : (
                <Link className="erp-button erp-button--secondary erp-button--compact" href={executionTracker.currentGate.action.href}>
                  {executionTracker.currentGate.action.label}
                </Link>
              )
            ) : null}
          </div>
          <DataTable
            columns={factoryExecutionColumns}
            rows={executionTracker.items}
            getRowKey={(item) => item.id}
            preserveColumnWidths
            emptyState={<EmptyState title="Chưa có công việc thực thi" />}
          />
        </section>
      ) : null}

      <section className="erp-masterdata-list-card" id="factory-dispatch">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Gửi nhà máy</h2>
            <p className="erp-page-description">
              Tạo bộ gửi lệnh, ghi nhận đã gửi thủ công và phản hồi xác nhận từ nhà máy. Chưa gửi qua email/Zalo trong sprint này.
            </p>
          </div>
          <FactoryDispatchActions
            busy={dispatchBusy}
            dispatch={latestDispatch}
            order={order}
            responseNote={responseNote}
            setResponseNote={setResponseNote}
            onCreate={() =>
              runDispatchAction(
                () => createSubcontractFactoryDispatch({ order, note: "Tạo bộ gửi nhà máy từ chi tiết lệnh." }),
                "Đã tạo bộ gửi nhà máy"
              )
            }
            onReady={(dispatch) =>
              runDispatchAction(() => markSubcontractFactoryDispatchReady(order, dispatch), "Bộ gửi đã sẵn sàng")
            }
            onSent={(dispatch) =>
              runDispatchAction(
                () =>
                  markSubcontractFactoryDispatchSent({
                    order,
                    dispatch,
                    sentBy: "subcontract-user",
                    sentAt: new Date().toISOString(),
                    note: "Đã gửi thủ công ngoài hệ thống.",
                    evidence: [
                      {
                        id: `${dispatch.id}-manual-send`,
                        evidenceType: "manual_send",
                        objectKey: `manual-factory-dispatch/${dispatch.id}`,
                        note: "Ghi nhận gửi thủ công; chưa tích hợp email/Zalo."
                      }
                    ]
                  }),
                "Đã ghi nhận gửi nhà máy"
              )
            }
            onResponse={(dispatch, status) =>
              runDispatchAction(
                () =>
                  recordSubcontractFactoryDispatchResponse({
                    order,
                    dispatch,
                    responseStatus: status,
                    responseBy: "factory-user",
                    respondedAt: new Date().toISOString(),
                    responseNote: responseNote.trim()
                  }),
                status === "confirmed" ? "Nhà máy đã xác nhận" : "Đã ghi nhận phản hồi nhà máy"
              )
            }
          />
        </header>
        {dispatchError ? (
          <p className="erp-form-error" role="alert">
            {dispatchError}
          </p>
        ) : null}
        {dispatchMessage ? (
          <p className="erp-form-success" role="status">
            {dispatchMessage}
          </p>
        ) : null}
        {latestDispatch ? (
          <FactoryDispatchSummary dispatch={latestDispatch} />
        ) : (
          <EmptyState title="Chưa có bộ gửi nhà máy cho lệnh này" />
        )}
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

const factoryExecutionColumns: DataTableColumn<FactoryExecutionWorkItem>[] = [
  {
    key: "task",
    header: "Công việc",
    render: (item) => (
      <span className="erp-masterdata-product-cell">
        <strong>{item.title}</strong>
        <small>{item.description}</small>
      </span>
    ),
    width: "380px"
  },
  {
    key: "status",
    header: "Trạng thái",
    render: (item) => <StatusChip tone={item.tone}>{factoryExecutionStatusLabel(item.status)}</StatusChip>,
    width: "140px"
  },
  {
    key: "metric",
    header: "Số liệu",
    render: (item) => item.metric,
    width: "180px"
  },
  {
    key: "action",
    header: "Thao tác",
    render: (item) =>
      item.action ? (
        item.action.disabled ? (
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" disabled>
            {item.action.label}
          </button>
        ) : (
          <Link className="erp-button erp-button--secondary erp-button--compact" href={item.action.href}>
            {item.action.label}
          </Link>
        )
      ) : (
        "-"
      ),
    width: "160px"
  }
];

function FactoryDispatchActions({
  busy,
  dispatch,
  order,
  responseNote,
  setResponseNote,
  onCreate,
  onReady,
  onSent,
  onResponse
}: {
  busy: boolean;
  dispatch?: SubcontractFactoryDispatch;
  order: SubcontractOrder;
  responseNote: string;
  setResponseNote: (value: string) => void;
  onCreate: () => void;
  onReady: (dispatch: SubcontractFactoryDispatch) => void;
  onSent: (dispatch: SubcontractFactoryDispatch) => void;
  onResponse: (dispatch: SubcontractFactoryDispatch, status: "confirmed" | "revision_requested" | "rejected") => void;
}) {
  if (!dispatch) {
    return (
      <button className="erp-button erp-button--primary" type="button" disabled={busy || order.status !== "approved"} onClick={onCreate}>
        Tạo bộ gửi
      </button>
    );
  }
  if (dispatch.status === "draft" || dispatch.status === "revision_requested") {
    return (
      <button className="erp-button erp-button--primary" type="button" disabled={busy} onClick={() => onReady(dispatch)}>
        Đánh dấu sẵn sàng gửi
      </button>
    );
  }
  if (dispatch.status === "ready") {
    return (
      <button className="erp-button erp-button--primary" type="button" disabled={busy} onClick={() => onSent(dispatch)}>
        Đánh dấu đã gửi
      </button>
    );
  }
  if (dispatch.status === "sent") {
    return (
      <div className="erp-form-inline-actions">
        <label className="erp-form-field erp-form-field--inline">
          <span>Ghi chú phản hồi</span>
          <input value={responseNote} onChange={(event) => setResponseNote(event.target.value)} />
        </label>
        <button className="erp-button erp-button--primary" type="button" disabled={busy} onClick={() => onResponse(dispatch, "confirmed")}>
          Xác nhận
        </button>
        <button
          className="erp-button erp-button--secondary"
          type="button"
          disabled={busy || responseNote.trim() === ""}
          onClick={() => onResponse(dispatch, "revision_requested")}
        >
          Cần chỉnh
        </button>
        <button
          className="erp-button erp-button--secondary"
          type="button"
          disabled={busy || responseNote.trim() === ""}
          onClick={() => onResponse(dispatch, "rejected")}
        >
          Từ chối
        </button>
      </div>
    );
  }

  return null;
}

function FactoryDispatchSummary({ dispatch }: { dispatch: SubcontractFactoryDispatch }) {
  return (
    <div className="erp-production-selected-plan-main">
      <div className="erp-production-selected-plan-badges">
        <StatusChip tone={subcontractFactoryDispatchStatusTone(dispatch.status)}>
          {formatSubcontractFactoryDispatchStatus(dispatch.status)}
        </StatusChip>
        <StatusChip tone="normal">{dispatch.dispatchNo}</StatusChip>
      </div>
      <dl className="erp-production-selected-plan-meta">
        <div>
          <dt>Nhà máy</dt>
          <dd>{dispatch.factoryName}</dd>
        </div>
        <div>
          <dt>Thành phẩm</dt>
          <dd>
            {dispatch.sku} / {formatProductionPlanQuantity(dispatch.plannedQty, dispatch.uomCode)}
          </dd>
        </div>
        <div>
          <dt>Đã gửi</dt>
          <dd>
            {dispatch.sentAt ? formatDate(dispatch.sentAt) : "-"}
            {dispatch.sentBy ? ` / ${dispatch.sentBy}` : ""}
          </dd>
        </div>
        <div>
          <dt>Phản hồi</dt>
          <dd>
            {dispatch.respondedAt ? formatDate(dispatch.respondedAt) : "-"}
            {dispatch.responseBy ? ` / ${dispatch.responseBy}` : ""}
          </dd>
        </div>
        <div>
          <dt>Sẵn sàng</dt>
          <dd>{dispatch.readyAt ? formatDate(dispatch.readyAt) : "-"}</dd>
        </div>
        <div>
          <dt>Bằng chứng gửi</dt>
          <dd>{dispatch.evidence.length} dòng</dd>
        </div>
      </dl>
      {dispatch.factoryResponseNote ? <p className="erp-page-description">{dispatch.factoryResponseNote}</p> : null}
      {dispatch.evidence.length > 0 ? (
        <div className="erp-page-description">
          <strong>Bằng chứng:</strong>{" "}
          {dispatch.evidence
            .map((evidence) => evidence.fileName || evidence.objectKey || evidence.externalURL || evidence.note)
            .filter(Boolean)
            .join("; ")}
        </div>
      ) : null}
      <DataTable
        columns={factoryDispatchLineColumns}
        rows={dispatch.lines}
        getRowKey={(line) => line.id}
        pagination
        preserveColumnWidths
        emptyState={<EmptyState title="Bộ gửi chưa có dòng vật tư" />}
      />
    </div>
  );
}

const factoryDispatchLineColumns: DataTableColumn<SubcontractFactoryDispatch["lines"][number]>[] = [
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
    header: "Cần gửi",
    render: (line) => formatProductionPlanQuantity(line.plannedQty, line.uomCode),
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

function factoryExecutionStatusLabel(status: FactoryExecutionWorkItem["status"]) {
  switch (status) {
    case "complete":
      return "Đã xong";
    case "current":
      return "Đang xử lý";
    case "blocked":
      return "Đang chặn";
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
