import type { LandingConnection, LandingNetworkNode } from "./landingTypes";

const VIEW_W = 400;
const VIEW_H = 280;
const NODE_R = 22;

type LandingNetworkProps = {
  nodes: LandingNetworkNode[];
  connections: LandingConnection[];
  variant?: "default" | "compressed";
};

function toX(percent: number) {
  return (percent / 100) * VIEW_W;
}

function toY(percent: number) {
  return (percent / 100) * VIEW_H;
}

/** Shorten a segment so lines meet the circle edge, not the center. */
function edgePoints(
  from: LandingNetworkNode,
  to: LandingNetworkNode,
): { x1: number; y1: number; x2: number; y2: number } {
  const x1 = toX(from.x);
  const y1 = toY(from.y);
  const x2 = toX(to.x);
  const y2 = toY(to.y);
  const dx = x2 - x1;
  const dy = y2 - y1;
  const length = Math.hypot(dx, dy) || 1;
  const inset = NODE_R + 2;
  const ux = dx / length;
  const uy = dy / length;
  return {
    x1: x1 + ux * inset,
    y1: y1 + uy * inset,
    x2: x2 - ux * inset,
    y2: y2 - uy * inset,
  };
}

export function LandingNetwork({
  nodes,
  connections,
  variant = "default",
}: LandingNetworkProps) {
  const nodesById = new Map(nodes.map((node) => [node.id, node]));
  const networkClassName =
    variant === "compressed" ? "network is-compressed" : "network";

  return (
    <div className={networkClassName} key={variant}>
      <svg
        className="network-svg"
        viewBox={`0 0 ${VIEW_W} ${VIEW_H}`}
        role="img"
        aria-label="Debt clearing graph"
      >
        {connections.map((connection, index) => {
          const from = nodesById.get(connection.from);
          const to = nodesById.get(connection.to);
          if (!from || !to) return null;

          const { x1, y1, x2, y2 } = edgePoints(from, to);
          const midX = (x1 + x2) / 2;
          const midY = (y1 + y2) / 2;
          const labelWidth = Math.max(36, connection.amount.length * 8);

          return (
            <g
              key={`${connection.from}-${connection.to}-${connection.amount}`}
              className="network-link"
              style={{ animationDelay: `${0.05 + index * 0.08}s` }}
            >
              <line
                className={`network-edge network-edge-${connection.color}`}
                x1={x1}
                y1={y1}
                x2={x2}
                y2={y2}
              />
              <rect
                className="network-amount-bg"
                x={midX - labelWidth / 2}
                y={midY - 11}
                width={labelWidth}
                height={22}
                rx={11}
              />
              <text
                className="network-amount"
                x={midX}
                y={midY + 4}
                textAnchor="middle"
              >
                {connection.amount}
              </text>
            </g>
          );
        })}

        {nodes.map((node, index) => {
          const cx = toX(node.x);
          const cy = toY(node.y);
          return (
            <g
              key={node.id}
              className="network-node"
              style={{ animationDelay: `${0.12 + index * 0.05}s` }}
            >
              <circle
                className={`network-avatar tone-${node.tone}`}
                cx={cx}
                cy={cy}
                r={NODE_R}
              />
              <text
                className="network-initials"
                x={cx}
                y={cy + 4}
                textAnchor="middle"
              >
                {node.initials}
              </text>
              <text
                className="network-name"
                x={cx}
                y={cy + NODE_R + 16}
                textAnchor="middle"
              >
                {node.name}
              </text>
              <text
                className="network-balance"
                x={cx}
                y={cy + NODE_R + 30}
                textAnchor="middle"
              >
                {node.balance}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
}
