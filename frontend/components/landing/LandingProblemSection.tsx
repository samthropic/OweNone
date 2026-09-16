import { PROBLEM_CARDS } from "./landingData";

export function LandingProblemSection() {
  return (
    <section id="problem" className="band block">
      <div className="landing-shell problem-grid">
        <div>
          <p className="section-kicker">Where other apps stop</p>
          <h2 className="section-title">
            Ledgers record the mess.
            <br />
            OweNone clears the harbor.
          </h2>
        </div>
        <div className="compare-list">
          {PROBLEM_CARDS.map((card) => (
            <article
              key={card.title}
              className={card.highlighted ? "compare-row is-highlight" : "compare-row"}
            >
              <h3>{card.title}</h3>
              <p>{card.body}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
