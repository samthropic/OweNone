export function HomeHeader() {
  return (
    <header className="home-header">
      <div>
        <h1>Overview</h1>
        <p>Saturday, 25 April 2026</p>
      </div>
      <div className="home-header-actions">
        <button type="button" className="home-btn">
          Add expense
        </button>
        <button type="button" className="home-btn home-btn-primary">
          Pay someone
        </button>
      </div>
    </header>
  );
}
