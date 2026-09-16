"use server";

import { getActivity } from "@/lib/api";
import type { ActivityFeed } from "@/lib/api-types";

export async function fetchGroupExpensePage(
  groupId: string,
  offset: number,
): Promise<ActivityFeed> {
  return getActivity({ limit: 50, offset, kind: "expense", groupId });
}
