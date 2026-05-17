import { useMemo, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { api } from "../lib/api";

export function VerifyEmailPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const storedId = sessionStorage.getItem("pendingVerificationId") ?? "";
  const verificationIdFromState = (location.state as { verificationId?: string } | undefined)?.verificationId ?? "";
  const initialVerificationId = useMemo(() => verificationIdFromState || storedId, [storedId, verificationIdFromState]);
  const hasStoredReference = Boolean(initialVerificationId);

  const [form, setForm] = useState({ verification_id: initialVerificationId, code: "" });
  const [feedback, setFeedback] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setFeedback("");
    setError("");

    try {
      await api.verifyEmail(form);
      sessionStorage.removeItem("pendingVerificationId");
      setFeedback("Email verified. Log in to start using the platform.");
      navigate("/login");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Failed to verify email.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="stack-page">
      <section className="panel-shell auth-page-shell">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Verify Email</p>
            <h1>Confirm your account</h1>
          </div>
          <Link className="text-link" to="/register">
            Back to register
          </Link>
        </div>

        <form className="stacked-form auth-form-shell" onSubmit={handleSubmit}>
          {hasStoredReference ? <p className="helper-text success">Verification reference ready.</p> : null}
          {!hasStoredReference ? (
            <label>
              <span>Verification reference</span>
              <input
                value={form.verification_id}
                onChange={(event) => setForm((current) => ({ ...current, verification_id: event.target.value }))}
                required
              />
            </label>
          ) : null}
          <label>
            <span>Code</span>
            <input
              value={form.code}
              onChange={(event) => setForm((current) => ({ ...current, code: event.target.value }))}
              required
            />
          </label>
          {feedback ? <p className="helper-text success">{feedback}</p> : null}
          {error ? <p className="form-error">{error}</p> : null}
          <button className="button secondary" disabled={submitting} type="submit">
            {submitting ? "Verifying..." : "Verify email"}
          </button>
        </form>
      </section>
    </div>
  );
}
