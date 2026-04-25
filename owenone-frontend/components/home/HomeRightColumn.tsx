import { groups } from "@/components/home/homeData";

const netPositions = [
  { name: "Jordan K.", amount: "-£12.00", tone: "is-negative" },
  { name: "Mike K.", amount: "+£18.50", tone: "is-positive" },
  { name: "Alex L.", amount: "+£9.00", tone: "is-positive" },
  { name: "Sam R.", amount: "-£8.00", tone: "is-negative" },
];

const stats = [
  { label: "Total owed", value: "£34", tone: "is-negative" },
  { label: "You are owed", value: "£22", tone: "is-positive" },
  { label: "Saved this month", value: "5", tone: "is-indigo" },
  { label: "Net balance", value: "-£12", tone: "is-negative" },
];

export function HomeRightColumn() {
  return (
    <div className="home-right-column">
      <section className="home-card">
        <div className="home-card-head">
          <div>
            <h3>Your groups</h3>
            <p>3 active</p>
          </div>
          <button type="button" className="home-btn home-btn-small">
            New group
          </button>
        </div>
        <div className="home-list">
          {groups.map((group) => (
            <article key={group.id} className="home-row">
              <div className="home-icon">{group.icon}</div>
              <div className="home-row-copy">
                <p className="home-row-title">{group.name}</p>
                <p className="home-row-sub">{group.members}</p>
              </div>
              <div className="home-row-right">
                <p
                  className={`home-row-amount ${group.tone === "positive" ? "is-positive" : ""} ${group.tone === "negative" ? "is-negative" : ""}`}
                >
                  {group.balance}
                </p>
                <p className="home-row-sub">{group.status}</p>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="home-card">
        <div className="home-card-head">
          <div>
            <h3>Net positions</h3>
            <p>After compression</p>
          </div>
        </div>
        <div className="home-list compact">
          {netPositions.map((position) => (
            <article key={position.name} className="home-compact-row">
              <p className="home-row-title">{position.name}</p>
              <p className={`home-row-amount ${position.tone}`}>{position.amount}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="home-stat-grid">
        {stats.map((stat) => (
          <article key={stat.label} className="home-stat-card">
            <p>{stat.label}</p>
            <strong className={stat.tone}>{stat.value}</strong>
          </article>
        ))}
      </section>
    </div>
  );
}
