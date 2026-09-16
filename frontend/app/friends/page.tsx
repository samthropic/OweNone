import type { Metadata } from "next";
import { redirect } from "next/navigation";
import "@/components/home/home.css";
import "@/components/friends/friends.css";
import { FriendsPage } from "@/components/friends/FriendsPage";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import { getDashboard, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

export const metadata: Metadata = {
  title: "Friends — OweNone",
  description: "Manage your friends and see shared balances.",
};

export default async function FriendsRoute() {
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
      <HomeSidebar user={dashboard.user} active="Friends" />
      <main className="home-main">
        <div className="home-main-inner">
          <FriendsPage dashboard={dashboard} />
        </div>
      </main>
    </div>
  );
}
