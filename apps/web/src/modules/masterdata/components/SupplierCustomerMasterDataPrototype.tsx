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
import { decimalScales, formatMoney, formatQuantity, formatRate } from "@/shared/format/numberFormat";
import { t } from "@/shared/i18n";
import { usePartyMasterData } from "../hooks/usePartyMasterData";
import {
  customerStatusOptions,
  customerTypeOptions,
  emptyCustomerInput,
  emptySupplierInput,
  partyStatusTone,
  supplierGroupOptions,
  supplierStatusOptions,
  toCustomerInput,
  toSupplierInput
} from "../services/partyMasterDataService";
import type {
  CustomerMasterDataInput,
  CustomerMasterDataItem,
  CustomerMasterDataQuery,
  CustomerStatus,
  CustomerType,
  SupplierMasterDataInput,
  SupplierMasterDataItem,
  SupplierMasterDataQuery,
  SupplierGroup,
  SupplierStatus
} from "../types";

const allSupplierStatusOptions = [{ label: partyCopy("filters.allSupplierStatuses"), value: "" }, ...supplierStatusOptions] as const;
const allSupplierGroupOptions = [{ label: partyCopy("filters.allSupplierGroups"), value: "" }, ...supplierGroupOptions] as const;
const allCustomerStatusOptions = [{ label: partyCopy("filters.allCustomerStatuses"), value: "" }, ...customerStatusOptions] as const;
const allCustomerTypeOptions = [{ label: partyCopy("filters.allCustomerTypes"), value: "" }, ...customerTypeOptions] as const;

type PartyMasterDataMode = "all" | "suppliers" | "customers";

export function SupplierCustomerMasterDataPrototype({ embedded = false, mode = "all" }: { embedded?: boolean; mode?: PartyMasterDataMode }) {
  const showSuppliers = mode !== "customers";
  const showCustomers = mode !== "suppliers";
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
      header: partyCopy("supplier.columns.supplier"),
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.supplierCode}</strong>
          <small>{row.supplierName}</small>
        </div>
      ),
      width: "240px"
    },
    { key: "group", header: partyCopy("supplier.columns.group"), render: (row) => supplierGroupDisplay(row.supplierGroup), width: "130px" },
    { key: "terms", header: partyCopy("supplier.columns.terms"), render: (row) => row.paymentTerms || "-", width: "90px" },
    { key: "lead", header: partyCopy("supplier.columns.lead"), render: (row) => partyCopy("supplier.detail.shortDays", { count: row.leadTimeDays ?? 0 }), width: "70px" },
    {
      key: "score",
      header: partyCopy("supplier.columns.score"),
      render: (row) => `${formatRate(row.qualityScore ?? "0.0000")}/${formatRate(row.deliveryScore ?? "0.0000")}`,
      width: "90px"
    },
    {
      key: "status",
      header: partyCopy("columns.status"),
      render: (row) => <StatusChip tone={partyStatusTone(row.status)}>{supplierStatusDisplay(row.status)}</StatusChip>,
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
            {commonAction("detail")}
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startSupplierEdit(row)}>
            {commonAction("edit")}
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving || row.status === "blacklisted"}
            onClick={() => toggleSupplierStatus(row)}
          >
            {row.status === "active" ? commonAction("inactivate") : commonAction("activate")}
          </button>
        </div>
      )
    }
  ];

  const customerColumns: DataTableColumn<CustomerMasterDataItem>[] = [
    {
      key: "customer",
      header: partyCopy("customer.columns.customer"),
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.customerCode}</strong>
          <small>{row.customerName}</small>
        </div>
      ),
      width: "240px"
    },
    { key: "type", header: partyCopy("customer.columns.type"), render: (row) => customerTypeDisplay(row.customerType), width: "135px" },
    { key: "channel", header: partyCopy("customer.columns.channel"), render: (row) => row.channelCode || "-", width: "100px" },
    { key: "price", header: partyCopy("customer.columns.priceList"), render: (row) => row.priceListCode || "-", width: "130px" },
    { key: "credit", header: partyCopy("customer.columns.credit"), render: (row) => formatMoney(row.creditLimit ?? "0.00"), width: "110px" },
    {
      key: "status",
      header: partyCopy("columns.status"),
      render: (row) => <StatusChip tone={partyStatusTone(row.status)}>{customerStatusDisplay(row.status)}</StatusChip>,
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
            {commonAction("detail")}
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startCustomerEdit(row)}>
            {commonAction("edit")}
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving || row.status === "blocked"}
            onClick={() => toggleCustomerStatus(row)}
          >
            {row.status === "active" ? commonAction("inactivate") : commonAction("activate")}
          </button>
        </div>
      )
    }
  ];

  const toolbarClassName = showSuppliers && showCustomers ? "erp-masterdata-toolbar erp-masterdata-toolbar--wide" : "erp-masterdata-toolbar";
  const kpiClassName = showSuppliers && showCustomers ? "erp-kpi-grid erp-masterdata-kpis" : "erp-kpi-grid erp-masterdata-kpis erp-masterdata-kpis--split";
  const workspaceClassName = showSuppliers && showCustomers ? "erp-masterdata-workspace" : "erp-masterdata-workspace erp-masterdata-workspace--single";
  const searchPlaceholder = showSuppliers && showCustomers ? "SUP-RM-BIO / CUS-DL-MINHANH" : showSuppliers ? "SUP-RM-BIO" : "CUS-DL-MINHANH";

  const content = (
    <>
      <section className={toolbarClassName} aria-label={partyCopy("filters.label")}>
        <label className="erp-field">
          <span>{partyCopy("filters.search")}</span>
          <input className="erp-input" type="search" value={search} placeholder={searchPlaceholder} onChange={(event) => setSearch(event.target.value.toUpperCase())} />
        </label>
        {showSuppliers ? (
          <>
            <label className="erp-field">
              <span>{partyCopy("filters.supplierStatus")}</span>
              <select className="erp-input" value={supplierStatus} onChange={(event) => setSupplierStatus(event.target.value as SupplierStatus | "")}>
                {allSupplierStatusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.value ? supplierStatusDisplay(option.value) : option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{partyCopy("filters.supplierGroup")}</span>
              <select className="erp-input" value={supplierGroup} onChange={(event) => setSupplierGroup(event.target.value as SupplierMasterDataQuery["supplierGroup"])}>
                {allSupplierGroupOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.value ? supplierGroupDisplay(option.value) : option.label}
                  </option>
                ))}
              </select>
            </label>
          </>
        ) : null}
        {showCustomers ? (
          <>
            <label className="erp-field">
              <span>{partyCopy("filters.customerStatus")}</span>
              <select className="erp-input" value={customerStatus} onChange={(event) => setCustomerStatus(event.target.value as CustomerStatus | "")}>
                {allCustomerStatusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.value ? customerStatusDisplay(option.value) : option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="erp-field">
              <span>{partyCopy("filters.customerType")}</span>
              <select className="erp-input" value={customerType} onChange={(event) => setCustomerType(event.target.value as CustomerMasterDataQuery["customerType"])}>
                {allCustomerTypeOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.value ? customerTypeDisplay(option.value) : option.label}
                  </option>
                ))}
              </select>
            </label>
          </>
        ) : null}
      </section>

      <section className={kpiClassName}>
        {showSuppliers ? (
          <>
            <MasterDataKPI label={partyCopy("kpi.suppliers")} value={summary.suppliers} tone="normal" />
            <MasterDataKPI label={partyCopy("kpi.activeSuppliers")} value={summary.activeSuppliers} tone="success" />
          </>
        ) : null}
        {showCustomers ? (
          <>
            <MasterDataKPI label={partyCopy("kpi.customers")} value={summary.customers} tone="info" />
            <MasterDataKPI label={partyCopy("kpi.activeCustomers")} value={summary.activeCustomers} tone="success" />
          </>
        ) : null}
      </section>

      <section className={workspaceClassName}>
        {showSuppliers ? (
          <section className="erp-card erp-card--padded erp-masterdata-list-card">
            <div className="erp-section-header">
              <h2 className="erp-section-title">{partyCopy("supplier.list.title")}</h2>
              <StatusChip tone={suppliers.length === 0 ? "warning" : "info"}>{partyCopy("list.rows", { count: suppliers.length })}</StatusChip>
            </div>
            <DataTable
              columns={supplierColumns}
              rows={suppliers}
              getRowKey={(row) => row.id}
              loading={loading}
              error={tableError(error, clearError)}
              emptyState={<EmptyState title={partyCopy("supplier.empty.title")} description={partyCopy("supplier.empty.description")} />}
            />
          </section>
        ) : null}

        {showCustomers ? (
          <section className="erp-card erp-card--padded erp-masterdata-list-card">
            <div className="erp-section-header">
              <h2 className="erp-section-title">{partyCopy("customer.list.title")}</h2>
              <StatusChip tone={customers.length === 0 ? "warning" : "info"}>{partyCopy("list.rows", { count: customers.length })}</StatusChip>
            </div>
            <DataTable
              columns={customerColumns}
              rows={customers}
              getRowKey={(row) => row.id}
              loading={loading}
              error={tableError(error, clearError)}
              emptyState={<EmptyState title={partyCopy("customer.empty.title")} description={partyCopy("customer.empty.description")} />}
            />
          </section>
        ) : null}
      </section>

      <section className={workspaceClassName}>
        {showSuppliers ? (
          <SupplierForm
            editingId={editingSupplierId}
            form={supplierForm}
            formError={formError}
            saving={saving}
            onChange={(patch) => setSupplierForm((current) => ({ ...current, ...patch }))}
            onClear={resetSupplierForm}
            onSubmit={submitSupplierForm}
          />
        ) : null}
        {showCustomers ? (
          <CustomerForm
            editingId={editingCustomerId}
            form={customerForm}
            formError={formError}
            saving={saving}
            onChange={(patch) => setCustomerForm((current) => ({ ...current, ...patch }))}
            onClear={resetCustomerForm}
            onSubmit={submitCustomerForm}
          />
        ) : null}
      </section>

      {showSuppliers ? (
        <DetailDrawer
          open={Boolean(selectedSupplier)}
          title={selectedSupplier?.supplierCode ?? partyCopy("supplier.detail.title")}
          subtitle={selectedSupplier?.supplierName}
          onClose={clearSelectedSupplier}
          footer={
            selectedSupplier ? (
              <button className="erp-button erp-button--secondary" type="button" onClick={() => startSupplierEdit(selectedSupplier)}>
                {commonAction("edit")}
              </button>
            ) : null
          }
        >
          {selectedSupplier ? <SupplierDetail item={selectedSupplier} /> : null}
        </DetailDrawer>
      ) : null}
      {showCustomers ? (
        <DetailDrawer
          open={Boolean(selectedCustomer)}
          title={selectedCustomer?.customerCode ?? partyCopy("customer.detail.title")}
          subtitle={selectedCustomer?.customerName}
          onClose={clearSelectedCustomer}
          footer={
            selectedCustomer ? (
              <button className="erp-button erp-button--secondary" type="button" onClick={() => startCustomerEdit(selectedCustomer)}>
                {commonAction("edit")}
              </button>
            ) : null
          }
        >
          {selectedCustomer ? <CustomerDetail item={selectedCustomer} /> : null}
        </DetailDrawer>
      ) : null}
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
          <h1 className="erp-page-title">{partyCopy("page.title")}</h1>
          <p className="erp-page-description">{partyCopy("page.description")}</p>
        </div>
        <StatusChip tone="info">{partyCopy("page.activeCustomers", { count: summary.activeCustomers })}</StatusChip>
      </header>
      {content}
    </section>
  );

  async function openSupplierDetail(supplierId: string) {
    try {
      await loadSupplierDetail(supplierId);
    } catch (detailError) {
      pushToast(partyCopy("toast.detailFailed"), errorText(detailError), "danger");
    }
  }

  async function openCustomerDetail(customerId: string) {
    try {
      await loadCustomerDetail(customerId);
    } catch (detailError) {
      pushToast(partyCopy("toast.detailFailed"), errorText(detailError), "danger");
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
      pushToast(editingSupplierId ? partyCopy("supplier.toast.updated") : partyCopy("supplier.toast.created"), partyCopy("toast.saved", { code: result.supplierCode }), "success");
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
      pushToast(editingCustomerId ? partyCopy("customer.toast.updated") : partyCopy("customer.toast.created"), partyCopy("toast.saved", { code: result.customerCode }), "success");
      resetCustomerForm();
    } catch (saveError) {
      setFormError(errorText(saveError));
    }
  }

  async function toggleSupplierStatus(item: SupplierMasterDataItem) {
    const nextStatus: SupplierStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveSupplierStatus(item.id, nextStatus);
      pushToast(partyCopy("supplier.toast.statusChanged"), partyCopy("toast.statusChangedDescription", { code: result.supplierCode, status: supplierStatusDisplay(result.status) }), "success");
    } catch (statusError) {
      pushToast(partyCopy("toast.statusFailed"), errorText(statusError), "danger");
    }
  }

  async function toggleCustomerStatus(item: CustomerMasterDataItem) {
    const nextStatus: CustomerStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveCustomerStatus(item.id, nextStatus);
      pushToast(partyCopy("customer.toast.statusChanged"), partyCopy("toast.statusChangedDescription", { code: result.customerCode, status: customerStatusDisplay(result.status) }), "success");
    } catch (statusError) {
      pushToast(partyCopy("toast.statusFailed"), errorText(statusError), "danger");
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
        title={editingId ? partyCopy("supplier.form.updateTitle") : partyCopy("supplier.form.createTitle")}
        description={partyCopy("supplier.form.description")}
        footer={
          <>
            <button className="erp-button erp-button--secondary" type="button" onClick={onClear}>
              {commonAction("clear")}
            </button>
            <button className="erp-button erp-button--primary" type="submit" disabled={saving}>
              {saving ? commonAction("saving") : editingId ? commonAction("update") : commonAction("create")}
            </button>
          </>
        }
      >
        {formError ? <p className="erp-form-error">{formError}</p> : null}
        <div className="erp-masterdata-form-grid">
          <TextField label={partyCopy("supplier.form.supplierCode")} value={form.supplierCode} onChange={(value) => onChange({ supplierCode: value.toUpperCase() })} />
          <TextField label={partyCopy("supplier.form.supplierName")} value={form.supplierName} onChange={(value) => onChange({ supplierName: value })} />
          <label className="erp-field">
            <span>{partyCopy("supplier.form.group")}</span>
            <select className="erp-input" value={form.supplierGroup} onChange={(event) => onChange({ supplierGroup: event.target.value as SupplierMasterDataInput["supplierGroup"] })}>
              {supplierGroupOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {supplierGroupDisplay(option.value)}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{partyCopy("supplier.form.status")}</span>
            <select className="erp-input" value={form.status} onChange={(event) => onChange({ status: event.target.value as SupplierStatus })}>
              {supplierStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {supplierStatusDisplay(option.value)}
                </option>
              ))}
            </select>
          </label>
          <TextField label={partyCopy("fields.contact")} value={form.contactName} onChange={(value) => onChange({ contactName: value })} />
          <TextField label={partyCopy("fields.phone")} value={form.phone} onChange={(value) => onChange({ phone: value })} />
          <TextField label={partyCopy("fields.email")} value={form.email} onChange={(value) => onChange({ email: value })} />
          <TextField label={partyCopy("fields.taxCode")} value={form.taxCode} onChange={(value) => onChange({ taxCode: value.toUpperCase() })} />
          <TextField label={partyCopy("fields.address")} value={form.address} onChange={(value) => onChange({ address: value })} />
          <TextField label={partyCopy("fields.paymentTerms")} value={form.paymentTerms} onChange={(value) => onChange({ paymentTerms: value.toUpperCase() })} />
          <NumberField label={partyCopy("supplier.form.leadDays")} value={form.leadTimeDays} onChange={(value) => onChange({ leadTimeDays: value })} />
          <DecimalInput label={partyCopy("supplier.form.moq")} scale={decimalScales.quantity} value={form.moq} onChange={(value) => onChange({ moq: value })} />
          <DecimalInput label={partyCopy("supplier.form.qualityScore")} scale={decimalScales.rate} suffix="%" value={form.qualityScore} onChange={(value) => onChange({ qualityScore: value })} />
          <DecimalInput label={partyCopy("supplier.form.deliveryScore")} scale={decimalScales.rate} suffix="%" value={form.deliveryScore} onChange={(value) => onChange({ deliveryScore: value })} />
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
        title={editingId ? partyCopy("customer.form.updateTitle") : partyCopy("customer.form.createTitle")}
        description={partyCopy("customer.form.description")}
        footer={
          <>
            <button className="erp-button erp-button--secondary" type="button" onClick={onClear}>
              {commonAction("clear")}
            </button>
            <button className="erp-button erp-button--primary" type="submit" disabled={saving}>
              {saving ? commonAction("saving") : editingId ? commonAction("update") : commonAction("create")}
            </button>
          </>
        }
      >
        {formError ? <p className="erp-form-error">{formError}</p> : null}
        <div className="erp-masterdata-form-grid">
          <TextField label={partyCopy("customer.form.customerCode")} value={form.customerCode} onChange={(value) => onChange({ customerCode: value.toUpperCase() })} />
          <TextField label={partyCopy("customer.form.customerName")} value={form.customerName} onChange={(value) => onChange({ customerName: value })} />
          <label className="erp-field">
            <span>{partyCopy("customer.form.type")}</span>
            <select className="erp-input" value={form.customerType} onChange={(event) => onChange({ customerType: event.target.value as CustomerMasterDataInput["customerType"] })}>
              {customerTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {customerTypeDisplay(option.value)}
                </option>
              ))}
            </select>
          </label>
          <label className="erp-field">
            <span>{partyCopy("customer.form.status")}</span>
            <select className="erp-input" value={form.status} onChange={(event) => onChange({ status: event.target.value as CustomerStatus })}>
              {customerStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {customerStatusDisplay(option.value)}
                </option>
              ))}
            </select>
          </label>
          <TextField label={partyCopy("customer.form.channel")} value={form.channelCode} onChange={(value) => onChange({ channelCode: value.toUpperCase() })} />
          <TextField label={partyCopy("customer.form.priceList")} value={form.priceListCode} onChange={(value) => onChange({ priceListCode: value.toUpperCase() })} />
          <TextField label={partyCopy("customer.form.discountGroup")} value={form.discountGroup} onChange={(value) => onChange({ discountGroup: value })} />
          <DecimalInput label={partyCopy("customer.form.creditLimit")} scale={decimalScales.money} suffix="VND" value={form.creditLimit} onChange={(value) => onChange({ creditLimit: value })} />
          <TextField label={partyCopy("fields.paymentTerms")} value={form.paymentTerms} onChange={(value) => onChange({ paymentTerms: value.toUpperCase() })} />
          <TextField label={partyCopy("fields.contact")} value={form.contactName} onChange={(value) => onChange({ contactName: value })} />
          <TextField label={partyCopy("fields.phone")} value={form.phone} onChange={(value) => onChange({ phone: value })} />
          <TextField label={partyCopy("fields.email")} value={form.email} onChange={(value) => onChange({ email: value })} />
          <TextField label={partyCopy("fields.taxCode")} value={form.taxCode} onChange={(value) => onChange({ taxCode: value.toUpperCase() })} />
          <TextField label={partyCopy("fields.address")} value={form.address} onChange={(value) => onChange({ address: value })} />
        </div>
      </FormSection>
    </form>
  );
}

function SupplierDetail({ item }: { item: SupplierMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label={partyCopy("detail.code")} value={item.supplierCode} />
      <MasterDataFact label={partyCopy("supplier.detail.group")} value={supplierGroupDisplay(item.supplierGroup)} />
      <MasterDataFact label={partyCopy("detail.status")} value={supplierStatusDisplay(item.status)} />
      <MasterDataFact label={partyCopy("fields.contact")} value={item.contactName || "-"} />
      <MasterDataFact label={partyCopy("fields.phone")} value={item.phone || "-"} />
      <MasterDataFact label={partyCopy("fields.email")} value={item.email || "-"} />
      <MasterDataFact label={partyCopy("fields.tax")} value={item.taxCode || "-"} />
      <MasterDataFact label={partyCopy("fields.terms")} value={item.paymentTerms || "-"} />
      <MasterDataFact label={partyCopy("supplier.detail.lead")} value={partyCopy("supplier.detail.days", { count: item.leadTimeDays ?? 0 })} />
      <MasterDataFact label={partyCopy("supplier.detail.moq")} value={formatQuantity(item.moq ?? "0.000000")} />
      <MasterDataFact label={partyCopy("supplier.detail.scores")} value={`${formatRate(item.qualityScore ?? "0.0000")}/${formatRate(item.deliveryScore ?? "0.0000")}`} />
      <MasterDataFact label={partyCopy("detail.updated")} value={formatDate(item.updatedAt)} />
      <MasterDataFact label={partyCopy("detail.audit")} value={item.auditLogId || partyCopy("detail.auditFallback")} />
    </div>
  );
}

function CustomerDetail({ item }: { item: CustomerMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label={partyCopy("detail.code")} value={item.customerCode} />
      <MasterDataFact label={partyCopy("customer.detail.type")} value={customerTypeDisplay(item.customerType)} />
      <MasterDataFact label={partyCopy("detail.status")} value={customerStatusDisplay(item.status)} />
      <MasterDataFact label={partyCopy("customer.detail.channel")} value={item.channelCode || "-"} />
      <MasterDataFact label={partyCopy("customer.detail.priceList")} value={item.priceListCode || "-"} />
      <MasterDataFact label={partyCopy("customer.detail.credit")} value={formatMoney(item.creditLimit ?? "0.00")} />
      <MasterDataFact label={partyCopy("fields.contact")} value={item.contactName || "-"} />
      <MasterDataFact label={partyCopy("fields.phone")} value={item.phone || "-"} />
      <MasterDataFact label={partyCopy("fields.email")} value={item.email || "-"} />
      <MasterDataFact label={partyCopy("fields.tax")} value={item.taxCode || "-"} />
      <MasterDataFact label={partyCopy("detail.updated")} value={formatDate(item.updatedAt)} />
      <MasterDataFact label={partyCopy("detail.audit")} value={item.auditLogId || partyCopy("detail.auditFallback")} />
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
      title={partyCopy("errors.loadTitle")}
      description={error}
      action={
        <button className="erp-button erp-button--secondary" type="button" onClick={clearError}>
          {commonAction("dismiss")}
        </button>
      }
    />
  ) : undefined;
}


function formatDate(value: string) {
  return new Intl.DateTimeFormat("vi-VN", { day: "2-digit", month: "2-digit" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : partyCopy("errors.requestFailed");
}

function supplierGroupDisplay(group: SupplierGroup) {
  return partyCopy(`supplier.group.${group}`);
}

function supplierStatusDisplay(status: SupplierStatus) {
  return partyCopy(`supplier.status.${status}`);
}

function customerTypeDisplay(type: CustomerType) {
  return partyCopy(`customer.type.${type}`);
}

function customerStatusDisplay(status: CustomerStatus) {
  return partyCopy(`customer.status.${status}`);
}

function commonAction(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.actions.${key}`, { values, fallback });
}

function partyCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.party.${key}`, { values, fallback });
}
