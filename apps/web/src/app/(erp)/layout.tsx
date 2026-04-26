import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { mockSession } from "@/shared/auth/mockSession";
import { AppShell } from "@/shared/layouts/AppShell";

type ERPLayoutProps = {
  children: ReactNode;
};

export default function ERPLayout({ children }: ERPLayoutProps) {
  if (!mockSession.isAuthenticated) {
    redirect("/login");
  }

  return <AppShell user={mockSession.user}>{children}</AppShell>;
}
