import { signInAction } from "./actions";

type LoginPageProps = {
  searchParams?: Promise<{ error?: string }>;
};

const errorCopy: Record<string, string> = {
  invalid_credentials: "Email or password is invalid.",
  password_policy: "Password does not meet the active policy."
};

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const params = await searchParams;
  const error = params?.error ? errorCopy[params.error] : null;

  return (
    <main className="erp-page erp-page--centered">
      <form action={signInAction} className="erp-card erp-form-card">
        <h1 className="erp-form-title">ERP Login</h1>
        {error ? <p className="erp-form-error">{error}</p> : null}
        <label className="erp-field">
          Email
          <input
            className="erp-input"
            name="email"
            type="email"
            defaultValue="admin@example.local"
          />
        </label>
        <label className="erp-field">
          Password
          <input className="erp-input" name="password" type="password" />
        </label>
        <button className="erp-button erp-button--primary erp-button--full" type="submit">
          Sign in
        </button>
      </form>
    </main>
  );
}
