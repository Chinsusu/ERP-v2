import { apiGet, apiPost } from "../api/client";
import type { components } from "../api/generated/schema";

export type AuthLoginResponse = components["schemas"]["AuthLoginResponse"];
export type AuthLogoutResponse = components["schemas"]["AuthLogoutResponse"];
export type AuthenticatedUser = components["schemas"]["UserProfile"];

export function loginWithBackend(email: string, password: string) {
  return apiPost<AuthLoginResponse, components["schemas"]["AuthLoginRequest"]>("/auth/login", {
    email,
    password
  });
}

export function refreshBackendSession(refreshToken: string) {
  return apiPost<AuthLoginResponse, components["schemas"]["AuthRefreshRequest"]>("/auth/refresh", {
    refresh_token: refreshToken
  });
}

export function logoutBackendSession(refreshToken: string) {
  return apiPost<AuthLogoutResponse, components["schemas"]["AuthLogoutRequest"]>("/auth/logout", {
    refresh_token: refreshToken
  });
}

export function fetchCurrentUser(accessToken: string) {
  return apiGet("/me", { accessToken });
}
