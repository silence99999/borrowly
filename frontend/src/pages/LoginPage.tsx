import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { api } from "../lib/api";

export function LoginPage() {
  const { user, refresh } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const redirectTo = (location.state as { from?: string } | undefined)?.from ?? "/dashboard";

  const [form, setForm] = useState({ email: "", password: "" });
  const [feedback, setFeedback] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setFeedback("");
    setError("");

    try {
      await api.login(form);
      await refresh();
      navigate(redirectTo, { replace: true });
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Failed to login.");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleLogout() {
    await api.logout();
    await refresh();
    setFeedback("Logged out.");
  }

  return (
    <div className="stack-page">
      <section className="panel-shell auth-page-shell">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Login</p>
            <h1>{user ? "Manage your session" : "Sign in to your account"}</h1>
          </div>
          <div className="inline-actions">
            <Link className="text-link" to="/register">
              Create account
            </Link>
            <Link className="text-link" to="/verify-email">
              Verify email
            </Link>
          </div>
        </div>

        <form className="stacked-form auth-form-shell" onSubmit={handleSubmit}>
          <label>
            <span>Email</span>
            <input
              type="email"
              value={form.email}
              onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))}
              required
            />
          </label>
          <label>
            <span>Password</span>
            <input
              type="password"
              value={form.password}
              onChange={(event) => setForm((current) => ({ ...current, password: event.target.value }))}
              required
            />
          </label>
          {user ? <p className="helper-text success">You are signed in.</p> : null}
          {feedback ? <p className="helper-text success">{feedback}</p> : null}
          {error ? <p className="form-error">{error}</p> : null}
          <div className="inline-actions">
            <button className="button" disabled={submitting} type="submit">
              {submitting ? "Signing in..." : "Login"}
            </button>
            {user ? (
              <button className="button tertiary" onClick={() => void handleLogout()} type="button">
                Logout
              </button>
            ) : null}
          </div>
        </form>
      </section>
    </div>
  );
}
