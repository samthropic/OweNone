import { redirect } from "next/navigation";

/** Keep /landing working; canonical marketing URL is /. */
export default function LandingRedirect() {
  redirect("/");
}
