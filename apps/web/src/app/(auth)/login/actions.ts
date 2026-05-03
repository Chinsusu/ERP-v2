"use server";

import { redirect } from "next/navigation";
import { signInBackendSession } from "@/shared/auth/serverSession";

export async function signInAction(formData: FormData) {
  const email = String(formData.get("email") || "");
  const password = String(formData.get("password") || "");
  const result = await signInBackendSession(email, password);
  if (!result.ok) {
    redirect(`/login?error=${result.reason}`);
  }

  redirect("/dashboard");
}
