const localStaticAccessToken = "local-dev-access-token";

let clientAccessToken: string | undefined;

export function rememberClientAccessToken(accessToken: string | undefined) {
  clientAccessToken = normalizeToken(accessToken);
}

export function clearClientAccessToken() {
  clientAccessToken = undefined;
}

export function resolveApiAccessToken(requestedAccessToken?: string) {
  const requested = normalizeToken(requestedAccessToken);

  if (requested && requested !== localStaticAccessToken) {
    return requested;
  }

  if (clientAccessToken) {
    return clientAccessToken;
  }

  if (isProductionLikeWebRuntime() && requested === localStaticAccessToken) {
    return undefined;
  }

  return requested;
}

export function isProductionLikeWebRuntime() {
  const appEnv = process.env.NEXT_PUBLIC_APP_ENV ?? process.env.APP_ENV ?? "";
  const normalized = appEnv.trim().toLowerCase();

  if (normalized === "local" || normalized === "dev" || normalized === "development" || normalized === "test") {
    return false;
  }

  return process.env.NODE_ENV === "production" || normalized === "staging" || normalized === "prod" || normalized === "production";
}

function normalizeToken(value: string | undefined) {
  const trimmed = value?.trim();
  return trimmed || undefined;
}
