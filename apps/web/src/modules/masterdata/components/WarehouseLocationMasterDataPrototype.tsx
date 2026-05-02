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
import { t } from "@/shared/i18n";
import { useWarehouseMasterData } from "../hooks/useWarehouseMasterData";
import {
  emptyLocationInput,
  emptyWarehouseInput,
  locationStatusOptions,
  locationTypeOptions,
  toLocationInput,
  toWarehouseInput,
  warehouseStatusOptions,
  warehouseStatusTone,
  warehouseTypeOptions
} from "../services/warehouseMasterDataService";
import type {
  WarehouseLocationMasterDataInput,
  WarehouseLocationMasterDataItem,
  WarehouseLocationMasterDataQuery,
  WarehouseLocationStatus,
  WarehouseLocationType,
  WarehouseMasterDataInput,
  WarehouseMasterDataItem,
  WarehouseMasterDataQuery,
  WarehouseStatus,
  WarehouseType
} from "../types";

const allWarehouseStatusOptions = [{ label: warehouseCopy("filters.allWarehouseStatuses"), value: "" }, ...warehouseStatusOptions] as const;
const allWarehouseTypeOptions = [{ label: warehouseCopy("filters.allWarehouseTypes"), value: "" }, ...warehouseTypeOptions] as const;
const allLocationStatusOptions = [{ label: warehouseCopy("filters.allLocationStatuses"), value: "" }, ...locationStatusOptions] as const;
const allLocationTypeOptions = [{ label: warehouseCopy("filters.allLocationTypes"), value: "" }, ...locationTypeOptions] as const;

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
      header: warehouseCopy("warehouse.columns.warehouse"),
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.warehouseCode}</strong>
          <small>{row.warehouseName}</small>
        </div>
      ),
      width: "240px"
    },
    { key: "type", header: warehouseCopy("columns.type"), render: (row) => warehouseTypeDisplay(row.warehouseType), width: "130px" },
    { key: "site", header: warehouseCopy("columns.site"), render: (row) => row.siteCode, width: "80px" },
    {
      key: "flow",
      header: warehouseCopy("columns.flow"),
      render: (row) => warehouseFlowLabels(row).join(", ") || "-",
      width: "160px"
    },
    {
      key: "status",
      header: warehouseCopy("columns.status"),
      render: (row) => <StatusChip tone={warehouseStatusTone(row.status)}>{warehouseStatusDisplay(row.status)}</StatusChip>,
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
            {commonAction("detail")}
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startWarehouseEdit(row)}>
            {commonAction("edit")}
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving}
            onClick={() => toggleWarehouseStatus(row)}
          >
            {row.status === "active" ? commonAction("inactivate") : commonAction("activate")}
          </button>
        </div>
      )
    }
  ];

  const locationColumns: DataTableColumn<WarehouseLocationMasterDataItem>[] = [
    {
      key: "location",
      header: warehouseCopy("location.columns.location"),
      render: (row) => (
        <div className="erp-masterdata-product-cell">
          <strong>{row.locationCode}</strong>
          <small>{row.locationName}</small>
        </div>
      ),
      width: "220px"
    },
    { key: "warehouse", header: warehouseCopy("location.columns.warehouse"), render: (row) => row.warehouseCode, width: "120px" },
    { key: "type", header: warehouseCopy("columns.type"), render: (row) => locationTypeDisplay(row.locationType), width: "120px" },
    {
      key: "flow",
      header: warehouseCopy("columns.flow"),
      render: (row) => locationFlowLabels(row).join(", ") || "-",
      width: "160px"
    },
    {
      key: "status",
      header: warehouseCopy("columns.status"),
      render: (row) => <StatusChip tone={warehouseStatusTone(row.status)}>{locationStatusDisplay(row.status)}</StatusChip>,
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
            {commonAction("detail")}
          </button>
          <button className="erp-button erp-button--secondary erp-button--compact" type="button" onClick={() => startLocationEdit(row)}>
            {commonAction("edit")}
          </button>
          <button
            className="erp-button erp-button--secondary erp-button--compact"
            type="button"
            disabled={saving}
            onClick={() => toggleLocationStatus(row)}
          >
            {row.status === "active" ? commonAction("inactivate") : commonAction("activate")}
          </button>
        </div>
      )
    }
  ];

  const content = (
    <>
      <section className="erp-masterdata-toolbar erp-masterdata-toolbar--wide" aria-label={warehouseCopy("filters.label")}>
        <label className="erp-field">
          <span>{warehouseCopy("filters.search")}</span>
          <input className="erp-input" type="search" value={search} placeholder="WH-HCM-FG" onChange={(event) => setSearch(event.target.value.toUpperCase())} />
        </label>
        <label className="erp-field">
          <span>{warehouseCopy("filters.warehouseStatus")}</span>
          <select className="erp-input" value={warehouseStatus} onChange={(event) => setWarehouseStatus(event.target.value as WarehouseStatus | "")}>
            {allWarehouseStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? warehouseStatusDisplay(option.value) : option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{warehouseCopy("filters.warehouseType")}</span>
          <select className="erp-input" value={warehouseType} onChange={(event) => setWarehouseType(event.target.value as WarehouseMasterDataQuery["warehouseType"])}>
            {allWarehouseTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? warehouseTypeDisplay(option.value) : option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{warehouseCopy("filters.locationStatus")}</span>
          <select className="erp-input" value={locationStatus} onChange={(event) => setLocationStatus(event.target.value as WarehouseLocationStatus | "")}>
            {allLocationStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? locationStatusDisplay(option.value) : option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="erp-field">
          <span>{warehouseCopy("filters.locationType")}</span>
          <select className="erp-input" value={locationType} onChange={(event) => setLocationType(event.target.value as WarehouseLocationMasterDataQuery["locationType"])}>
            {allLocationTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.value ? locationTypeDisplay(option.value) : option.label}
              </option>
            ))}
          </select>
        </label>
      </section>

      <section className="erp-kpi-grid erp-masterdata-kpis">
        <MasterDataKPI label={warehouseCopy("kpi.warehouses")} value={summary.warehouses} tone="normal" />
        <MasterDataKPI label={warehouseCopy("kpi.activeWarehouses")} value={summary.activeWarehouses} tone="success" />
        <MasterDataKPI label={warehouseCopy("kpi.activeLocations")} value={summary.activeLocations} tone="info" />
        <MasterDataKPI label={warehouseCopy("kpi.receiving")} value={summary.receivingLocations} tone="warning" />
      </section>

      <section className="erp-masterdata-workspace">
        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{warehouseCopy("warehouse.list.title")}</h2>
            <StatusChip tone={warehouses.length === 0 ? "warning" : "info"}>{warehouseCopy("list.rows", { count: warehouses.length })}</StatusChip>
          </div>
          <DataTable
            columns={warehouseColumns}
            rows={warehouses}
            getRowKey={(row) => row.id}
            loading={loading}
            error={tableError(error, clearError)}
            emptyState={<EmptyState title={warehouseCopy("warehouse.empty.title")} description={warehouseCopy("warehouse.empty.description")} />}
          />
        </section>

        <section className="erp-card erp-card--padded erp-masterdata-list-card">
          <div className="erp-section-header">
            <h2 className="erp-section-title">{warehouseCopy("location.list.title")}</h2>
            <StatusChip tone={locations.length === 0 ? "warning" : "info"}>{warehouseCopy("list.rows", { count: locations.length })}</StatusChip>
          </div>
          <label className="erp-field erp-masterdata-inline-filter">
            <span>{warehouseCopy("fields.warehouse")}</span>
            <select className="erp-input" value={locationWarehouseId} onChange={(event) => setLocationWarehouseId(event.target.value)}>
              <option value="">{warehouseCopy("filters.allWarehouses")}</option>
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
            emptyState={<EmptyState title={warehouseCopy("location.empty.title")} description={warehouseCopy("location.empty.description")} />}
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
        title={selectedWarehouse?.warehouseCode ?? warehouseCopy("warehouse.detail.title")}
        subtitle={selectedWarehouse?.warehouseName}
        onClose={clearSelectedWarehouse}
        footer={
          selectedWarehouse ? (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => startWarehouseEdit(selectedWarehouse)}>
              {commonAction("edit")}
            </button>
          ) : null
        }
      >
        {selectedWarehouse ? <WarehouseDetail item={selectedWarehouse} /> : null}
      </DetailDrawer>
      <DetailDrawer
        open={Boolean(selectedLocation)}
        title={selectedLocation?.locationCode ?? warehouseCopy("location.detail.title")}
        subtitle={selectedLocation?.locationName}
        onClose={clearSelectedLocation}
        footer={
          selectedLocation ? (
            <button className="erp-button erp-button--secondary" type="button" onClick={() => startLocationEdit(selectedLocation)}>
              {commonAction("edit")}
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
          <h1 className="erp-page-title">{warehouseCopy("page.title")}</h1>
          <p className="erp-page-description">{warehouseCopy("page.description")}</p>
        </div>
        <StatusChip tone="info">{warehouseCopy("page.activeLocations", { count: summary.activeLocations })}</StatusChip>
      </header>
      {content}
    </section>
  );

  async function openWarehouseDetail(warehouseId: string) {
    try {
      await loadWarehouseDetail(warehouseId);
    } catch (detailError) {
      pushToast(warehouseCopy("toast.detailFailed"), errorText(detailError), "danger");
    }
  }

  async function openLocationDetail(locationId: string) {
    try {
      await loadLocationDetail(locationId);
    } catch (detailError) {
      pushToast(warehouseCopy("toast.detailFailed"), errorText(detailError), "danger");
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
      pushToast(editingWarehouseId ? warehouseCopy("warehouse.toast.updated") : warehouseCopy("warehouse.toast.created"), warehouseCopy("toast.saved", { code: result.warehouseCode }), "success");
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
      pushToast(editingLocationId ? warehouseCopy("location.toast.updated") : warehouseCopy("location.toast.created"), warehouseCopy("toast.saved", { code: result.locationCode }), "success");
      resetLocationForm();
    } catch (saveError) {
      setFormError(errorText(saveError));
    }
  }

  async function toggleWarehouseStatus(item: WarehouseMasterDataItem) {
    const nextStatus: WarehouseStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveWarehouseStatus(item.id, nextStatus);
      pushToast(warehouseCopy("warehouse.toast.statusChanged"), warehouseCopy("toast.statusChangedDescription", { code: result.warehouseCode, status: warehouseStatusDisplay(result.status) }), "success");
    } catch (statusError) {
      pushToast(warehouseCopy("toast.statusFailed"), errorText(statusError), "danger");
    }
  }

  async function toggleLocationStatus(item: WarehouseLocationMasterDataItem) {
    const nextStatus: WarehouseLocationStatus = item.status === "active" ? "inactive" : "active";
    try {
      const result = await saveLocationStatus(item.id, nextStatus);
      pushToast(warehouseCopy("location.toast.statusChanged"), warehouseCopy("toast.statusChangedDescription", { code: result.locationCode, status: locationStatusDisplay(result.status) }), "success");
    } catch (statusError) {
      pushToast(warehouseCopy("toast.statusFailed"), errorText(statusError), "danger");
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
        title={editingId ? warehouseCopy("warehouse.form.updateTitle") : warehouseCopy("warehouse.form.createTitle")}
        description={warehouseCopy("warehouse.form.description")}
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
          <TextField label={warehouseCopy("warehouse.form.warehouseCode")} value={form.warehouseCode} onChange={(value) => onChange({ warehouseCode: value.toUpperCase() })} />
          <TextField label={warehouseCopy("warehouse.form.warehouseName")} value={form.warehouseName} onChange={(value) => onChange({ warehouseName: value })} />
          <label className="erp-field">
            <span>{warehouseCopy("warehouse.form.warehouseType")}</span>
            <select className="erp-input" value={form.warehouseType} onChange={(event) => onChange({ warehouseType: event.target.value as WarehouseMasterDataInput["warehouseType"] })}>
              {warehouseTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {warehouseTypeDisplay(option.value)}
                </option>
              ))}
            </select>
          </label>
          <TextField label={warehouseCopy("fields.site")} value={form.siteCode} onChange={(value) => onChange({ siteCode: value.toUpperCase() })} />
          <TextField label={warehouseCopy("fields.address")} value={form.address} onChange={(value) => onChange({ address: value })} />
          <label className="erp-field">
            <span>{warehouseCopy("fields.status")}</span>
            <select className="erp-input" value={form.status} onChange={(event) => onChange({ status: event.target.value as WarehouseStatus })}>
              {warehouseStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {warehouseStatusDisplay(option.value)}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="erp-masterdata-toggle-grid">
          <ToggleField label={warehouseCopy("flow.saleIssue")} checked={form.allowSaleIssue} onChange={(value) => onChange({ allowSaleIssue: value })} />
          <ToggleField label={warehouseCopy("flow.productionIssue")} checked={form.allowProdIssue} onChange={(value) => onChange({ allowProdIssue: value })} />
          <ToggleField label={warehouseCopy("flow.quarantine")} checked={form.allowQuarantine} onChange={(value) => onChange({ allowQuarantine: value })} />
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
        title={editingId ? warehouseCopy("location.form.updateTitle") : warehouseCopy("location.form.createTitle")}
        description={warehouseCopy("location.form.description")}
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
          <label className="erp-field">
            <span>{warehouseCopy("fields.warehouse")}</span>
            <select className="erp-input" value={form.warehouseId} onChange={(event) => onChange({ warehouseId: event.target.value })}>
              <option value="">{warehouseCopy("location.form.selectWarehouse")}</option>
              {warehouses.map((warehouse) => (
                <option key={warehouse.id} value={warehouse.id}>
                  {warehouse.warehouseCode}
                </option>
              ))}
            </select>
          </label>
          <TextField label={warehouseCopy("location.form.locationCode")} value={form.locationCode} onChange={(value) => onChange({ locationCode: value.toUpperCase() })} />
          <TextField label={warehouseCopy("location.form.locationName")} value={form.locationName} onChange={(value) => onChange({ locationName: value })} />
          <label className="erp-field">
            <span>{warehouseCopy("location.form.locationType")}</span>
            <select className="erp-input" value={form.locationType} onChange={(event) => onChange({ locationType: event.target.value as WarehouseLocationMasterDataInput["locationType"] })}>
              {locationTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {locationTypeDisplay(option.value)}
                </option>
              ))}
            </select>
          </label>
          <TextField label={warehouseCopy("fields.zone")} value={form.zoneCode} onChange={(value) => onChange({ zoneCode: value.toUpperCase() })} />
          <label className="erp-field">
            <span>{warehouseCopy("fields.status")}</span>
            <select className="erp-input" value={form.status} onChange={(event) => onChange({ status: event.target.value as WarehouseLocationStatus })}>
              {locationStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {locationStatusDisplay(option.value)}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="erp-masterdata-toggle-grid">
          <ToggleField label={warehouseCopy("flow.receive")} checked={form.allowReceive} onChange={(value) => onChange({ allowReceive: value })} />
          <ToggleField label={warehouseCopy("flow.pick")} checked={form.allowPick} onChange={(value) => onChange({ allowPick: value })} />
          <ToggleField label={warehouseCopy("flow.store")} checked={form.allowStore} onChange={(value) => onChange({ allowStore: value })} />
          <ToggleField label={warehouseCopy("flow.default")} checked={form.isDefault} onChange={(value) => onChange({ isDefault: value })} />
        </div>
      </FormSection>
    </form>
  );
}

function WarehouseDetail({ item }: { item: WarehouseMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label={warehouseCopy("detail.code")} value={item.warehouseCode} />
      <MasterDataFact label={warehouseCopy("columns.type")} value={warehouseTypeDisplay(item.warehouseType)} />
      <MasterDataFact label={warehouseCopy("fields.site")} value={item.siteCode} />
      <MasterDataFact label={warehouseCopy("fields.status")} value={warehouseStatusDisplay(item.status)} />
      <MasterDataFact label={warehouseCopy("columns.flow")} value={warehouseFlowLabels(item).join(", ") || "-"} />
      <MasterDataFact label={warehouseCopy("fields.address")} value={item.address || "-"} />
      <MasterDataFact label={warehouseCopy("detail.updated")} value={formatDate(item.updatedAt)} />
      <MasterDataFact label={warehouseCopy("detail.audit")} value={item.auditLogId || warehouseCopy("detail.auditFallback")} />
    </div>
  );
}

function LocationDetail({ item }: { item: WarehouseLocationMasterDataItem }) {
  return (
    <div className="erp-masterdata-detail-grid">
      <MasterDataFact label={warehouseCopy("detail.code")} value={item.locationCode} />
      <MasterDataFact label={warehouseCopy("fields.warehouse")} value={item.warehouseCode} />
      <MasterDataFact label={warehouseCopy("columns.type")} value={locationTypeDisplay(item.locationType)} />
      <MasterDataFact label={warehouseCopy("fields.zone")} value={item.zoneCode || "-"} />
      <MasterDataFact label={warehouseCopy("fields.status")} value={locationStatusDisplay(item.status)} />
      <MasterDataFact label={warehouseCopy("columns.flow")} value={locationFlowLabels(item).join(", ") || "-"} />
      <MasterDataFact label={warehouseCopy("detail.updated")} value={formatDate(item.updatedAt)} />
      <MasterDataFact label={warehouseCopy("detail.audit")} value={item.auditLogId || warehouseCopy("detail.auditFallback")} />
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
      title={warehouseCopy("errors.loadTitle")}
      description={error}
      action={
        <button className="erp-button erp-button--secondary" type="button" onClick={clearError}>
          {commonAction("dismiss")}
        </button>
      }
    />
  ) : undefined;
}

function warehouseFlowLabels(item: WarehouseMasterDataItem) {
  return [
    item.allowSaleIssue ? warehouseCopy("flow.saleIssue") : undefined,
    item.allowProdIssue ? warehouseCopy("flow.productionIssue") : undefined,
    item.allowQuarantine ? warehouseCopy("flow.quarantine") : undefined
  ].filter((label): label is string => Boolean(label));
}

function locationFlowLabels(item: WarehouseLocationMasterDataItem) {
  return [
    item.allowReceive ? warehouseCopy("flow.receive") : undefined,
    item.allowPick ? warehouseCopy("flow.pick") : undefined,
    item.allowStore ? warehouseCopy("flow.store") : undefined,
    item.isDefault ? warehouseCopy("flow.default") : undefined
  ].filter((label): label is string => Boolean(label));
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("vi-VN", { day: "2-digit", month: "2-digit" }).format(new Date(value));
}

function errorText(error: unknown) {
  return error instanceof Error ? error.message : warehouseCopy("errors.requestFailed");
}

function warehouseTypeDisplay(type: WarehouseType) {
  return warehouseCopy(`warehouse.type.${type}`);
}

function warehouseStatusDisplay(status: WarehouseStatus) {
  return warehouseCopy(`status.${status}`);
}

function locationTypeDisplay(type: WarehouseLocationType) {
  return warehouseCopy(`location.type.${type}`);
}

function locationStatusDisplay(status: WarehouseLocationStatus) {
  return warehouseCopy(`status.${status}`);
}

function commonAction(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.actions.${key}`, { values, fallback });
}

function warehouseCopy(key: string, values?: Record<string, string | number>, fallback?: string) {
  return t(`masterdata.warehouse.${key}`, { values, fallback });
}
