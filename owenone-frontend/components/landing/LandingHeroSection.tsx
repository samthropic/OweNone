import {
  AFTER_CONNECTIONS,
  AFTER_FLOWS,
  BEFORE_CONNECTIONS,
  BEFORE_FLOWS,
  HERO_STATS,
  MINI_STATS,
  NETWORK_NODES,
} from "./landingData";
import { LandingNetwork } from "./LandingNetwork";

export function LandingHeroSection() {
  return (
    <section className="hero">
      <div className="hero-copy">
        <div className="eyebrow">
          <span className="eyebrow-dot" />
          Financial graph compression for modern social networks
        </div>
      </div>
      <div className="hero-grid">
        <div>
          <h1>
            Shared money,
            <br />
            <span className="subtle">reduced to elegance.</span>
          </h1>
          <p className="lead">
            OweNone transforms fragmented group expenses into the minimum
            number of payments across your entire relationship graph with
            real-time recomputation, cross-group netting, and settlement paths
            that actually make sense.
          </p>
          <div className="hero-actions">
            <button type="button" className="btn btn-dark">
              Join the private beta
            </button>
            <button type="button" className="btn">
              View live model
            </button>
          </div>
          <div className="stats">
            {HERO_STATS.map((stat) => (
              <div key={stat.value} className="stat">
                <div className="value">{stat.value}</div>
                <div className="label">{stat.label}</div>
              </div>
            ))}
          </div>
        </div>

        <aside className="panel-wrap">
          <div className="float-metric">
            <div className="tiny">Network efficiency</div>
            <div className="big">-50%</div>
            <div className="small">transfers in live example</div>
          </div>
          <div className="graph-panel">
            <div className="panel-head">
              <div>
                <div className="panel-title">Live network calculation</div>
                <div className="panel-subtitle">
                  Raw obligations recompute into the minimum transfer set
                </div>
              </div>
              <div className="live-pill">Live</div>
            </div>
            <div className="panel-grid">
              <div className="graph-card">
                <div className="graph-card-head">
                  <div>
                    <div className="graph-card-title">Before compression</div>
                    <div className="graph-card-sub">
                      3 transfers across overlapping balances
                    </div>
                  </div>
                  <div className="graph-badge">Raw graph</div>
                </div>
                <LandingNetwork
                  nodes={NETWORK_NODES}
                  connections={BEFORE_CONNECTIONS}
                />
                <div className="flow-list">
                  {BEFORE_FLOWS.map((flow) => (
                    <div key={flow.label} className="flow-item">
                      <span>{flow.label}</span>
                      <strong>{flow.amount}</strong>
                    </div>
                  ))}
                </div>
              </div>
              <div className="compress-pill">Compress</div>
              <div className="graph-card indigo">
                <div className="graph-card-head">
                  <div>
                    <div className="graph-card-title">After compression</div>
                    <div className="graph-card-sub">
                      2 transfers after global netting
                    </div>
                  </div>
                  <div className="graph-badge">Optimal state</div>
                </div>
                <LandingNetwork
                  nodes={NETWORK_NODES}
                  connections={AFTER_CONNECTIONS}
                  variant="indigo"
                />
                <div className="flow-list">
                  {AFTER_FLOWS.map((flow) => (
                    <div key={flow.label} className="flow-item">
                      <span>{flow.label}</span>
                      <strong>{flow.amount}</strong>
                    </div>
                  ))}
                </div>
              </div>
            </div>
            <div className="mini-stats">
              {MINI_STATS.map((stat) => (
                <div key={stat.value} className="mini-stat">
                  <div className="v">{stat.value}</div>
                  <div className="l">{stat.label}</div>
                </div>
              ))}
            </div>
          </div>
        </aside>
      </div>
    </section>
  );
}
