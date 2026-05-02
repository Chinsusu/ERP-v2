import { describe, expect, it } from "vitest";
import { defaultLocale, fallbackLocale } from "./config";
import { getActionLabel } from "./action-labels";
import { getErrorLabel } from "./error-labels";
import { formatERPDate, formatERPMoney, formatERPPercent, formatERPQuantity } from "./formatters";
import { getNavigationGroupLabel, getNavigationItemLabel } from "./navigation-labels";
import { getStatusLabel } from "./status-labels";
import { getUnitLabel } from "./units";
import { getValidationLabel } from "./validation-labels";
import { t } from "./index";

describe("Vietnamese-first i18n foundation", () => {
  it("uses Vietnamese as the default locale with English fallback", () => {
    expect(defaultLocale).toBe("vi");
    expect(fallbackLocale).toBe("en");
    expect(t("common.appName")).toBe("ERP Mỹ phẩm");
    expect(t("common.appName", { locale: "en" })).toBe("ERP Platform");
    expect(t("common.missing", { fallback: "Fallback copy" })).toBe("Fallback copy");
  });

  it("localizes navigation and actions without changing routes", () => {
    expect(getNavigationGroupLabel("Operations")).toBe("Vận hành");
    expect(getNavigationItemLabel({ href: "/warehouse", label: "Warehouse Daily Board" })).toBe(
      "Bảng công việc kho"
    );
    expect(getNavigationItemLabel({ href: "/sales", label: "Sales Orders" })).toBe("Đơn bán hàng");
    expect(getActionLabel("Quick create")).toBe("Tạo nhanh");
    expect(getActionLabel("Export CSV")).toBe("Xuất CSV");
    expect(getActionLabel("Download")).toBe("Tải xuống");
    expect(t("common.noRecordsYet")).toBe("Chưa có dữ liệu.");
    expect(t("common.scanCode")).toBe("Quét mã");
    expect(t("warehouse.dailyBoard")).toBe("Bảng công việc kho");
    expect(t("warehouse.taskBoard.rows", { values: { count: 3 } })).toBe("3 dòng");
    expect(t("warehouse.shiftClosingPanel.closeShift")).toBe("Đóng ca");
    expect(t("sales.orders")).toBe("Đơn bán hàng");
    expect(t("sales.actions.confirm")).toBe("Xác nhận");
    expect(t("shipping.picking.title")).toBe("Trạm soạn hàng");
    expect(t("shipping.picking.scan.readyTitle")).toBe("Sẵn sàng quét");
    expect(t("shipping.packing.title")).toBe("Trạm đóng hàng");
    expect(t("shipping.packing.scan.readyTitle")).toBe("Sẵn sàng đóng hàng");
    expect(t("shipping.operations.tabs.handover")).toBe("Bàn giao ĐVVC");
    expect(t("shipping.handover.title")).toBe("Bảng kê bàn giao ĐVVC");
    expect(t("shipping.handover.scanResult.MANIFEST_MISMATCH")).toBe("Sai bảng kê");
    expect(t("returns.receiving.title")).toBe("Nhận hàng hoàn");
    expect(t("returns.receiving.disposition.needs_inspection")).toBe("Cần kiểm tra");
  });

  it("centralizes status, error, validation, and unit labels", () => {
    expect(getStatusLabel("QC_HOLD")).toBe("Đang giữ kiểm");
    expect(getStatusLabel("PACKED")).toBe("Đã đóng hàng");
    expect(getErrorLabel("INSUFFICIENT_STOCK")).toBe("Tồn khả dụng không đủ.");
    expect(getValidationLabel("positiveQuantity")).toBe("Số lượng phải lớn hơn 0.");
    expect(getUnitLabel("carton")).toBe("thùng");
  });

  it("re-exports approved ERP display formatters", () => {
    expect(formatERPMoney("1250000")).toBe("1.250.000 ₫");
    expect(formatERPQuantity("10.5", "kg")).toBe("10,5 KG");
    expect(formatERPPercent("2.5")).toBe("2,5%");
    expect(formatERPDate("2026-04-26T00:00:00+07:00")).toBe("26/04/2026");
  });
});
