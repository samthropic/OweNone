"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { AddExpenseDialog } from "@/components/expenses/AddExpenseDialog";
import { fetchGroupExpensePage } from "@/components/groups/groupDetailFetcher";
import {
  balanceState,
  categoryIcon,
  formatMoney,
  shortName,
} from "@/components/home/homeData";
import type { ActivityFeed, ActivityItem, Dashboard } from "@/lib/api-types";

type Group = Dashboard["groups"][number];

type Props = {
  dashboard: Dashboard;
  group: Group;
  initialExpenses: ActivityFeed;
  initialSettlements: ActivityFeed;
};

export function GroupDetailPage({
  dashboard,
  group,
  initialExpenses,
  initialSettlements,
}: Props) {
  const router = useRouter();
  const [items, setItems] = useState<ActivityItem[]>(initialExpenses.activity);
  const [hasMore, setHasMore] = useState(initialExpenses.hasMore);
  const [offset, setOffset] = useState(initialExpenses.activity.length);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [expenseOpen, setExpenseOpen] = useState(false);
  const [dialogKey, setDialogKey] = useState(0);

  const prevFirstIdRef = useRef(initialExpenses.activity[0]?.id ?? null);

  // Sync items when the server re-renders after router.refresh()
  useEffect(() => {
    const newFirstId = initialExpenses.activity[0]?.id ?? null;
    if (newFirstId !== prevFirstIdRef.current) {
      prevFirstIdRef.current = newFirstId;
      setItems(initialExpenses.activity);
      setHasMore(initialExpenses.hasMore);
      setOffset(initialExpenses.activity.length);
    }
  }, [initialExpenses]);

  function openExpenseDialog() {
    setDialogKey((k) => k + 1);
    setExpenseOpen(true);
  }

  async function handleLoadMore() {
    setLoadingMore(true);
    setLoadError(null);
    try {
      const data = await fetchGroupExpensePage(group.id, offset);
      setItems((prev) => [...prev, ...data.activity]);
      setHasMore(data.hasMore);
      setOffset((prev) => prev + data.activity.length);
    } catch (err) {
      setLoadError(
        err instanceof Error ? err.message : "Could not load more items.",
      );
    } finally {
      setLoadingMore(false);
    }
  }

  function handleExpenseSuccess() {
    setExpenseOpen(false);
    router.refresh();
  }

  const balanceInfo = balanceState(group.balance.amountMinor);
  const memberNames = group.members.map(shortName).join(", ");
  const dayGroups = groupByDay(items);
  const hasSettlements = initialSettlements.activity.length > 0;

  return (
    <>
      {/* Header */}
      <div className="group-detail-head">
        <Link href="/groups" className="group-detail-back">
          ← Groups
        </Link>
        <div className="group-detail-hero">
          <div className="group-detail-title-row">
            <div className="group-detail-icon" aria-hidden="true">
              {group.icon}
            </div>
            <div>
              <h1 className="group-detail-name">{group.name}</h1>
              <p className="group-detail-members">{memberNames}</p>
            </div>
          </div>
          <div className="group-detail-actions">
            <div className="group-detail-balance-col">
              <p
                className={`group-detail-amount ${
                  balanceInfo.tone === "green"
                    ? "is-positive"
                    : balanceInfo.tone === "rose"
                      ? "is-negative"
                      : ""
                }`}
              >
                {formatMoney(group.balance)}
              </p>
              <p className="home-row-sub">{balanceInfo.direction}</p>
            </div>
            <button
              type="button"
              className="home-btn home-btn-small"
              onClick={openExpenseDialog}
            >
              Add expense
            </button>
          </div>
        </div>
      </div>

      {/* Expenses */}
      {items.length === 0 ? (
        <EmptyState onAddExpense={openExpenseDialog} />
      ) : (
        <div className="activity-feed">
          {dayGroups.map((dayGroup) => (
            <section key={dayGroup.heading} className="activity-day-group">
              <h2 className="activity-day-heading">{dayGroup.heading}</h2>
              <div className="home-card">
                <div className="home-list">
                  {dayGroup.items.map((item) => (
                    <ExpenseRow
                      key={item.id}
                      item={item}
                      userId={dashboard.user.id}
                    />
                  ))}
                </div>
              </div>
            </section>
          ))}

          <div className="activity-load-more" role="status" aria-live="polite">
            {loadError && (
              <p className="home-action-message is-error">{loadError}</p>
            )}
            {hasMore && (
              <button
                type="button"
                className="home-btn"
                onClick={handleLoadMore}
                disabled={loadingMore}
              >
                {loadingMore ? "Loading…" : "Load more"}
              </button>
            )}
          </div>
        </div>
      )}

      {/* Settlements (only when this group has any) */}
      {hasSettlements && (
        <div className="group-detail-settlements">
          <section className="activity-day-group">
            <h2 className="activity-day-heading">Payments</h2>
            <div className="home-card">
              <div className="home-list">
                {initialSettlements.activity.map((item) => (
                  <ExpenseRow
                    key={item.id}
                    item={item}
                    userId={dashboard.user.id}
                  />
                ))}
              </div>
            </div>
          </section>
        </div>
      )}

      {/* Add expense dialog */}
      {expenseOpen && (
        <AddExpenseDialog
          key={dialogKey}
          groups={[group]}
          fixedGroupId={group.id}
          currentUserId={dashboard.user.id}
          onClose={() => setExpenseOpen(false)}
          onSuccess={handleExpenseSuccess}
        />
      )}
    </>
  );
}

// — Expense row ————————————————————————————————————————————————————

function ExpenseRow({
  item,
  userId,
}: {
  item: ActivityItem;
  userId: string;
}) {
  const isOwn = item.actor.id === userId;
  const actorName = isOwn ? "You" : shortName(item.actor);
  const verb = item.kind === "expense" ? "added" : "settled";

  const subParts: string[] = [];
  if (item.kind === "expense") {
    const split =
      item.splitMethod === "equal" ? "Split equally" : "Exact split";
    subParts.push(
      `${split}, ${item.peopleCount ?? 0} ${item.peopleCount === 1 ? "person" : "people"}`,
    );
  } else {
    subParts.push("Marked paid");
  }

  return (
    <article className="home-row">
      <div className="home-icon" aria-hidden="true">
        {categoryIcon(item.category, item.kind)}
      </div>
      <div className="home-row-copy">
        <p className="home-row-title">
          {actorName} {verb} &ldquo;{item.description}&rdquo;
        </p>
        <p className="home-row-sub">{subParts.join(", ")}</p>
      </div>
      <div className="home-row-right">
        <p
          className={`home-row-amount ${
            item.impact.amountMinor >= 0 ? "is-positive" : "is-negative"
          }`}
        >
          {formatMoney(item.impact)}
        </p>
        <p className="home-row-sub" suppressHydrationWarning>
          {timeOfDay(item.occurredAt)}
        </p>
      </div>
    </article>
  );
}

// — Empty state ———————————————————————————————————————————————————

function EmptyState({ onAddExpense }: { onAddExpense: () => void }) {
  return (
    <div
      className="activity-empty"
      role="region"
      aria-label="No expenses yet"
    >
      <p className="activity-empty-title">No expenses yet</p>
      <p className="activity-empty-sub">
        Add an expense to start tracking what this group owes.
      </p>
      <button
        type="button"
        className="home-btn home-btn-primary"
        onClick={onAddExpense}
      >
        Add expense
      </button>
    </div>
  );
}

// — Helpers ———————————————————————————————————————————————————————

function localDateKey(isoString: string): string {
  const d = new Date(isoString);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

const dateFormatter = new Intl.DateTimeFormat("en-GB", {
  day: "numeric",
  month: "long",
  year: "numeric",
});

function timeOfDay(value: string) {
  return new Date(value).toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  });
}

function groupByDay(
  items: ActivityItem[],
): Array<{ heading: string; items: ActivityItem[] }> {
  const groups = new Map<string, ActivityItem[]>();

  for (const item of items) {
    const key = localDateKey(item.occurredAt);
    let bucket = groups.get(key);
    if (!bucket) {
      bucket = [];
      groups.set(key, bucket);
    }
    bucket.push(item);
  }

  const today = new Date();
  const todayKey = localDateKey(today.toISOString());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);
  const yesterdayKey = localDateKey(yesterday.toISOString());

  return Array.from(groups.entries()).map(([key, dayItems]) => {
    let heading: string;
    if (key === todayKey) {
      heading = "Today";
    } else if (key === yesterdayKey) {
      heading = "Yesterday";
    } else {
      const [y, m, d] = key.split("-").map(Number);
      heading = dateFormatter.format(new Date(y, m - 1, d));
    }
    return { heading, items: dayItems };
  });
}
