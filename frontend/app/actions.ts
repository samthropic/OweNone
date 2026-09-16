"use server";

import { revalidatePath } from "next/cache";
import { authenticatedDelete, authenticatedPost, authenticatedUpload } from "@/lib/api";
import type { ActionState } from "@/lib/action-state";

export async function createGroupAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Group created.", async () => {
    await authenticatedPost<{ id: string }>("/api/v1/groups", {
      name: formData.get("name"),
      icon: formData.get("icon"),
      memberIds: formData.getAll("memberIds"),
    });
    revalidatePath("/groups");
    revalidatePath("/");
  });
}

export async function updateGroupAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Group updated.", async () => {
    await authenticatedPost<{ group: { id: string } }>("/api/v1/groups/update", {
      groupId: formData.get("groupId"),
      name: formData.get("name"),
      icon: formData.get("icon"),
      memberIds: formData.getAll("memberIds"),
    });
    revalidatePath("/groups");
    revalidatePath("/");
  });
}

export async function deleteGroupAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Group deleted.", async () => {
    await authenticatedPost<{ deleted: boolean }>("/api/v1/groups/delete", {
      groupId: formData.get("groupId"),
    });
    revalidatePath("/groups");
    revalidatePath("/");
  });
}

export async function addFriendAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Friend added.", async () => {
    await authenticatedPost("/api/v1/friends", { email: formData.get("email") });
    revalidatePath("/");
    revalidatePath("/friends");
  });
}

export async function removeFriendAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  try {
    const result = await authenticatedPost<{ removed: boolean; user: { displayName: string } }>(
      "/api/v1/friends/remove",
      { email: formData.get("email") },
    );
    revalidatePath("/friends");
    revalidatePath("/");
    return { status: "success", message: `Removed ${result.user.displayName}.` };
  } catch (error) {
    return {
      status: "error",
      message: error instanceof Error ? error.message : "The request could not be completed.",
    };
  }
}

export async function createExpenseAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Expense added and balances recomputed.", async () => {
    const amountMinor = parseMoney(formData.get("amount"));
    const splitMethod = String(formData.get("splitMethod") || "equal");
    const split =
      splitMethod === "exact"
        ? {
            method: "exact" as const,
            shares: parseExactShares(formData.get("shares"), amountMinor),
          }
        : {
            method: "equal" as const,
            participantIds: formData.getAll("participantIds"),
          };

    await authenticatedPost<{ id: string }>(
      "/api/v1/expenses",
      {
        groupId: formData.get("groupId"),
        description: formData.get("description"),
        category: formData.get("category"),
        amountMinor,
        currency: "USD",
        paidByUserId: formData.get("paidByUserId"),
        split,
      },
      crypto.randomUUID(),
    );
    revalidatePath("/");
    revalidatePath("/activity");
    revalidatePath("/groups");
  });
}

function parseExactShares(value: FormDataEntryValue | null, totalMinor: number) {
  if (typeof value !== "string" || !value.trim()) {
    throw new Error("Receipt shares are missing.");
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(value);
  } catch {
    throw new Error("Receipt shares are invalid.");
  }
  if (!Array.isArray(parsed) || parsed.length === 0) {
    throw new Error("Assign each item to at least one person.");
  }
  const shares = parsed.map((entry) => {
    if (!entry || typeof entry !== "object") {
      throw new Error("Receipt shares are invalid.");
    }
    const row = entry as { userId?: unknown; amountMinor?: unknown };
    const userId = typeof row.userId === "string" ? row.userId : "";
    const amountMinor = typeof row.amountMinor === "number" ? row.amountMinor : Number.NaN;
    if (!userId || !Number.isSafeInteger(amountMinor) || amountMinor <= 0) {
      throw new Error("Receipt shares are invalid.");
    }
    return { userId, amountMinor };
  });
  const sum = shares.reduce((total, share) => total + share.amountMinor, 0);
  if (sum !== totalMinor) {
    throw new Error("Item shares must add up to the receipt total.");
  }
  return shares;
}

export async function createSettlementAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Payment recorded and balances recomputed.", async () => {
    const groupID = formData.get("groupId");
    const paymentMethod = formData.get("paymentMethod");
    await authenticatedPost<{ id: string }>(
      "/api/v1/settlements",
      {
        groupId: groupID || null,
        toUserId: formData.get("toUserId"),
        amountMinor: parseMoney(formData.get("amount")),
        currency: "USD",
        paymentMethod: typeof paymentMethod === "string" ? paymentMethod : "",
      },
      crypto.randomUUID(),
    );
    revalidatePath("/", "layout");
    revalidatePath("/settle-up");
    revalidatePath("/friends");
    revalidatePath("/groups");
    revalidatePath("/activity");
  });
}

export async function sendReminderAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Reminder sent.", async () => {
    await authenticatedPost<{ id: string }>("/api/v1/reminders", {
      recipientId: formData.get("recipientId"),
      message: formData.get("message") ?? "",
    });
  });
}

export async function updateProfileAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Profile updated.", async () => {
    await authenticatedPost("/api/v1/profile", {
      displayName: formData.get("displayName"),
      email: formData.get("email"),
      preferredCurrency: "USD",
      paymentApps: {
        venmo: formData.get("venmo") ?? "",
        paypal: formData.get("paypal") ?? "",
        cashApp: formData.get("cashApp") ?? "",
        zelle: formData.get("zelle") ?? "",
      },
    });
    revalidatePath("/profile");
    revalidatePath("/");
  });
}

export async function uploadAvatarAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const file = formData.get("avatar");
  if (!(file instanceof File) || file.size === 0) {
    return { status: "error", message: "Choose a photo to upload." };
  }
  return runAction("Profile photo updated.", async () => {
    const body = new FormData();
    body.append("avatar", file);
    await authenticatedUpload("/api/v1/profile/avatar", body);
    revalidatePath("/profile");
    revalidatePath("/");
    revalidatePath("/friends");
    revalidatePath("/settle-up");
  });
}

export async function removeAvatarAction(
  _previousState: ActionState,
  _formData: FormData,
): Promise<ActionState> {
  return runAction("Profile photo removed.", async () => {
    await authenticatedDelete("/api/v1/profile/avatar");
    revalidatePath("/profile");
    revalidatePath("/");
    revalidatePath("/friends");
    revalidatePath("/settle-up");
  });
}

// Deliberately uses runAction, which catches all errors including UnauthorizedError
// and surfaces them as form messages rather than redirecting to /login.
// POST /api/v1/profile/password returns 401 when the current password is wrong;
// apiRequest converts that 401 into UnauthorizedError("current password is incorrect"),
// and runAction returns it as { status: "error", message: "current password is incorrect" }
// so the user sees the API message inline — not a redirect to /login.
export async function changePasswordAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const newPassword = formData.get("newPassword");
  if (newPassword !== formData.get("confirmPassword")) {
    return { status: "error", message: "New passwords do not match." };
  }
  return runAction("Password changed.", async () => {
    await authenticatedPost("/api/v1/profile/password", {
      currentPassword: formData.get("currentPassword"),
      newPassword,
    });
    revalidatePath("/profile");
  });
}

async function runAction(
  successMessage: string,
  action: () => Promise<void>,
): Promise<ActionState> {
  try {
    await action();
    return { status: "success", message: successMessage };
  } catch (error) {
    return {
      status: "error",
      message: error instanceof Error ? error.message : "The request could not be completed.",
    };
  }
}

function parseMoney(value: FormDataEntryValue | null) {
  const text = typeof value === "string" ? value.trim() : "";
  if (!/^\d{1,10}(?:\.\d{1,2})?$/.test(text)) {
    throw new Error("Enter a valid positive amount with up to two decimal places.");
  }
  const [whole, fraction = ""] = text.split(".");
  const amountMinor = Number(whole) * 100 + Number(fraction.padEnd(2, "0"));
  if (!Number.isSafeInteger(amountMinor) || amountMinor <= 0) {
    throw new Error("Enter a valid positive amount.");
  }
  return amountMinor;
}