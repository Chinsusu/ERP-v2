import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const i18nDir = path.dirname(fileURLToPath(import.meta.url));
const sourceRoot = path.resolve(i18nDir, "..", "..");

const componentFiles = [
  "modules/audit/components/AuditLogPrototype.tsx",
  "modules/purchase/components/PurchaseOrderPrototype.tsx",
  "modules/qc/components/InboundQCPrototype.tsx",
  "modules/receiving/components/WarehouseReceivingPrototype.tsx",
  "modules/returns/components/ReturnInspectionPanel.tsx",
  "modules/returns/components/ReturnReceivingPrototype.tsx",
  "modules/returns/components/SupplierRejectionPanel.tsx",
  "shared/design-system/pageTemplates.tsx",
  "shared/layouts/AppShell.tsx"
];

const blockedUIPhrases = [
  "Audit Log",
  "Immutable trace of sensitive ERP actions",
  "All actions",
  "Audit events",
  "PO attachments",
  "Upload file",
  "Warehouse Receiving",
  "Receiving attachments",
  "QC attachments",
  "Return Receiving",
  "Inspection attachments",
  "Supplier rejection",
  "Supplier rejection attachments",
  "No supplier rejection evidence attached.",
  "Evidence file",
  "Audit trail",
  "No audit events loaded",
  "Download file",
  "Delete file"
];

describe("hardcoded English UI checklist", () => {
  it("keeps common English UI phrases out of localized business components", () => {
    const source = componentFiles
      .map((file) => fs.readFileSync(path.join(sourceRoot, file), "utf8"))
      .join("\n");

    for (const phrase of blockedUIPhrases) {
      expect(source, phrase).not.toContain(phrase);
    }
  });
});
