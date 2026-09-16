"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { AddExpenseDialog } from "@/components/expenses/AddExpenseDialog";
import { fetchActivityPage } from "@/components/activity/activityFetcher";
import {
  categoryIcon,
  formatMoney,
  shortName,
} from "@/components/home/homeData";
import type { ActivityFeed, ActivityItem, Dashboard } from "@/lib/api-types";

type Props = {
  dashboard: Dashboard;
  initialFeed: ActivityFeed;
};

export function ActivityPage({ dashboard, initialFeed }: Props) {
  const router = useRouter();
  const [items, setItems] = useState<ActivityItem[]>(initialFeed.activity);
  const [hasMore, setHasMore] = useState(initialFeed.hasMore);
  const [offset, setOffset] = useState(initialFeed.activity.length);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [expenseOpen, setExpenseOpen] = useState(false);
  const [dialogKey, setDialogKey] = useState(0);

  const prevFirstIdRef = useRef(initialFeed.activity[0]?.id ?? null);

  // Sync items state when the server refreshes (e.g., after router.refresh())
  useEffect(() => {
    const newFirstId = initialFeed.activity[0]?.id ?? null;
    if (newFirstId !== prevFirstIdRef.current) {
      prevFirstIdRef.current = newFirstId;
      setItems(initialFeed.activity);
      setHasMore(initialFeed.hasMore);
      setOffset(initialFeed.activity.length);
    }
  }, [initialFeed]);

  const hasNoGroups = dashboard.groups.length === 0;

  function openExpenseDialog() {
    setDialogKey((k) => k + 1);
    setExpenseOpen(true);
  }

  async function handleLoadMore() {
    setLoadingMore(true);
    setLoadError(null);
    try {
      const data = await fetchActivityPage(offset);
      setItems((prev) => [...prev, ...data.activity]);
      setHasMore(data.hasMore);
      setOffset((prev) => prev + data.activity.length);
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : "Could not load more items.");
    } finally {
      setLoadingMore(false);
    }
  }

  function handleExpenseSuccess() {
    setExpenseOpen(false);
    router.refresh();
  }

  const dayGroups = groupByDay(items);

  return (
    <>
      <div className="activity-page-head">
        <div>
          <h1>Activity</h1>
          <p>Everything that happened across all your groups.</p>
        </div>
        <div className="activity-head-actions">
          <button
            type="button"
            className="home-btn home-btn-small"
            onClick={openExpenseDialog}
            disabled={hasNoGroups}
            aria-disabled={hasNoGroups}
            title={hasNoGroups ? "Create a group first to add expenses" : undefined}
          >
            Add expense
          </button>
          {hasNoGroups && (
            <p className="activity-no-groups-hint">
              <Link href="/groups" className="activity-link">Create a group</Link> to start recording expenses.
            </p>
          )}
        </div>
      </div>

      {items.length === 0 ? (
        <EmptyState hasNoGroups={hasNoGroups} onAddExpense={openExpenseDialog} />
      ) : (
        <div className="activity-feed">
          {dayGroups.map((group) => (
            <section key={group.heading} className="activity-day-group">
              <h2 className="activity-day-heading">{group.heading}</h2>
              <div className="home-card">
                <div className="home-list">
                  {group.items.map((item) => (
                    <ActivityRow
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

      {expenseOpen && (
        <AddExpenseDialog
          key={dialogKey}
          groups={dashboard.groups}
          currentUserId={dashboard.user.id}
          onClose={() => setExpenseOpen(false)}
          onSuccess={handleExpenseSuccess}
        />
      )}
    </>
  );
}

function ActivityRow({
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
    subParts.push(`${split}, ${item.peopleCount ?? 0} ${item.peopleCount === 1 ? "person" : "people"}`);
  } else {
    subParts.push("Marked paid");
  }
  if (item.groupName) subParts.push(item.groupName);

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

function EmptyState({
  hasNoGroups,
  onAddExpense,
}: {
  hasNoGroups: boolean;
  onAddExpense: () => void;
}) {
  return (
    <div className="activity-empty" role="region" aria-label="No activity yet">
      <p className="activity-empty-title">No activity yet</p>
      <p className="activity-empty-sub">
        {hasNoGroups
          ? "Create a group, then add expenses to start tracking shared costs."
          : "Add an expense to start tracking shared costs."}
      </p>
      {hasNoGroups ? (
        <Link href="/groups" className="home-btn home-btn-primary activity-empty-cta">
          Create a group
        </Link>
      ) : (
        <button
          type="button"
          className="home-btn home-btn-primary"
          onClick={onAddExpense}
        >
          Add expense
        </button>
      )}
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

// The date already sits in the day heading, so each row shows only its clock time.
// Using relative wording here contradicted the heading: an item two calendar days
// old but under 48 hours read as "Yesterday" beneath a dated heading.
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
