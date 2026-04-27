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
import { useWarehouseMasterData } from "../hooks/useWarehouseMasterData";
import {
  emptyLocationInput,
  emptyWarehouseInput,
  locationStatusLabel,
  locationStatusOptions,
  locationTypeLabel,
  locationTypeOptions,
  toLocationInput,
  toWarehouseInput,
  warehouseStatusLabel,
  warehouseStatusOptions,
  warehouseStatusTone,
  warehouseTypeLabel,
  warehouseTypeOptions
} from "../services/warehouseMasterDataService";
import type {
  WarehouseLocationMasterDataInput,
  WarehouseLocationMasterDataItem,
  WarehouseLocationMasterDataQuery,
  WarehouseLocationStatus,
  WarehouseMasterDataInput,
  WarehouseMasterDataItem,
  WarehouseMasterDataQuery,
  WarehouseStatus
} from "../types";

const allWarehouseStatusOptions = [{ label: "All warehouse statuses", value: "" }, ...warehouseStatusOptions] as const;
const allWarehouseTypeOptions = [{ label: "All warehouse types", value: "" }, ...warehouseTypeOptions] as const;
const allLocationStatusOptions = [{ label: "All location statuses", value: "" }, ...locationStatusOptions] as const;
const allLocationTypeOptions = [{ label: "All location types", value: "" }, ...locationTypeOptions] as const;

export function WarehouseLocationMasterDataPrototype({ embedded = false }: { embedded?: boolean }) {
  const [search, setSearch] = useState("");
  const [warehouseStatus, setWarehouseStatus] = useState<WarehouseStatus | "">("");
  const [warehouseType, setWarehouseType] = useState<WarehouseMasterDataQuery["warehouseType"]>("");
  const [locationStatus, setLocationStatus] = useState<WarehouseLocationStatus | "">("");
  const [locationType, setLocationType] = useState<WarehouseLocationMasterDataQuery["locationType"]>("");
  const [locationWarehouseId, setLocationWarehouseId] = useState("");
  const [editingWarehouseId, setEditingWarehouseId] = useState<string | null>(null);
  const [editingLocationId, setEditingLocationId] = useState<string | null>(null);
  const [warehouseForm, setWarehouseForm] = useState<WarehouseMasterDataInput>(emptyWarehouseInput);
  const [locationForm, setLocationForm] = useState<WarehouseLocationMasterDataInput>(emptyLocationInput);
  const [formError, setFormError] = useState<string | undefined>();
  const [toast, setToast] = useState<ToastMessage[]>([]);

  const warehouseQuery = useMemo<WarehouseMasterDataQuery>(
    () => ({
      search: search || undefined,
      status: warehouseStatus,
      warehouseType
    }),
    [search, warehouseStatus, warehouseType]
  );
  const locationQuery = useMemo<WarehouseLocationMasterDataQuery>(
    () => ({
      search: search || undefined,
      warehouseId: locationWarehouseId || undefined,
      status: locationStatus,
      locationType
    }),
    [search, locationWarehouseId, locationStatus, locationType]
  );
  const {
    warehouses,
    locations,
    selectedWarehouse,
    selectedLocation,
    loading,
    saving,
    error,
    summary,
    clearError,
    clearSelectedWarehouse,
    clearSelectedLocation,
    loadWarehouseDetail,
    loadLocationDetail,
    saveNewWarehouse,
    saveWarehouse,
    saveWarehouseStatus,
    saveNewLocation,
    saveLocation,
    saveLocationStatus
  } = useWarehouseMasterData(warehouseQuery, locationQuery);

  const warehouseColumns: DataTableColumn<WarehouseMasterDataItem>[] = [
    {
      key: "warehouse",
      header: "Warehouse",
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.warehouseCode}</strong>
          <small>{row.warehouseName}</small>
        </div>
      ),
      width: "240px"
    },
    { key: "type", header: "Type", render: (row) => warehouseTypeLabel(row.warehouseType), width: "130px" },
    { key: "site", header: "Site", render: (row) => row.siteCode, width: "80px" },
    {
      key: "flow",
      header: "Flow",
      render: (row) => warehouseFlowLabels(row).join(", ") || "-",
      width: "160px"
    },
    {
      key: "status",
      header: "Status",
      render: (row) => <StatusChip tone={warehouseStatusTone(row.status)}>{warehouseStatusLabel(row.status)}</StatusChip>,
      width: "110px"
    },
    {
      key: "actions",
      header: "",
      align: "right",
      sticky: true,
      render: (row) => (
        <div className="erp-masterdata-row-actions">
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => openWarehouseDetail(row.id)}>
            Detail
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startWarehouseEdit(row)}>
            Edit
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving}
            onClick={() => toggleWarehouseStatus(row)}
          >
            {row.status === "active" ? "Inactivate" : "Activate"}
          </button>
        </div>
      )
    }
  ];

  const locationColumns: DataTableColumn<WarehouseLocationMasterDataItem>[] = [
    {
      key: "location",
      header: "Location",
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.locationCode}</strong>
          <small>{row.locationName}</small>
        </div>
      ),
      width: "220px"
    },
    { key: "warehouse", header: "WH", render: (row) => row.warehouseCode, width: "120px" },
    { key: "type", header: "Type", render: (row) => locationTypeLabel(row.locationType), width: "120px" },
    {
      key: "flow",
      header: "Flow",
      render: (row) => locationFlowLabels(row).join(", ") || "-",
      width: "160px"
    },
    {
      key: "status",
      header: "Status",
      render: (row) => <StatusChip tone={warehouseStatusTone(row.status)}>{locationStatusLabel(row.status)}</StatusChip>,
      width: "110px"
    },
    {
      key: "actions",
      header: "",
      align: "right",
      sticky: true,
      render: (row) => (
        <div className="erp-masterdata-row-actions">
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => openLocationDetail(row.id)}>
            Detail
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startLocationEdit(row)}>
            Edit
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving}
            onClick={() => toggleLocationStatus(row)}
          >
            {row.status === "active" ? "Inactivate" : "Activate"}
          </button>
        </div>
      )
    }
  ];

  const content = (
    <>
      <section className="erp-masterdata-toolbar erp-masterdata-toolbar--wide" aria-label="Warehouse master data filters">
        <label className="erp-field">
          <span>Search</span>
          <input className="erp-input" type="search" value={search} placeholder="WH-HCM-FG" onChange={(event) => setSearch(event.target.value.toUpperCase())} />
        </label>
        <label className="erp-field">
          <span>Warehouse status</span>
          <select className="erp-input" value={warehouseStatus} onChange={(event) => setWarehouseStatus(event.target.value as WarehouseStatus | "")}>
            {allWarehouseStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Warehouse type</span>
          <select className="erp-input" value={warehouseType} onChange={(event) => setWarehouseType(event.target.value as WarehouseMasterDataQuery["warehouseType"])}>
            {allWarehouseTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Location status</span>
          <select className="erp-input" value={locationStatus} onChange={(event) => setLocationStatus(event.target.value as WarehouseLocationStatus | "")}>
            {allLocationStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>Location type</span>
          <select className="erp-input" value={locationType} onChange={(event) => setLocationType(event.target.value as WarehouseLocationMasterDataQuery["locationType"])}>
            {allLocationTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-masterdata-kpis">
        <MasterDataKPI label="Warehouses" value={summary.warehouses} tone="normal" />
        <MasterDataKPI label="Active WH" value={summary.activeWarehouses} tone="success" />
        <MasterDataKPI label="Active locations" value={summary.activeLocations} tone="info" />
        <MasterDataKPI label="Receiving" value={summary.receivingLocations} tone="warning" />
      </section>

      <section className="erp-masterdata-workspace">
        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Warehouse master</h2>
            <StatusChip tone={warehouses.length === 0 ? "warning" : "info"}>{warehouses.length} rows</StatusChip>
          </div>
          <DataTable
            columns={warehouseColumns}
            rows={warehouses}
            getRowKey={(row) => row.id}
            loading={loading}
            error={tableError(error, clearError)}
            emptyState={<EmptyState title="No warehouses" description="Adjust the filters or create a warehouse." />}
          />
        </section>

        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">Location master</h2>
            <StatusChip tone={locations.length === 0 ? "warning" : "info"}>{locations.length} rows</StatusChip>
          </div>
          <label className="erp-field erp-masterdata-inline-filter">
            <span>Warehouse</span>
            <select className="erp-input" value={locationWarehouseId} onChange={(event) => setLocationWarehouseId(event.target.value)}>
              <option value="">All warehouses</option>
              {warehouses.map((warehouse) => (
                <option key={warehouse.id} value={warehouse.id}>
                  {warehouse.warehouseCode}
                </option>
              ))}
            </select>
          </label>
          <DataTable
            columns={locationColumns}
            rows={locations}
            getRowKey={(row) => row.id}
            loading={loading}
            error={tableError(error, clearError)}
            emptyState={<EmptyState title="No locations" description="Adjust the filters or create a location." />}
          />
        </section>
      </section>

      <section className="erp-masterdata-workspace">
        <WarehouseForm
          editingId={editingWarehouseId}
          form={warehouseForm}
          formError={formError}
          saving={saving}
          onChange={(patch) => setWarehouseForm((current) => ({ ...current, ...patch }))}
          onClear={resetWarehouseForm}
          onSubmit={submitWarehouseForm}
        />
        <LocationForm
          editingId={editingLocationId}
          form={locationForm}
          formError={formError}
          saving={saving}
          warehouses={warehouses}
          onChange={(patch) => setLocationForm((current) => ({ ...current, ...patch }))}
          onClear={resetLocationForm}
          onSubmit={submitLocationForm}
        />
      </section>

      <DetailDrawer
        open={Boolean(selectedWarehouse)}
        title={selectedWarehouse?.warehouseCode ?? "Warehouse detail"}
        subtitle={selectedWarehouse?.warehouseName}
        onClose={clearSelectedWarehouse}
        footer={
          selectedWarehouse ? (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => startWarehouseEdit(selectedWarehouse)}>
              Edit
            </button>
          ) : null
        }
      >
        {selectedWarehouse ? <WarehouseDetail item={selectedWarehouse} /> : null}
      </DetailDrawer>
      <DetailDrawer
        open={Boolean(selectedLocation)}
        title={selectedLocation?.locationCode ?? "Location detail"}
        subtitle={selectedLocation?.locationName}
        onClose={clearSelectedLocation}
        footer={
          selectedLocation ? (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => startLocationEdit(selectedLocation)}>
              Edit
            </button>
          ) : null
        }
      >
        {selectedLocation ? <LocationDetail item={selectedLocation} /> : null}
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
          <h1 className="erp-page-title">Warehouse Master Data</h1>
          <p className="erp-page-description">Warehouse and location catalog for receiving and stock ledger flows</p>
        </div>
        <StatusChip tone="info">{summary.activeLocations} active locations</StatusChip>
      </header>
      {content}
    </section>
  );

  async function openWarehouseDetail(warehouseId: string) {
    try {
      await loadWarehouseDetail(warehouseId);
    } catch (detailError) {
      pushToast("Detail failed", errorText(detailError), "danger");
    }
  }

  async function openLocationDetail(locationId: string) {
    try {
      await loadLocationDetail(locationId);
    } catch (detailError) {
      pushToast("Detail failed", errorText(detailError), "danger");
    }
  }

  function startWarehouseEdit(item: WarehouseMasterDataItem) {
    setEditingWarehouseId(item.id);
    setWarehouseForm(toWarehouseInput(item));
    setFormError(undefined);
  }

  function startLocationEdit(item: WarehouseLocationMasterDataItem) {
    setEditingLocationId(item.id);
    setLocationForm(toLocationInput(item));
    setFormError(undefined);
  }

  function resetWarehouseForm() {
    setEditingWarehouseId(null);
    setWarehouseForm(emptyWarehouseInput);
    setFormError(undefined);
  }

  function resetLocationForm() {
    setEditingLocationId(null);
    setLocationForm({ ...emptyLocationInput, warehouseId: warehouses[0]?.id ?? "" });
    setFormError(undefined);
  }

  async function submitWarehouseForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(undefined);
    try {
      const result = editingWarehouseId ? await saveWarehouse(editingWarehouseId, warehouseForm) : await saveNewWarehouse(warehouseForm);
      pushToast(editingWarehouseId ? "Warehouse updated" : "Warehouse created", `${result.warehouseCode} saved`, "success");
      resetWarehouseForm();
    } catch (saveError) {
      setFormError(errorText(saveError));
    }
  }

  async function submitLocationForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(undefined);
    try {
      const result = editingLocationId ? await saveLocation(editingLocationId, locationForm) : await saveNewLocation(locationForm);
      pushToast(editingLocationId ? "Location updated" : "Location created", `${result.locationCode} saved`, "success");
      resetLocationForm();
    } catch (saveError) {
      setFormError(errorText(saveError));
    }
  }

  async function toggleWarehouseStatus(item: WarehouseMasterDataItem) {
    const nextStatus: WarehouseStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveWarehouseStatus(item.id, nextStatus);
      pushToast("Warehouse status changed", `${result.warehouseCode} is ${warehouseStatusLabel(result.status)}`, "success");
    } catch (statusError) {
      pushToast("Status failed", errorText(statusError), "danger");
    }
  }

  async function toggleLocationStatus(item: WarehouseLocationMasterDataItem) {
    const nextStatus: WarehouseLocationStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveLocationStatus(item.id, nextStatus);
      pushToast("Location status changed", `${result.locationCode} is ${locationStatusLabel(result.status)}`, "success");
    } catch (statusError) {
      pushToast("Status failed", errorText(statusError), "danger");
    }
  }

  function pushToast(title: string, description: string, tone: ToastMessage["tone"]) {
    setToast([{ id: `${Date.now()}`, title, description, tone }]);
  }
}

function WarehouseForm({
  editingId,
  form,
  formError,
  saving,
  onChange,
  onClear,
  onSubmit
}: {
  editingId: string | null;
  form: WarehouseMasterDataInput;
  formError?: string;
  saving: boolean;
  onChange: (patch: Partial<WarehouseMasterDataInput>) => void;
  onClear: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return (
    <form onSubmit={onSubmit}>
      <FormSection
        title={editingId ? "Update warehouse" : "Create warehouse"}
        description="Warehouse identity, lifecycle, and operational issue controls"
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
          <TextField label="Warehouse code" value={form.warehouseCode} onChange={(value) => onChange({ warehouseCode: value.toUpperCase() })} />
          <TextField label="Warehouse name" value={form.warehouseName} onChange={(value) => onChange({ warehouseName: value })} />
          <label className="erp-field">
            <span>Warehouse type</span>
            <select className="erp-input" value={form.warehouseType} onChange={(event) => onChange({ warehouseType: event.target.value as WarehouseMasterDataInput["warehouseType"] })}>
              {warehouseTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <TextField label="Site" value={form.siteCode} onChange={(value) => onChange({ siteCode: value.toUpperCase() })} />
          <TextField label="Address" value={form.address} onChange={(value) => onChange({ address: value })} />
          <label className="erp-field">
            <span>Status</span>
            <select className="erp-input" value={form.status} onChange={(event) => onChange({ status: event.target.value as WarehouseStatus })}>
              {warehouseStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="erp-masterdata-toggle-grid">
          <ToggleField label="Sale issue" checked={form.allowSaleIssue} onChange={(value) => onChange({ allowSaleIssue: value })} />
          <ToggleField label="Production issue" checked={form.allowProdIssue} onChange={(value) => onChange({ allowProdIssue: value })} />
          <ToggleField label="Quarantine" checked={form.allowQuarantine} onChange={(value) => onChange({ allowQuarantine: value })} />
        </div>
      </FormSection>
    </form>
  );
}

function LocationForm({
  editingId,
  form,
  formError,
  saving,
  warehouses,
  onChange,
  onClear,
  onSubmit
}: {
  editingId: string | null;
  form: WarehouseLocationMasterDataInput;
  formError?: string;
  saving: boolean;
  warehouses: WarehouseMasterDataItem[];
  onChange: (patch: Partial<WarehouseLocationMasterDataInput>) => void;
  onClear: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return (
    <form onSubmit={onSubmit}>
      <FormSection
        title={editingId ? "Update location" : "Create location"}
        description="Location/bin setup for receiving, pick, QC hold, and stock ledger references"
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
          <label className="erp-field">
            <span>Warehouse</span>
            <select className="erp-input" value={form.warehouseId} onChange={(event) => onChange({ warehouseId: event.target.value })}>
              <option value="">Select warehouse</option>
              {warehouses.map((warehouse) => (
                <option key={warehouse.id} value={warehouse.id}>
                  {warehouse.warehouseCode}
                </option>
              ))}
            </select>
          </label>
          <TextField label="Location code" value={form.locationCode} onChange={(value) => onChange({ locationCode: value.toUpperCase() })} />
          <TextField label="Location name" value={form.locationName} onChange={(value) => onChange({ locationName: value })} />
          <label className="erp-field">
            <span>Location type</span>
            <select className="erp-input" value={form.locationType} onChange={(event) => onChange({ locationType: event.target.value as WarehouseLocationMasterDataInput["locationType"] })}>
              {locationTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <TextField label="Zone" value={form.zoneCode} onChange={(value) => onChange({ zoneCode: value.toUpperCase() })} />
          <label className="erp-field">
            <span>Status</span>
            <select className="erp-input" value={form.status} onChange={(event) => onChange({ status: event.target.value as WarehouseLocationStatus })}>
              {locationStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="erp-masterdata-toggle-grid">
          <ToggleField label="Receive" checked={form.allowReceive} onChange={(value) => onChange({ allowReceive: value })} />
          <ToggleField label="Pick" checked={form.allowPick} onChange={(value) => onChange({ allowPick: value })} />
          <ToggleField label="Store" checked={form.allowStore} onChange={(value) => onChange({ allowStore: value })} />
          <ToggleField label="Default" checked={form.isDefault} onChange={(value) => onChange({ isDefault: value })} />
        </div>
      </FormSection>
    </form>
  );
}

function WarehouseDetail({ item }: { item: WarehouseMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label="Code" value={item.warehouseCode} />
      <MasterDataFact label="Type" value={warehouseTypeLabel(item.warehouseType)} />
      <MasterDataFact label="Site" value={item.siteCode} />
      <MasterDataFact label="Status" value={warehouseStatusLabel(item.status)} />
      <MasterDataFact label="Flow" value={warehouseFlowLabels(item).join(", ") || "-"} />
      <MasterDataFact label="Address" value={item.address || "-"} />
      <MasterDataFact label="Updated" value={formatDate(item.updatedAt)} />
      <MasterDataFact label="Audit" value={item.auditLogId || "Tracked on write"} />
    </div>
  );
}

function LocationDetail({ item }: { item: WarehouseLocationMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label="Code" value={item.locationCode} />
      <MasterDataFact label="Warehouse" value={item.warehouseCode} />
      <MasterDataFact label="Type" value={locationTypeLabel(item.locationType)} />
      <MasterDataFact label="Zone" value={item.zoneCode || "-"} />
      <MasterDataFact label="Status" value={locationStatusLabel(item.status)} />
      <MasterDataFact label="Flow" value={locationFlowLabels(item).join(", ") || "-"} />
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

function ToggleField({ label, checked, onChange }: { label: string; checked: boolean; onChange: (value: boolean) => void }) {
  return (
    <label className="erp-masterdata-toggle">
      <input type="checkbox" checked={checked} onChange={(event) => onChange(event.target.checked)} />
      <span>{label}</span>
    </label>
  );
}

function tableError(error: string | undefined, clearError: () => void) {
  return error ? (
    <ErrorState
      title="Warehouse master data could not load"
      description={error}
      action={
        <button className="erp-button erp-button--secondary" type="button" onClick={clearError}>
          Dismiss
        </button>
      }
    />
  ) : undefined;
}

function warehouseFlowLabels(item: WarehouseMasterDataItem) {
  return [
    item.allowSaleIssue ? "Sale" : undefined,
    item.allowProdIssue ? "Production" : undefined,
    item.allowQuarantine ? "QC hold" : undefined
  ].filter((label): label is string => Boolean(label));
}

function locationFlowLabels(item: WarehouseLocationMasterDataItem) {
  return [
    item.allowReceive ? "Receive" : undefined,
    item.allowPick ? "Pick" : undefined,
    item.allowStore ? "Store" : undefined,
    item.isDefault ? "Default" : undefined
  ].filter((label): label is string => Boolean(label));
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-US", { month: "short", day: "2-digit" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : "Warehouse master data request failed";
}
