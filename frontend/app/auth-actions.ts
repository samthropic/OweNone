"use server";

import { redirect } from "next/navigation";
import { logIn, logOut, signUp } from "@/lib/api";
import { clearSessionCookie, setSessionCookie } from "@/lib/session";
import type { ActionState } from "@/lib/action-state";

export async function signUpAction(
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const email = (formData.get("email") as string | null)?.trim() ?? "";
  const displayName = (formData.get("displayName") as string | null)?.trim() ?? "";
  const password = (formData.get("password") as string | null) ?? "";
  const confirmPassword = (formData.get("confirmPassword") as string | null) ?? "";

  if (!email) return { status: "error", message: "Email is required." };
  if (!displayName) return { status: "error", message: "Display name is required." };
  if (password.length < 8)
    return { status: "error", message: "Password must be at least 8 characters." };
  if (password !== confirmPassword)
    return { status: "error", message: "Passwords do not match." };

  let token: string;
  let expiresAt: string;

  try {
    const result = await signUp({ email, displayName, password });
    token = result.token;
    expiresAt = result.expiresAt;
  } catch (err) {
    return {
      status: "error",
      message:
        err instanceof Error
          ? err.message
          : "Could not create account. Please try again.",
    };
  }

  await setSessionCookie(token, expiresAt);
  redirect("/");
}

export async function logInAction(
  _prev: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const email = (formData.get("email") as string | null)?.trim() ?? "";
  const password = (formData.get("password") as string | null) ?? "";

  if (!email) return { status: "error", message: "Email is required." };
  if (password.length < 8)
    return { status: "error", message: "Password must be at least 8 characters." };

  let token: string;
  let expiresAt: string;

  try {
    const result = await logIn({ email, password });
    token = result.token;
    expiresAt = result.expiresAt;
  } catch (err) {
    return {
      status: "error",
      message:
        err instanceof Error
          ? err.message
          : "Could not log in. Please try again.",
    };
  }

  await setSessionCookie(token, expiresAt);
  redirect("/");
}

export async function logOutAction(): Promise<void> {
  await logOut();
  await clearSessionCookie();
  redirect("/");
}
