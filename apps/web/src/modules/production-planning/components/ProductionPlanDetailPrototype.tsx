"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  DataTable,
  EmptyState,
  ErrorState,
  LoadingState,
  StatusChip,
  type DataTableColumn
} from "@/shared/design-system/components";
import {
  createWarehouseIssueFromProductionPlan,
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
import { getGoodsReceipts } from "../../receiving/services/warehouseReceivingService";
import {
  buildPurchaseOrderReceiptRows,
  type PurchaseOrderReceiptRow
} from "../../purchase/services/purchaseOrderReceivingTraceability";
import {
  formatSubcontractOrderStatus,
  getSubcontractOrders,
  subcontractOrderStatusTone
} from "../../subcontract/services/subcontractOrderService";
import { buildProductionPlanWorkflowContext } from "../services/productionPlanWorkflowContext";
import { buildProductionPlanWorklist, type ProductionPlanWorkTask } from "../services/productionPlanWorklist";
import { t } from "@/shared/i18n";
import type { PurchaseOrder, PurchaseOrderStatus } from "../../purchase/types";
import type { SubcontractOrder } from "../../subcontract/types";
import type { ProductionPlan, ProductionPlanLine, PurchaseRequestDraftLine } from "../types";

type MaterialIssueActionState = {
  lineId?: string;
  loading?: boolean;
  message?: string;
  error?: string;
};

type ProductionPlanReceiptRow = PurchaseOrderReceiptRow & {
  poNo: string;
  supplierName: string;
};

type ProductionPlanDetailPrototypeProps = {
  planId: string;
};

export function ProductionPlanDetailPrototype({ planId }: ProductionPlanDetailPrototypeProps) {
  const [plan, setPlan] = useState<ProductionPlan>();
  const [relatedPurchaseOrders, setRelatedPurchaseOrders] = useState<PurchaseOrder[]>([]);
  const [relatedPurchaseOrdersLoading, setRelatedPurchaseOrdersLoading] = useState(false);
  const [relatedPurchaseOrdersError, setRelatedPurchaseOrdersError] = useState<string | undefined>();
  const [relatedReceiptRows, setRelatedReceiptRows] = useState<ProductionPlanReceiptRow[]>([]);
  const [relatedReceiptsLoading, setRelatedReceiptsLoading] = useState(false);
  const [relatedReceiptsError, setRelatedReceiptsError] = useState<string | undefined>();
  const [relatedSubcontractOrders, setRelatedSubcontractOrders] = useState<SubcontractOrder[]>([]);
  const [relatedSubcontractOrdersLoading, setRelatedSubcontractOrdersLoading] = useState(false);
  const [relatedSubcontractOrdersError, setRelatedSubcontractOrdersError] = useState<string | undefined>();
  const [materialIssueAction, setMaterialIssueAction] = useState<MaterialIssueActionState>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();

  const loadPlan = useCallback(() => getProductionPlan(planId), [planId]);

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    loadPlan()
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
  }, [loadPlan]);

  useEffect(() => {
    if (!plan?.planNo) {
      setRelatedPurchaseOrders([]);
      setRelatedPurchaseOrdersLoading(false);
      setRelatedPurchaseOrdersError(undefined);
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

  useEffect(() => {
    if (relatedPurchaseOrders.length === 0) {
      setRelatedReceiptRows([]);
      setRelatedReceiptsLoading(false);
      setRelatedReceiptsError(undefined);
      return;
    }

    let active = true;
    setRelatedReceiptsLoading(true);
    setRelatedReceiptsError(undefined);

    getGoodsReceipts()
      .then((receipts) => {
        const nextRows = relatedPurchaseOrders.flatMap((order) =>
          buildPurchaseOrderReceiptRows(order, receipts).map((receipt) => ({
            ...receipt,
            poNo: order.poNo,
            supplierName: order.supplierName
          }))
        );

        if (active) {
          setRelatedReceiptRows(nextRows);
        }
      })
      .catch((loadError) => {
        if (active) {
          setRelatedReceiptsError(errorText(loadError));
        }
      })
      .finally(() => {
        if (active) {
          setRelatedReceiptsLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [relatedPurchaseOrders]);

  useEffect(() => {
    if (!plan?.id) {
      setRelatedSubcontractOrders([]);
      setRelatedSubcontractOrdersLoading(false);
      setRelatedSubcontractOrdersError(undefined);
      return;
    }

    let active = true;
    setRelatedSubcontractOrdersLoading(true);
    setRelatedSubcontractOrdersError(undefined);

    getSubcontractOrders({ sourceProductionPlanId: plan.id, search: plan.planNo })
      .then((orders) => (orders.length === 0 ? getSubcontractOrders({ search: plan.planNo }) : orders))
      .then((orders) => {
        if (active) {
          setRelatedSubcontractOrders(orders);
        }
      })
      .catch((loadError) => {
        if (active) {
          setRelatedSubcontractOrdersError(errorText(loadError));
        }
      })
      .finally(() => {
        if (active) {
          setRelatedSubcontractOrdersLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [plan?.id, plan?.planNo]);

  const handleCreateWarehouseIssue = useCallback(
    async (line: ProductionPlanLine) => {
      setMaterialIssueAction({ lineId: line.id, loading: true });
      try {
        const issue = await createWarehouseIssueFromProductionPlan(planId, {
          lineIds: [line.id],
          destinationType: "factory",
          destinationName: "Factory",
          reasonCode: "production_plan_issue"
        });
        const nextPlan = await loadPlan();
        setPlan(nextPlan);
        setMaterialIssueAction({ message: `Đã tạo phiếu xuất ${issue.issueNo} từ ${line.componentSku}.` });
      } catch (createError) {
        setMaterialIssueAction({ lineId: line.id, error: errorText(createError) });
      }
    },
    [loadPlan, planId]
  );

  const workflowContext = useMemo(() => (plan ? buildProductionPlanWorkflowContext(plan) : undefined), [plan]);
  const workTasks = useMemo(() => (plan ? buildProductionPlanWorklist(plan) : []), [plan]);
  const materialColumns = useMemo(
    () => buildMaterialColumns(handleCreateWarehouseIssue, materialIssueAction),
    [handleCreateWarehouseIssue, materialIssueAction]
  );

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
            <h2 className="erp-section-title">Gia công / thành phẩm</h2>
            <p className="erp-page-description">
              Lệnh gia công liên kết với {plan.planNo}; theo dõi nhận thành phẩm, QC, claim nhà máy và sẵn sàng thanh toán cuối.
            </p>
          </div>
          <Link className="erp-button erp-button--secondary" href={subcontractHref(plan)}>
            Mở gia công
          </Link>
        </header>
        <DataTable
          columns={relatedSubcontractOrderColumns}
          rows={relatedSubcontractOrders}
          getRowKey={(order) => order.id}
          loading={relatedSubcontractOrdersLoading}
          error={relatedSubcontractOrdersError}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="Chưa có lệnh gia công liên kết với kế hoạch này" />}
        />
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Phiếu nhập theo kế hoạch</h2>
            <p className="erp-page-description">
              Các phiếu nhập kho được gom từ những PO liên quan tới {plan.planNo}; dùng để kiểm soát vật tư đã về kho và trạng thái QC.
            </p>
          </div>
          <Link className="erp-button erp-button--secondary" href="/receiving#receiving-list">
            Mở nhập hàng
          </Link>
        </header>
        <DataTable
          columns={relatedReceiptColumns}
          rows={relatedReceiptRows}
          getRowKey={(receipt) => receipt.id}
          loading={relatedReceiptsLoading}
          error={relatedReceiptsError}
          pagination
          preserveColumnWidths
          emptyState={<EmptyState title="Chưa có phiếu nhập liên quan đến các PO của kế hoạch này" />}
        />
      </section>

      <section className="erp-masterdata-list-card">
        <header className="erp-section-header">
          <div>
            <h2 className="erp-section-title">Nhu cầu vật tư</h2>
            <p className="erp-page-description">Công thức tính cho 1 thành phẩm; nhu cầu dưới đây đã nhân theo số lượng kế hoạch.</p>
            {materialIssueAction.message || materialIssueAction.error ? (
              <p className={materialIssueAction.error ? "erp-form-error" : "erp-page-description"}>
                {materialIssueAction.error ?? materialIssueAction.message}
                <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => setMaterialIssueAction({})}>
                  Đóng
                </button>
              </p>
            ) : null}
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
            <h2 className="erp-section-title">Đề nghị mua từ kế hoạch</h2>
            <p className="erp-page-description">
              Mở đề nghị mua để gửi duyệt, duyệt và tạo PO vật tư thiếu cho kế hoạch này.
            </p>
          </div>
          {plan.purchaseRequestDraft.id ? (
            <Link
              className="erp-button erp-button--secondary"
              href={`/purchase/requests/${encodeURIComponent(plan.purchaseRequestDraft.id)}`}
            >
              Mở đề nghị mua
            </Link>
          ) : (
            <Link className="erp-button erp-button--secondary" href={`/purchase?search=${encodeURIComponent(plan.planNo)}#purchase-list`}>
              Mở mua hàng
            </Link>
          )}
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

function buildMaterialColumns(
  onCreateWarehouseIssue: (line: ProductionPlanLine) => void,
  actionState: MaterialIssueActionState
): DataTableColumn<ProductionPlanLine>[] {
  return [
    {
      key: "sku",
      header: "Vật tư",
      render: (line) => (
        <div className="erp-masterdata-product-cell">
          <strong>{line.componentSku}</strong>
          <small>{line.componentName}</small>
        </div>
      ),
      width: "240px"
    },
    {
      key: "formula",
      header: "Định lượng/TP",
      render: (line) => formatProductionPlanQuantity(line.formulaQty, line.formulaUomCode),
      width: "130px"
    },
    {
      key: "required",
      header: "Nhu cầu",
      render: (line) => formatProductionPlanQuantity(line.requiredStockBaseQty, line.stockBaseUomCode),
      width: "130px"
    },
    {
      key: "available",
      header: "Tồn khả dụng",
      render: (line) => formatProductionPlanQuantity(line.availableQty, line.stockBaseUomCode),
      width: "130px"
    },
    {
      key: "issued",
      header: "Đã xuất",
      render: (line) => formatProductionPlanQuantity(line.issuedQty, line.stockBaseUomCode),
      width: "130px"
    },
    {
      key: "remaining",
      header: "Còn phải xuất",
      render: (line) => formatProductionPlanQuantity(line.remainingIssueQty, line.stockBaseUomCode),
      width: "140px"
    },
    {
      key: "shortage",
      header: "Cần mua",
      render: (line) => formatProductionPlanQuantity(line.shortageQty, line.stockBaseUomCode),
      width: "130px"
    },
    {
      key: "status",
      header: "Trạng thái xuất",
      render: (line) => <StatusChip tone={issueStatusTone(line.issueStatus)}>{issueStatusLabel(line.issueStatus)}</StatusChip>,
      width: "160px"
    },
    {
      key: "action",
      header: "Thao tác",
      align: "right",
      sticky: true,
      render: (line) => renderMaterialIssueAction(line, actionState, onCreateWarehouseIssue),
      width: "170px"
    }
  ];
}

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

const relatedReceiptColumns: DataTableColumn<ProductionPlanReceiptRow>[] = [
  {
    key: "receiptNo",
    header: "Phiếu nhập",
    render: (receipt) => (
      <div className="erp-masterdata-product-cell">
        <strong>{receipt.receiptNo}</strong>
        <small>{formatPurchaseDate(receipt.createdAt)}</small>
      </div>
    ),
    width: "220px"
  },
  {
    key: "po",
    header: "PO nguồn",
    render: (receipt) => (
      <div className="erp-masterdata-product-cell">
        <strong>{receipt.poNo}</strong>
        <small>{receipt.supplierName}</small>
      </div>
    ),
    width: "220px"
  },
  {
    key: "status",
    header: "Trạng thái",
    render: (receipt) => <StatusChip tone={receipt.statusTone}>{receipt.statusLabel}</StatusChip>,
    width: "150px"
  },
  {
    key: "lines",
    header: "Dòng",
    render: (receipt) => receipt.lineCount,
    align: "right",
    width: "90px"
  },
  {
    key: "qc",
    header: "QC",
    render: (receipt) => receipt.qcSummary,
    width: "180px"
  },
  {
    key: "action",
    header: "Thao tác",
    align: "right",
    sticky: true,
    render: (receipt) => (
      <Link className="erp-button erp-button--secondary erp-button--compact" href={receipt.href}>
        Mở nhập hàng
      </Link>
    ),
    width: "140px"
  }
];

const relatedSubcontractOrderColumns: DataTableColumn<SubcontractOrder>[] = [
  {
    key: "order",
    header: "Lệnh gia công",
    render: (order) => (
      <div className="erp-masterdata-product-cell">
        <strong>{order.orderNo}</strong>
        <small>{order.factoryName}</small>
      </div>
    ),
    width: "230px"
  },
  {
    key: "status",
    header: "Trạng thái",
    render: (order) => <StatusChip tone={subcontractOrderStatusTone(order.status)}>{formatSubcontractOrderStatus(order.status)}</StatusChip>,
    width: "180px"
  },
  {
    key: "qty",
    header: "Kế hoạch / đã nhận",
    render: (order) =>
      `${formatProductionPlanQuantity(String(order.quantity), order.uomCode ?? "PCS")} / ${formatProductionPlanQuantity(
        order.receivedQty ?? "0",
        order.uomCode ?? "PCS"
      )}`,
    width: "180px"
  },
  {
    key: "qc",
    header: "QC thành phẩm",
    render: (order) => subcontractCloseoutLabel(order),
    width: "190px"
  },
  {
    key: "expected",
    header: "Ngày dự kiến",
    render: (order) => formatPurchaseDate(order.expectedDeliveryDate),
    width: "130px"
  },
  {
    key: "payment",
    header: "Thanh toán cuối",
    render: (order) => subcontractFinalPaymentLabel(order),
    width: "160px"
  },
  {
    key: "action",
    header: "Thao tác",
    align: "right",
    sticky: true,
    render: (order) => (
      <Link className="erp-button erp-button--secondary erp-button--compact" href={subcontractOrderHref(order)}>
        Mở gia công
      </Link>
    ),
    width: "130px"
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

function subcontractCloseoutLabel(order: SubcontractOrder) {
  const uomCode = order.uomCode ?? "PCS";
  const accepted = formatProductionPlanQuantity(order.acceptedQty ?? "0", uomCode);
  const rejected = formatProductionPlanQuantity(order.rejectedQty ?? "0", uomCode);
  if (order.status === "accepted" || order.status === "final_payment_ready" || order.status === "closed") {
    return `Đạt QC ${accepted}`;
  }
  if (order.status === "rejected_with_factory_issue") {
    return `Claim nhà máy ${rejected}`;
  }
  if (order.status === "finished_goods_received" || order.status === "qc_in_progress") {
    return `Chờ QC; đã nhận ${formatProductionPlanQuantity(order.receivedQty ?? "0", uomCode)}`;
  }

  return "Chưa nhận thành phẩm";
}

function subcontractFinalPaymentLabel(order: SubcontractOrder) {
  if (order.status === "closed") {
    return "Đã đóng";
  }
  if (order.status === "final_payment_ready") {
    return "Sẵn sàng";
  }
  if (order.status === "accepted") {
    return "Chờ xác nhận";
  }

  return "Chưa sẵn sàng";
}

function subcontractHref(plan: ProductionPlan) {
  return `/subcontract?source_production_plan_id=${encodeURIComponent(plan.id)}&search=${encodeURIComponent(plan.planNo)}#subcontract-orders`;
}

function subcontractOrderHref(order: SubcontractOrder) {
  const params = new URLSearchParams();
  if (order.sourceProductionPlanId) {
    params.set("source_production_plan_id", order.sourceProductionPlanId);
  }
  params.set("search", order.sourceProductionPlanNo || order.orderNo);

  return `/subcontract?${params.toString()}#subcontract-orders`;
}

function renderMaterialIssueAction(
  line: ProductionPlanLine,
  actionState: MaterialIssueActionState,
  onCreateWarehouseIssue: (line: ProductionPlanLine) => void
) {
  if (actionState.lineId === line.id && actionState.loading) {
    return (
      <button className="erp-button erp-button--secondary erp-button--compact" type="button" disabled>
        Đang tạo
      </button>
    );
  }

  if (actionState.lineId === line.id && actionState.error) {
    return (
      <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => onCreateWarehouseIssue(line)}>
        Thử lại
      </button>
    );
  }

  if (line.issueStatus === "ready_to_issue" || line.issueStatus === "partially_issued") {
    return (
      <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => onCreateWarehouseIssue(line)}>
        Tạo phiếu xuất
      </button>
    );
  }

  if (line.warehouseIssues.length > 0) {
    return (
      <Link className="erp-button erp-button--secondary erp-button--compact" href="/inventory#warehouse-issues">
        Mở phiếu xuất
      </Link>
    );
  }

  return (
    <button className="erp-button erp-button--secondary erp-button--compact" type="button" disabled>
      {line.issueStatus === "shortage" ? "Chờ vật tư" : "Chờ xử lý"}
    </button>
  );
}

function issueStatusLabel(status: ProductionPlanLine["issueStatus"]) {
  switch (status) {
    case "ready_to_issue":
      return "Sẵn sàng xuất";
    case "issue_draft":
      return "Phiếu nháp";
    case "issue_submitted":
      return "Chờ duyệt xuất";
    case "issue_approved":
      return "Đã duyệt xuất";
    case "partially_issued":
      return "Xuất một phần";
    case "issued":
      return "Đã xuất đủ";
    case "waived":
      return "Đã waiver";
    case "blocked":
      return "Bị chặn";
    case "shortage":
    default:
      return "Thiếu vật tư";
  }
}

function issueStatusTone(status: ProductionPlanLine["issueStatus"]) {
  switch (status) {
    case "issued":
    case "waived":
      return "success" as const;
    case "ready_to_issue":
    case "partially_issued":
      return "warning" as const;
    case "issue_draft":
    case "issue_submitted":
    case "issue_approved":
      return "info" as const;
    case "blocked":
      return "danger" as const;
    case "shortage":
    default:
      return "warning" as const;
  }
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
