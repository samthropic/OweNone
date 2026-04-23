function DebtEdge({
  x1,
  y1,
  x2,
  y2,
  label,
  midX,
  midY,
}: {
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  label: string;
  midX: number;
  midY: number;
}) {
  return (
    <g>
      <line
        x1={x1}
        y1={y1}
        x2={x2}
        y2={y2}
        stroke="#F43F5E"
        strokeWidth="1.5"
        strokeDasharray="6 5"
        opacity={0.85}
      />
      <text
        x={midX}
        y={midY}
        fill="#F43F5E"
        fontSize="11"
        fontWeight="600"
        textAnchor="middle"
        style={{ fontFamily: "inherit" }}
      >
        {label}
      </text>
    </g>
  );
}

function PersonNode({
  cx,
  cy,
  label,
  sub,
  active = true,
}: {
  cx: number;
  cy: number;
  label: string;
  sub: string;
  active?: boolean;
}) {
  const opacity = active ? 1 : 0.22;
  return (
    <g className="node-group" opacity={opacity}>
      <circle
        cx={cx}
        cy={cy}
        r="22"
        fill="rgba(15,23,42,0.9)"
        stroke="rgba(148,163,184,0.35)"
        strokeWidth="1.2"
      />
      <text
        x={cx}
        y={cy - 2}
        fill="#F1F5F9"
        fontSize="11"
        fontWeight="700"
        textAnchor="middle"
        style={{ fontFamily: "var(--font-syne), Syne, sans-serif" }}
      >
        {label}
      </text>
      <text
        x={cx}
        y={cy + 10}
        fill="#94A3B8"
        fontSize="8"
        textAnchor="middle"
        style={{ fontFamily: "inherit" }}
      >
        {sub}
      </text>
    </g>
  );
}

function GraphBefore() {
  return (
    <svg
      className="graph-svg"
      viewBox="0 0 420 300"
      role="img"
      aria-label="Before: six payments needed between four people"
    >
      <DebtEdge x1={95} y1={88} x2={325} y2={88} label="£18" midX={210} midY={78} />
      <DebtEdge x1={95} y1={88} x2={95} y2={218} label="£11" midX={72} midY={150} />
      <DebtEdge x1={95} y1={88} x2={325} y2={218} label="£22" midX={200} midY={120} />
      <DebtEdge x1={325} y1={88} x2={95} y2={218} label="£15" midX={240} midY={175} />
      <DebtEdge x1={325} y1={88} x2={325} y2={218} label="£9" midX={348} midY={150} />
      <DebtEdge x1={95} y1={218} x2={325} y2={218} label="£14" midX={210} midY={238} />
      <PersonNode cx={95} cy={88} label="Sarah" sub="J." />
      <PersonNode cx={325} cy={88} label="Alex" sub="L." />
      <PersonNode cx={95} cy={218} label="Mike" sub="K." />
      <PersonNode cx={325} cy={218} label="Jordan" sub="K." />
    </svg>
  );
}

function GraphAfter() {
  return (
    <svg
      className="graph-svg"
      viewBox="0 0 420 300"
      role="img"
      aria-label="After: one payment from Sarah to Jordan"
    >
      <defs>
        <filter id="landing-glow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="4" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      <line
        x1={95}
        y1={88}
        x2={325}
        y2={218}
        stroke="#14B8A6"
        strokeWidth="5"
        strokeLinecap="round"
        opacity={0.95}
        filter="url(#landing-glow)"
      />
      <text
        x={210}
        y={148}
        fill="#5EEAD4"
        fontSize="12"
        fontWeight="800"
        textAnchor="middle"
        style={{ fontFamily: "var(--font-syne), Syne, sans-serif" }}
      >
        £12
      </text>
      <PersonNode cx={95} cy={88} label="Sarah" sub="J." active />
      <PersonNode cx={325} cy={88} label="Alex" sub="L." active={false} />
      <PersonNode cx={95} cy={218} label="Mike" sub="K." active={false} />
      <PersonNode cx={325} cy={218} label="Jordan" sub="K." active />
    </svg>
  );
}

export function DebtGraphPreview() {
  return (
    <section className="viz-section" aria-labelledby="viz-title">
      <div className="container">
        <div className="viz-container">
          <div className="viz-header">
            <p className="viz-title" id="viz-title">
              Live debt graph — Holiday &apos;24 · Flatmates · Dinners
            </p>
            <div className="ws-indicator">
              <span className="ws-dot" aria-hidden />
              <span className="ws-text">WebSocket Synced · 3 connections</span>
            </div>
          </div>
          <div className="viz-body">
            <div className="graph-panel">
              <p className="graph-label before">Before — 6 payments needed</p>
              <GraphBefore />
            </div>
            <div className="sweep-col">
              <div className="sweep-badge">
                <div className="sweep-pct">−60%</div>
                <div className="sweep-pct-label">payments reduced</div>
              </div>
              <div className="sweep-arrow" aria-hidden>
                <span className="sweep-icon">→</span>
              </div>
              <div className="sweep-label">Min-flow algorithm</div>
            </div>
            <div className="graph-panel">
              <p className="graph-label after">After — 1 payment needed</p>
              <GraphAfter />
              <div className="result-card">
                <div className="result-label">Next step: Pay Jordan K.</div>
                <div className="result-amount">£12.00</div>
                <div className="result-sub">
                  Clears Holiday &apos;24 & Flatmates · 1 of 1 payment
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
