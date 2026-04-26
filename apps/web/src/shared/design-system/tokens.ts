export const erpTokens = {
  primitive: {
    color: {
      red: "#D50C2D",
      darkGrey: "#3C3C3B",
      white: "#FFFFFF",
      grey50: "#FAFAFA",
      grey100: "#F5F5F5",
      grey150: "#F0F0F0",
      grey200: "#E8E8E8",
      grey250: "#E5E5E5",
      grey300: "#D9D9D9",
      grey500: "#8C8C8C",
      grey700: "#666666",
      grey900: "#1F1F1F",
      green700: "#2E7D32",
      green50: "#EAF5EC",
      amber700: "#B26A00",
      amber50: "#FFF4E0",
      red50: "#FFE8EC",
      red25: "#FFF1F3",
      blue600: "#246BFE",
      blue50: "#EAF1FF"
    },
    spacing: {
      "2": 2,
      "4": 4,
      "8": 8,
      "12": 12,
      "16": 16,
      "20": 20,
      "24": 24,
      "32": 32,
      "40": 40,
      "48": 48
    },
    radius: {
      small: 2,
      chip: 3,
      default: 4,
      large: 6
    },
    shadow: {
      subtle: "0 1px 2px rgba(0, 0, 0, 0.04)",
      raised: "0 8px 24px rgba(31, 31, 31, 0.08)"
    }
  },
  semantic: {
    color: {
      primary: "#D50C2D",
      danger: "#D50C2D",
      dangerBackground: "#FFE8EC",
      success: "#2E7D32",
      successBackground: "#EAF5EC",
      warning: "#B26A00",
      warningBackground: "#FFF4E0",
      info: "#246BFE",
      infoBackground: "#EAF1FF",
      neutral: "#666666",
      neutralBackground: "#F0F0F0",
      appBackground: "#F5F5F5",
      sectionBackground: "#FAFAFA",
      surface: "#FFFFFF",
      surfaceHover: "#F7F7F7",
      surfaceSelected: "#FFF1F3",
      text: "#1F1F1F",
      textSecondary: "#666666",
      textMuted: "#8C8C8C",
      border: "#D9D9D9",
      borderSoft: "#E8E8E8"
    },
    typography: {
      fontFamily: 'Inter, "Roboto", "Helvetica Neue", Arial, sans-serif',
      monoFamily: '"Roboto Mono", "SFMono-Regular", Consolas, monospace',
      weightRegular: 400,
      weightMedium: 500,
      weightSemibold: 600,
      pageTitle: { fontSize: 24, lineHeight: 32, fontWeight: 600 },
      sectionTitle: { fontSize: 18, lineHeight: 26, fontWeight: 600 },
      cardTitle: { fontSize: 14, lineHeight: 20, fontWeight: 600 },
      body: { fontSize: 14, lineHeight: 22, fontWeight: 400 },
      table: { fontSize: 13, lineHeight: 20, fontWeight: 400 },
      helper: { fontSize: 12, lineHeight: 18, fontWeight: 400 },
      label: { fontSize: 13, lineHeight: 20, fontWeight: 500 },
      button: { fontSize: 14, lineHeight: 20, fontWeight: 500 }
    },
    layout: {
      pagePaddingDesktop: 24,
      pagePaddingTablet: 16,
      cardPadding: 16,
      sectionGap: 24,
      fieldGap: 12,
      tableToolbarGap: 12,
      topbarHeight: 56,
      sidebarWidth: 240,
      sidebarCollapsedWidth: 72,
      tableRowHeight: 44,
      controlHeight: 36
    }
  },
  component: {
    button: {
      primary: {
        background: "#D50C2D",
        color: "#FFFFFF",
        border: "#D50C2D",
        height: 36,
        radius: 4
      },
      secondary: {
        background: "#FFFFFF",
        color: "#3C3C3B",
        border: "#D9D9D9",
        height: 36,
        radius: 4
      },
      danger: {
        background: "#FFFFFF",
        color: "#D50C2D",
        border: "#D50C2D",
        height: 36,
        radius: 4
      }
    },
    card: {
      background: "#FFFFFF",
      border: "#E5E5E5",
      radius: 4,
      padding: 16,
      shadow: "0 1px 2px rgba(0, 0, 0, 0.04)"
    },
    table: {
      headerBackground: "#F5F5F5",
      headerColor: "#3C3C3B",
      rowHoverBackground: "#F7F7F7",
      border: "#EFEFEF",
      fontSize: 13,
      rowHeight: 44
    },
    status: {
      normal: {
        color: "#666666",
        background: "#F0F0F0"
      },
      success: {
        color: "#2E7D32",
        background: "#EAF5EC"
      },
      warning: {
        color: "#B26A00",
        background: "#FFF4E0"
      },
      danger: {
        color: "#D50C2D",
        background: "#FFE8EC"
      },
      info: {
        color: "#246BFE",
        background: "#EAF1FF"
      }
    }
  }
} as const;

export const tokens = {
  color: {
    background: erpTokens.semantic.color.appBackground,
    surface: erpTokens.semantic.color.surface,
    text: erpTokens.semantic.color.text,
    muted: erpTokens.semantic.color.textSecondary,
    border: erpTokens.semantic.color.border,
    primary: erpTokens.semantic.color.primary
  },
  radius: {
    card: erpTokens.component.card.radius
  }
} as const;

export type ErpTokens = typeof erpTokens;
