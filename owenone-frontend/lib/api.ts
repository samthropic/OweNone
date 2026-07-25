import type { APIError, APIResponse, Dashboard } from "@/lib/api-types";

const DEFAULT_API_URL = "http://localhost:8080";
export const DEMO_USER_ID = "10000000-0000-0000-0000-000000000001";

function apiURL(path: string) {
  return new URL(path, process.env.OWENONE_API_URL ?? DEFAULT_API_URL);
}

export function currentUserID() {
  return process.env.OWENONE_USER_ID ?? DEMO_USER_ID;
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

  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as APIError;
    throw new Error(body.error?.message ?? `OweNone API returned ${response.status}`);
  }

  return ((await response.json()) as APIResponse<T>).data;
}

export function getDashboard() {
  return apiRequest<Dashboard>("/api/v1/dashboard", {
    headers: { "X-User-ID": currentUserID() },
  });
}

export function authenticatedPost<T>(
  path: string,
  body: unknown,
  idempotencyKey?: string,
) {
  return apiRequest<T>(path, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-User-ID": currentUserID(),
      ...(idempotencyKey ? { "Idempotency-Key": idempotencyKey } : {}),
    },
    body: JSON.stringify(body),
  });
}