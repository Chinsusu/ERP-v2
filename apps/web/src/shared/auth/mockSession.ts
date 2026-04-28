import { cookies } from "next/headers";
import { getPermissionsForRole, type PermissionKey, type RoleKey } from "@/shared/permissions/menu";
import { localAuthPolicy, validateLocalCredentials } from "./sessionPolicy";

const mockAuthCookieName = "erp_mock_session";

export type MockUser = {
  id: string;
  name: string;
  email: string;
  role: RoleKey;
  permissions: PermissionKey[];
};

export type MockSession =
  | {
      isAuthenticated: true;
      accessToken: string;
      expiresAt: string;
      user: MockUser;
    }
  | {
      isAuthenticated: false;
      user: null;
    };

export const mockUser: MockUser = {
  id: "user-erp-admin",
  name: "ERP Admin",
  email: "admin@example.local",
  role: "ERP_ADMIN",
  permissions: getPermissionsForRole("ERP_ADMIN")
};

export async function getMockSession(): Promise<MockSession> {
  const cookieStore = await cookies();
  const cookieSession = cookieStore.get(mockAuthCookieName)?.value;
  const forcedSignedOut = process.env.NEXT_PUBLIC_MOCK_AUTH_STATE === "signed-out";

  if (forcedSignedOut && !cookieSession) {
    return { isAuthenticated: false, user: null };
  }

  const payload = parseSessionCookie(cookieSession);
  if (!payload || Date.parse(payload.expiresAt) <= Date.now()) {
    return { isAuthenticated: false, user: null };
  }

  return {
    isAuthenticated: true,
    accessToken: payload.accessToken,
    expiresAt: payload.expiresAt,
    user: mockUser
  };
}

export async function signInMockUser(email: string, password: string) {
  const validation = validateLocalCredentials(email, password);
  if (!validation.ok) {
    return validation;
  }

  const expiresAt = new Date(Date.now() + localAuthPolicy.accessTokenMaxAgeSeconds * 1000).toISOString();
  const accessToken = process.env.NEXT_PUBLIC_MOCK_ACCESS_TOKEN || "local-dev-access-token";
  const cookieStore = await cookies();
  cookieStore.set(mockAuthCookieName, encodeURIComponent(JSON.stringify({ accessToken, expiresAt })), {
    httpOnly: true,
    maxAge: localAuthPolicy.accessTokenMaxAgeSeconds,
    path: "/",
    sameSite: "lax",
    secure: process.env.AUTH_COOKIE_SECURE === "true"
  });

  return { ok: true } as const;
}

function parseSessionCookie(value: string | undefined): { accessToken: string; expiresAt: string } | null {
  if (!value) {
    return null;
  }

  try {
    const parsed = JSON.parse(decodeURIComponent(value)) as Partial<{ accessToken: string; expiresAt: string }>;
    if (!parsed.accessToken || !parsed.expiresAt) {
      return null;
    }
    return { accessToken: parsed.accessToken, expiresAt: parsed.expiresAt };
  } catch {
    return null;
  }
}
