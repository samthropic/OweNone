import type { Metadata } from "next";
import { redirect } from "next/navigation";
import "@/components/home/home.css";
import "@/components/groups/groups.css";
import { GroupsPage } from "@/components/groups/GroupsPage";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import { getDashboard, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

export const metadata: Metadata = {
  title: "Groups — OweNone",
  description: "Create and manage groups to split expenses with friends.",
};

export default async function GroupsRoute() {
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

  return (
    <div className="home-root">
      <HomeSidebar user={dashboard.user} active="Groups" />
      <main className="home-main">
        <div className="home-main-inner">
          <GroupsPage dashboard={dashboard} />
        </div>
      </main>
    </div>
  );
}
