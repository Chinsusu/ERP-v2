import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { getBackendSession } from "@/shared/auth/serverSession";
import { AppShell } from "@/shared/layouts/AppShell";
import { signOutAction } from "./actions";

type ERPLayoutProps = {
  children: ReactNode;
};

export default async function ERPLayout({ children }: ERPLayoutProps) {
  const session = await getBackendSession();

  if (!session.isAuthenticated) {
    redirect("/login");
  }

  return (
    <AppShell
      accessToken={session.accessToken}
      expiresAt={session.expiresAt}
      signOutAction={signOutAction}
      user={session.user}
    >
      {children}
    </AppShell>
  );
}
