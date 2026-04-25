import Link from "next/link";

const menuItems = ["Overview", "Friends", "Groups", "Activity", "Settle Up"];

export function HomeSidebar() {
  return (
    <aside className="home-sidebar">
      <Link href="/landing" className="home-brand">
        <div className="home-brand-mark">ON</div>
        <div>
          <p className="home-brand-name">OweNone</p>
          <p className="home-brand-tag">Untangle group IOUs</p>
        </div>
      </Link>

      <p className="home-sidebar-label">Menu</p>
      <nav className="home-nav">
        {menuItems.map((item, index) => (
          <button
            key={item}
            type="button"
            className={`home-nav-item ${index === 0 ? "is-active" : ""}`}
          >
            {item}
          </button>
        ))}
      </nav>

      <div className="home-sidebar-user">
        <div className="home-avatar tone-indigo">SJ</div>
        <div>
          <p className="home-user-name">Sarah J.</p>
          <p className="home-user-email">sarah@example.com</p>
        </div>
      </div>
    </aside>
  );
}
