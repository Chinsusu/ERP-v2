"use server";

import { redirect } from "next/navigation";
import { logoutCurrentBackendSession } from "@/shared/auth/serverSession";

export async function signOutAction() {
  await logoutCurrentBackendSession();
  redirect("/login");
}
