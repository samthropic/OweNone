"use client";

import Link from "next/link";
import { useActionState } from "react";
import { useFormStatus } from "react-dom";
import { logInAction } from "@/app/auth-actions";
import { initialActionState } from "@/lib/action-state";

export function LoginForm() {
  const [state, action] = useActionState(logInAction, initialActionState);

  return (
    <form action={action} className="auth-form" noValidate>
      <div className="auth-field">
        <label htmlFor="login-email" className="auth-label">
          Email
        </label>
        <input
          id="login-email"
          className="auth-input"
          name="email"
          type="email"
          placeholder="you@example.com"
          aria-label="Email address"
          autoComplete="email"
          required
        />
      </div>

      <div className="auth-field">
        <label htmlFor="login-password" className="auth-label">
          Password
        </label>
        <input
          id="login-password"
          className="auth-input"
          name="password"
          type="password"
          placeholder="••••••••"
          aria-label="Password"
          autoComplete="current-password"
          required
          minLength={8}
        />
      </div>

      <p
        className={`auth-status ${state.status !== "idle" ? `is-${state.status}` : ""}`}
        role="status"
        aria-live="polite"
      >
        {state.status !== "idle" ? state.message : ""}
      </p>

      <LoginButton />

      <hr className="auth-divider" />

      <p className="auth-demo-hint">
        Demo account: <code>sarah@example.com</code> / <code>password123</code>
      </p>

      <p className="auth-footer">
        New to OweNone?{" "}
        <Link href="/signup">Create an account</Link>
      </p>
    </form>
  );
}

function LoginButton() {
  const { pending } = useFormStatus();
  return (
    <button
      type="submit"
      className="auth-btn"
      disabled={pending}
      aria-label="Log in to your account"
    >
      {pending ? "Logging in…" : "Log in"}
    </button>
  );
}
