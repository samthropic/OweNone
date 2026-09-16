import type { Metadata } from "next";
import { redirect } from "next/navigation";
import "@/components/home/home.css";
import "@/components/settle/settle.css";
import { SettlePage } from "@/components/settle/SettlePage";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import { getDashboard, getActivity, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

// Always fetch fresh data so payments appear immediately
export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Settle up — OweNone",
  description: "Make the minimum transfers to clear every balance in your network.",
};

export default async function SettleUpRoute() {
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

  let recentPayments;
  try {
    recentPayments = await getActivity({ kind: "settlement", limit: 5 });
  } catch {
    // Backend may not yet support the kind filter — degrade gracefully
    recentPayments = { activity: [], total: 0, hasMore: false };
  }

  return (
    <div className="home-root">
      <HomeSidebar user={dashboard.user} active="Settle Up" />
      <main className="home-main">
        <div className="home-main-inner">
          <SettlePage dashboard={dashboard} recentPayments={recentPayments} />
        </div>
      </main>
    </div>
  );
}
