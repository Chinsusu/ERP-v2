"use client";

import { ConfigProvider } from "antd";
import type { ReactNode } from "react";
import { erpTheme } from "./theme";

type DesignSystemProviderProps = {
  children: ReactNode;
};

export function DesignSystemProvider({ children }: DesignSystemProviderProps) {
  return <ConfigProvider theme={erpTheme}>{children}</ConfigProvider>;
}
