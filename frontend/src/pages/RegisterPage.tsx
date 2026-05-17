import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../lib/api";

export function RegisterPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState({ email: "", password: "" });
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError("");

    try {
      const response = await api.signUp(form);
      sessionStorage.setItem("pendingVerificationId", response.verification_id);
      navigate("/verify-email", { state: { verificationId: response.verification_id } });
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Failed to sign up.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="stack-page">
      <section className="panel-shell auth-page-shell">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Register</p>
            <h1>Create your account</h1>
          </div>
          <Link className="text-link" to="/login">
            Already have an account?
          </Link>
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
              minLength={6}
              value={form.password}
              onChange={(event) => setForm((current) => ({ ...current, password: event.target.value }))}
              required
            />
          </label>
          {error ? <p className="form-error">{error}</p> : null}
          <button className="button" disabled={submitting} type="submit">
            {submitting ? "Creating account..." : "Create account"}
          </button>
        </form>
      </section>
    </div>
  );
}
