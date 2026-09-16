import type { Metadata } from "next";
import { notFound, redirect } from "next/navigation";
import "@/components/home/home.css";
import "@/components/activity/activity.css";
import "@/components/groups/groups.css";
import { GroupDetailPage } from "@/components/groups/GroupDetailPage";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import { getDashboard, getActivity, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

// Always fetch fresh data — new expenses must appear immediately after add
export const dynamic = "force-dynamic";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ groupId: string }>;
}): Promise<Metadata> {
  try {
    const { groupId } = await params;
    const dashboard = await getDashboard();
    const group = dashboard.groups.find((g) => g.id === groupId);
    if (group) return { title: `${group.name} — OweNone` };
  } catch {
    // session absent or API unavailable — fall through to static title
  }
  return { title: "Group — OweNone" };
}

export default async function GroupDetailRoute({
  params,
}: {
  params: Promise<{ groupId: string }>;
}) {
  const token = await getSessionToken();
  if (!token) redirect("/login");

  const { groupId } = await params;

  let dashboard;
  try {
    dashboard = await getDashboard();
  } catch (err) {
    if (!(err instanceof UnauthorizedError)) throw err;
    // UnauthorizedError → fall through; dashboard stays undefined
  }

  // redirect is called outside the try/catch so the NEXT_REDIRECT
  // control-flow error cannot be swallowed by the catch block above
  if (!dashboard) redirect("/login");

  const group = dashboard.groups.find((g) => g.id === groupId);
  // notFound outside any try/catch so NEXT_NOT_FOUND is never swallowed
  if (!group) notFound();

  let initialExpenses;
  try {
    initialExpenses = await getActivity({
      groupId,
      kind: "expense",
      limit: 50,
      offset: 0,
    });
  } catch {
    initialExpenses = { activity: [], total: 0, hasMore: false };
  }

  let initialSettlements;
  try {
    initialSettlements = await getActivity({
      groupId,
      kind: "settlement",
      limit: 50,
      offset: 0,
    });
  } catch {
    initialSettlements = { activity: [], total: 0, hasMore: false };
  }

  return (
    <div className="home-root">
      <HomeSidebar user={dashboard.user} active="Groups" />
      <main className="home-main">
        <div className="home-main-inner">
          <GroupDetailPage
            dashboard={dashboard}
            group={group}
            initialExpenses={initialExpenses}
            initialSettlements={initialSettlements}
          />
        </div>
      </main>
    </div>
  );
}
