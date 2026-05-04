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
import { t } from "@/shared/i18n";
import {
  activateFormula,
  calculateFormulaRequirement,
  createFormula,
  emptyFormulaInput,
  formulaInputForParentItem,
  formatFormulaQuantity,
  formulaComponentTypeOptions,
  formulaStatusOptions,
  getFormulas,
  summarizeFormulas
} from "../services/formulaMasterDataService";
import { productUomOptions } from "../services/productMasterDataService";
import type {
  FormulaComponentType,
  FormulaLineInput,
  FormulaMasterDataInput,
  FormulaMasterDataItem,
  FormulaMasterDataQuery,
  FormulaRequirementPreview,
  FormulaStatus,
  ProductMasterDataItem
} from "../types";

const allStatusOptions = [{ label: formulaCopy("formula.filters.allStatuses"), value: "" }, ...formulaStatusOptions] as const;

type FormulaMasterDataPrototypeProps = {
  parentItems?: ProductMasterDataItem[];
  selectedParentItemId?: string;
};

export function FormulaMasterDataPrototype({ parentItems = [], selectedParentItemId = "" }: FormulaMasterDataPrototypeProps) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<FormulaStatus | "">("");
  const [items, setItems] = useState<FormulaMasterDataItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [form, setForm] = useState<FormulaMasterDataInput>(freshFormulaInput());
  const [formError, setFormError] = useState<string | undefined>();
  const [preview, setPreview] = useState<FormulaRequirementPreview | null>(null);
  const [toast, setToast] = useState<ToastMessage[]>([]);
  const parentOptions = useMemo(
    () => parentItems.filter((item) => item.status === "active" && (item.itemType === "finished_good" || item.itemType === "semi_finished")),
    [parentItems]
  );
  const selectedParent = useMemo(
    () => parentOptions.find((item) => item.id === selectedParentItemId) ?? parentOptions.find((item) => item.id === form.finishedItemId),
    [form.finishedItemId, parentOptions, selectedParentItemId]
  );
  const query = useMemo<FormulaMasterDataQuery>(
    () => ({
      search: search || undefined,
      status,
      finishedItemId: selectedParentItemId || undefined
    }),
    [search, selectedParentItemId, status]
  );
  const summary = useMemo(() => summarizeFormulas(items), [items]);

  useEffect(() => {
    if (!selectedParentItemId) {
      return;
    }
    const parent = parentOptions.find((item) => item.id === selectedParentItemId);
    if (!parent) {
      return;
    }
    setForm((current) => formulaInputForParentItem(current, parent));
  }, [parentOptions, selectedParentItemId]);

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    getFormulas(query)
      .then((data) => {
        if (active) {
          setItems(data);
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
  }, [query]);

  const columns: DataTableColumn<FormulaMasterDataItem>[] = [
    {
      key: "formula",
      header: formulaCopy("formula.columns.formula"),
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.formulaCode}</strong>
          <small>{row.formulaVersion}</small>
        </div>
      ),
      width: "180px"
    },
    {
      key: "parent",
      header: formulaCopy("formula.columns.parent"),
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.finishedSku}</strong>
          <small>{row.finishedItemName}</small>
        </div>
      ),
      width: "260px"
    },
    {
      key: "batch",
      header: formulaCopy("formula.columns.batch"),
      render: (row) => formatFormulaQuantity(row.batchQty, row.batchUomCode),
      width: "120px"
    },
    {
      key: "lines",
      header: formulaCopy("formula.columns.lines"),
      render: (row) => row.lines.length,
      width: "80px"
    },
    {
      key: "status",
      header: formulaCopy("formula.columns.status"),
      render: (row) => <StatusChip tone={formulaStatusTone(row.status)}>{formulaStatusDisplay(row.status)}</StatusChip>,
      width: "120px"
    },
    {
      key: "updated",
      header: formulaCopy("formula.columns.updated"),
      render: (row) => formatDate(row.updatedAt),
      width: "120px"
    },
    {
      key: "actions",
      header: "",
      align: "right",
      sticky: true,
      render: (row) => (
        <div className="erp-masterdata-row-actions">
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => calculateBatchRequirements(row)}>
            {formulaCopy("formula.actions.calculate")}
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving || row.status === "active"}
            onClick={() => activate(row)}
          >
            {formulaCopy("formula.actions.activate")}
          </button>
        </div>
      ),
      width: "230px"
    }
  ];

  async function refresh() {
    setLoading(true);
    setError(undefined);
    try {
      setItems(await getFormulas(query));
    } catch (loadError) {
      setError(errorText(loadError));
    } finally {
      setLoading(false);
    }
  }

  async function submitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setFormError(undefined);
    try {
      const created = await createFormula(form);
      await refresh();
      setForm(freshFormulaInput(selectedParent));
      pushToast(formulaCopy("formula.toast.created"), created.formulaCode, "success");
    } catch (saveError) {
      setFormError(errorText(saveError));
    } finally {
      setSaving(false);
    }
  }

  async function activate(item: FormulaMasterDataItem) {
    setSaving(true);
    try {
      const active = await activateFormula(item.id);
      await refresh();
      pushToast(formulaCopy("formula.toast.activated"), active.formulaCode, "success");
    } catch (activationError) {
      pushToast(formulaCopy("formula.errors.requestFailed"), errorText(activationError), "danger");
    } finally {
      setSaving(false);
    }
  }

  async function calculateBatchRequirements(item: FormulaMasterDataItem) {
    try {
      const result = await calculateFormulaRequirement(item.id, {
        plannedQty: item.batchQty,
        plannedUomCode: item.batchUomCode
      });
      setPreview(result);
      pushToast(formulaCopy("formula.toast.calculated"), item.formulaCode, "info");
    } catch (previewError) {
      pushToast(formulaCopy("formula.errors.requestFailed"), errorText(previewError), "danger");
    }
  }

  function resetForm() {
    setForm(freshFormulaInput(selectedParent));
    setFormError(undefined);
  }

  function pushToast(title: string, description: string, tone: ToastMessage["tone"]) {
    setToast([{ id: `${Date.now()}`, title, description, tone }]);
  }

  return (
    <>
      <section className="erp-masterdata-toolbar" aria-label={formulaCopy("formula.filters.label")}>
        <label className="erp-field">
          <span>{formulaCopy("formula.filters.search")}</span>
          <input
            className="erp-input"
            type="search"
            value={search}
            placeholder="XFF-150ML"
            onChange={(event) => setSearch(event.target.value.toUpperCase())}
          />
        </label>
        <label className="erp-field">
          <span>{formulaCopy("formula.filters.status")}</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as FormulaStatus | "")}>
            {allStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? formulaStatusDisplay(option.value) : option.label}
              </option>
            ))}
          </select>
        </label>
        <button className="erp-button erp-button--secondary" type="button" onClick={resetForm}>
          {formulaCopy("formula.actions.newFormula")}
        </button>
      </section>

      <section className="erp-kpi-grid erp-masterdata-kpis">
        <FormulaKPI label={formulaCopy("formula.kpi.total")} value={summary.total} tone="normal" />
        <FormulaKPI label={formulaStatusDisplay("active")} value={summary.active} tone="success" />
        <FormulaKPI label={formulaStatusDisplay("draft")} value={summary.draft} tone="info" />
        <FormulaKPI label={formulaCopy("formula.kpi.lines")} value={summary.lines} tone="warning" />
      </section>

      <section className="erp-masterdata-workspace">
        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{formulaCopy("formula.list.title")}</h2>
            <StatusChip tone={items.length === 0 ? "warning" : "info"}>{formulaCopy("formula.list.rows", { count: items.length })}</StatusChip>
          </div>
          <DataTable
            columns={columns}
            rows={items}
            getRowKey={(row) => row.id}
            loading={loading}
            pagination
            preserveColumnWidths
            error={
              error ? (
                <ErrorState
                  title={formulaCopy("formula.errors.loadTitle")}
                  description={error}
                  action={
                    <button className="erp-button erp-button--secondary" type="button" onClick={() => setError(undefined)}>
                      {formulaCopy("actions.dismiss")}
                    </button>
                  }
                />
              ) : undefined
            }
            emptyState={<EmptyState title={formulaCopy("formula.empty.title")} description={formulaCopy("formula.empty.description")} />}
          />
        </section>

        <form onSubmit={submitForm}>
          <FormSection
            title={formulaCopy("formula.form.createTitle")}
            description={formulaCopy("formula.form.description")}
            footer={
              <>
                <button className="erp-button erp-button--secondary" type="button" onClick={resetForm}>
                  {formulaCopy("actions.clear")}
                </button>
                <button className="erp-button erp-button--primary" type="submit" disabled={saving || !form.finishedItemId}>
                  {saving ? formulaCopy("formula.actions.saving") : formulaCopy("formula.actions.create")}
                </button>
              </>
            }
          >
            {formError ? <p className="erp-form-error">{formError}</p> : null}
            <div className="erp-masterdata-form-grid">
              <TextField label={formulaCopy("formula.form.formulaCode")} value={form.formulaCode} onChange={(value) => updateForm({ formulaCode: value.toUpperCase() })} />
              <TextField label={formulaCopy("formula.form.formulaVersion")} value={form.formulaVersion} onChange={(value) => updateForm({ formulaVersion: value })} />
              <label className="erp-field">
                <span>{formulaCopy("formula.form.parentItem")}</span>
                <select
                  className="erp-input"
                  value={form.finishedItemId}
                  disabled={parentOptions.length === 0 || Boolean(selectedParentItemId)}
                  onChange={(event) => {
                    const parent = parentOptions.find((item) => item.id === event.target.value);
                    if (parent) {
                      updateForm(formulaInputForParentItem(form, parent));
                    }
                  }}
                >
                  <option value="">{formulaCopy("formula.form.selectParentItem")}</option>
                  {parentOptions.map((option) => (
                    <option key={option.id} value={option.id}>
                      {option.skuCode} - {option.name}
                    </option>
                  ))}
                </select>
              </label>
              <DecimalInput label={formulaCopy("formula.form.batchQty")} value={form.batchQty} onChange={(value) => updateForm({ batchQty: value })} />
              <UomField label={formulaCopy("formula.form.uom")} value={form.batchUomCode} onChange={(value) => updateForm({ batchUomCode: value, baseBatchUomCode: value })} />
              <DecimalInput label={formulaCopy("formula.form.baseBatchQty")} value={form.baseBatchQty} onChange={(value) => updateForm({ baseBatchQty: value })} />
              <TextField label={formulaCopy("formula.form.note")} value={form.note} onChange={(value) => updateForm({ note: value })} />
            </div>

            <section className="erp-masterdata-form-lines" aria-label={formulaCopy("formula.form.lineTitle")}>
              <div className="erp-section-header">
                <h3 className="erp-section-title">{formulaCopy("formula.form.lineTitle")}</h3>
                <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={addLine}>
                  {formulaCopy("formula.form.addLine")}
                </button>
              </div>
              {form.lines.map((line, index) => (
                <div className="erp-masterdata-form-line" key={index}>
                  <TextField label={formulaCopy("formula.form.componentSku")} value={line.componentSku} onChange={(value) => updateLine(index, { componentSku: value.toUpperCase() })} />
                  <TextField label={formulaCopy("formula.form.componentName")} value={line.componentName} onChange={(value) => updateLine(index, { componentName: value })} />
                  <label className="erp-field">
                    <span>{formulaCopy("formula.form.componentType")}</span>
                    <select className="erp-input" value={line.componentType} onChange={(event) => updateLine(index, { componentType: event.target.value as FormulaComponentType })}>
                      {formulaComponentTypeOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {formulaComponentTypeDisplay(option.value)}
                        </option>
                      ))}
                    </select>
                  </label>
                  <DecimalInput label={formulaCopy("formula.form.enteredQty")} value={line.enteredQty} suffix={line.enteredUomCode} onChange={(value) => updateLine(index, { enteredQty: value })} />
                  <UomField label={formulaCopy("formula.form.uom")} value={line.enteredUomCode} onChange={(value) => updateLine(index, { enteredUomCode: value })} />
                  <DecimalInput label={formulaCopy("formula.form.calcQty")} value={line.calcQty} suffix={line.calcUomCode} onChange={(value) => updateLine(index, { calcQty: value })} />
                  <UomField label={formulaCopy("formula.form.uom")} value={line.calcUomCode} onChange={(value) => updateLine(index, { calcUomCode: value })} />
                  <DecimalInput label={formulaCopy("formula.form.stockBaseQty")} value={line.stockBaseQty} suffix={line.stockBaseUomCode} onChange={(value) => updateLine(index, { stockBaseQty: value })} />
                  <UomField label={formulaCopy("formula.form.uom")} value={line.stockBaseUomCode} onChange={(value) => updateLine(index, { stockBaseUomCode: value })} />
                  <DecimalInput label={formulaCopy("formula.form.wastePercent")} scale={decimalScales.rate} value={line.wastePercent} suffix="%" onChange={(value) => updateLine(index, { wastePercent: value })} />
                  <ToggleField label={formulaCopy("formula.form.required")} checked={line.isRequired} onChange={(value) => updateLine(index, { isRequired: value })} />
                  <ToggleField label={formulaCopy("formula.form.stockManaged")} checked={line.isStockManaged} onChange={(value) => updateLine(index, { isStockManaged: value })} />
                  <button className="erp-button erp-button--secondary erp-button--compact" type="button" disabled={form.lines.length === 1} onClick={() => removeLine(index)}>
                    {formulaCopy("formula.form.removeLine")}
                  </button>
                </div>
              ))}
            </section>
          </FormSection>
        </form>
      </section>

      <section className="erp-card erp-card--padded">
        <div className="erp-section-header">
          <h2 className="erp-section-title">{formulaCopy("formula.preview.title")}</h2>
          {preview ? (
            <StatusChip tone="info">
              {formulaCopy("formula.preview.planned", {
                quantity: `${formatFormulaQuantity(preview.plannedQty, preview.plannedUomCode)}`
              })}
            </StatusChip>
          ) : null}
        </div>
        {preview ? (
          <DataTable
            columns={previewColumns}
            rows={preview.requirements}
            getRowKey={(row) => row.formulaLineId}
            preserveColumnWidths
          />
        ) : (
          <EmptyState title={formulaCopy("formula.preview.empty")} />
        )}
      </section>
      <ToastStack messages={toast} />
    </>
  );

  function updateForm(patch: Partial<FormulaMasterDataInput>) {
    setForm((current) => ({ ...current, ...patch }));
  }

  function updateLine(index: number, patch: Partial<FormulaLineInput>) {
    setForm((current) => ({
      ...current,
      lines: current.lines.map((line, lineIndex) => (lineIndex === index ? { ...line, ...patch } : line))
    }));
  }

  function addLine() {
    setForm((current) => ({
      ...current,
      lines: [...current.lines, { ...emptyFormulaInput.lines[0], lineNo: current.lines.length + 1 }]
    }));
  }

  function removeLine(index: number) {
    setForm((current) => ({
      ...current,
      lines: current.lines
        .filter((_, lineIndex) => lineIndex !== index)
        .map((line, lineIndex) => ({ ...line, lineNo: lineIndex + 1 }))
    }));
  }
}

const previewColumns: DataTableColumn<FormulaRequirementPreview["requirements"][number]>[] = [
  {
    key: "component",
    header: formulaCopy("formula.form.componentSku"),
    render: (row) => (
      <div className="erp-masterdata-product-cell">
        <strong>{row.componentSku}</strong>
        <small>{row.componentName}</small>
      </div>
    ),
    width: "280px"
  },
  {
    key: "calc",
    header: formulaCopy("formula.form.calcQty"),
    render: (row) => formatFormulaQuantity(row.requiredCalcQty, row.calcUomCode),
    width: "140px"
  },
  {
    key: "stock",
    header: formulaCopy("formula.form.stockBaseQty"),
    render: (row) => formatFormulaQuantity(row.requiredStockBaseQty, row.stockBaseUomCode),
    width: "160px"
  },
  {
    key: "type",
    header: formulaCopy("formula.form.componentType"),
    render: (row) => formulaComponentTypeDisplay(row.componentType),
    width: "150px"
  }
];

function FormulaKPI({ label, value, tone }: { label: string; value: number; tone: "normal" | "success" | "warning" | "info" }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function TextField({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="erp-field">
      <span>{label}</span>
      <input className="erp-input" value={value} onChange={(event) => onChange(event.target.value)} />
    </label>
  );
}

function UomField({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  const normalizedValue = value.trim().toUpperCase();
  const options = productUomOptions.some((option) => option.value === normalizedValue)
    ? productUomOptions
    : [...productUomOptions, { value: normalizedValue }];

  return (
    <label className="erp-field">
      <span>{label}</span>
      <select className="erp-input" value={normalizedValue} onChange={(event) => onChange(event.target.value)}>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.value}
          </option>
        ))}
      </select>
    </label>
  );
}

function ToggleField({ label, checked, onChange }: { label: string; checked: boolean; onChange: (value: boolean) => void }) {
  return (
    <label className="erp-masterdata-toggle">
      <input type="checkbox" checked={checked} onChange={(event) => onChange(event.target.checked)} />
      <span>{label}</span>
    </label>
  );
}

function formulaStatusDisplay(status: FormulaStatus) {
  return formulaCopy(`formula.status.${status}`);
}

function formulaComponentTypeDisplay(type: FormulaComponentType) {
  return formulaCopy(`formula.componentType.${type}`);
}

function formulaStatusTone(status: FormulaStatus): "normal" | "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "active":
      return "success";
    case "draft":
      return "info";
    case "inactive":
      return "warning";
    case "archived":
      return "danger";
    default:
      return "normal";
  }
}

function freshFormulaInput(parent?: ProductMasterDataItem): FormulaMasterDataInput {
  const input = {
    ...emptyFormulaInput,
    lines: emptyFormulaInput.lines.map((line) => ({ ...line }))
  };
  return parent ? formulaInputForParentItem(input, parent) : input;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("vi-VN", { day: "2-digit", month: "2-digit" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : formulaCopy("formula.errors.requestFailed");
}

function formulaCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.${key}`, { values, fallback });
}
