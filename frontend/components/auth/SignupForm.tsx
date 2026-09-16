"use client";

import Link from "next/link";
import { useActionState } from "react";
import { useFormStatus } from "react-dom";
import { signUpAction } from "@/app/auth-actions";
import { initialActionState } from "@/lib/action-state";

export function SignupForm() {
  const [state, action] = useActionState(signUpAction, initialActionState);

  return (
    <form action={action} className="auth-form" noValidate>
      <div className="auth-field">
        <label htmlFor="signup-name" className="auth-label">
          Display name
        </label>
        <input
          id="signup-name"
          className="auth-input"
          name="displayName"
          type="text"
          placeholder="Sarah Chen"
          aria-label="Display name"
          autoComplete="name"
          required
          maxLength={100}
        />
      </div>

      <div className="auth-field">
        <label htmlFor="signup-email" className="auth-label">
          Email
        </label>
        <input
          id="signup-email"
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
        <label htmlFor="signup-password" className="auth-label">
          Password
        </label>
        <input
          id="signup-password"
          className="auth-input"
          name="password"
          type="password"
          placeholder="At least 8 characters"
          aria-label="Password"
          autoComplete="new-password"
          required
          minLength={8}
          maxLength={128}
        />
      </div>

      <div className="auth-field">
        <label htmlFor="signup-confirm" className="auth-label">
          Confirm password
        </label>
        <input
          id="signup-confirm"
          className="auth-input"
          name="confirmPassword"
          type="password"
          placeholder="Repeat your password"
          aria-label="Confirm password"
          autoComplete="new-password"
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

      <SignupButton />

      <p className="auth-footer">
        Already have an account?{" "}
        <Link href="/login">Log in</Link>
      </p>
    </form>
  );
}

function SignupButton() {
  const { pending } = useFormStatus();
  return (
    <button
      type="submit"
      className="auth-btn"
      disabled={pending}
      aria-label="Create your OweNone account"
    >
      {pending ? "Creating account…" : "Create account"}
    </button>
  );
}
