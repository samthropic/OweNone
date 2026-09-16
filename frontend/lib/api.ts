import type { ActivityFeed, APIError, APIResponse, AuthToken, AuthUser, Dashboard } from "@/lib/api-types";
import { getSessionToken } from "@/lib/session";

const DEFAULT_API_URL = "http://localhost:8080";

export class UnauthorizedError extends Error {
  constructor(message = "Session expired. Please log in again.") {
    super(message);
    this.name = "UnauthorizedError";
  }
}

function defaultAPIURL() {
  if (process.env.OWENONE_API_URL) return process.env.OWENONE_API_URL;
  // Same-deployment Vercel URL so /api/* hits the Go service via rewrites.
  if (process.env.VERCEL_URL) return `https://${process.env.VERCEL_URL}`;
  return DEFAULT_API_URL;
}

function apiURL(path: string) {
  return new URL(path, defaultAPIURL());
}

export async function apiRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(apiURL(path), {
    ...init,
    cache: "no-store",
    headers: {
      Accept: "application/json",
      ...init.headers,
    },
  });

  if (response.status === 401) {
    const body = (await response.json().catch(() => ({}))) as APIError;
    const message = body.error?.message ?? "Session expired. Please log in again.";
    throw new UnauthorizedError(message);
  }

  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as APIError;
    throw new Error(body.error?.message ?? `OweNone API returned ${response.status}`);
  }

  return ((await response.json()) as APIResponse<T>).data;
}

async function bearerHeaders(): Promise<Record<string, string>> {
  const token = await getSessionToken();
  if (!token) throw new UnauthorizedError();
  return { Authorization: `Bearer ${token}` };
}

export async function getDashboard() {
  return apiRequest<Dashboard>("/api/v1/dashboard", {
    headers: await bearerHeaders(),
  });
}

export async function getActivity(params: { limit?: number; offset?: number; kind?: string; groupId?: string } = {}) {
  const { limit = 50, offset = 0, kind, groupId } = params;
  const qs = new URLSearchParams({ limit: String(limit), offset: String(offset) });
  if (kind) qs.set("kind", kind);
  if (groupId) qs.set("groupId", groupId);
  return apiRequest<ActivityFeed>(`/api/v1/activity?${qs}`, {
    headers: await bearerHeaders(),
  });
}

export async function authenticatedPost<T>(
  path: string,
  body: unknown,
  idempotencyKey?: string,
) {
  return apiRequest<T>(path, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(await bearerHeaders()),
      ...(idempotencyKey ? { "Idempotency-Key": idempotencyKey } : {}),
    },
    body: JSON.stringify(body),
  });
}

export async function authenticatedUpload<T>(path: string, formData: FormData) {
  return apiRequest<T>(path, {
    method: "POST",
    headers: await bearerHeaders(),
    body: formData,
  });
}

export async function authenticatedDelete<T>(path: string) {
  return apiRequest<T>(path, {
    method: "DELETE",
    headers: await bearerHeaders(),
  });
}

// Auth helpers

export function signUp(params: {
  email: string;
  displayName: string;
  password: string;
}) {
  return apiRequest<AuthToken>("/api/v1/auth/signup", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(params),
  });
}

export function logIn(params: { email: string; password: string }) {
  return apiRequest<AuthToken>("/api/v1/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(params),
  });
}

export async function logOut() {
  try {
    const token = await getSessionToken();
    if (!token) return;
    await apiRequest<{ loggedOut: boolean }>("/api/v1/auth/logout", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch {
    // best-effort — intentionally swallow all errors including network failures
  }
}

export async function getSessionUser() {
  return apiRequest<{ user: AuthToken["user"] }>("/api/v1/auth/session", {
    headers: await bearerHeaders(),
  });
}

// Landing is a public page, so a missing, expired or unreachable session must
// degrade to "signed out" rather than breaking the render.
export async function getOptionalSessionUser(): Promise<AuthUser | null> {
  try {
    const { user } = await getSessionUser();
    return user;
  } catch {
    return null;
  }
}