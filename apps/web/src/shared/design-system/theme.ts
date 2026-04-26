import type { ThemeConfig } from "antd";
import { erpTokens } from "./tokens";

export const erpTheme: ThemeConfig = {
  token: {
    colorPrimary: erpTokens.semantic.color.primary,
    colorText: erpTokens.semantic.color.text,
    colorTextSecondary: erpTokens.semantic.color.textSecondary,
    colorTextTertiary: erpTokens.semantic.color.textMuted,
    colorBgLayout: erpTokens.semantic.color.appBackground,
    colorBgContainer: erpTokens.semantic.color.surface,
    colorBgElevated: erpTokens.semantic.color.surface,
    colorBorder: erpTokens.semantic.color.border,
    colorBorderSecondary: erpTokens.semantic.color.borderSoft,
    colorError: erpTokens.semantic.color.danger,
    colorErrorBg: erpTokens.semantic.color.dangerBackground,
    colorSuccess: erpTokens.semantic.color.success,
    colorSuccessBg: erpTokens.semantic.color.successBackground,
    colorWarning: erpTokens.semantic.color.warning,
    colorWarningBg: erpTokens.semantic.color.warningBackground,
    colorInfo: erpTokens.semantic.color.info,
    colorInfoBg: erpTokens.semantic.color.infoBackground,
    borderRadius: erpTokens.primitive.radius.default,
    borderRadiusSM: erpTokens.primitive.radius.small,
    borderRadiusLG: erpTokens.primitive.radius.large,
    fontFamily: erpTokens.semantic.typography.fontFamily,
    fontSize: erpTokens.semantic.typography.body.fontSize,
    controlHeight: erpTokens.semantic.layout.controlHeight,
    wireframe: false
  },
  components: {
    Button: {
      borderRadius: erpTokens.component.button.primary.radius,
      controlHeight: erpTokens.component.button.primary.height
    },
    Card: {
      borderRadiusLG: erpTokens.component.card.radius,
      paddingLG: erpTokens.component.card.padding
    },
    Input: {
      controlHeight: erpTokens.semantic.layout.controlHeight
    },
    Select: {
      controlHeight: erpTokens.semantic.layout.controlHeight
    },
    Table: {
      headerBg: erpTokens.component.table.headerBackground,
      headerColor: erpTokens.component.table.headerColor,
      rowHoverBg: erpTokens.component.table.rowHoverBackground,
      fontSize: erpTokens.component.table.fontSize
    },
    Tag: {
      borderRadiusSM: erpTokens.primitive.radius.chip
    }
  }
};
