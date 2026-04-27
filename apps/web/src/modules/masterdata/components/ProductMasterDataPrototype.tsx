"use client";

import { useMemo, useState, type FormEvent } from "react";
import {
  DataTable,
  DetailDrawer,
  EmptyState,
  ErrorState,
  FormSection,
  StatusChip,
  ToastStack,
  type DataTableColumn,
  type ToastMessage
} from "@/shared/design-system/components";
import { useProductMasterData } from "../hooks/useProductMasterData";
import {
  emptyProductInput,
  productStatusOptions,
  productStatusTone,
  productTypeLabel,
  productTypeOptions,
  statusLabel,
  toProductInput
} from "../services/productMasterDataService";
import type { ProductMasterDataInput, ProductMasterDataItem, ProductMasterDataQuery, ProductStatus } from "../types";

const allStatusOptions = [{ label: "All statuses", value: "" }, ...productStatusOptions] as const;
const allTypeOptions = [{ label: "All item types", value: "" }, ...productTypeOptions] as const;

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
      header: "SKU",
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
      header: "Item",
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.name}</strong>
          <small>{[productTypeLabel(row.itemType), row.itemGroup].filter(Boolean).join(" / ")}</small>
        </div>
      ),
      width: "260px"
    },
    {
      key: "uom",
      header: "UOM",
      render: (row) => row.uomBase,
      width: "80px"
    },
    {
      key: "controls",
      header: "Controls",
      render: (row) => controlLabels(row).join(", "),
      width: "190px"
    },
    {
      key: "status",
      header: "Status",
      render: (row) => <StatusChip tone={productStatusTone(row.status)}>{statusLabel(row.status)}</StatusChip>,
      width: "120px"
    },
    {
      key: "updated",
      header: "Updated",
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
            Detail
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startEdit(row)}>
            Edit
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving}
            onClick={() => toggleStatus(row)}
          >
            {row.status === "active" ? "Inactivate" : "Activate"}
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
      pushToast("Detail failed", "Could not load product detail", "danger");
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
      pushToast("Status changed", `${result.skuCode} is ${statusLabel(result.status)}`, "success");
    } catch (statusError) {
      pushToast("Status failed", errorText(statusError), "danger");
    }
  }

  async function submitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(undefined);
    try {
      const result = editingId ? await saveProduct(editingId, form) : await saveNewProduct(form);
      pushToast(editingId ? "Item updated" : "Item created", `${result.skuCode} saved`, "success");
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
      <section className="erp-masterdata-toolbar" aria-label="Master data filters">
        <label className="erp-field">
          <span>Search</span>
          <input
            className="erp-input"
            type="search"
            value={search}
            placeholder="SERUM-30ML"
            onChange={(event) => setSearch(event.target.value.toUpperCase())}
          />
        </label>
        <label className="erp-field">
          <span>Status</span>
          <select className="erp-input" value={status} onChange={(event) => setStatus(event.target.value as ProductStatus | "")}>
            {allStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Item type</span>
          <select
            className="erp-input"
            value={itemType}
            onChange={(event) => setItemType(event.target.value as ProductMasterDataQuery["itemType"])}
          >
            {allTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <button className="erp-button erp-button--secondary" type="button" onClick={resetForm}>
          New item
        </button>
      </section>

      <section className="erp-kpi-grid erp-masterdata-kpis">
        <MasterDataKPI label="Total" value={summary.total} tone="normal" />
        <MasterDataKPI label="Active" value={summary.active} tone="success" />
        <MasterDataKPI label="Draft" value={summary.draft} tone="info" />
        <MasterDataKPI label="Controlled" value={summary.controlled} tone="warning" />
      </section>

      <section className="erp-masterdata-workspace">
        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Item & SKU master</h2>
            <StatusChip tone={items.length === 0 ? "warning" : "info"}>{items.length} rows</StatusChip>
          </div>
          <DataTable
            columns={columns}
            rows={items}
            getRowKey={(row) => row.id}
            loading={loading}
            error={
              error ? (
                <ErrorState
                  title="Master data could not load"
                  description={error}
                  action={
                    <button className="erp-button erp-button--secondary" type="button" onClick={clearError}>
                      Dismiss
                    </button>
                  }
                />
              ) : undefined
            }
            emptyState={<EmptyState title="No product master data" description="Adjust the filters or create a SKU." />}
          />
        </section>

        <form onSubmit={submitForm}>
          <FormSection
            title={editingId ? "Update SKU" : "Create SKU"}
            description="Required cosmetic master data, lifecycle, and control flags"
            footer={
              <>
                <button className="erp-button erp-button--secondary" type="button" onClick={resetForm}>
                  Clear
                </button>
                <button className="erp-button erp-button--primary" type="submit" disabled={saving}>
                  {saving ? "Saving" : editingId ? "Update" : "Create"}
                </button>
              </>
            }
          >
            {formError ? <p className="erp-form-error">{formError}</p> : null}
            <div className="erp-masterdata-form-grid">
              <TextField label="Item code" value={form.itemCode} onChange={(value) => updateForm({ itemCode: value.toUpperCase() })} />
              <TextField label="SKU code" value={form.skuCode} onChange={(value) => updateForm({ skuCode: value.toUpperCase() })} />
              <TextField label="Name" value={form.name} onChange={(value) => updateForm({ name: value })} />
              <label className="erp-field">
                <span>Item type</span>
                <select className="erp-input" value={form.itemType} onChange={(event) => updateForm({ itemType: event.target.value as ProductMasterDataInput["itemType"] })}>
                  {productTypeOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>
              <TextField label="Group" value={form.itemGroup} onChange={(value) => updateForm({ itemGroup: value })} />
              <TextField label="Brand" value={form.brandCode} onChange={(value) => updateForm({ brandCode: value.toUpperCase() })} />
              <TextField label="Base UOM" value={form.uomBase} onChange={(value) => updateForm({ uomBase: value.toUpperCase() })} />
              <TextField label="Purchase UOM" value={form.uomPurchase} onChange={(value) => updateForm({ uomPurchase: value.toUpperCase() })} />
              <TextField label="Issue UOM" value={form.uomIssue} onChange={(value) => updateForm({ uomIssue: value.toUpperCase() })} />
              <NumberField label="Shelf life days" value={form.shelfLifeDays} onChange={(value) => updateForm({ shelfLifeDays: value })} />
              <NumberField label="Standard cost" value={form.standardCost} onChange={(value) => updateForm({ standardCost: value })} />
              <TextField label="Spec version" value={form.specVersion} onChange={(value) => updateForm({ specVersion: value })} />
            </div>
            <div className="erp-masterdata-toggle-grid">
              <ToggleField label="Lot" checked={form.lotControlled} onChange={(value) => updateForm({ lotControlled: value })} />
              <ToggleField label="Expiry" checked={form.expiryControlled} onChange={(value) => updateForm({ expiryControlled: value })} />
              <ToggleField label="QC" checked={form.qcRequired} onChange={(value) => updateForm({ qcRequired: value })} />
              <ToggleField label="Sellable" checked={form.isSellable} onChange={(value) => updateForm({ isSellable: value })} />
              <ToggleField label="Purchasable" checked={form.isPurchasable} onChange={(value) => updateForm({ isPurchasable: value })} />
              <ToggleField label="Producible" checked={form.isProducible} onChange={(value) => updateForm({ isProducible: value })} />
            </div>
            <label className="erp-field">
              <span>Status</span>
              <select className="erp-input" value={form.status} onChange={(event) => updateForm({ status: event.target.value as ProductStatus })}>
                {productStatusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <p className="erp-masterdata-audit-hint">Audit metadata records create, update, and status changes.</p>
          </FormSection>
        </form>
      </section>

      <DetailDrawer
        open={Boolean(selectedItem)}
        title={selectedItem?.skuCode ?? "SKU detail"}
        subtitle={selectedItem?.name}
        onClose={clearSelectedItem}
        footer={
          selectedItem ? (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => startEdit(selectedItem)}>
              Edit
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
          <h1 className="erp-page-title">Master Data</h1>
          <p className="erp-page-description">Item and SKU catalog for cosmetic operations</p>
        </div>
        <StatusChip tone="info">{summary.total} SKUs</StatusChip>
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
      <MasterDataFact label="Item code" value={item.itemCode} />
      <MasterDataFact label="SKU" value={item.skuCode} />
      <MasterDataFact label="Type" value={productTypeLabel(item.itemType)} />
      <MasterDataFact label="Status" value={statusLabel(item.status)} />
      <MasterDataFact label="UOM" value={item.uomBase} />
      <MasterDataFact label="Shelf life" value={`${item.shelfLifeDays ?? 0} days`} />
      <MasterDataFact label="Spec" value={item.specVersion || "-"} />
      <MasterDataFact label="Audit" value={item.auditLogId || "Tracked on write"} />
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
    item.lotControlled ? "Lot" : undefined,
    item.expiryControlled ? "Expiry" : undefined,
    item.qcRequired ? "QC" : undefined
  ].filter((label): label is string => Boolean(label));
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-US", { month: "short", day: "2-digit" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Product master data request failed";
}
