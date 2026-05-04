"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  DataTable,
  DecimalInput,
  EmptyState,
  ErrorState,
  FormSection,
  StatusChip,
  ToastStack,
  type DataTableColumn,
  type ToastMessage
} from "@/shared/design-system/components";
import { decimalScales } from "@/shared/format/numberFormat";
import { getFormulas } from "@/modules/masterdata/services/formulaMasterDataService";
import { finishedProductTypes, getProducts } from "@/modules/masterdata/services/productMasterDataService";
import type { FormulaMasterDataItem, ProductMasterDataItem } from "@/modules/masterdata/types";
import {
  createProductionPlan,
  formatProductionPlanQuantity,
  getProductionPlans,
  productionPlanStatusDisplay,
  productionPlanStatusTone,
  summarizeProductionPlans
} from "../services/productionPlanService";
import { defaultProductionPlanUom, findFormulaForProduct, formulaBelongsToProduct } from "../services/productionPlanFormDefaults";
import type { ProductionPlan, ProductionPlanInput, ProductionPlanLine } from "../types";

const emptyInput: ProductionPlanInput = {
  outputItemId: "",
  formulaId: "",
  plannedQty: "1.000000",
  uomCode: "PCS",
  plannedStartDate: "",
  plannedEndDate: ""
};

export function ProductionPlanPrototype() {
  const [plans, setPlans] = useState<ProductionPlan[]>([]);
  const [products, setProducts] = useState<ProductMasterDataItem[]>([]);
  const [formulas, setFormulas] = useState<FormulaMasterDataItem[]>([]);
  const [form, setForm] = useState<ProductionPlanInput>(emptyInput);
  const [selectedPlanId, setSelectedPlanId] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [formError, setFormError] = useState<string | undefined>();
  const [toast, setToast] = useState<ToastMessage[]>([]);

  const activeFinishedProducts = useMemo(
    () =>
      products.filter(
        (product) => product.status === "active" && finishedProductTypes.includes(product.itemType) && product.isProducible
      ),
    [products]
  );
  const activeFormulas = useMemo(() => formulas.filter((formula) => formula.status === "active"), [formulas]);
  const selectedProduct = useMemo(
    () => activeFinishedProducts.find((product) => product.id === form.outputItemId),
    [activeFinishedProducts, form.outputItemId]
  );
  const formulasForSelectedProduct = useMemo(
    () => activeFormulas.filter((formula) => formulaBelongsToProduct(formula, selectedProduct)),
    [activeFormulas, selectedProduct]
  );
  const summary = useMemo(() => summarizeProductionPlans(plans), [plans]);
  const selectedPlan = useMemo(
    () => plans.find((plan) => plan.id === selectedPlanId) ?? plans[0],
    [plans, selectedPlanId]
  );

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
        const firstProduct = productRows.find(
          (product) => product.status === "active" && finishedProductTypes.includes(product.itemType) && product.isProducible
        );
        if (firstProduct) {
          const firstFormula = findFormulaForProduct(
            formulaRows.filter((formula) => formula.status === "active"),
            firstProduct
          );
          setForm((current) => ({
            ...current,
            outputItemId: current.outputItemId || firstProduct.id,
            formulaId: current.formulaId || firstFormula?.id || "",
            uomCode: current.outputItemId ? current.uomCode : defaultProductionPlanUom(firstProduct, firstFormula)
          }));
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

  const planColumns: DataTableColumn<ProductionPlan>[] = [
    {
      key: "plan",
      header: "Kế hoạch",
      render: (plan) => (
        <div className="erp-masterdata-product-cell">
          <strong>{plan.planNo}</strong>
          <small>{formatDate(plan.createdAt)}</small>
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
      render: (plan) => (
        <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => setSelectedPlanId(plan.id)}>
          Chi tiết
        </button>
      ),
      width: "110px"
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
      const created = await createProductionPlan(form);
      await refreshPlans();
      setSelectedPlanId(created.id);
      pushToast("Đã tạo kế hoạch", `${created.planNo} - ${created.outputSku}`, "success");
    } catch (saveError) {
      setFormError(errorText(saveError));
    } finally {
      setSaving(false);
    }
  }

  function changeProduct(productId: string) {
    const product = activeFinishedProducts.find((candidate) => candidate.id === productId);
    const formula = findFormulaForProduct(activeFormulas, product);
    setForm((current) => ({
      ...current,
      outputItemId: productId,
      formulaId: formula?.id ?? "",
      uomCode: defaultProductionPlanUom(product, formula) || current.uomCode
    }));
  }

  function changeFormula(formulaId: string) {
    const formula = activeFormulas.find((candidate) => candidate.id === formulaId);
    setForm((current) => ({
      ...current,
      formulaId,
      uomCode: defaultProductionPlanUom(selectedProduct, formula ?? formulasForSelectedProduct[0]) || current.uomCode
    }));
  }

  function pushToast(title: string, description: string, tone: ToastMessage["tone"]) {
    setToast([{ id: `${Date.now()}`, title, description, tone }]);
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

      <section className="erp-masterdata-workspace">
        <article className="erp-masterdata-list-card">
          <DataTable
            columns={planColumns}
            rows={plans}
            getRowKey={(plan) => plan.id}
            loading={loading}
            pagination
            preserveColumnWidths
            emptyState={<EmptyState title="Chưa có kế hoạch sản xuất" />}
          />
        </article>

        <form onSubmit={submit}>
          <FormSection
            title="Tạo kế hoạch sản xuất"
            description="Chọn thành phẩm và công thức đang active để snapshot nhu cầu vật tư. Kết quả chỉ tạo đề nghị mua nháp nội bộ."
            footer={
              <>
                {formError ? <span className="erp-form-error">{formError}</span> : null}
                <button className="erp-button erp-button--primary" type="submit" disabled={saving || loading}>
                  {saving ? "Đang tạo" : "Tạo kế hoạch"}
                </button>
              </>
            }
          >
            <div className="erp-masterdata-form-grid">
              <label className="erp-field">
                <span>Thành phẩm</span>
                <select className="erp-input" value={form.outputItemId} onChange={(event) => changeProduct(event.currentTarget.value)}>
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
                  value={form.formulaId ?? ""}
                  onChange={(event) => changeFormula(event.currentTarget.value)}
                >
                  <option value="">Tự chọn công thức active</option>
                  {formulasForSelectedProduct.map((formula) => (
                    <option key={formula.id} value={formula.id}>
                      {formula.formulaCode} - {formula.formulaVersion}
                    </option>
                  ))}
                </select>
              </label>
              <DecimalInput
                label="Số lượng cần sản xuất"
                value={form.plannedQty}
                scale={decimalScales.quantity}
                suffix={form.uomCode}
                onChange={(value) => setForm((current) => ({ ...current, plannedQty: value }))}
              />
              <label className="erp-field">
                <span>Đơn vị</span>
                <input
                  className="erp-input"
                  value={form.uomCode}
                  onChange={(event) => setForm((current) => ({ ...current, uomCode: event.currentTarget.value.toUpperCase() }))}
                />
              </label>
              <label className="erp-field">
                <span>Ngày bắt đầu</span>
                <input
                  className="erp-input"
                  type="date"
                  value={form.plannedStartDate ?? ""}
                  onChange={(event) => setForm((current) => ({ ...current, plannedStartDate: event.currentTarget.value }))}
                />
              </label>
              <label className="erp-field">
                <span>Ngày kết thúc</span>
                <input
                  className="erp-input"
                  type="date"
                  value={form.plannedEndDate ?? ""}
                  onChange={(event) => setForm((current) => ({ ...current, plannedEndDate: event.currentTarget.value }))}
                />
              </label>
            </div>
          </FormSection>
        </form>

        <article className="erp-masterdata-list-card">
          <header className="erp-section-header">
            <div>
              <h2 className="erp-section-title">Nhu cầu vật tư</h2>
              <p className="erp-page-description">
                {selectedPlan
                  ? `${selectedPlan.planNo} - ${selectedPlan.outputSku} - ${formatProductionPlanQuantity(
                      selectedPlan.plannedQty,
                      selectedPlan.uomCode
                    )}`
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
      </section>

      <ToastStack messages={toast} />
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
