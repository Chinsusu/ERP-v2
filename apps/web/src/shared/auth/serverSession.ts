import { cookies } from "next/headers";
import { ApiError } from "../api/client";
import { isProductionLikeWebRuntime } from "./clientSessionToken";
import { loginErrorReasonFromUnknown, type LoginErrorReason } from "./authErrors";
import {
  fetchCurrentUser,
  loginWithBackend,
  logoutBackendSession,
  refreshBackendSession,
  type AuthLoginResponse
} from "./backendAuth";
import type { BackendAuthSession } from "./session";

const accessTokenCookieName = "erp_access_token";
const refreshTokenCookieName = "erp_refresh_token";
const accessExpiresAtCookieName = "erp_access_expires_at";

export async function getBackendSession(): Promise<BackendAuthSession> {
  assertMockAuthNotForcedInProduction();

  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessTokenCookieName)?.value;
  const refreshToken = cookieStore.get(refreshTokenCookieName)?.value;
  const expiresAt = cookieStore.get(accessExpiresAtCookieName)?.value;

  if (!accessToken) {
    return unauthenticatedSession();
  }

  try {
    const user = await fetchCurrentUser(accessToken);
    return {
      isAuthenticated: true,
      accessToken,
      refreshToken: refreshToken ?? "",
      expiresAt: expiresAt ?? "",
      user
    };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return unauthenticatedSession();
    }
    if (isProductionLikeWebRuntime()) {
      throw error;
    }
    return unauthenticatedSession();
  }
}

export async function signInBackendSession(
  email: string,
  password: string
): Promise<{ ok: true } | { ok: false; reason: LoginErrorReason }> {
  try {
    await setBackendSessionCookies(await loginWithBackend(email, password));
    return { ok: true };
  } catch (error) {
    return { ok: false, reason: loginErrorReasonFromUnknown(error) };
  }
}

export async function logoutCurrentBackendSession() {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(refreshTokenCookieName)?.value;
  if (refreshToken) {
    try {
      await logoutBackendSession(refreshToken);
    } catch {
      // Local cookies are still cleared so stale privileged UI cannot survive logout.
    }
  }
  await clearBackendSessionCookies();
}

export async function refreshCurrentBackendSession() {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(refreshTokenCookieName)?.value;
  if (!refreshToken) {
    return { ok: false } as const;
  }

  try {
    const refreshed = await refreshBackendSession(refreshToken);
    await setBackendSessionCookies(refreshed);
    return {
      ok: true,
      accessToken: refreshed.access_token,
      expiresAt: refreshed.expires_at
    } as const;
  } catch {
    await clearBackendSessionCookies();
    return { ok: false } as const;
  }
}

async function setBackendSessionCookies(session: AuthLoginResponse) {
  const cookieOptions = {
    httpOnly: true,
    path: "/",
    sameSite: "lax" as const,
    secure: process.env.AUTH_COOKIE_SECURE === "true" || isProductionLikeWebRuntime()
  };
  const cookieStore = await cookies();

  cookieStore.set(accessTokenCookieName, session.access_token, {
    ...cookieOptions,
    maxAge: session.expires_in
  });
  cookieStore.set(refreshTokenCookieName, session.refresh_token, {
    ...cookieOptions,
    maxAge: session.refresh_expires_in
  });
  cookieStore.set(accessExpiresAtCookieName, session.expires_at, {
    ...cookieOptions,
    maxAge: session.expires_in
  });
}

async function clearBackendSessionCookies() {
  const cookieStore = await cookies();
  cookieStore.delete(accessTokenCookieName);
  cookieStore.delete(refreshTokenCookieName);
  cookieStore.delete(accessExpiresAtCookieName);
}

function unauthenticatedSession(): BackendAuthSession {
  return { isAuthenticated: false, user: null };
}

function assertMockAuthNotForcedInProduction() {
  const mockState = process.env.NEXT_PUBLIC_MOCK_AUTH_STATE;
  if (isProductionLikeWebRuntime() && mockState && mockState !== "signed-out") {
    throw new Error("Mock auth state is not allowed in production-like runtime.");
  }
}
