"use client";

import { useActionState } from "react";
import { useFormStatus } from "react-dom";
import { initialActionState, joinWaitlistAction } from "@/app/actions";

export function WaitlistForm() {
  const [state, action] = useActionState(joinWaitlistAction, initialActionState);
  return (
    <form action={action} className="input-row">
      <input
        className="input"
        name="email"
        type="email"
        placeholder="Enter your email"
        aria-label="Email address"
        autoComplete="email"
        required
      />
      <WaitlistButton />
      {state.status !== "idle" ? (
        <p className={`waitlist-status is-${state.status}`} role="status">{state.message}</p>
      ) : null}
    </form>
  );
}

function WaitlistButton() {
  const { pending } = useFormStatus();
  return (
    <button type="submit" className="btn btn-dark waitlist-btn" disabled={pending}>
      {pending ? "Joining..." : "Join waitlist"}
    </button>
  );
}