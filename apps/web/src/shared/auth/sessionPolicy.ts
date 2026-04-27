export const localAuthPolicy = {
  passwordMinLength: 10,
  passwordRequiresLetter: true,
  passwordRequiresNumberOrSymbol: true,
  commonPasswordsBlocked: true,
  maxFailedAttempts: 5,
  lockoutWindowSeconds: 15 * 60,
  lockoutDurationSeconds: 15 * 60,
  accessTokenMaxAgeSeconds: 8 * 60 * 60
} as const;

const localEmail = "admin@example.local";
const localPassword = "local-only-mock-password";
const commonPasswords = new Set(["password", "password1", "password123", "1234567890", "admin123456"]);

export type LocalCredentialResult =
  | { ok: true }
  | { ok: false; reason: "invalid_credentials" | "password_policy"; message: string };

export function validateLocalPasswordPolicy(password: string): string | null {
  if (password.length < localAuthPolicy.passwordMinLength) {
    return "Password does not meet the minimum length policy.";
  }
  if (commonPasswords.has(password.trim().toLowerCase())) {
    return "Password is too common.";
  }

  const hasLetter = /\p{L}/u.test(password);
  const hasNumberOrSymbol = /[\p{N}\p{P}\p{S}]/u.test(password);

  if (localAuthPolicy.passwordRequiresLetter && !hasLetter) {
    return "Password must include at least one letter.";
  }
  if (localAuthPolicy.passwordRequiresNumberOrSymbol && !hasNumberOrSymbol) {
    return "Password must include at least one number or symbol.";
  }

  return null;
}

export function validateLocalCredentials(email: string, password: string): LocalCredentialResult {
  const policyError = validateLocalPasswordPolicy(password);
  if (policyError) {
    return { ok: false, reason: "password_policy", message: policyError };
  }
  if (email.trim().toLowerCase() !== localEmail || password !== localPassword) {
    return { ok: false, reason: "invalid_credentials", message: "Invalid email or password." };
  }

  return { ok: true };
}
