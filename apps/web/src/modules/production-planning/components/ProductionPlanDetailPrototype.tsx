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
import {
  formatProductionPlanQuantity,
  getProductionPlan,
  productionPlanStatusDisplay,
  productionPlanStatusTone
} from "../services/productionPlanService";
import {
  formatPurchaseDate,
  formatPurchaseMoney,
  getPurchaseOrders,
  purchaseOrderStatusTone
} from "../../purchase/services/purchaseOrderService";
import { buildProductionPlanWorkflowContext } from "../services/productionPlanWorkflowContext";
import { buildProductionPlanWorklist, type ProductionPlanWorkTask } from "../services/productionPlanWorklist";
import { t } from "@/shared/i18n";
import type { PurchaseOrder, PurchaseOrderStatus } from "../../purchase/types";
import type { ProductionPlan, ProductionPlanLine, PurchaseRequestDraftLine } from "../types";

type ProductionPlanDetailPrototypeProps = {
  planId: string;
};

export function ProductionPlanDetailPrototype({ planId }: ProductionPlanDetailPrototypeProps) {
  const [plan, setPlan] = useState<ProductionPlan>();
  const [relatedPurchaseOrders, setRelatedPurchaseOrders] = useState<PurchaseOrder[]>([]);
  const [relatedPurchaseOrdersLoading, setRelatedPurchaseOrdersLoading] = useState(false);
  const [relatedPurchaseOrdersError, setRelatedPurchaseOrdersError] = useState<string | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    getProductionPlan(planId)
      .then((nextPlan) => {
        if (active) {
          setPlan(nextPlan);
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
  }, [planId]);

  useEffect(() => {
    if (!plan?.planNo) {
      setRelatedPurchaseOrders([]);
      return;
    }

    let active = true;
    setRelatedPurchaseOrdersLoading(true);
    setRelatedPurchaseOrdersError(undefined);

    getPurchaseOrders({ search: plan.planNo })
      .then((orders) => {
        if (active) {
          setRelatedPurchaseOrders(orders);
        }
      })
      .catch((loadError) => {
        if (active) {
          setRelatedPurchaseOrdersError(errorText(loadError));
        }
      })
      .finally(() => {
        if (active) {
          setRelatedPurchaseOrdersLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [plan?.planNo]);

  const workflowContext = useMemo(() => (plan ? buildProductionPlanWorkflowContext(plan) : undefined), [plan]);
  const workTasks = useMemo(() => (plan ? buildProductionPlanWorklist(plan) : []), [plan]);

  if (loading) {
    return <LoadingState title="Đang tải kế hoạch sản xuất" />;
  }

  if (error || !plan || !workflowContext) {
    return (
      <ErrorState
        title="Không tải được kế hoạch sản xuất"
        description={error ?? "Không tìm thấy kế hoạch sản xuất."}
        action={
          <Link className="erp-button erp-button--secondary" href="/production">
            Quay lại sản xuất
          </Link>
        }
      />
    );
  }

  return (
    <main className="erp-masterdata-page">
      <header className="erp-page-header">
        <div>
          <span className="erp-production-step-label">Kế hoạch sản xuất</span>
          <h1 className="erp-page-title">{plan.planNo}</h1>
          <p className="erp-page-description">{workflowContext.outputLabel}</p>
        </div>
        <div className="erp-page-actions">
          <Link className="erp-button erp-button--secondary" href="/production">
            Quay lại danh sách
          </Link>
        </div>
      </header>

      <section className="erp-production-selected-plan-card" aria-label="Tóm tắt kế hoạch sản xuất">
        <div className="erp-production-selected-plan-main">
          <span className="erp-production-step-label">Tóm tắt</span>
          <h2>{plan.outputSku}</h2>
          <p>{plan.outputItemName}</p>
          <div className="erp-production-selected-plan-badges">
            <StatusChip tone={productionPlanStatusTone(plan.status)}>{productionPlanStatusDisplay(plan.status)}</StatusChip>
            <StatusChip tone={workflowContext.materialStatusTone}>{workflowContext.materialStatusLabel}</StatusChip>
            <StatusChip tone={workflowContext.purchaseLineCount > 0 ? "warning" : "success"}>
              {workflowContext.purchaseLineCount} dòng đề nghị mua
            </StatusChip>
          </div>
        </div>
        <dl className="erp-production-selected-plan-meta">
          <div>
            <dt>Số lượng</dt>
            <dd>{workflowContext.quantityLabel}</dd>
          </div>
          <div>
            <dt>Công thức</dt>
            <dd>{workflowContext.formulaLabel}</dd>
          </div>
          <div>
            <dt>Ngày kế hoạch</dt>
            <dd>{dateRangeLabel(plan.plannedStartDate, plan.plannedEndDate)}</dd>
          </div>
        </dl>
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Danh sách công việc của kế hoạch</h2>
            <p className="erp-page-description">
              Theo dõi một kế hoạch từ tạo kế hoạch, vật tư, mua hàng, nhập kho, QC đến gia công.
            </p>
          </div>
        </header>
        <DataTable columns={workTaskColumns} rows={workTasks} getRowKey={(task) => task.id} preserveColumnWidths />
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">PO liên quan</h2>
            <p className="erp-page-description">
              Các PO có ghi chú nguồn từ {plan.planNo}; mở từng PO để xem trạng thái, timeline và dòng hàng.
            </p>
          </div>
          <Link className="erp-button erp-button--secondary" href={`/purchase?search=${encodeURIComponent(plan.planNo)}#purchase-list`}>
            Mở danh sách PO
          </Link>
        </header>
        <DataTable
          columns={relatedPurchaseOrderColumns}
          rows={relatedPurchaseOrders}
          getRowKey={(order) => order.id}
          loading={relatedPurchaseOrdersLoading}
          error={relatedPurchaseOrdersError}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="Chưa có PO liên quan đến kế hoạch này" />}
        />
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Nhu cầu vật tư</h2>
            <p className="erp-page-description">Công thức tính cho 1 thành phẩm; nhu cầu dưới đây đã nhân theo số lượng kế hoạch.</p>
          </div>
        </header>
        <DataTable
          columns={materialColumns}
          rows={plan.lines}
          getRowKey={(line) => line.id}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="Kế hoạch chưa có dòng nhu cầu vật tư" />}
        />
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Đề nghị mua nháp từ kế hoạch</h2>
            <p className="erp-page-description">
              Các dòng này là nguồn để tạo PO vật tư thiếu. PO/Nhập kho/QC thật hiện theo dõi ở module tương ứng.
            </p>
          </div>
          <Link className="erp-button erp-button--secondary" href={`/purchase?search=${encodeURIComponent(plan.planNo)}#purchase-list`}>
            Mở mua hàng
          </Link>
        </header>
        <DataTable
          columns={purchaseDraftColumns}
          rows={plan.purchaseRequestDraft.lines}
          getRowKey={(line) => line.id}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="Kế hoạch này không có đề nghị mua nháp" />}
        />
      </section>
    </main>
  );
}

const workTaskColumns: DataTableColumn<ProductionPlanWorkTask>[] = [
  {
    key: "step",
    header: "Bước",
    render: (task) => `Bước ${task.step}`,
    width: "90px"
  },
  {
    key: "task",
    header: "Công việc",
    render: (task) => (
      <div className="erp-masterdata-product-cell">
        <strong>{task.title}</strong>
        <small>{task.detail}</small>
      </div>
    ),
    width: "420px"
  },
  {
    key: "status",
    header: "Trạng thái",
    render: (task) => <StatusChip tone={task.statusTone}>{task.statusLabel}</StatusChip>,
    width: "180px"
  },
  {
    key: "action",
    header: "Thao tác",
    align: "right",
    sticky: true,
    render: (task) => renderTaskAction(task),
    width: "160px"
  }
];

const materialColumns: DataTableColumn<ProductionPlanLine>[] = [
  {
    key: "sku",
    header: "Vật tư",
    render: (line) => (
      <div className="erp-masterdata-product-cell">
        <strong>{line.componentSku}</strong>
        <small>{line.componentName}</small>
      </div>
    ),
    width: "260px"
  },
  {
    key: "formula",
    header: "Định lượng/TP",
    render: (line) => formatProductionPlanQuantity(line.formulaQty, line.formulaUomCode),
    width: "140px"
  },
  {
    key: "required",
    header: "Nhu cầu",
    render: (line) => formatProductionPlanQuantity(line.requiredStockBaseQty, line.stockBaseUomCode),
    width: "140px"
  },
  {
    key: "available",
    header: "Tồn khả dụng",
    render: (line) => formatProductionPlanQuantity(line.availableQty, line.stockBaseUomCode),
    width: "140px"
  },
  {
    key: "shortage",
    header: "Cần mua",
    render: (line) => formatProductionPlanQuantity(line.shortageQty, line.stockBaseUomCode),
    width: "140px"
  },
  {
    key: "status",
    header: "Xử lý",
    render: (line) => (
      <StatusChip tone={line.needsPurchase ? "warning" : "success"}>{line.needsPurchase ? "Đề nghị mua nháp" : "Đủ tồn"}</StatusChip>
    ),
    width: "160px"
  }
];

const purchaseDraftColumns: DataTableColumn<PurchaseRequestDraftLine>[] = [
  {
    key: "sku",
    header: "Vật tư mua",
    render: (line) => (
      <div className="erp-masterdata-product-cell">
        <strong>{line.sku}</strong>
        <small>{line.itemName}</small>
      </div>
    ),
    width: "260px"
  },
  {
    key: "qty",
    header: "Số lượng đề nghị",
    render: (line) => formatProductionPlanQuantity(line.requestedQty, line.uomCode),
    width: "160px"
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
    render: (line) => line.note ?? "",
    width: "220px"
  }
];

const relatedPurchaseOrderColumns: DataTableColumn<PurchaseOrder>[] = [
  {
    key: "po",
    header: "PO",
    render: (order) => (
      <div className="erp-masterdata-product-cell">
        <strong>{order.poNo}</strong>
        <small>{order.supplierName}</small>
      </div>
    ),
    width: "230px"
  },
  {
    key: "status",
    header: "Trạng thái",
    render: (order) => <StatusChip tone={purchaseOrderStatusTone(order.status)}>{purchaseOrderStatusLabel(order.status)}</StatusChip>,
    width: "160px"
  },
  {
    key: "expected",
    header: "Ngày dự kiến",
    render: (order) => formatPurchaseDate(order.expectedDate),
    width: "130px"
  },
  {
    key: "lines",
    header: "Dòng",
    render: (order) => order.lineCount ?? order.lines.length,
    align: "right",
    width: "90px"
  },
  {
    key: "received",
    header: "Đã nhận",
    render: (order) => order.receivedLineCount ?? 0,
    align: "right",
    width: "100px"
  },
  {
    key: "total",
    header: "Tổng tiền",
    render: (order) => formatPurchaseMoney(order.totalAmount, order.currencyCode),
    align: "right",
    width: "150px"
  },
  {
    key: "action",
    header: "Thao tác",
    align: "right",
    sticky: true,
    render: (order) => (
      <Link className="erp-button erp-button--secondary erp-button--compact" href={`/purchase/orders/${order.id}`}>
        Mở PO
      </Link>
    ),
    width: "120px"
  }
];

function renderTaskAction(task: ProductionPlanWorkTask) {
  if (!task.action) {
    return <span>-</span>;
  }

  if (task.action.disabled || !task.action.href) {
    return (
      <button className="erp-button erp-button--secondary erp-button--compact" type="button" disabled>
        {task.action.label}
      </button>
    );
  }

  return (
    <Link className="erp-button erp-button--secondary erp-button--compact" href={task.action.href}>
      {task.action.label}
    </Link>
  );
}

function dateRangeLabel(start?: string, end?: string) {
  if (!start && !end) {
    return "Chưa đặt ngày";
  }
  if (start && end) {
    return `${formatDate(start)} - ${formatDate(end)}`;
  }

  return formatDate(start ?? end ?? "");
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("vi-VN", { day: "2-digit", month: "2-digit", year: "numeric" }).format(new Date(value));
}

function purchaseOrderStatusLabel(status: PurchaseOrderStatus) {
  return t(`purchase.status.${status}`);
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Request failed";
}
