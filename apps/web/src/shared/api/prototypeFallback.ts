import { ApiError } from "./client";

export function shouldUsePrototypeFallback(reason: unknown): boolean {
  if (!isPrototypeFallbackRuntimeEnabled()) {
    return false;
  }
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

export function isPrototypeFallbackRuntimeEnabled(): boolean {
  return process.env.NODE_ENV !== "production";
}
