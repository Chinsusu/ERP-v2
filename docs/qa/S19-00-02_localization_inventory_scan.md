# S19-00-02 Localization Inventory Scan

Date: 2026-05-02
Scope: Frontend hardcoded English user-facing UI labels before Sprint 19 localization.

## Current Coverage

The first localization slice adds the shared i18n foundation, default Vietnamese locale, Ant Design `vi_VN` provider, app shell labels, navigation labels, action labels, status/error/validation mappings, unit labels, and shared display formatter exports.

## Modules With Remaining Hardcoded English UI Copy

| Area | Representative files | Follow-up task |
| --- | --- | --- |
| Dashboard / warehouse board | `apps/web/src/modules/warehouse/components/WarehouseDailyBoard.tsx`, `ShiftClosingPanel.tsx` | S19-05-01, S19-05-02, S19-05-03 |
| Sales order | `apps/web/src/modules/sales/components/SalesOrderPrototype.tsx` | S19-06-01 |
| Picking / packing / shipping | `apps/web/src/modules/shipping/components/*Prototype.tsx` | S19-06-02, S19-06-03, S19-06-04, S19-06-05 |
| Returns | `apps/web/src/modules/returns/components/*Panel.tsx`, `ReturnReceivingPrototype.tsx` | S19-07-01, S19-07-02, S19-07-03 |
| Purchase / receiving / inbound QC | `apps/web/src/modules/purchase/components/PurchaseOrderPrototype.tsx`, `apps/web/src/modules/receiving/components/WarehouseReceivingPrototype.tsx`, `apps/web/src/modules/qc/components/InboundQCPrototype.tsx` | S19-08-01, S19-08-02, S19-08-03 |
| Master data / inventory / batch QC | `apps/web/src/modules/masterdata/components/*Prototype.tsx`, `apps/web/src/modules/inventory/components/AvailableStockPrototype.tsx` | S19-09-01, S19-09-02, S19-09-03 |
| Finance / audit / attachments | `apps/web/src/modules/finance/components/*Panel.tsx`, `apps/web/src/modules/audit/components/AuditLogPrototype.tsx`, shared attachment panel | S19-10-02, S19-10-03 |
| Auth | `apps/web/src/app/(auth)/login/page.tsx` | S19-10-01 |
| Shared templates | `apps/web/src/shared/design-system/pageTemplates.tsx`, `apps/web/src/shared/layouts/ModulePlaceholder.tsx` | S19-03, S19-10 |

## Guardrails Confirmed

```text
Routes remain English technical paths.
Permission keys remain unchanged.
Backend enum, OpenAPI schema, audit event, and error code names remain unchanged.
User-facing labels should be pulled from dictionaries or centralized label mapping as each module is localized.
```
