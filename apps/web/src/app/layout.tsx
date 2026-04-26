import type { Metadata } from "next";
import type { ReactNode } from "react";
import "antd/dist/reset.css";
import { DesignSystemProvider } from "@/shared/design-system/DesignSystemProvider";
import "./globals.css";

export const metadata: Metadata = {
  title: "ERP Platform",
  description: "Cosmetics ERP Phase 1",
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
    <html lang="en">
      <body>
        <DesignSystemProvider>{children}</DesignSystemProvider>
      </body>
    </html>
  );
}
