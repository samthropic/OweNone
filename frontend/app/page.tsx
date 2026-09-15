import "@/components/home/home.css";
import { HomePage } from "@/components/home/HomePage";
import { getDashboard } from "@/lib/api";

export default async function Home() {
  const dashboard = await getDashboard();
  return <HomePage dashboard={dashboard} />;
}
