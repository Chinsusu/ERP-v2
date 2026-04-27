"use server";

import { redirect } from "next/navigation";
import { signInMockUser } from "@/shared/auth/mockSession";

export async function signInAction(formData: FormData) {
  const email = String(formData.get("email") || "");
  const password = String(formData.get("password") || "");
  const result = await signInMockUser(email, password);
  if (!result.ok) {
    redirect(`/login?error=${result.reason}`);
  }

  redirect("/dashboard");
}
