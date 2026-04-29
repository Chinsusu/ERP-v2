import { describe, expect, it } from "vitest";
import {
  decimalScales,
  erpCurrencyCode,
  erpNumberLocale,
  erpTimezone,
  formatMoney,
  formatQuantity,
  formatRate,
  normalizeDecimalInput,
  normalizeUOMCode
} from "./numberFormat";

describe("ERP decimal and vi-VN number format helpers", () => {
  it("keeps the approved locale, timezone, and currency constants", () => {
    expect(erpCurrencyCode).toBe("VND");
    expect(erpNumberLocale).toBe("vi-VN");
    expect(erpTimezone).toBe("Asia/Ho_Chi_Minh");
  });

  it("normalizes API decimal strings to fixed scales", () => {
    expect(normalizeDecimalInput("1250000", decimalScales.money)).toBe("1250000.00");
    expect(normalizeDecimalInput("10.5", decimalScales.quantity)).toBe("10.500000");
    expect(normalizeDecimalInput("8", decimalScales.rate)).toBe("8.0000");
  });

  it("normalizes vi-VN input separators", () => {
    expect(normalizeDecimalInput("1.250.000,5", decimalScales.money)).toBe("1250000.50");
    expect(normalizeDecimalInput("10,1254567", decimalScales.quantity)).toBe("10.125457");
  });

  it("keeps scaled decimal rounding cases exact", () => {
    expect(normalizeDecimalInput("0,105", decimalScales.money)).toBe("0.11");
    expect(normalizeDecimalInput("0.1000004", decimalScales.quantity)).toBe("0.100000");
    expect(normalizeDecimalInput("0.1000005", decimalScales.quantity)).toBe("0.100001");
    expect(normalizeDecimalInput("12.34505", decimalScales.rate)).toBe("12.3451");
  });

  it("formats money, quantity, and rate displays", () => {
    expect(formatMoney("1250000.00")).toBe("1.250.000 ₫");
    expect(formatMoney("1.250.000 ₫")).toBe("1.250.000 ₫");
    expect(formatQuantity("10.500000", "kg")).toBe("10,5 KG");
    expect(formatRate("8.1250")).toBe("8,125%");
  });

  it("normalizes UOM codes", () => {
    expect(normalizeUOMCode(" carton ")).toBe("CARTON");
  });
});
