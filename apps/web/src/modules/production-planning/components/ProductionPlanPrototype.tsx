"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  DataTable,
  EmptyState,
  ErrorState,
  FormSection,
  StatusChip,
  ToastStack,
  type DataTableColumn,
  type ToastMessage
} from "@/shared/design-system/components";
import { getFormulas } from "@/modules/masterdata/services/formulaMasterDataService";
import { finishedProductTypes, getProducts } from "@/modules/masterdata/services/productMasterDataService";
import type { FormulaMasterDataItem, ProductMasterDataItem } from "@/modules/masterdata/types";
import { purchaseSupplierOptions, purchaseWarehouseOptions } from "@/modules/purchase/services/purchaseOrderService";
import { subcontractFactoryOptions } from "@/modules/subcontract/services/subcontractOrderService";
import {
  createProductionPlans,
  formatProductionPlanQuantity,
  getProductionPlans,
  productionPlanStatusDisplay,
  productionPlanStatusTone,
  summarizeProductionPlans
} from "../services/productionPlanService";
import {
  buildProductionPlanWorkflowContext,
  productionPlanWorkflowSteps
} from "../services/productionPlanWorkflowContext";
import {
  applyFormulaToProductionPlanDraftLine,
  applyProductToProductionPlanDraftLine,
  createProductionPlanDraftLine,
  formulaBelongsToProduct
} from "../services/productionPlanFormDefaults";
import {
  createPurchaseOrderFromProductionPlan,
  createSubcontractOrderFromProductionPlan
} from "../services/productionPlanNextActions";
import type { ProductionPlan, ProductionPlanDraftLine, ProductionPlanLine } from "../types";

export function ProductionPlanPrototype() {
  const [plans, setPlans] = useState<ProductionPlan[]>([]);
  const [products, setProducts] = useState<ProductMasterDataItem[]>([]);
  const [formulas, setFormulas] = useState<FormulaMasterDataItem[]>([]);
  const [draftLines, setDraftLines] = useState<ProductionPlanDraftLine[]>([createProductionPlanDraftLine(newDraftLineID())]);
  const [selectedPlanId, setSelectedPlanId] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [nextActionSaving, setNextActionSaving] = useState<"" | "purchase" | "subcontract">("");
  const [showCreatePlanForm, setShowCreatePlanForm] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [formError, setFormError] = useState<string | undefined>();
  const [nextActionError, setNextActionError] = useState<string | undefined>();
  const [purchaseSupplierId, setPurchaseSupplierId] = useState(purchaseSupplierOptions[0]?.value ?? "");
  const [purchaseWarehouseId, setPurchaseWarehouseId] = useState(purchaseWarehouseOptions[0]?.value ?? "");
  const [purchaseExpectedDate, setPurchaseExpectedDate] = useState(defaultDateOffset(3));
  const [subcontractFactoryId, setSubcontractFactoryId] = useState(subcontractFactoryOptions[0]?.id ?? "");
  const [subcontractExpectedDate, setSubcontractExpectedDate] = useState(defaultDateOffset(14));
  const [toast, setToast] = useState<ToastMessage[]>([]);

  const activeFinishedProducts = useMemo(
    () =>
      products.filter(
        (product) => product.status === "active" && finishedProductTypes.includes(product.itemType) && product.isProducible
      ),
    [products]
  );
  const activeFormulas = useMemo(() => formulas.filter((formula) => formula.status === "active"), [formulas]);
  const summary = useMemo(() => summarizeProductionPlans(plans), [plans]);
  const selectedPlan = useMemo(
    () => plans.find((plan) => plan.id === selectedPlanId) ?? plans[0],
    [plans, selectedPlanId]
  );
  const selectedWorkflowContext = useMemo(
    () => (selectedPlan ? buildProductionPlanWorkflowContext(selectedPlan) : undefined),
    [selectedPlan]
  );
  const selectedPlanPurchaseLineCount = selectedPlan?.purchaseRequestDraft.lines.length ?? 0;
  const selectedPlanShortageLineCount =
    selectedPlan?.lines.filter((line) => line.needsPurchase || Number(line.shortageQty) > 0).length ?? 0;

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    Promise.all([
      getProductionPlans(),
      getProducts({ status: "active", itemTypes: finishedProductTypes }),
      getFormulas({ status: "active" })
    ])
      .then(([planRows, productRows, formulaRows]) => {
        if (!active) {
          return;
        }
        setPlans(planRows);
        setProducts(productRows);
        setFormulas(formulaRows);
        if (planRows[0]) {
          setSelectedPlanId(planRows[0].id);
        }
        setShowCreatePlanForm(planRows.length === 0);
        const firstProduct = productRows.find(
          (product) => product.status === "active" && finishedProductTypes.includes(product.itemType) && product.isProducible
        );
        if (firstProduct) {
          const activeFormulaRows = formulaRows.filter((formula) => formula.status === "active");
          setDraftLines((current) => {
            if (current.some((line) => line.outputItemId)) {
              return current;
            }

            return [createProductionPlanDraftLine(current[0]?.rowId ?? newDraftLineID(), firstProduct, activeFormulaRows)];
          });
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
  }, []);

  useEffect(() => {
    setPurchaseExpectedDate(selectedPlan?.plannedStartDate ?? defaultDateOffset(3));
    setSubcontractExpectedDate(selectedPlan?.plannedEndDate ?? defaultDateOffset(14));
  }, [selectedPlan?.id, selectedPlan?.plannedEndDate, selectedPlan?.plannedStartDate]);

  const planColumns: DataTableColumn<ProductionPlan>[] = [
    {
      key: "plan",
      header: "Kế hoạch",
      render: (plan) => (
        <div className="erp-masterdata-product-cell">
          <strong>{plan.planNo}</strong>
          <small>{formatDate(plan.createdAt)}</small>
          {plan.id === selectedPlan?.id ? <StatusChip tone="info">Đang xử lý</StatusChip> : null}
        </div>
      ),
      width: "160px"
    },
    {
      key: "output",
      header: "Thành phẩm",
      render: (plan) => (
        <div className="erp-masterdata-product-cell">
          <strong>{plan.outputSku}</strong>
          <small>{plan.outputItemName}</small>
        </div>
      ),
      width: "260px"
    },
    {
      key: "formula",
      header: "Công thức",
      render: (plan) => (
        <div className="erp-masterdata-product-cell">
          <strong>{plan.formulaCode}</strong>
          <small>{plan.formulaVersion}</small>
        </div>
      ),
      width: "170px"
    },
    {
      key: "qty",
      header: "Số lượng",
      render: (plan) => formatProductionPlanQuantity(plan.plannedQty, plan.uomCode),
      width: "120px"
    },
    {
      key: "shortage",
      header: "Thiếu vật tư",
      render: (plan) => plan.lines.filter((line) => line.needsPurchase).length,
      align: "right",
      width: "110px"
    },
    {
      key: "status",
      header: "Trạng thái",
      render: (plan) => <StatusChip tone={productionPlanStatusTone(plan.status)}>{productionPlanStatusDisplay(plan.status)}</StatusChip>,
      width: "150px"
    },
    {
      key: "actions",
      header: "",
      align: "right",
      sticky: true,
      render: (plan) => {
        const isSelected = plan.id === selectedPlan?.id;

        return (
          <button
            aria-current={isSelected ? "true" : undefined}
            className={`erp-button erp-button--${isSelected ? "primary" : "secondary"} erp-button--compact`}
            type="button"
            onClick={() => setSelectedPlanId(plan.id)}
          >
            {isSelected ? "Đang chọn" : "Chọn"}
          </button>
        );
      },
      width: "120px"
    }
  ];

  const demandColumns: DataTableColumn<ProductionPlanLine>[] = [
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

  async function refreshPlans() {
    const nextPlans = await getProductionPlans();
    setPlans(nextPlans);
    if (nextPlans[0]) {
      setSelectedPlanId(nextPlans[0].id);
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setFormError(undefined);
    try {
      const readyLines = draftLines.filter((line) => line.outputItemId.trim() !== "");
      const created = await createProductionPlans(readyLines.map(toProductionPlanInput));
      await refreshPlans();
      if (created[0]) {
        setSelectedPlanId(created[0].id);
      }
      setShowCreatePlanForm(false);
      pushToast("Đã tạo kế hoạch", `${created.length} thành phẩm: ${created.map((plan) => plan.outputSku).join(", ")}`, "success");
    } catch (saveError) {
      setFormError(errorText(saveError));
    } finally {
      setSaving(false);
    }
  }

  function changeProduct(rowId: string, productId: string) {
    const product = activeFinishedProducts.find((candidate) => candidate.id === productId);
    setDraftLines((current) =>
      current.map((line) =>
        line.rowId === rowId ? applyProductToProductionPlanDraftLine(line, product, activeFormulas) : line
      )
    );
  }

  function changeFormula(line: ProductionPlanDraftLine, formulaId: string) {
    const formula = activeFormulas.find((candidate) => candidate.id === formulaId);
    const product = productForLine(line);
    setDraftLines((current) =>
      current.map((candidate) =>
        candidate.rowId === line.rowId
          ? applyFormulaToProductionPlanDraftLine(candidate, product, formula ?? formulasForLine(line)[0])
          : candidate
      )
    );
  }

  function updateDraftLine(rowId: string, patch: Partial<ProductionPlanDraftLine>) {
    setDraftLines((current) => current.map((line) => (line.rowId === rowId ? { ...line, ...patch } : line)));
  }

  function addDraftLine() {
    setDraftLines((current) => [...current, createProductionPlanDraftLine(newDraftLineID())]);
  }

  function removeDraftLine(rowId: string) {
    setDraftLines((current) =>
      current.length === 1
        ? [createProductionPlanDraftLine(current[0]?.rowId ?? newDraftLineID())]
        : current.filter((line) => line.rowId !== rowId)
    );
  }

  function productForLine(line: ProductionPlanDraftLine) {
    return activeFinishedProducts.find((product) => product.id === line.outputItemId);
  }

  function formulasForLine(line: ProductionPlanDraftLine) {
    return activeFormulas.filter((formula) => formulaBelongsToProduct(formula, productForLine(line)));
  }

  async function createPurchaseOrderFromSelectedPlan() {
    if (!selectedPlan) {
      return;
    }
    setNextActionSaving("purchase");
    setNextActionError(undefined);
    try {
      const order = await createPurchaseOrderFromProductionPlan(selectedPlan, {
        supplierId: purchaseSupplierId,
        warehouseId: purchaseWarehouseId,
        expectedDate: purchaseExpectedDate
      });
      pushToast("Đã tạo PO", `${order.poNo} từ ${selectedPlan.planNo}`, "success");
    } catch (actionError) {
      setNextActionError(errorText(actionError));
    } finally {
      setNextActionSaving("");
    }
  }

  async function createSubcontractOrderFromSelectedPlan() {
    if (!selectedPlan) {
      return;
    }
    setNextActionSaving("subcontract");
    setNextActionError(undefined);
    try {
      const order = await createSubcontractOrderFromProductionPlan(selectedPlan, {
        factoryId: subcontractFactoryId,
        expectedDeliveryDate: subcontractExpectedDate
      });
      pushToast("Đã tạo lệnh gia công", `${order.orderNo} từ ${selectedPlan.planNo}`, "success");
    } catch (actionError) {
      setNextActionError(errorText(actionError));
    } finally {
      setNextActionSaving("");
    }
  }

  function pushToast(title: string, description: string, tone: ToastMessage["tone"]) {
    setToast([{ id: `${Date.now()}`, title, description, tone }]);
  }

  function dismissToast(id: string) {
    setToast((current) => current.filter((message) => message.id !== id));
  }

  if (error) {
    return <ErrorState title="Không tải được dữ liệu sản xuất" description={error} />;
  }

  return (
    <main className="erp-masterdata-page" aria-busy={loading}>
      <header className="erp-page-header">
        <div>
          <h1 className="erp-page-title">Sản xuất</h1>
          <p className="erp-page-description">Tạo kế hoạch từ thành phẩm và công thức, tính nhu cầu vật tư trước khi mua.</p>
        </div>
      </header>

      <section className="erp-kpi-grid erp-masterdata-kpis" aria-label="Tổng quan kế hoạch sản xuất">
        <article className="erp-card erp-card--padded erp-kpi-card">
          <span className="erp-kpi-label">Kế hoạch</span>
          <strong className="erp-kpi-value">{summary.total}</strong>
          <StatusChip>Tổng</StatusChip>
        </article>
        <article className="erp-card erp-card--padded erp-kpi-card">
          <span className="erp-kpi-label">Nháp</span>
          <strong className="erp-kpi-value">{summary.draft}</strong>
          <StatusChip tone="info">Nháp</StatusChip>
        </article>
        <article className="erp-card erp-card--padded erp-kpi-card">
          <span className="erp-kpi-label">Dòng thiếu vật tư</span>
          <strong className="erp-kpi-value">{summary.shortageLines}</strong>
          <StatusChip tone={summary.shortageLines > 0 ? "warning" : "success"}>MRP</StatusChip>
        </article>
        <article className="erp-card erp-card--padded erp-kpi-card">
          <span className="erp-kpi-label">Dòng đề nghị mua</span>
          <strong className="erp-kpi-value">{summary.purchaseDraftLines}</strong>
          <StatusChip tone="warning">Draft</StatusChip>
        </article>
      </section>

      <section className="erp-production-workflow-steps" aria-label="Luồng xử lý kế hoạch sản xuất">
        {productionPlanWorkflowSteps.map((step) => (
          <article className="erp-production-workflow-step" key={step.number}>
            <span>Bước {step.number}</span>
            <strong>{step.label}</strong>
            <small>{step.description}</small>
          </article>
        ))}
      </section>

      <section className="erp-masterdata-workspace">
        <article className="erp-masterdata-list-card">
          <header className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Kế hoạch sản xuất</h2>
              <p className="erp-page-description">Chọn một kế hoạch để tính nhu cầu vật tư và tạo chứng từ tiếp theo.</p>
            </div>
            <button
              aria-expanded={showCreatePlanForm}
              className="erp-button erp-button--secondary"
              type="button"
              onClick={() => setShowCreatePlanForm((current) => !current)}
            >
              {showCreatePlanForm ? "Ẩn form tạo" : "Tạo kế hoạch mới"}
            </button>
          </header>
          <DataTable
            columns={planColumns}
            rows={plans}
            getRowKey={(plan) => plan.id}
            loading={loading}
            pagination
            preserveColumnWidths
            rowClassName={(plan) => (plan.id === selectedPlan?.id ? "erp-ds-table-row--selected" : undefined)}
            emptyState={<EmptyState title="Chưa có kế hoạch sản xuất" />}
          />
        </article>

        {showCreatePlanForm ? (
          <form onSubmit={submit}>
            <FormSection
              title="Tạo kế hoạch mới"
              description="Form này tạo thêm kế hoạch sản xuất mới. Nếu đang xử lý kế hoạch đã chọn, tiếp tục ở phần nhu cầu vật tư và tạo PO bên dưới."
              footer={
                <>
                  {formError ? <span className="erp-form-error">{formError}</span> : null}
                  <button className="erp-button erp-button--secondary" type="button" onClick={addDraftLine} disabled={saving || loading}>
                    Thêm thành phẩm
                  </button>
                  <button className="erp-button erp-button--primary" type="submit" disabled={saving || loading}>
                    {saving ? "Đang tạo" : `Tạo ${draftLines.filter((line) => line.outputItemId).length || 1} kế hoạch`}
                  </button>
                </>
              }
            >
              <div className="erp-production-plan-lines">
                {draftLines.map((line, index) => {
                  const lineFormulas = formulasForLine(line);

                  return (
                    <div className="erp-production-plan-line" key={line.rowId}>
                      <label className="erp-field">
                        <span>Thành phẩm {index + 1}</span>
                        <select
                          className="erp-input"
                          value={line.outputItemId}
                          onChange={(event) => changeProduct(line.rowId, event.currentTarget.value)}
                        >
                          <option value="">Chọn thành phẩm</option>
                          {activeFinishedProducts.map((product) => (
                            <option key={product.id} value={product.id}>
                              {product.skuCode} - {product.name}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label className="erp-field">
                        <span>Công thức</span>
                        <select
                          className="erp-input"
                          value={line.formulaId ?? ""}
                          onChange={(event) => changeFormula(line, event.currentTarget.value)}
                        >
                          <option value="">Tự chọn công thức active</option>
                          {lineFormulas.map((formula) => (
                            <option key={formula.id} value={formula.id}>
                              {formula.formulaCode} - {formula.formulaVersion}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label className="erp-field">
                        <span>Số lượng</span>
                        <input
                          className="erp-input"
                          inputMode="numeric"
                          pattern="[0-9]*"
                          value={line.plannedQty}
                          onChange={(event) => updateDraftLine(line.rowId, { plannedQty: integerText(event.currentTarget.value) })}
                        />
                      </label>
                      <label className="erp-field">
                        <span>Đơn vị</span>
                        <input
                          className="erp-input"
                          value={line.uomCode}
                          onChange={(event) => updateDraftLine(line.rowId, { uomCode: event.currentTarget.value.toUpperCase() })}
                        />
                      </label>
                      <label className="erp-field">
                        <span>Ngày bắt đầu</span>
                        <input
                          className="erp-input"
                          type="date"
                          value={line.plannedStartDate ?? ""}
                          onChange={(event) => updateDraftLine(line.rowId, { plannedStartDate: event.currentTarget.value })}
                        />
                      </label>
                      <label className="erp-field">
                        <span>Ngày kết thúc</span>
                        <input
                          className="erp-input"
                          type="date"
                          value={line.plannedEndDate ?? ""}
                          onChange={(event) => updateDraftLine(line.rowId, { plannedEndDate: event.currentTarget.value })}
                        />
                      </label>
                      <div className="erp-production-plan-line-actions">
                        <button
                          className="erp-button erp-button--secondary erp-button--compact"
                          type="button"
                          onClick={() => removeDraftLine(line.rowId)}
                          disabled={saving}
                        >
                          Xóa
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            </FormSection>
          </form>
        ) : null}

        {selectedPlan && selectedWorkflowContext ? (
          <article className="erp-production-selected-plan-card" aria-label="Kế hoạch sản xuất đang xử lý">
            <div className="erp-production-selected-plan-main">
              <span className="erp-production-step-label">Kế hoạch đang xử lý</span>
              <h2>{selectedPlan.planNo}</h2>
              <p>{selectedWorkflowContext.outputLabel}</p>
              <div className="erp-production-selected-plan-badges">
                <StatusChip tone={selectedWorkflowContext.materialStatusTone}>{selectedWorkflowContext.materialStatusLabel}</StatusChip>
                <StatusChip tone={selectedWorkflowContext.purchaseLineCount > 0 ? "warning" : "info"}>
                  {selectedWorkflowContext.purchaseLineCount} dòng đề nghị mua
                </StatusChip>
              </div>
            </div>
            <dl className="erp-production-selected-plan-meta">
              <div>
                <dt>Số lượng</dt>
                <dd>{selectedWorkflowContext.quantityLabel}</dd>
              </div>
              <div>
                <dt>Công thức</dt>
                <dd>{selectedWorkflowContext.formulaLabel}</dd>
              </div>
              <div>
                <dt>Bước tiếp theo</dt>
                <dd>{selectedWorkflowContext.shortageLineCount > 0 ? "Tạo PO xử lý thiếu vật tư" : "Tạo lệnh gia công"}</dd>
              </div>
            </dl>
          </article>
        ) : null}

        <article className="erp-masterdata-list-card">
          <header className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Bước 2: Tính nhu cầu vật tư</h2>
              <p className="erp-page-description">
                {selectedWorkflowContext
                  ? `${selectedWorkflowContext.planLabel}; công thức ${selectedWorkflowContext.formulaLabel}; công thức tính cho 1 thành phẩm.`
                  : "Chọn một kế hoạch để xem nhu cầu vật tư."}
              </p>
            </div>
          </header>
          <DataTable
            columns={demandColumns}
            rows={selectedPlan?.lines ?? []}
            getRowKey={(line) => line.id}
            pagination
            preserveColumnWidths
            emptyState={<EmptyState title="Chưa có dòng nhu cầu vật tư" />}
          />
        </article>

        {selectedPlan && selectedWorkflowContext ? (
          <FormSection
            title="Bước 3 và 4: Tạo chứng từ tiếp theo"
            description={`Các chứng từ bên dưới luôn tạo từ kế hoạch đang xử lý: ${selectedPlan.planNo}.`}
          >
            <div className="erp-production-next-actions">
              <article className="erp-production-next-action-panel">
                <header>
                  <span className="erp-production-step-label">Bước 3</span>
                  <h3>{selectedWorkflowContext.purchaseTitle}</h3>
                  <p>{selectedWorkflowContext.purchaseSummary}</p>
                </header>
                <div className="erp-production-next-action-fields">
                  <label className="erp-field">
                    <span>Nhà cung cấp</span>
                    <select
                      className="erp-input"
                      value={purchaseSupplierId}
                      onChange={(event) => setPurchaseSupplierId(event.currentTarget.value)}
                    >
                      {purchaseSupplierOptions.map((supplier) => (
                        <option key={supplier.value} value={supplier.value}>
                          {supplier.label}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="erp-field">
                    <span>Kho nhận</span>
                    <select
                      className="erp-input"
                      value={purchaseWarehouseId}
                      onChange={(event) => setPurchaseWarehouseId(event.currentTarget.value)}
                    >
                      {purchaseWarehouseOptions.map((warehouse) => (
                        <option key={warehouse.value} value={warehouse.value}>
                          {warehouse.label}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="erp-field">
                    <span>Ngày dự kiến</span>
                    <input
                      className="erp-input"
                      type="date"
                      value={purchaseExpectedDate}
                      onChange={(event) => setPurchaseExpectedDate(event.currentTarget.value)}
                    />
                  </label>
                </div>
                <footer className="erp-production-next-action-footer">
                  <button
                    className="erp-button erp-button--primary"
                    type="button"
                    onClick={createPurchaseOrderFromSelectedPlan}
                    disabled={nextActionSaving !== "" || selectedPlanPurchaseLineCount === 0}
                  >
                    {nextActionSaving === "purchase" ? "Đang tạo PO" : selectedWorkflowContext.purchaseButtonLabel}
                  </button>
                </footer>
              </article>

              <article className="erp-production-next-action-panel">
                <header>
                  <span className="erp-production-step-label">Bước 4</span>
                  <h3>{selectedWorkflowContext.subcontractTitle}</h3>
                  <p>{selectedWorkflowContext.subcontractSummary}</p>
                </header>
                <div className="erp-production-next-action-fields">
                  <label className="erp-field">
                    <span>Nhà máy</span>
                    <select
                      className="erp-input"
                      value={subcontractFactoryId}
                      onChange={(event) => setSubcontractFactoryId(event.currentTarget.value)}
                    >
                      {subcontractFactoryOptions.map((factory) => (
                        <option key={factory.id} value={factory.id}>
                          {factory.name}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="erp-field">
                    <span>Ngày nhận dự kiến</span>
                    <input
                      className="erp-input"
                      type="date"
                      value={subcontractExpectedDate}
                      onChange={(event) => setSubcontractExpectedDate(event.currentTarget.value)}
                    />
                  </label>
                </div>
                <footer className="erp-production-next-action-footer">
                  <button
                    className="erp-button erp-button--primary"
                    type="button"
                    onClick={createSubcontractOrderFromSelectedPlan}
                    disabled={nextActionSaving !== "" || selectedPlanShortageLineCount > 0}
                  >
                    {nextActionSaving === "subcontract" ? "Đang tạo lệnh" : selectedWorkflowContext.subcontractButtonLabel}
                  </button>
                </footer>
              </article>
            </div>
            {nextActionError ? <p className="erp-form-error">{nextActionError}</p> : null}
          </FormSection>
        ) : null}
      </section>

      <ToastStack messages={toast} onDismiss={dismissToast} />
    </main>
  );
}

function formatDate(value?: string) {
  if (!value) {
    return "";
  }
  return new Intl.DateTimeFormat("vi-VN", { day: "2-digit", month: "2-digit" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Request failed";
}

function integerText(value: string) {
  return value.replace(/\D/g, "");
}

function defaultDateOffset(days: number) {
  const date = new Date();
  date.setDate(date.getDate() + days);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function newDraftLineID() {
  return `production-plan-line-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function toProductionPlanInput(line: ProductionPlanDraftLine) {
  return {
    outputItemId: line.outputItemId,
    formulaId: line.formulaId,
    plannedQty: line.plannedQty,
    uomCode: line.uomCode,
    plannedStartDate: line.plannedStartDate,
    plannedEndDate: line.plannedEndDate
  };
}
