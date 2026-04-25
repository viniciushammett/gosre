import { type FormEvent, useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { getAccessToken } from "../api/auth";
import { useAuth } from "../hooks/useAuth";

const AUTH_BASE = (import.meta.env.VITE_AUTH_URL ?? "http://localhost:8081").replace(/\/$/, "");

export default function LoginPage() {
  const navigate = useNavigate();
  const { login, isLoginPending, loginError } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  if (getAccessToken()) {
    return <Navigate to="/" replace />;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    try {
      await login({ email, password });
      navigate("/");
    } catch {
      // loginError reflects the failure
    }
  }

  return (
    <div
      className="min-h-screen bg-surface flex items-center justify-center px-4"
      style={{
        backgroundImage: "radial-gradient(circle, #1e2233 1px, transparent 1px)",
        backgroundSize: "28px 28px",
      }}
    >
      <div className="w-full max-w-sm">
        {/* Brand mark */}
        <div className="mb-8 text-center select-none">
          <p className="font-mono text-2xl tracking-[0.2em] text-white uppercase">
            GoSRE<span className="text-brand animate-pulse">_</span>
          </p>
          <p className="text-gray-600 text-[10px] font-mono tracking-[0.3em] uppercase mt-1">
            SRE Control Plane
          </p>
        </div>

        {/* Card */}
        <div
          className="bg-surface-raised rounded-lg shadow-2xl overflow-hidden"
          style={{ border: "1px solid #1e2233", borderTop: "2px solid #4f8ef7" }}
        >
          <div className="px-8 py-7">
            <h1 className="text-sm font-medium text-gray-300 mb-6">
              Sign in to your account
            </h1>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-[10px] font-mono text-gray-500 tracking-widest uppercase mb-1.5">
                  Email
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  autoComplete="email"
                  placeholder="you@company.com"
                  className="w-full bg-surface border border-surface-border rounded px-3 py-2.5 text-sm text-gray-200 placeholder-gray-700 focus:outline-none focus:border-brand transition-colors"
                />
              </div>

              <div>
                <label className="block text-[10px] font-mono text-gray-500 tracking-widest uppercase mb-1.5">
                  Password
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  autoComplete="current-password"
                  className="w-full bg-surface border border-surface-border rounded px-3 py-2.5 text-sm text-gray-200 focus:outline-none focus:border-brand transition-colors"
                />
              </div>

              {loginError && (
                <p className="text-status-fail text-xs font-mono">
                  {loginError.message}
                </p>
              )}

              <button
                type="submit"
                disabled={isLoginPending}
                className="w-full bg-brand hover:bg-brand-dim text-white text-sm font-medium py-2.5 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoginPending ? "Signing in…" : "Sign in"}
              </button>
            </form>

            <div className="flex items-center gap-3 my-5">
              <div className="flex-1 border-t border-surface-border" />
              <span className="text-[10px] text-gray-600 font-mono tracking-widest">OR</span>
              <div className="flex-1 border-t border-surface-border" />
            </div>

            <a
              href={`${AUTH_BASE}/auth/github/login`}
              className="flex items-center justify-center gap-2.5 w-full border border-surface-border rounded py-2.5 text-sm text-gray-400 hover:border-gray-600 hover:text-gray-200 transition-colors"
            >
              <GitHubIcon />
              Continue with GitHub
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}

function GitHubIcon() {
  return (
    <svg className="w-4 h-4 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
    </svg>
  );
}
