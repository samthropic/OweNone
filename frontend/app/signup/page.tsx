import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { redirect, unstable_rethrow } from "next/navigation";
import "@/components/auth/auth.css";
import { SignupForm } from "@/components/auth/SignupForm";
import { getSessionUser, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

export const metadata: Metadata = {
  title: "Create account — OweNone",
  description: "Join OweNone and start untangling your group expenses.",
};

export default async function SignupPage() {
  const token = await getSessionToken();
  let sessionValid = false;

  if (token) {
    try {
      await getSessionUser();
      sessionValid = true;
    } catch (err) {
      unstable_rethrow(err); // re-throw any Next.js control-flow errors (redirect, notFound…)
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

        <h1 className="auth-heading">Create your account</h1>
        <p className="auth-subheading">
          Start settling debts the smart way.
        </p>

        <SignupForm />
      </div>
    </main>
  );
}
