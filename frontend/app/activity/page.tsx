import type { Metadata } from "next";
import { redirect } from "next/navigation";
import "@/components/home/home.css";
import "@/components/activity/activity.css";
import { ActivityPage } from "@/components/activity/ActivityPage";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import { getDashboard, getActivity, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

// Always fetch fresh data so new expenses appear immediately
export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Activity — OweNone",
  description: "Everything that happened across all your groups.",
};

export default async function ActivityRoute() {
  const token = await getSessionToken();
  if (!token) redirect("/login");

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

  let initialFeed;
  try {
    initialFeed = await getActivity({ limit: 50, offset: 0 });
  } catch {
    initialFeed = { activity: [], total: 0, hasMore: false };
  }

  return (
    <div className="home-root">
      <HomeSidebar user={dashboard.user} active="Activity" />
      <main className="home-main">
        <div className="home-main-inner">
          <ActivityPage dashboard={dashboard} initialFeed={initialFeed} />
        </div>
      </main>
    </div>
  );
}
