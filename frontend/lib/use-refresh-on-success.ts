"use client";

import { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import type { ActionState } from "@/lib/action-state";

/** Refresh RSC data after a successful server action so balances/icons update. */
export function useRefreshOnSuccess(state: ActionState) {
  const router = useRouter();
  const lastMessage = useRef<string | null>(null);

  useEffect(() => {
    if (state.status !== "success") return;
    // Avoid refreshing twice for the same success payload
    if (lastMessage.current === state.message) return;
    lastMessage.current = state.message;
    router.refresh();
  }, [state.status, state.message, router]);
}
