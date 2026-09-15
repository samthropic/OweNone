import type { LandingConnection, LandingNetworkNode } from "./landingTypes";

type LandingNetworkProps = {
  nodes: LandingNetworkNode[];
  connections: LandingConnection[];
  variant?: "default" | "indigo";
};

export function LandingNetwork({
  nodes,
  connections,
  variant = "default",
}: LandingNetworkProps) {
  const nodesById = new Map(nodes.map((node) => [node.id, node]));
  const networkClassName = variant === "indigo" ? "network indigo" : "network";

  return (
    <div className={networkClassName}>
      {connections.map((connection) => {
        const from = nodesById.get(connection.from);
        const to = nodesById.get(connection.to);
        if (!from || !to) {
          return null;
        }

        const dx = to.x - from.x;
        const dy = to.y - from.y;
        const length = Math.sqrt(dx * dx + dy * dy);
        const angle = Math.atan2(dy, dx) * (180 / Math.PI);

        return (
          <div key={`${connection.from}-${connection.to}-${connection.amount}`}>
            <div
              className={`line ${connection.color}`}
              style={{
                left: `${from.x}%`,
                top: `${from.y}%`,
                width: `${length}%`,
                transform: `translateY(-50%) rotate(${angle}deg)`,
              }}
            />
            <div
              className="amount"
              style={{
                left: `${(from.x + to.x) / 2}%`,
                top: `${(from.y + to.y) / 2}%`,
              }}
            >
              {connection.amount}
            </div>
          </div>
        );
      })}

      {nodes.map((node) => (
        <div
          key={node.id}
          className="person"
          style={{ left: `${node.x}%`, top: `${node.y}%` }}
        >
          <div className={`person-avatar tone-${node.tone}`}>{node.initials}</div>
          <div className="person-name">{node.name}</div>
          <div className="person-balance">{node.balance}</div>
        </div>
      ))}
    </div>
  );
}
