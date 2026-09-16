import "@/components/home/home.css";
import { HomePage } from "@/components/home/HomePage";
import { getDashboard, UnauthorizedError } from "@/lib/api";
import { getSessionToken } from "@/lib/session";
import { redirect } from "next/navigation";

// Always fetch fresh balances after settlements elsewhere in the app
export const dynamic = "force-dynamic";

export default async function Home() {
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

  return <HomePage dashboard={dashboard} />;
}
