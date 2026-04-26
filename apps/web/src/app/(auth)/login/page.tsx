export default function LoginPage() {
  return (
    <main className="erp-page erp-page--centered">
      <form className="erp-card erp-form-card">
        <h1 className="erp-form-title">ERP Login</h1>
        <label className="erp-field">
          Email
          <input className="erp-input" name="email" type="email" />
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
