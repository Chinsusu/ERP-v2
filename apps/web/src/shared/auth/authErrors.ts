import { ApiError } from "@/shared/api/client";

export type LoginErrorReason = "invalid_credentials" | "password_policy" | "locked" | "auth_unavailable";

export function loginErrorReasonFromUnknown(reason: unknown): LoginErrorReason {
  if (reason instanceof ApiError) {
    const backendReason = typeof reason.details?.reason === "string" ? reason.details.reason : "";
    if (backendReason === "password_policy") {
      return "password_policy";
    }
    if (backendReason === "locked") {
      return "locked";
    }
    if (backendReason === "invalid_credentials") {
      return "invalid_credentials";
    }
    if (reason.status === 401) {
      return "invalid_credentials";
    }
  }

  return "auth_unavailable";
}
