"use server";

import { revalidatePath } from "next/cache";
import { apiRequest, authenticatedPost } from "@/lib/api";

export type ActionState = {
  status: "idle" | "success" | "error";
  message: string;
};

export const initialActionState: ActionState = { status: "idle", message: "" };

export async function joinWaitlistAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("You are on the waitlist.", async () => {
    await apiRequest<{ joined: boolean }>("/api/v1/waitlist", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: formData.get("email") }),
    });
  });
}

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
  });
}

export async function createExpenseAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Expense added and balances recomputed.", async () => {
    const amountMinor = parseMoney(formData.get("amount"));
    await authenticatedPost<{ id: string }>(
      "/api/v1/expenses",
      {
        groupId: formData.get("groupId"),
        description: formData.get("description"),
        category: formData.get("category"),
        amountMinor,
        currency: formData.get("currency"),
        paidByUserId: formData.get("paidByUserId"),
        split: {
          method: "equal",
          participantIds: formData.getAll("participantIds"),
        },
      },
      crypto.randomUUID(),
    );
    revalidatePath("/");
  });
}

export async function createSettlementAction(
  _previousState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  return runAction("Payment recorded and balances recomputed.", async () => {
    const groupID = formData.get("groupId");
    await authenticatedPost<{ id: string }>(
      "/api/v1/settlements",
      {
        groupId: groupID || null,
        toUserId: formData.get("toUserId"),
        amountMinor: parseMoney(formData.get("amount")),
        currency: formData.get("currency"),
      },
      crypto.randomUUID(),
    );
    revalidatePath("/");
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