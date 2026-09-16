import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { redirect, unstable_rethrow } from "next/navigation";
import "@/components/auth/auth.css";
import { LoginForm } from "@/components/auth/LoginForm";
import { getSessionUser, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

export const metadata: Metadata = {
  title: "Log in — OweNone",
  description: "Sign in to OweNone to manage your group expenses and settlements.",
};

export default async function LoginPage() {
  const token = await getSessionToken();
  let sessionValid = false;

  if (token) {
    try {
      await getSessionUser();
      sessionValid = true;
    } catch (err) {
      unstable_rethrow(err); // re-throw any Next.js control-flow errors (redirect, notFound…)
      // UnauthorizedError or network errors → stale/revoked token, render the form
      if (!(err instanceof UnauthorizedError)) throw err;
    }
  }

  if (sessionValid) redirect("/");

  return (
    <main className="auth-root">
      <div className="auth-card">
        <Link href="/landing" className="auth-brand">
          <Image
            src="/owenone-logo.png"
            alt="OweNone"
            width={390}
            height={336}
            className="auth-brand-logo"
            preload={true}
          />
        </Link>

        <h1 className="auth-heading">Welcome back</h1>
        <p className="auth-subheading">Log in to your account to continue.</p>

        <LoginForm />
      </div>
    </main>
  );
}
