import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { getMockSession } from "@/shared/auth/mockSession";
import { AppShell } from "@/shared/layouts/AppShell";

type ERPLayoutProps = {
  children: ReactNode;
};

export default async function ERPLayout({ children }: ERPLayoutProps) {
  const session = await getMockSession();

  if (!session.isAuthenticated) {
    redirect("/login");
  }

  return <AppShell user={session.user}>{children}</AppShell>;
}
