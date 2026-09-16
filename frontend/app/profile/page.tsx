import type { Metadata } from "next";
import { redirect } from "next/navigation";
import "@/components/home/home.css";
import "@/components/profile/profile.css";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import { ProfilePage } from "@/components/profile/ProfilePage";
import { getSessionUser, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

export const metadata: Metadata = {
  title: "Profile — OweNone",
  description: "Update your details and manage your account.",
};

export default async function ProfileRoute() {
  const token = await getSessionToken();
  if (!token) redirect("/login");

  let sessionData;
  try {
    sessionData = await getSessionUser();
  } catch (err) {
    if (!(err instanceof UnauthorizedError)) throw err;
    // UnauthorizedError → fall through; sessionData stays undefined
  }

  // redirect is called outside the try/catch so the NEXT_REDIRECT
  // control-flow error cannot be swallowed by the catch block above
  if (!sessionData) redirect("/login");

  return (
    <div className="home-root">
      <HomeSidebar user={sessionData.user} active="Profile" />
      <main className="home-main">
        <div className="home-main-inner">
          <ProfilePage user={sessionData.user} />
        </div>
      </main>
    </div>
  );
}
