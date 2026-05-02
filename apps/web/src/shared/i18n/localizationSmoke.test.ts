import { describe, expect, it } from "vitest";
import type { MockUser } from "../auth/mockSession";
import { getVisibleMenuGroups, rolePermissions } from "../permissions/menu";
import { getNavigationGroupLabel, getNavigationItemLabel } from "./navigation-labels";
import { t } from "./index";

const erpAdminUser: MockUser = {
  id: "localization-smoke-admin",
  name: "ERP Admin",
  email: "admin@example.local",
  role: "ERP_ADMIN",
  permissions: rolePermissions.ERP_ADMIN
};

describe("Vietnamese UI localization smoke", () => {
  it("resolves app shell menu and topbar labels in Vietnamese while keeping routes English", () => {
    const groups = getVisibleMenuGroups(erpAdminUser);
    const groupLabels = groups.map((group) => getNavigationGroupLabel(group.label));
    const itemLabels = groups.flatMap((group) => group.items.map((item) => getNavigationItemLabel(item)));
    const routes = groups.flatMap((group) => group.items.map((item) => item.href));

    expect(t("common.appName")).toBe("ERP Mỹ phẩm");
    expect(t("common.globalSearch")).toBe("Tìm kiếm toàn hệ thống");
    expect(t("common.quickActions")).toBe("Thao tác nhanh");
    expect(groupLabels).toContain("Vận hành");
    expect(groupLabels).toContain("Dữ liệu gốc");
    expect(groupLabels).toContain("Kiểm soát");
    expect(itemLabels).toContain("Bảng công việc kho");
    expect(itemLabels).toContain("Nhập kho");
    expect(itemLabels).toContain("Nhật ký thao tác");
    expect(itemLabels).not.toContain("Warehouse Daily Board");
    expect(itemLabels).not.toContain("Audit Log");
    expect(routes).toContain("/warehouse");
    expect(routes).toContain("/audit-log");
  });

  it("covers warehouse-facing screens with Vietnamese operational labels", () => {
    expect(t("warehouse.dailyBoard")).toBe("Bảng công việc kho");
    expect(t("warehouse.shiftClosingPanel.closeShift")).toBe("Đóng ca");
    expect(t("shipping.handover.title")).toBe("Bảng kê bàn giao ĐVVC");
    expect(t("returns.receiving.title")).toBe("Nhận hàng hoàn");
    expect(t("returns.inspection.sections.attachments")).toBe("Tệp kiểm hàng hoàn");
    expect(t("purchase.receiving.title")).toBe("Nhập kho");
    expect(t("qc.inbound.title")).toBe("QC hàng nhập");
  });
});
