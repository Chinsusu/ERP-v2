import type { components } from "../api/generated/schema";

export type AuthenticatedUser = components["schemas"]["UserProfile"];

export type BackendAuthSession =
  | {
      isAuthenticated: true;
      accessToken: string;
      refreshToken: string;
      expiresAt: string;
      user: AuthenticatedUser;
    }
  | {
      isAuthenticated: false;
      user: null;
    };
