import { dashboardDate } from "@/components/home/homeData";
import { HeaderActions } from "@/components/home/HomeActionForms";
import type { Dashboard } from "@/lib/api-types";

export function HomeHeader({ dashboard }: { dashboard: Dashboard }) {
  return (
    <header className="home-header">
      <div>
        <h1>Overview</h1>
        <p>{dashboardDate(dashboard)}</p>
      </div>
      <HeaderActions dashboard={dashboard} />
    </header>
  );
}
