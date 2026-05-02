"use client";

import { useMemo, useState, type FormEvent } from "react";
import {
  DataTable,
  DecimalInput,
  DetailDrawer,
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
import { useProductMasterData } from "../hooks/useProductMasterData";
import {
  emptyProductInput,
  productStatusOptions,
  productStatusTone,
  productTypeOptions,
  toProductInput
} from "../services/productMasterDataService";
import type { ProductMasterDataInput, ProductMasterDataItem, ProductMasterDataQuery, ProductStatus, ProductType } from "../types";

const allStatusOptions = [{ label: productCopy("filters.allStatuses"), value: "" }, ...productStatusOptions] as const;
const allTypeOptions = [{ label: productCopy("filters.allItemTypes"), value: "" }, ...productTypeOptions] as const;

export function ProductMasterDataPrototype({ embedded = false }: { embedded?: boolean }) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<ProductStatus | "">("");
  const [itemType, setItemType] = useState<ProductMasterDataQuery["itemType"]>("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<ProductMasterDataInput>(emptyProductInput);
  const [formError, setFormError] = useState<string | undefined>();
  const [toast, setToast] = useState<ToastMessage[]>([]);
  const query = useMemo<ProductMasterDataQuery>(
    () => ({
      search: search || undefined,
      status,
      itemType
    }),
    [search, status, itemType]
  );
  const {
    items,
    selectedItem,
    loading,
    saving,
    error,
    summary,
    clearError,
    clearSelectedItem,
    loadProductDetail,
    saveNewProduct,
    saveProduct,
    saveProductStatus
  } = useProductMasterData(query);

  const columns: DataTableColumn<ProductMasterDataItem>[] = [
    {
      key: "sku",
      header: productCopy("columns.sku"),
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.skuCode}</strong>
          <small>{row.itemCode}</small>
        </div>
      ),
      width: "180px"
    },
    {
      key: "name",
      header: productCopy("columns.item"),
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.name}</strong>
          <small>{[productTypeDisplay(row.itemType), row.itemGroup].filter(Boolean).join(" / ")}</small>
        </div>
      ),
      width: "260px"
    },
    {
      key: "uom",
      header: productCopy("columns.uom"),
      render: (row) => row.uomBase,
      width: "80px"
    },
    {
      key: "controls",
      header: productCopy("columns.controls"),
      render: (row) => controlLabels(row).join(", "),
      width: "190px"
    },
    {
      key: "status",
      header: productCopy("columns.status"),
      render: (row) => <StatusChip tone={productStatusTone(row.status)}>{productStatusDisplay(row.status)}</StatusChip>,
      width: "120px"
    },
    {
      key: "updated",
      header: productCopy("columns.updated"),
      render: (row) => formatDate(row.updatedAt),
      width: "130px"
    },
    {
      key: "actions",
      header: "",
      align: "right",
      sticky: true,
      render: (row) => (
        <div className="erp-masterdata-row-actions">
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => openDetail(row.id)}>
            {productCopy("actions.detail")}
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startEdit(row)}>
            {productCopy("actions.edit")}
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving}
            onClick={() => toggleStatus(row)}
          >
            {row.status === "active" ? productCopy("actions.inactivate") : productCopy("actions.activate")}
          </button>
        </div>
      ),
      width: "250px"
    }
  ];

  async function openDetail(productId: string) {
    try {
      await loadProductDetail(productId);
    } catch {
      pushToast(productCopy("toast.detailFailed"), productCopy("toast.detailFailedDescription"), "danger");
    }
  }

  function startEdit(item: ProductMasterDataItem) {
    setEditingId(item.id);
    setForm(toProductInput(item));
    setFormError(undefined);
  }

  function resetForm() {
    setEditingId(null);
    setForm(emptyProductInput);
    setFormError(undefined);
  }

  async function toggleStatus(item: ProductMasterDataItem) {
    const nextStatus: ProductStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveProductStatus(item.id, nextStatus);
      pushToast(productCopy("toast.statusChanged"), productCopy("toast.statusChangedDescription", { sku: result.skuCode, status: productStatusDisplay(result.status) }), "success");
    } catch (statusError) {
      pushToast(productCopy("toast.statusFailed"), errorText(statusError), "danger");
    }
  }

  async function submitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(undefined);
    try {
      const result = editingId ? await saveProduct(editingId, form) : await saveNewProduct(form);
      pushToast(editingId ? productCopy("toast.itemUpdated") : productCopy("toast.itemCreated"), productCopy("toast.saved", { sku: result.skuCode }), "success");
      resetForm();
    } catch (saveError) {
      setFormError(errorText(saveError));
    }
  }

  function pushToast(title: string, description: string, tone: ToastMessage["tone"]) {
    setToast([{ id: `${Date.now()}`, title, description, tone }]);
  }

  const content = (
    <>
      <section className="erp-masterdata-toolbar" aria-label={productCopy("filters.label")}>
        <label className="erp-field">
          <span>{productCopy("filters.search")}</span>
          <input
            className="erp-input"
            type="search"
            value={search}
            placeholder="SERUM-30ML"
            onChange={(event) => setSearch(event.target.value.toUpperCase())}
          />
        </label>
        <label className="erp-field">
          <span>{productCopy("filters.status")}</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as ProductStatus | "")}>
            {allStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? productStatusDisplay(option.value) : option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{productCopy("filters.itemType")}</span>
          <select
            className="erp-input"
            value={itemType}
            onChange={(event) => setItemType(event.target.value as ProductMasterDataQuery["itemType"])}
          >
            {allTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? productTypeDisplay(option.value) : option.label}
              </option>
            ))}
          </select>
        </label>
        <button className="erp-button erp-button--secondary" type="button" onClick={resetForm}>
          {productCopy("actions.newItem")}
        </button>
      </section>

      <section className="erp-kpi-grid erp-masterdata-kpis">
        <MasterDataKPI label={productCopy("kpi.total")} value={summary.total} tone="normal" />
        <MasterDataKPI label={productStatusDisplay("active")} value={summary.active} tone="success" />
        <MasterDataKPI label={productStatusDisplay("draft")} value={summary.draft} tone="info" />
        <MasterDataKPI label={productCopy("kpi.controlled")} value={summary.controlled} tone="warning" />
      </section>

      <section className="erp-masterdata-workspace">
        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{productCopy("list.title")}</h2>
            <StatusChip tone={items.length === 0 ? "warning" : "info"}>{productCopy("list.rows", { count: items.length })}</StatusChip>
          </div>
          <DataTable
            columns={columns}
            rows={items}
            getRowKey={(row) => row.id}
            loading={loading}
            error={
              error ? (
                <ErrorState
                  title={productCopy("errors.loadTitle")}
                  description={error}
                  action={
                    <button className="erp-button erp-button--secondary" type="button" onClick={clearError}>
                      {productCopy("actions.dismiss")}
                    </button>
                  }
                />
              ) : undefined
            }
            emptyState={<EmptyState title={productCopy("empty.title")} description={productCopy("empty.description")} />}
          />
        </section>

        <form onSubmit={submitForm}>
          <FormSection
            title={editingId ? productCopy("form.updateTitle") : productCopy("form.createTitle")}
            description={productCopy("form.description")}
            footer={
              <>
                <button className="erp-button erp-button--secondary" type="button" onClick={resetForm}>
                  {productCopy("actions.clear")}
                </button>
                <button className="erp-button erp-button--primary" type="submit" disabled={saving}>
                  {saving ? productCopy("actions.saving") : editingId ? productCopy("actions.update") : productCopy("actions.create")}
                </button>
              </>
            }
          >
            {formError ? <p className="erp-form-error">{formError}</p> : null}
            <div className="erp-masterdata-form-grid">
              <TextField label={productCopy("form.itemCode")} value={form.itemCode} onChange={(value) => updateForm({ itemCode: value.toUpperCase() })} />
              <TextField label={productCopy("form.skuCode")} value={form.skuCode} onChange={(value) => updateForm({ skuCode: value.toUpperCase() })} />
              <TextField label={productCopy("form.name")} value={form.name} onChange={(value) => updateForm({ name: value })} />
              <label className="erp-field">
                <span>{productCopy("form.itemType")}</span>
                <select className="erp-input" value={form.itemType} onChange={(event) => updateForm({ itemType: event.target.value as ProductMasterDataInput["itemType"] })}>
                  {productTypeOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {productTypeDisplay(option.value)}
                    </option>
                  ))}
                </select>
              </label>
              <TextField label={productCopy("form.group")} value={form.itemGroup} onChange={(value) => updateForm({ itemGroup: value })} />
              <TextField label={productCopy("form.brand")} value={form.brandCode} onChange={(value) => updateForm({ brandCode: value.toUpperCase() })} />
              <TextField label={productCopy("form.baseUom")} value={form.uomBase} onChange={(value) => updateForm({ uomBase: value.toUpperCase() })} />
              <TextField label={productCopy("form.purchaseUom")} value={form.uomPurchase} onChange={(value) => updateForm({ uomPurchase: value.toUpperCase() })} />
              <TextField label={productCopy("form.issueUom")} value={form.uomIssue} onChange={(value) => updateForm({ uomIssue: value.toUpperCase() })} />
              <NumberField label={productCopy("form.shelfLifeDays")} value={form.shelfLifeDays} onChange={(value) => updateForm({ shelfLifeDays: value })} />
              <DecimalInput label={productCopy("form.standardCost")} scale={decimalScales.unitCost} suffix="VND" value={form.standardCost} onChange={(value) => updateForm({ standardCost: value })} />
              <TextField label={productCopy("form.specVersion")} value={form.specVersion} onChange={(value) => updateForm({ specVersion: value })} />
            </div>
            <div className="erp-masterdata-toggle-grid">
              <ToggleField label={productCopy("controls.lot")} checked={form.lotControlled} onChange={(value) => updateForm({ lotControlled: value })} />
              <ToggleField label={productCopy("controls.expiry")} checked={form.expiryControlled} onChange={(value) => updateForm({ expiryControlled: value })} />
              <ToggleField label={productCopy("controls.qc")} checked={form.qcRequired} onChange={(value) => updateForm({ qcRequired: value })} />
              <ToggleField label={productCopy("controls.sellable")} checked={form.isSellable} onChange={(value) => updateForm({ isSellable: value })} />
              <ToggleField label={productCopy("controls.purchasable")} checked={form.isPurchasable} onChange={(value) => updateForm({ isPurchasable: value })} />
              <ToggleField label={productCopy("controls.producible")} checked={form.isProducible} onChange={(value) => updateForm({ isProducible: value })} />
            </div>
            <label className="erp-field">
              <span>{productCopy("form.status")}</span>
              <select className="erp-input" value={form.status} onChange={(event) => updateForm({ status: event.target.value as ProductStatus })}>
                {productStatusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {productStatusDisplay(option.value)}
                  </option>
                ))}
              </select>
            </label>
            <p className="erp-masterdata-audit-hint">{productCopy("form.auditHint")}</p>
          </FormSection>
        </form>
      </section>

      <DetailDrawer
        open={Boolean(selectedItem)}
        title={selectedItem?.skuCode ?? productCopy("detail.title")}
        subtitle={selectedItem?.name}
        onClose={clearSelectedItem}
        footer={
          selectedItem ? (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => startEdit(selectedItem)}>
              {productCopy("actions.edit")}
            </button>
          ) : null
        }
      >
        {selectedItem ? <ProductDetail item={selectedItem} /> : null}
      </DetailDrawer>
      <ToastStack messages={toast} />
    </>
  );

  if (embedded) {
    return content;
  }

  return (
    <section className="erp-module-page erp-masterdata-page">
      <header className="erp-page-header">
        <div>
          <p className="erp-module-eyebrow">MD</p>
          <h1 className="erp-page-title">{productCopy("page.title")}</h1>
          <p className="erp-page-description">{productCopy("page.description")}</p>
        </div>
        <StatusChip tone="info">{productCopy("page.skuCount", { count: summary.total })}</StatusChip>
      </header>
      {content}
    </section>
  );

  function updateForm(patch: Partial<ProductMasterDataInput>) {
    setForm((current) => ({ ...current, ...patch }));
  }
}

function ProductDetail({ item }: { item: ProductMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label={productCopy("detail.itemCode")} value={item.itemCode} />
      <MasterDataFact label={productCopy("detail.sku")} value={item.skuCode} />
      <MasterDataFact label={productCopy("detail.type")} value={productTypeDisplay(item.itemType)} />
      <MasterDataFact label={productCopy("detail.status")} value={productStatusDisplay(item.status)} />
      <MasterDataFact label={productCopy("detail.uom")} value={item.uomBase} />
      <MasterDataFact label={productCopy("detail.shelfLife")} value={productCopy("detail.days", { count: item.shelfLifeDays ?? 0 })} />
      <MasterDataFact label={productCopy("detail.spec")} value={item.specVersion || "-"} />
      <MasterDataFact label={productCopy("detail.audit")} value={item.auditLogId || productCopy("detail.auditFallback")} />
    </div>
  );
}

function MasterDataKPI({ label, value, tone }: { label: string; value: number; tone: "normal" | "success" | "warning" | "info" }) {
  return (
    <article className="erp-card erp-card--padded erp-kpi-card">
      <div className="erp-kpi-label">{label}</div>
      <strong className="erp-kpi-value">{value}</strong>
      <StatusChip tone={tone}>{label}</StatusChip>
    </article>
  );
}

function MasterDataFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="erp-masterdata-fact">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
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

function NumberField({ label, value, onChange }: { label: string; value: number; onChange: (value: number) => void }) {
  return (
    <label className="erp-field">
      <span>{label}</span>
      <input
        className="erp-input"
        type="number"
        min="0"
        value={value}
        onChange={(event) => onChange(Number(event.target.value))}
      />
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

function controlLabels(item: ProductMasterDataItem) {
  return [
    item.lotControlled ? productCopy("controls.lot") : undefined,
    item.expiryControlled ? productCopy("controls.expiry") : undefined,
    item.qcRequired ? productCopy("controls.qc") : undefined
  ].filter((label): label is string => Boolean(label));
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("vi-VN", { day: "2-digit", month: "2-digit" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : productCopy("errors.requestFailed");
}

function productStatusDisplay(status: ProductStatus) {
  return productCopy(`product.status.${status}`);
}

function productTypeDisplay(type: ProductType) {
  return productCopy(`product.type.${type}`);
}

function productCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.${key}`, { values, fallback });
}
