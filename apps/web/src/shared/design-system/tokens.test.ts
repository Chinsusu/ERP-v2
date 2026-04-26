import { describe, expect, it } from "vitest";
import { erpTheme } from "./theme";
import { erpTokens, tokens } from "./tokens";

describe("ERP design tokens", () => {
  it("maps the file 39 primary palette into semantic tokens", () => {
    expect(erpTokens.semantic.color.primary).toBe("#D50C2D");
    expect(erpTokens.semantic.color.appBackground).toBe("#F5F5F5");
    expect(erpTokens.semantic.color.surface).toBe("#FFFFFF");
    expect(erpTokens.semantic.color.border).toBe("#D9D9D9");
    expect(tokens.color.primary).toBe("#D50C2D");
  });

  it("keeps the industrial minimal shape and density scale", () => {
    expect(erpTokens.primitive.radius.default).toBe(4);
    expect(erpTokens.component.card.radius).toBe(4);
    expect(erpTokens.semantic.layout.controlHeight).toBe(36);
    expect(erpTokens.semantic.layout.tableRowHeight).toBe(44);
  });

  it("exports an Ant Design theme aligned with ERP tokens", () => {
    expect(erpTheme.token?.colorPrimary).toBe(erpTokens.semantic.color.primary);
    expect(erpTheme.token?.colorBgLayout).toBe(erpTokens.semantic.color.appBackground);
    expect(erpTheme.components?.Table?.headerBg).toBe(erpTokens.component.table.headerBackground);
  });
});
