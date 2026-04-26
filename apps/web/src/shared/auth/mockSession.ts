import { cookies } from "next/headers";
import type { PermissionKey, RoleKey } from "@/shared/permissions/menu";

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
  permissions: [
    "dashboard:view",
    "warehouse:view",
    "inventory:view",
    "purchase:view",
    "qc:view",
    "production:view",
    "sales:view",
    "shipping:view",
    "returns:view",
    "master-data:view",
    "approvals:view",
    "audit-log:view",
    "reports:view",
    "settings:view"
  ]
};

export async function getMockSession(): Promise<MockSession> {
  const cookieStore = await cookies();
  const cookieSession = cookieStore.get(mockAuthCookieName)?.value;
  const forcedSignedOut = process.env.NEXT_PUBLIC_MOCK_AUTH_STATE === "signed-out";

  if (forcedSignedOut && cookieSession !== "authenticated") {
    return { isAuthenticated: false, user: null };
  }

  return { isAuthenticated: true, user: mockUser };
}

export async function signInMockUser() {
  const cookieStore = await cookies();
  cookieStore.set(mockAuthCookieName, "authenticated", {
    httpOnly: true,
    maxAge: 60 * 60 * 8,
    path: "/",
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production"
  });
}
