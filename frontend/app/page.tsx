import "@/components/home/home.css";
import "@/components/landing/landing.css";
import { HomePage } from "@/components/home/HomePage";
import { LandingPage } from "@/components/landing/LandingPage";
import { getDashboard, getOptionalSessionUser, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

// Always fetch fresh balances after settlements elsewhere in the app
export const dynamic = "force-dynamic";

export default async function Home() {
  const token = await getSessionToken();
  if (!token) {
    const user = await getOptionalSessionUser();
    return <LandingPage user={user} />;
  }

  let dashboard;
  try {
    dashboard = await getDashboard();
  } catch (err) {
    if (!(err instanceof UnauthorizedError)) throw err;
  }

  // Stale/invalid session → marketing page, not login
  if (!dashboard) {
    const user = await getOptionalSessionUser();
    return <LandingPage user={user} />;
  }

  return <HomePage dashboard={dashboard} />;
}
