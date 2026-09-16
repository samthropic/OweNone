import "@/components/landing/landing.css";
import { LandingPage } from "@/components/landing/LandingPage";
import { getOptionalSessionUser } from "@/lib/api";

export default async function Landing() {
  const user = await getOptionalSessionUser();
  return <LandingPage user={user} />;
}
