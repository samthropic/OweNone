import { activity } from "@/components/home/homeData";

export function HomeActivityCard() {
  return (
    <section className="home-card">
      <div className="home-card-head">
        <div>
          <h3>Recent activity</h3>
          <p>Last 7 days</p>
        </div>
        <button type="button" className="home-btn home-btn-small">
          View all
        </button>
      </div>
      <div className="home-list">
        {activity.map((item) => (
          <article key={item.id} className="home-row">
            <div className="home-icon">{item.icon}</div>
            <div className="home-row-copy">
              <p className="home-row-title">{item.title}</p>
              <p className="home-row-sub">{item.meta}</p>
            </div>
            <div className="home-row-right">
              <p
                className={`home-row-amount ${item.amountTone === "positive" ? "is-positive" : "is-negative"}`}
              >
                {item.amount}
              </p>
              <p className="home-row-sub">{item.time}</p>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
