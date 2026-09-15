import { PROBLEM_CARDS } from "./landingData";

export function LandingProblemSection() {
  return (
    <section id="problem" className="band block">
      <div className="landing-shell problem-grid">
        <div>
          <div className="section-kicker">The category mistake</div>
          <h2 className="section-title">
            Expense apps record transactions.
            <br />
            OweNone optimizes the system.
          </h2>
        </div>
        <div className="cards-2">
          {PROBLEM_CARDS.map((card) => (
            <article
              key={card.title}
              className={card.highlighted ? "card soft-indigo" : "card"}
            >
              <h3>{card.title}</h3>
              <p className="card-text">{card.body}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
