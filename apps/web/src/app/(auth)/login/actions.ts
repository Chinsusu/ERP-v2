"use server";

import { redirect } from "next/navigation";
import { signInMockUser } from "@/shared/auth/mockSession";

export async function signInAction() {
  await signInMockUser();
  redirect("/dashboard");
}
