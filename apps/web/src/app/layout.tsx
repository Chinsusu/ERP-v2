import type { Metadata } from "next";
import type { ReactNode } from "react";
import "antd/dist/reset.css";
import { DesignSystemProvider } from "@/shared/design-system/DesignSystemProvider";
import { t } from "@/shared/i18n";
import "./globals.css";

export const metadata: Metadata = {
  title: t("common.appName"),
  description: "ERP mỹ phẩm Phase 1",
  icons: {
    icon: "/icon.svg"
  }
};

export default function RootLayout({
  children
}: Readonly<{
  children: ReactNode;
}>) {
  return (
    <html lang="vi">
      <body>
        <DesignSystemProvider>{children}</DesignSystemProvider>
      </body>
    </html>
  );
}
