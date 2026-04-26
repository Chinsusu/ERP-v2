import { describe, expect, it } from "vitest";
import {
  overlayTemplateClassName,
  pageTemplateClassName,
  pageTemplateNames,
  templatePanelClassName
} from "./pageTemplates";

describe("UI page templates", () => {
  it("exports the S0-03-04 template set", () => {
    expect(pageTemplateNames).toEqual([
      "AppShell",
      "PageHeader",
      "FilterBar",
      "TablePageTemplate",
      "FormPageTemplate",
      "DetailPageTemplate",
      "ModalTemplate",
      "DrawerTemplate",
      "PopoverTemplate",
      "EmptyState",
      "LoadingState",
      "ErrorState",
      "AuditLogPanel",
      "AttachmentPanel"
    ]);
  });

  it("maps templates to stable class names", () => {
    expect(pageTemplateClassName("detail")).toBe("erp-ds-page-template erp-ds-page-template--detail");
    expect(templatePanelClassName("warning")).toBe("erp-ds-template-panel erp-ds-template-panel--warning");
    expect(overlayTemplateClassName("popover")).toBe("erp-ds-overlay-template erp-ds-overlay-template--popover");
  });
});
