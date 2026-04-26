export default function LoginPage() {
  return (
    <main style={{ maxWidth: 360, margin: "80px auto", padding: 24 }}>
      <h1>ERP Login</h1>
      <form>
        <label>
          Email
          <input name="email" type="email" style={{ display: "block", width: "100%", marginTop: 8 }} />
        </label>
        <label style={{ display: "block", marginTop: 16 }}>
          Password
          <input name="password" type="password" style={{ display: "block", width: "100%", marginTop: 8 }} />
        </label>
        <button type="submit" style={{ marginTop: 20 }}>
          Sign in
        </button>
      </form>
    </main>
  );
}
