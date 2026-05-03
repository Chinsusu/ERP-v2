import { signInAction } from "./actions";
import { isProductionLikeWebRuntime } from "@/shared/auth/clientSessionToken";
import { t } from "@/shared/i18n";

type LoginPageProps = {
  searchParams?: Promise<{ error?: string }>;
};

const errorCopy: Record<string, string> = {
  invalid_credentials: authCopy("errors.invalidCredentials"),
  password_policy: authCopy("errors.passwordPolicy"),
  locked: authCopy("errors.locked"),
  session_expired: authCopy("errors.sessionExpired"),
  auth_unavailable: authCopy("errors.authUnavailable")
};

const devDefaultEmail = isProductionLikeWebRuntime() ? undefined : "admin@example.local";

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const params = await searchParams;
  const error = params?.error ? errorCopy[params.error] : null;

  return (
    <main className="erp-page erp-page--centered">
      <form action={signInAction} className="erp-card erp-form-card">
        <h1 className="erp-form-title">{authCopy("loginTitle")}</h1>
        {error ? <p className="erp-form-error">{error}</p> : null}
        <label className="erp-field">
          {authCopy("email")}
          <input
            className="erp-input"
            name="email"
            type="email"
            defaultValue={devDefaultEmail}
          />
        </label>
        <label className="erp-field">
          {authCopy("password")}
          <input className="erp-input" name="password" type="password" />
        </label>
        <button className="erp-button erp-button--primary erp-button--full" type="submit">
          {authCopy("signIn")}
        </button>
      </form>
    </main>
  );
}

function authCopy(key: string) {
  return t(`auth.${key}`);
}
