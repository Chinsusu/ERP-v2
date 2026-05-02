"use client";

import { ConfigProvider } from "antd";
import viVN from "antd/locale/vi_VN";
import type { ReactNode } from "react";
import { erpTheme } from "./theme";

type DesignSystemProviderProps = {
  children: ReactNode;
};

export function DesignSystemProvider({ children }: DesignSystemProviderProps) {
  return (
    <ConfigProvider locale={viVN} theme={erpTheme}>
      {children}
    </ConfigProvider>
  );
}
