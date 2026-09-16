"use client";

import { useFormStatus } from "react-dom";
import { logOutAction } from "@/app/auth-actions";

function LogoutButton() {
  const { pending } = useFormStatus();
  return (
    <button
      type="submit"
      className="home-btn home-btn-small home-logout-btn"
      disabled={pending}
      aria-label="Log out of OweNone"
    >
      {pending ? "Logging out…" : "Log out"}
    </button>
  );
}

export function LogoutForm() {
  return (
    <form action={logOutAction}>
      <LogoutButton />
    </form>
  );
}
