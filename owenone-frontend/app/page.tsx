export default function Home() {
  const markup = `
  <style>
  :root{--bg:#080B12;--bg2:#0D1120;--bg3:#111826;--blue:#3B82F6;--blue-light:#60A5FA;--blue-glow:rgba(59,130,246,0.18);--teal:#14B8A6;--teal-glow:rgba(20,184,166,0.15);--red:#F43F5E;--red-glow:rgba(244,63,94,0.15);--amber:#F59E0B;--green:#10B981;--text:#F1F5F9;--text2:#94A3B8;--text3:#475569;--glass:rgba(255,255,255,0.04);--glass-border:rgba(255,255,255,0.08);--glass-hover:rgba(255,255,255,0.07)}*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}body{background:var(--bg);color:var(--text);font-family:'DM Sans',sans-serif;font-size:15px;line-height:1.6;overflow-x:hidden}.bg-atmosphere{position:fixed;inset:0;pointer-events:none;z-index:0;overflow:hidden}.orb{position:absolute;border-radius:50%;filter:blur(120px);opacity:.35}.orb-1{width:700px;height:700px;background:radial-gradient(circle,#1d4ed8 0%,transparent 70%);top:-200px;left:-100px}.orb-2{width:500px;height:500px;background:radial-gradient(circle,#0f766e 0%,transparent 70%);top:40%;right:-100px}.orb-3{width:400px;height:400px;background:radial-gradient(circle,#7c3aed 0%,transparent 70%);bottom:-100px;left:30%;opacity:.2}.noise{position:fixed;inset:0;pointer-events:none;z-index:1;opacity:.03;background-image:url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E")}.container{max-width:1240px;margin:0 auto;padding:0 32px;position:relative;z-index:2}nav{position:sticky;top:0;z-index:100;backdrop-filter:blur(24px);background:rgba(8,11,18,.8);border-bottom:1px solid var(--glass-border)}.nav-inner{display:flex;align-items:center;gap:0;height:64px;max-width:1240px;margin:0 auto;padding:0 32px}.logo{display:flex;align-items:center;gap:10px;text-decoration:none;margin-right:auto}.logo-mark{width:34px;height:34px;border-radius:10px;background:linear-gradient(135deg,#3B82F6 0%,#14B8A6 100%);display:flex;align-items:center;justify-content:center;box-shadow:0 0 20px rgba(59,130,246,.4)}.logo-mark svg{width:18px;height:18px}.logo-name{font-family:'Syne',sans-serif;font-weight:800;font-size:18px;letter-spacing:-.02em;color:var(--text)}.logo-name span{color:var(--blue-light)}.nav-links{display:flex;gap:2px;align-items:center}.nav-links a{padding:6px 14px;border-radius:8px;font-size:14px;font-weight:400;color:var(--text2);text-decoration:none;transition:all .2s}.nav-links a:hover{color:var(--text);background:var(--glass)}.nav-right{display:flex;align-items:center;gap:12px;margin-left:20px}.avatar-pill{display:flex;align-items:center;gap:8px;padding:4px 10px 4px 4px;border:1px solid var(--glass-border);border-radius:100px;background:var(--glass);cursor:pointer}.avatar{width:28px;height:28px;border-radius:50%;background:linear-gradient(135deg,#3B82F6,#8B5CF6);display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:600;color:#fff}.avatar-name{font-size:13px;color:var(--text2)}.btn-cta{background:var(--blue);color:#fff;border:none;padding:9px 18px;border-radius:10px;font-family:'DM Sans',sans-serif;font-size:14px;font-weight:500;cursor:pointer;transition:all .2s;box-shadow:0 0 24px rgba(59,130,246,.35)}.btn-cta:hover{background:#2563eb;box-shadow:0 0 32px rgba(59,130,246,.5);transform:translateY(-1px)}.hero{padding:88px 0 0}.hero-eyebrow{display:inline-flex;align-items:center;gap:8px;background:rgba(59,130,246,.1);border:1px solid rgba(59,130,246,.25);border-radius:100px;padding:5px 14px;margin-bottom:28px}.eyebrow-dot{width:6px;height:6px;border-radius:50%;background:var(--blue);box-shadow:0 0 8px var(--blue);animation:pulse 2s infinite}@keyframes pulse{0%,100%{opacity:1}50%{opacity:.4}}.eyebrow-text{font-size:12px;font-weight:500;color:var(--blue-light);letter-spacing:.06em;text-transform:uppercase}.hero h1{font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(38px,5.5vw,66px);line-height:1.08;letter-spacing:-.03em;color:var(--text);max-width:780px;margin-bottom:22px}.hero h1 em{font-style:normal;color:var(--blue-light)}.hero-sub{font-size:18px;color:var(--text2);max-width:560px;line-height:1.65;font-weight:300;margin-bottom:40px}.hero-sub strong{color:var(--text);font-weight:500}.hero-actions{display:flex;align-items:center;gap:16px;margin-bottom:80px}.btn-primary{background:linear-gradient(135deg,#3B82F6 0%,#2563eb 100%);color:#fff;border:none;padding:14px 28px;border-radius:12px;font-family:'DM Sans',sans-serif;font-size:16px;font-weight:500;cursor:pointer;transition:all .2s;box-shadow:0 4px 32px rgba(59,130,246,.4),0 0 0 1px rgba(59,130,246,.3);display:flex;align-items:center;gap:8px}.btn-primary:hover{transform:translateY(-2px);box-shadow:0 8px 40px rgba(59,130,246,.5)}.btn-ghost{background:var(--glass);color:var(--text2);border:1px solid var(--glass-border);padding:14px 24px;border-radius:12px;font-family:'DM Sans',sans-serif;font-size:15px;cursor:pointer;display:flex;align-items:center;gap:8px;transition:all .2s}.btn-ghost:hover{color:var(--text);background:var(--glass-hover)}.trust-bar{display:flex;align-items:center;gap:20px}.trust-avatars{display:flex}.trust-avatars .av{width:28px;height:28px;border-radius:50%;border:2px solid var(--bg);margin-left:-8px;font-size:10px;font-weight:600;display:flex;align-items:center;justify-content:center}.trust-avatars .av:first-child{margin-left:0}.trust-text{font-size:13px;color:var(--text3)}.trust-text strong{color:var(--text2)}.viz-section{position:relative;margin:0 -32px 0;padding:0 32px 80px}.viz-container{position:relative;background:linear-gradient(180deg,rgba(13,17,32,.5) 0%,rgba(8,11,18,.9) 100%);border:1px solid var(--glass-border);border-radius:24px;overflow:hidden;backdrop-filter:blur(12px)}.viz-header{display:flex;align-items:center;justify-content:space-between;padding:20px 28px 0}.viz-title{font-family:'Syne',sans-serif;font-size:13px;font-weight:600;color:var(--text3);letter-spacing:.08em;text-transform:uppercase}.ws-indicator{display:flex;align-items:center;gap:7px;background:rgba(16,185,129,.1);border:1px solid rgba(16,185,129,.25);border-radius:100px;padding:4px 12px}.ws-dot{width:6px;height:6px;border-radius:50%;background:var(--green);animation:pulse 1.5s infinite}.ws-text{font-size:11px;color:var(--green);font-weight:500;letter-spacing:.04em}.viz-body{display:grid;grid-template-columns:1fr auto 1fr;align-items:center;gap:0;padding:32px 20px 36px;min-height:380px}.graph-panel{position:relative;padding:0 10px}.graph-label{font-family:'Syne',sans-serif;font-size:12px;font-weight:700;letter-spacing:.1em;text-transform:uppercase;text-align:center;margin-bottom:16px}.graph-label.before{color:var(--red)}.graph-label.after{color:var(--teal)}.graph-svg{width:100%;max-width:440px;margin:0 auto;display:block}.sweep-col{display:flex;flex-direction:column;align-items:center;gap:12px;padding:0 16px;flex-shrink:0;width:140px}.sweep-label{font-family:'Syne',sans-serif;font-size:11px;font-weight:700;color:var(--blue-light);letter-spacing:.06em;text-align:center;text-transform:uppercase}.sweep-arrow{display:flex;flex-direction:column;align-items:center;gap:6px}.sweep-badge{background:linear-gradient(135deg,rgba(59,130,246,.2),rgba(20,184,166,.15));border:1px solid rgba(59,130,246,.3);border-radius:10px;padding:10px 14px;text-align:center}.sweep-pct{font-family:'Syne',sans-serif;font-size:26px;font-weight:800;color:var(--blue-light);line-height:1}.sweep-pct-label{font-size:10px;color:var(--text3);margin-top:2px}.sweep-icon{font-size:28px;color:var(--blue-light);filter:drop-shadow(0 0 12px rgba(59,130,246,.6))}.result-card{position:absolute;bottom:-10px;left:50%;transform:translateX(-50%);background:linear-gradient(135deg,rgba(20,184,166,.12),rgba(16,185,129,.08));border:1px solid rgba(20,184,166,.3);border-radius:14px;padding:12px 18px;text-align:center;backdrop-filter:blur(8px);white-space:nowrap;box-shadow:0 0 30px rgba(20,184,166,.15)}.result-label{font-size:11px;color:var(--teal);font-weight:500;letter-spacing:.04em;margin-bottom:3px}.result-amount{font-family:'Syne',sans-serif;font-size:18px;font-weight:800;color:#fff}.result-sub{font-size:10px;color:var(--text3);margin-top:2px}.node-group{cursor:default}.node-bg{filter:url(#glow)}.features{padding:80px 0}.features-eyebrow{font-family:'Syne',sans-serif;font-size:11px;font-weight:700;color:var(--blue-light);letter-spacing:.1em;text-transform:uppercase;margin-bottom:16px}.features h2{font-family:'Syne',sans-serif;font-weight:800;font-size:36px;letter-spacing:-.03em;margin-bottom:12px}.features-sub{color:var(--text2);font-size:16px;max-width:480px;margin-bottom:52px;font-weight:300}.features-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:16px}.feat-card{background:var(--glass);border:1px solid var(--glass-border);border-radius:20px;padding:28px;transition:all .25s;position:relative;overflow:hidden}.feat-card::before{content:'';position:absolute;inset:0;border-radius:20px;background:linear-gradient(135deg,transparent 0%,rgba(255,255,255,.02) 100%);pointer-events:none}.feat-card:hover{border-color:rgba(59,130,246,.3);background:var(--glass-hover);transform:translateY(-3px)}.feat-icon{width:48px;height:48px;border-radius:14px;margin-bottom:20px;display:flex;align-items:center;justify-content:center}.icon-blue{background:rgba(59,130,246,.12);border:1px solid rgba(59,130,246,.2)}.icon-teal{background:rgba(20,184,166,.12);border:1px solid rgba(20,184,166,.2)}.icon-purple{background:rgba(139,92,246,.12);border:1px solid rgba(139,92,246,.2)}.feat-title{font-family:'Syne',sans-serif;font-weight:700;font-size:17px;margin-bottom:10px;letter-spacing:-.01em}.feat-desc{font-size:14px;color:var(--text2);line-height:1.65;font-weight:300}.feat-stat{margin-top:20px;padding-top:16px;border-top:1px solid var(--glass-border);display:flex;align-items:baseline;gap:6px}.feat-stat-num{font-family:'Syne',sans-serif;font-size:24px;font-weight:800}.feat-stat-label{font-size:12px;color:var(--text3)}.num-blue{color:var(--blue-light)}.num-teal{color:var(--teal)}.num-purple{color:#A78BFA}.bottom-section{display:grid;grid-template-columns:1fr 380px;gap:24px;padding-bottom:100px}.security-panel{background:var(--glass);border:1px solid var(--glass-border);border-radius:20px;padding:32px}.security-title{font-family:'Syne',sans-serif;font-weight:700;font-size:20px;margin-bottom:8px;letter-spacing:-.02em}.security-sub{font-size:14px;color:var(--text2);margin-bottom:28px;font-weight:300}.security-items{display:flex;flex-direction:column;gap:16px}.sec-item{display:flex;align-items:flex-start;gap:14px}.sec-icon{width:40px;height:40px;border-radius:12px;flex-shrink:0;display:flex;align-items:center;justify-content:center;background:rgba(59,130,246,.1);border:1px solid rgba(59,130,246,.2)}.sec-text-title{font-size:14px;font-weight:500;color:var(--text);margin-bottom:2px}.sec-text-desc{font-size:12px;color:var(--text3);line-height:1.5}.cta-panel{background:linear-gradient(160deg,rgba(37,99,235,.15) 0%,rgba(13,17,32,.8) 60%);border:1px solid rgba(59,130,246,.2);border-radius:20px;padding:36px 32px;display:flex;flex-direction:column;gap:0;position:relative;overflow:hidden}.cta-panel::before{content:'';position:absolute;top:-60px;right:-60px;width:200px;height:200px;border-radius:50%;background:radial-gradient(circle,rgba(59,130,246,.2),transparent 70%);pointer-events:none}.cta-tagline{font-family:'Syne',sans-serif;font-size:13px;font-weight:700;color:var(--blue-light);letter-spacing:.08em;text-transform:uppercase;margin-bottom:16px}.cta-heading{font-family:'Syne',sans-serif;font-weight:800;font-size:26px;line-height:1.2;letter-spacing:-.03em;margin-bottom:12px}.cta-sub{font-size:14px;color:var(--text2);margin-bottom:28px;line-height:1.6;font-weight:300}.cta-perks{display:flex;flex-direction:column;gap:10px;margin-bottom:28px}.perk{display:flex;align-items:center;gap:10px;font-size:14px;color:var(--text2)}.perk-check{width:20px;height:20px;border-radius:6px;flex-shrink:0;background:rgba(16,185,129,.15);border:1px solid rgba(16,185,129,.3);display:flex;align-items:center;justify-content:center}.btn-signup{width:100%;background:linear-gradient(135deg,#3B82F6 0%,#2563eb 100%);color:#fff;border:none;padding:16px 24px;border-radius:14px;font-family:'Syne',sans-serif;font-size:17px;font-weight:700;cursor:pointer;letter-spacing:-.01em;box-shadow:0 4px 40px rgba(59,130,246,.45),0 0 0 1px rgba(59,130,246,.3);transition:all .2s;display:flex;align-items:center;justify-content:center;gap:8px}.btn-signup:hover{transform:translateY(-2px);box-shadow:0 8px 50px rgba(59,130,246,.55)}.cta-disclaimer{text-align:center;font-size:11px;color:var(--text3);margin-top:12px}footer{border-top:1px solid var(--glass-border);padding:24px 0;position:relative;z-index:2}.footer-inner{display:flex;align-items:center;justify-content:space-between;max-width:1240px;margin:0 auto;padding:0 32px}.footer-copy{font-size:13px;color:var(--text3)}.footer-links{display:flex;gap:24px}.footer-links a{font-size:13px;color:var(--text3);text-decoration:none}.footer-links a:hover{color:var(--text2)}@keyframes fadeInUp{from{opacity:0;transform:translateY(24px)}to{opacity:1;transform:translateY(0)}}.hero h1{animation:fadeInUp .7s ease both}.hero-sub{animation:fadeInUp .7s .1s ease both}.hero-actions{animation:fadeInUp .7s .2s ease both}.viz-container{animation:fadeInUp .8s .3s ease both}
  </style>

  <div class="bg-atmosphere">
    <div class="orb orb-1"></div>
    <div class="orb orb-2"></div>
    <div class="orb orb-3"></div>
  </div>
  <div class="noise"></div>

  <!-- NAV -->
  <nav>
    <div class="nav-inner">
      <a class="logo" href="#">
        <div class="logo-mark">
          <svg viewBox="0 0 20 20" fill="none">
            <circle cx="5" cy="5" r="2.5" fill="white" opacity="0.9"/>
            <circle cx="15" cy="5" r="2.5" fill="white" opacity="0.9"/>
            <circle cx="5" cy="15" r="2.5" fill="white" opacity="0.9"/>
            <circle cx="15" cy="15" r="2.5" fill="white" opacity="0.9"/>
            <line x1="5" y1="5" x2="15" y2="15" stroke="white" stroke-width="1.5" opacity="0.5"/>
            <line x1="15" y1="5" x2="5" y2="15" stroke="white" stroke-width="1.5" opacity="0.5"/>
            <circle cx="10" cy="10" r="2" fill="white"/>
          </svg>
        </div>
        <span class="logo-name">Owe<span>No</span></span>
      </a>
      <div class="nav-links">
        <a href="#">Features</a>
        <a href="#">How It Works</a>
        <a href="#">Security</a>
        <a href="#">About Us</a>
      </div>
      <div class="nav-right">
        <div class="avatar-pill">
          <div class="avatar">SJ</div>
          <span class="avatar-name">Sarah J.</span>
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" style="color:#475569;margin-left:2px">
            <path d="M3 4.5L6 7.5L9 4.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
        </div>
        <button class="btn-cta">Get Early Access</button>
      </div>
    </div>
  </nav>

  <!-- HERO -->
  <section class="hero">
    <div class="container">
      <div class="hero-eyebrow">
        <div class="eyebrow-dot"></div>
        <span class="eyebrow-text">Introducing the debt graph compressor</span>
      </div>
      <h1>Drowning in Group <em>IOUs?</em><br/>OweNo Untangles the Mess.</h1>
      <p class="hero-sub">Stop making five payments across three apps. The first <strong>social graph debt compressor</strong> uses smart math to find your single optimal path. Instant sync, zero friction.</p>
      <div class="hero-actions">
        <button class="btn-primary">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M8 1.5L14.5 8L8 14.5M1.5 8H14.5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
          Start for Free
        </button>
        <button class="btn-ghost">
          <svg width="15" height="15" viewBox="0 0 15 15" fill="none"><circle cx="7.5" cy="7.5" r="6.5" stroke="currentColor" stroke-width="1.3"/><path d="M6 5.5L9.5 7.5L6 9.5V5.5Z" fill="currentColor"/></svg>
          Watch demo
        </button>
      </div>
      <div class="trust-bar">
        <div class="trust-avatars">
          <div class="av" style="background:linear-gradient(135deg,#3B82F6,#8B5CF6)">SJ</div>
          <div class="av" style="background:linear-gradient(135deg,#14B8A6,#10B981)">MK</div>
          <div class="av" style="background:linear-gradient(135deg,#F59E0B,#EF4444)">JK</div>
          <div class="av" style="background:linear-gradient(135deg,#8B5CF6,#EC4899)">AL</div>
          <div class="av" style="background:linear-gradient(135deg,#06B6D4,#3B82F6)">RP</div>
        </div>
        <span class="trust-text"><strong>2,400+ people</strong> on the waitlist</span>
      </div>
    </div>
  </section>

  <!-- rest of page retained as markup... -->
`;

  return <div dangerouslySetInnerHTML={{ __html: markup }} />;
}