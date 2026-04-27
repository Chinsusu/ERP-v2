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
import { usePartyMasterData } from "../hooks/usePartyMasterData";
import {
  customerStatusLabel,
  customerStatusOptions,
  customerTypeLabel,
  customerTypeOptions,
  emptyCustomerInput,
  emptySupplierInput,
  partyStatusTone,
  supplierGroupLabel,
  supplierGroupOptions,
  supplierStatusLabel,
  supplierStatusOptions,
  toCustomerInput,
  toSupplierInput
} from "../services/partyMasterDataService";
import type {
  CustomerMasterDataInput,
  CustomerMasterDataItem,
  CustomerMasterDataQuery,
  CustomerStatus,
  SupplierMasterDataInput,
  SupplierMasterDataItem,
  SupplierMasterDataQuery,
  SupplierStatus
} from "../types";

const allSupplierStatusOptions = [{ label: "All supplier statuses", value: "" }, ...supplierStatusOptions] as const;
const allSupplierGroupOptions = [{ label: "All supplier groups", value: "" }, ...supplierGroupOptions] as const;
const allCustomerStatusOptions = [{ label: "All customer statuses", value: "" }, ...customerStatusOptions] as const;
const allCustomerTypeOptions = [{ label: "All customer types", value: "" }, ...customerTypeOptions] as const;

export function SupplierCustomerMasterDataPrototype({ embedded = false }: { embedded?: boolean }) {
  const [search, setSearch] = useState("");
  const [supplierStatus, setSupplierStatus] = useState<SupplierStatus | "">("");
  const [supplierGroup, setSupplierGroup] = useState<SupplierMasterDataQuery["supplierGroup"]>("");
  const [customerStatus, setCustomerStatus] = useState<CustomerStatus | "">("");
  const [customerType, setCustomerType] = useState<CustomerMasterDataQuery["customerType"]>("");
  const [editingSupplierId, setEditingSupplierId] = useState<string | null>(null);
  const [editingCustomerId, setEditingCustomerId] = useState<string | null>(null);
  const [supplierForm, setSupplierForm] = useState<SupplierMasterDataInput>(emptySupplierInput);
  const [customerForm, setCustomerForm] = useState<CustomerMasterDataInput>(emptyCustomerInput);
  const [formError, setFormError] = useState<string | undefined>();
  const [toast, setToast] = useState<ToastMessage[]>([]);

  const supplierQuery = useMemo<SupplierMasterDataQuery>(
    () => ({
      search: search || undefined,
      status: supplierStatus,
      supplierGroup
    }),
    [search, supplierStatus, supplierGroup]
  );
  const customerQuery = useMemo<CustomerMasterDataQuery>(
    () => ({
      search: search || undefined,
      status: customerStatus,
      customerType
    }),
    [search, customerStatus, customerType]
  );
  const {
    suppliers,
    customers,
    selectedSupplier,
    selectedCustomer,
    loading,
    saving,
    error,
    summary,
    clearError,
    clearSelectedSupplier,
    clearSelectedCustomer,
    loadSupplierDetail,
    loadCustomerDetail,
    saveNewSupplier,
    saveSupplier,
    saveSupplierStatus,
    saveNewCustomer,
    saveCustomer,
    saveCustomerStatus
  } = usePartyMasterData(supplierQuery, customerQuery);

  const supplierColumns: DataTableColumn<SupplierMasterDataItem>[] = [
    {
      key: "supplier",
      header: "Supplier",
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.supplierCode}</strong>
          <small>{row.supplierName}</small>
        </div>
      ),
      width: "240px"
    },
    { key: "group", header: "Group", render: (row) => supplierGroupLabel(row.supplierGroup), width: "130px" },
    { key: "terms", header: "Terms", render: (row) => row.paymentTerms || "-", width: "90px" },
    { key: "lead", header: "Lead", render: (row) => `${row.leadTimeDays ?? 0}d`, width: "70px" },
    {
      key: "score",
      header: "Score",
      render: (row) => `${row.qualityScore ?? 0}/${row.deliveryScore ?? 0}`,
      width: "90px"
    },
    {
      key: "status",
      header: "Status",
      render: (row) => <StatusChip tone={partyStatusTone(row.status)}>{supplierStatusLabel(row.status)}</StatusChip>,
      width: "115px"
    },
    {
      key: "actions",
      header: "",
      align: "right",
      sticky: true,
      render: (row) => (
        <div className="erp-masterdata-row-actions">
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => openSupplierDetail(row.id)}>
            Detail
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startSupplierEdit(row)}>
            Edit
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving || row.status === "blacklisted"}
            onClick={() => toggleSupplierStatus(row)}
          >
            {row.status === "active" ? "Inactivate" : "Activate"}
          </button>
        </div>
      )
    }
  ];

  const customerColumns: DataTableColumn<CustomerMasterDataItem>[] = [
    {
      key: "customer",
      header: "Customer",
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.customerCode}</strong>
          <small>{row.customerName}</small>
        </div>
      ),
      width: "240px"
    },
    { key: "type", header: "Type", render: (row) => customerTypeLabel(row.customerType), width: "135px" },
    { key: "channel", header: "Channel", render: (row) => row.channelCode || "-", width: "100px" },
    { key: "price", header: "Price list", render: (row) => row.priceListCode || "-", width: "130px" },
    { key: "credit", header: "Credit", render: (row) => money(row.creditLimit ?? 0), width: "110px" },
    {
      key: "status",
      header: "Status",
      render: (row) => <StatusChip tone={partyStatusTone(row.status)}>{customerStatusLabel(row.status)}</StatusChip>,
      width: "115px"
    },
    {
      key: "actions",
      header: "",
      align: "right",
      sticky: true,
      render: (row) => (
        <div className="erp-masterdata-row-actions">
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => openCustomerDetail(row.id)}>
            Detail
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startCustomerEdit(row)}>
            Edit
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving || row.status === "blocked"}
            onClick={() => toggleCustomerStatus(row)}
          >
            {row.status === "active" ? "Inactivate" : "Activate"}
          </button>
        </div>
      )
    }
  ];

  const content = (
    <>
      <section className="erp-masterdata-toolbar erp-masterdata-toolbar--wide" aria-label="Party master data filters">
        <label className="erp-field">
          <span>Search</span>
          <input className="erp-input" type="search" value={search} placeholder="SUP-RM-BIO" onChange={(event) => setSearch(event.target.value.toUpperCase())} />
        </label>
        <label className="erp-field">
          <span>Supplier status</span>
          <select className="erp-input" value={supplierStatus} onChange={(event) => setSupplierStatus(event.target.value as SupplierStatus | "")}>
            {allSupplierStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Supplier group</span>
          <select className="erp-input" value={supplierGroup} onChange={(event) => setSupplierGroup(event.target.value as SupplierMasterDataQuery["supplierGroup"])}>
            {allSupplierGroupOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Customer status</span>
          <select className="erp-input" value={customerStatus} onChange={(event) => setCustomerStatus(event.target.value as CustomerStatus | "")}>
            {allCustomerStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Customer type</span>
          <select className="erp-input" value={customerType} onChange={(event) => setCustomerType(event.target.value as CustomerMasterDataQuery["customerType"])}>
            {allCustomerTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-masterdata-kpis">
        <MasterDataKPI label="Suppliers" value={summary.suppliers} tone="normal" />
        <MasterDataKPI label="Active suppliers" value={summary.activeSuppliers} tone="success" />
        <MasterDataKPI label="Customers" value={summary.customers} tone="info" />
        <MasterDataKPI label="Active customers" value={summary.activeCustomers} tone="success" />
      </section>

      <section className="erp-masterdata-workspace">
        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Supplier master</h2>
            <StatusChip tone={suppliers.length === 0 ? "warning" : "info"}>{suppliers.length} rows</StatusChip>
          </div>
          <DataTable
            columns={supplierColumns}
            rows={suppliers}
            getRowKey={(row) => row.id}
            loading={loading}
            error={tableError(error, clearError)}
            emptyState={<EmptyState title="No suppliers" description="Adjust the filters or create a supplier." />}
          />
        </section>

        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Customer master</h2>
            <StatusChip tone={customers.length === 0 ? "warning" : "info"}>{customers.length} rows</StatusChip>
          </div>
          <DataTable
            columns={customerColumns}
            rows={customers}
            getRowKey={(row) => row.id}
            loading={loading}
            error={tableError(error, clearError)}
            emptyState={<EmptyState title="No customers" description="Adjust the filters or create a customer." />}
          />
        </section>
      </section>

      <section className="erp-masterdata-workspace">
        <SupplierForm
          editingId={editingSupplierId}
          form={supplierForm}
          formError={formError}
          saving={saving}
          onChange={(patch) => setSupplierForm((current) => ({ ...current, ...patch }))}
          onClear={resetSupplierForm}
          onSubmit={submitSupplierForm}
        />
        <CustomerForm
          editingId={editingCustomerId}
          form={customerForm}
          formError={formError}
          saving={saving}
          onChange={(patch) => setCustomerForm((current) => ({ ...current, ...patch }))}
          onClear={resetCustomerForm}
          onSubmit={submitCustomerForm}
        />
      </section>

      <DetailDrawer
        open={Boolean(selectedSupplier)}
        title={selectedSupplier?.supplierCode ?? "Supplier detail"}
        subtitle={selectedSupplier?.supplierName}
        onClose={clearSelectedSupplier}
        footer={
          selectedSupplier ? (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => startSupplierEdit(selectedSupplier)}>
              Edit
            </button>
          ) : null
        }
      >
        {selectedSupplier ? <SupplierDetail item={selectedSupplier} /> : null}
      </DetailDrawer>
      <DetailDrawer
        open={Boolean(selectedCustomer)}
        title={selectedCustomer?.customerCode ?? "Customer detail"}
        subtitle={selectedCustomer?.customerName}
        onClose={clearSelectedCustomer}
        footer={
          selectedCustomer ? (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => startCustomerEdit(selectedCustomer)}>
              Edit
            </button>
          ) : null
        }
      >
        {selectedCustomer ? <CustomerDetail item={selectedCustomer} /> : null}
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
          <h1 className="erp-page-title">Supplier Customer Master Data</h1>
          <p className="erp-page-description">Party catalog for purchasing, sales, and channel setup</p>
        </div>
        <StatusChip tone="info">{summary.activeCustomers} active customers</StatusChip>
      </header>
      {content}
    </section>
  );

  async function openSupplierDetail(supplierId: string) {
    try {
      await loadSupplierDetail(supplierId);
    } catch (detailError) {
      pushToast("Detail failed", errorText(detailError), "danger");
    }
  }

  async function openCustomerDetail(customerId: string) {
    try {
      await loadCustomerDetail(customerId);
    } catch (detailError) {
      pushToast("Detail failed", errorText(detailError), "danger");
    }
  }

  function startSupplierEdit(item: SupplierMasterDataItem) {
    setEditingSupplierId(item.id);
    setSupplierForm(toSupplierInput(item));
    setFormError(undefined);
  }

  function startCustomerEdit(item: CustomerMasterDataItem) {
    setEditingCustomerId(item.id);
    setCustomerForm(toCustomerInput(item));
    setFormError(undefined);
  }

  function resetSupplierForm() {
    setEditingSupplierId(null);
    setSupplierForm(emptySupplierInput);
    setFormError(undefined);
  }

  function resetCustomerForm() {
    setEditingCustomerId(null);
    setCustomerForm(emptyCustomerInput);
    setFormError(undefined);
  }

  async function submitSupplierForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(undefined);
    try {
      const result = editingSupplierId ? await saveSupplier(editingSupplierId, supplierForm) : await saveNewSupplier(supplierForm);
      pushToast(editingSupplierId ? "Supplier updated" : "Supplier created", `${result.supplierCode} saved`, "success");
      resetSupplierForm();
    } catch (saveError) {
      setFormError(errorText(saveError));
    }
  }

  async function submitCustomerForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(undefined);
    try {
      const result = editingCustomerId ? await saveCustomer(editingCustomerId, customerForm) : await saveNewCustomer(customerForm);
      pushToast(editingCustomerId ? "Customer updated" : "Customer created", `${result.customerCode} saved`, "success");
      resetCustomerForm();
    } catch (saveError) {
      setFormError(errorText(saveError));
    }
  }

  async function toggleSupplierStatus(item: SupplierMasterDataItem) {
    const nextStatus: SupplierStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveSupplierStatus(item.id, nextStatus);
      pushToast("Supplier status changed", `${result.supplierCode} is ${supplierStatusLabel(result.status)}`, "success");
    } catch (statusError) {
      pushToast("Status failed", errorText(statusError), "danger");
    }
  }

  async function toggleCustomerStatus(item: CustomerMasterDataItem) {
    const nextStatus: CustomerStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveCustomerStatus(item.id, nextStatus);
      pushToast("Customer status changed", `${result.customerCode} is ${customerStatusLabel(result.status)}`, "success");
    } catch (statusError) {
      pushToast("Status failed", errorText(statusError), "danger");
    }
  }

  function pushToast(title: string, description: string, tone: ToastMessage["tone"]) {
    setToast([{ id: `${Date.now()}`, title, description, tone }]);
  }
}

function SupplierForm({
  editingId,
  form,
  formError,
  saving,
  onChange,
  onClear,
  onSubmit
}: {
  editingId: string | null;
  form: SupplierMasterDataInput;
  formError?: string;
  saving: boolean;
  onChange: (patch: Partial<SupplierMasterDataInput>) => void;
  onClear: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return (
    <form onSubmit={onSubmit}>
      <FormSection
        title={editingId ? "Update supplier" : "Create supplier"}
        description="Supplier identity, tax, contact, terms, and operational score fields"
        footer={
          <>
            <button className="erp-button erp-button--secondary" type="button" onClick={onClear}>
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
          <TextField label="Supplier code" value={form.supplierCode} onChange={(value) => onChange({ supplierCode: value.toUpperCase() })} />
          <TextField label="Supplier name" value={form.supplierName} onChange={(value) => onChange({ supplierName: value })} />
          <label className="erp-field">
            <span>Group</span>
            <select className="erp-input" value={form.supplierGroup} onChange={(event) => onChange({ supplierGroup: event.target.value as SupplierMasterDataInput["supplierGroup"] })}>
              {supplierGroupOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>Status</span>
            <select className="erp-input" value={form.status} onChange={(event) => onChange({ status: event.target.value as SupplierStatus })}>
              {supplierStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <TextField label="Contact" value={form.contactName} onChange={(value) => onChange({ contactName: value })} />
          <TextField label="Phone" value={form.phone} onChange={(value) => onChange({ phone: value })} />
          <TextField label="Email" value={form.email} onChange={(value) => onChange({ email: value })} />
          <TextField label="Tax code" value={form.taxCode} onChange={(value) => onChange({ taxCode: value.toUpperCase() })} />
          <TextField label="Address" value={form.address} onChange={(value) => onChange({ address: value })} />
          <TextField label="Payment terms" value={form.paymentTerms} onChange={(value) => onChange({ paymentTerms: value.toUpperCase() })} />
          <NumberField label="Lead days" value={form.leadTimeDays} onChange={(value) => onChange({ leadTimeDays: value })} />
          <NumberField label="MOQ" value={form.moq} onChange={(value) => onChange({ moq: value })} />
          <NumberField label="Quality score" value={form.qualityScore} onChange={(value) => onChange({ qualityScore: value })} />
          <NumberField label="Delivery score" value={form.deliveryScore} onChange={(value) => onChange({ deliveryScore: value })} />
        </div>
      </FormSection>
    </form>
  );
}

function CustomerForm({
  editingId,
  form,
  formError,
  saving,
  onChange,
  onClear,
  onSubmit
}: {
  editingId: string | null;
  form: CustomerMasterDataInput;
  formError?: string;
  saving: boolean;
  onChange: (patch: Partial<CustomerMasterDataInput>) => void;
  onClear: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return (
    <form onSubmit={onSubmit}>
      <FormSection
        title={editingId ? "Update customer" : "Create customer"}
        description="Customer identity, channel, tax, price list, credit, and payment terms"
        footer={
          <>
            <button className="erp-button erp-button--secondary" type="button" onClick={onClear}>
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
          <TextField label="Customer code" value={form.customerCode} onChange={(value) => onChange({ customerCode: value.toUpperCase() })} />
          <TextField label="Customer name" value={form.customerName} onChange={(value) => onChange({ customerName: value })} />
          <label className="erp-field">
            <span>Type</span>
            <select className="erp-input" value={form.customerType} onChange={(event) => onChange({ customerType: event.target.value as CustomerMasterDataInput["customerType"] })}>
              {customerTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>Status</span>
            <select className="erp-input" value={form.status} onChange={(event) => onChange({ status: event.target.value as CustomerStatus })}>
              {customerStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <TextField label="Channel" value={form.channelCode} onChange={(value) => onChange({ channelCode: value.toUpperCase() })} />
          <TextField label="Price list" value={form.priceListCode} onChange={(value) => onChange({ priceListCode: value.toUpperCase() })} />
          <TextField label="Discount group" value={form.discountGroup} onChange={(value) => onChange({ discountGroup: value })} />
          <NumberField label="Credit limit" value={form.creditLimit} onChange={(value) => onChange({ creditLimit: value })} />
          <TextField label="Payment terms" value={form.paymentTerms} onChange={(value) => onChange({ paymentTerms: value.toUpperCase() })} />
          <TextField label="Contact" value={form.contactName} onChange={(value) => onChange({ contactName: value })} />
          <TextField label="Phone" value={form.phone} onChange={(value) => onChange({ phone: value })} />
          <TextField label="Email" value={form.email} onChange={(value) => onChange({ email: value })} />
          <TextField label="Tax code" value={form.taxCode} onChange={(value) => onChange({ taxCode: value.toUpperCase() })} />
          <TextField label="Address" value={form.address} onChange={(value) => onChange({ address: value })} />
        </div>
      </FormSection>
    </form>
  );
}

function SupplierDetail({ item }: { item: SupplierMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label="Code" value={item.supplierCode} />
      <MasterDataFact label="Group" value={supplierGroupLabel(item.supplierGroup)} />
      <MasterDataFact label="Status" value={supplierStatusLabel(item.status)} />
      <MasterDataFact label="Contact" value={item.contactName || "-"} />
      <MasterDataFact label="Phone" value={item.phone || "-"} />
      <MasterDataFact label="Email" value={item.email || "-"} />
      <MasterDataFact label="Tax" value={item.taxCode || "-"} />
      <MasterDataFact label="Terms" value={item.paymentTerms || "-"} />
      <MasterDataFact label="Lead" value={`${item.leadTimeDays ?? 0} days`} />
      <MasterDataFact label="Scores" value={`${item.qualityScore ?? 0}/${item.deliveryScore ?? 0}`} />
      <MasterDataFact label="Updated" value={formatDate(item.updatedAt)} />
      <MasterDataFact label="Audit" value={item.auditLogId || "Tracked on write"} />
    </div>
  );
}

function CustomerDetail({ item }: { item: CustomerMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label="Code" value={item.customerCode} />
      <MasterDataFact label="Type" value={customerTypeLabel(item.customerType)} />
      <MasterDataFact label="Status" value={customerStatusLabel(item.status)} />
      <MasterDataFact label="Channel" value={item.channelCode || "-"} />
      <MasterDataFact label="Price list" value={item.priceListCode || "-"} />
      <MasterDataFact label="Credit" value={money(item.creditLimit ?? 0)} />
      <MasterDataFact label="Contact" value={item.contactName || "-"} />
      <MasterDataFact label="Phone" value={item.phone || "-"} />
      <MasterDataFact label="Email" value={item.email || "-"} />
      <MasterDataFact label="Tax" value={item.taxCode || "-"} />
      <MasterDataFact label="Updated" value={formatDate(item.updatedAt)} />
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
      <input className="erp-input" min="0" type="number" value={value} onChange={(event) => onChange(Number(event.target.value))} />
    </label>
  );
}

function tableError(error: string | undefined, clearError: () => void) {
  return error ? (
    <ErrorState
      title="Party master data could not load"
      description={error}
      action={
        <button className="erp-button erp-button--secondary" type="button" onClick={clearError}>
          Dismiss
        </button>
      }
    />
  ) : undefined;
}

function money(value: number) {
  return new Intl.NumberFormat("en-US", { notation: "compact", maximumFractionDigits: 1 }).format(value);
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-US", { month: "short", day: "2-digit" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Party master data request failed";
}
