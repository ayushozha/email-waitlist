package handler

import "net/http"

func HomepageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(homepageHTML))
}

const homepageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Email Waitlist - Multi-tenant email collection microservice</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0a0a0a; color: #e0e0e0; line-height: 1.6; overflow-x: hidden; }
  a { color: #60a5fa; text-decoration: none; }
  a:hover { text-decoration: underline; }
  code { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.85em; background: #1a1a2e; padding: 0.15em 0.4em; border-radius: 4px; }

  /* Nav */
  nav { display: flex; align-items: center; justify-content: space-between; max-width: 1100px; margin: 0 auto; padding: 1.25rem 1.5rem; }
  .nav-brand { display: flex; align-items: center; gap: 0.6rem; font-weight: 700; font-size: 1.05rem; color: #fff; }
  .nav-brand svg { flex-shrink: 0; }
  .nav-links { display: flex; gap: 1.5rem; align-items: center; }
  .nav-links a { color: #999; font-size: 0.9rem; transition: color 0.15s; }
  .nav-links a:hover { color: #fff; text-decoration: none; }
  .nav-links .btn-sm { background: #1e2a4a; color: #60a5fa; padding: 0.4rem 1rem; border-radius: 6px; font-weight: 600; font-size: 0.85rem; }
  .nav-links .btn-sm:hover { background: #253561; }

  /* Hero */
  .hero { max-width: 1100px; margin: 0 auto; padding: 5rem 1.5rem 4rem; text-align: center; }
  .badge { display: inline-block; background: #0d3320; color: #34d399; font-size: 0.78rem; font-weight: 600; padding: 0.3rem 0.85rem; border-radius: 100px; margin-bottom: 1.5rem; letter-spacing: 0.3px; }
  .hero h1 { font-size: clamp(2.25rem, 5vw, 3.5rem); color: #fff; line-height: 1.15; margin-bottom: 1.25rem; font-weight: 800; letter-spacing: -0.02em; }
  .hero h1 span { background: linear-gradient(135deg, #60a5fa, #a78bfa); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
  .hero p { font-size: 1.15rem; color: #888; max-width: 600px; margin: 0 auto 2.5rem; }
  .hero-actions { display: flex; gap: 1rem; justify-content: center; flex-wrap: wrap; }
  .btn { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.75rem; border-radius: 8px; font-weight: 600; font-size: 0.95rem; transition: all 0.15s; }
  .btn-primary { background: #fff; color: #0a0a0a; }
  .btn-primary:hover { background: #e0e0e0; text-decoration: none; }
  .btn-secondary { background: #161616; color: #ccc; border: 1px solid #333; }
  .btn-secondary:hover { background: #1a1a1a; border-color: #444; text-decoration: none; }

  /* Code preview */
  .code-preview { max-width: 680px; margin: 3.5rem auto 0; background: #111; border: 1px solid #222; border-radius: 12px; overflow: hidden; text-align: left; }
  .code-tabs { display: flex; border-bottom: 1px solid #222; }
  .code-tab { padding: 0.6rem 1.25rem; font-size: 0.8rem; color: #666; cursor: pointer; border-bottom: 2px solid transparent; transition: all 0.15s; }
  .code-tab.active { color: #60a5fa; border-bottom-color: #60a5fa; }
  .code-block { padding: 1.25rem; font-size: 0.82rem; line-height: 1.65; overflow-x: auto; }
  .code-block pre { margin: 0; white-space: pre; }
  .code-block .kw { color: #c084fc; }
  .code-block .fn { color: #60a5fa; }
  .code-block .str { color: #34d399; }
  .code-block .cm { color: #555; }
  .code-panel { display: none; }
  .code-panel.active { display: block; }

  /* Features */
  .features { max-width: 1100px; margin: 0 auto; padding: 4rem 1.5rem; }
  .features h2 { text-align: center; font-size: 1.75rem; color: #fff; margin-bottom: 0.5rem; }
  .features .sub { text-align: center; color: #666; margin-bottom: 3rem; font-size: 1rem; }
  .features-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.25rem; }
  .feature-card { background: #111; border: 1px solid #1a1a1a; border-radius: 10px; padding: 1.5rem; transition: border-color 0.2s; }
  .feature-card:hover { border-color: #333; }
  .feature-icon { width: 2.25rem; height: 2.25rem; border-radius: 8px; display: flex; align-items: center; justify-content: center; margin-bottom: 1rem; font-size: 1.1rem; }
  .fi-blue { background: #0f1d3d; color: #60a5fa; }
  .fi-green { background: #0d3320; color: #34d399; }
  .fi-purple { background: #2d0d2e; color: #c084fc; }
  .fi-amber { background: #2a1f0d; color: #fbbf24; }
  .fi-red { background: #3b1111; color: #f87171; }
  .fi-cyan { background: #0d2d33; color: #22d3ee; }
  .feature-card h3 { color: #fff; font-size: 1rem; margin-bottom: 0.4rem; }
  .feature-card p { color: #777; font-size: 0.88rem; }

  /* How it works */
  .how { max-width: 1100px; margin: 0 auto; padding: 4rem 1.5rem; }
  .how h2 { text-align: center; font-size: 1.75rem; color: #fff; margin-bottom: 0.5rem; }
  .how .sub { text-align: center; color: #666; margin-bottom: 3rem; font-size: 1rem; }
  .steps { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; }
  .step { background: #111; border: 1px solid #1a1a1a; border-radius: 10px; padding: 1.75rem; position: relative; }
  .step-num { position: absolute; top: -0.75rem; left: 1.5rem; background: #1e2a4a; color: #60a5fa; font-weight: 800; font-size: 0.8rem; width: 1.75rem; height: 1.75rem; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
  .step h3 { color: #fff; font-size: 1rem; margin-bottom: 0.5rem; margin-top: 0.5rem; }
  .step p { color: #777; font-size: 0.88rem; margin-bottom: 0.75rem; }
  .step pre { background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 6px; padding: 0.75rem; font-size: 0.78rem; overflow-x: auto; line-height: 1.55; }

  /* Endpoints table */
  .endpoints { max-width: 1100px; margin: 0 auto; padding: 4rem 1.5rem; }
  .endpoints h2 { text-align: center; font-size: 1.75rem; color: #fff; margin-bottom: 0.5rem; }
  .endpoints .sub { text-align: center; color: #666; margin-bottom: 2.5rem; font-size: 1rem; }
  .ep-table { width: 100%; border-collapse: collapse; background: #111; border-radius: 10px; overflow: hidden; border: 1px solid #1a1a1a; }
  .ep-table th { text-align: left; color: #888; font-weight: 600; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.5px; padding: 0.85rem 1.25rem; background: #0d0d0d; border-bottom: 1px solid #1a1a1a; }
  .ep-table td { padding: 0.75rem 1.25rem; border-bottom: 1px solid #141414; font-size: 0.88rem; }
  .ep-table tr:last-child td { border-bottom: none; }
  .ep-table tr:hover td { background: #141414; }
  .method-badge { font-size: 0.72rem; font-weight: 700; padding: 0.2rem 0.5rem; border-radius: 4px; font-family: monospace; letter-spacing: 0.3px; }
  .mb-get { background: #0d3320; color: #34d399; }
  .mb-post { background: #1e2a4a; color: #60a5fa; }
  .mb-delete { background: #3b1111; color: #f87171; }
  .auth-badge { font-size: 0.72rem; padding: 0.15rem 0.45rem; border-radius: 3px; font-family: monospace; }
  .ab-api { background: #2a1f0d; color: #fbbf24; }
  .ab-admin { background: #2d0d2e; color: #c084fc; }
  .ab-none { background: #1a1a1a; color: #666; }

  /* CTA */
  .cta { max-width: 1100px; margin: 0 auto; padding: 4rem 1.5rem; text-align: center; }
  .cta-box { background: linear-gradient(135deg, #111 0%, #0d1525 100%); border: 1px solid #1e2a4a; border-radius: 16px; padding: 3.5rem 2rem; }
  .cta-box h2 { font-size: 1.75rem; color: #fff; margin-bottom: 0.75rem; }
  .cta-box p { color: #888; font-size: 1rem; margin-bottom: 2rem; max-width: 500px; margin-left: auto; margin-right: auto; }

  /* Footer */
  footer { max-width: 1100px; margin: 0 auto; padding: 2rem 1.5rem; display: flex; justify-content: space-between; align-items: center; border-top: 1px solid #1a1a1a; }
  footer p { color: #555; font-size: 0.82rem; }
  footer a { color: #555; font-size: 0.82rem; }
  footer a:hover { color: #888; }

  @media (max-width: 640px) {
    .hero { padding: 3rem 1.25rem 2.5rem; }
    .nav-links a:not(.btn-sm) { display: none; }
    .steps { grid-template-columns: 1fr; }
    .ep-table { font-size: 0.82rem; }
    .ep-table th, .ep-table td { padding: 0.6rem 0.75rem; }
    footer { flex-direction: column; gap: 0.5rem; }
  }
</style>
</head>
<body>

<nav>
  <div class="nav-brand">
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#60a5fa" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
    Email Waitlist
  </div>
  <div class="nav-links">
    <a href="/docs">API Docs</a>
    <a href="/health">Status</a>
    <a href="https://github.com/ayushozha/email-waitlist" target="_blank">GitHub</a>
    <a href="/docs" class="btn-sm">Get Started</a>
  </div>
</nav>

<section class="hero">
  <div class="badge">Open Source Microservice</div>
  <h1>Email collection for<br><span>any project, in minutes</span></h1>
  <p>A multi-tenant, rate-limited, CORS-aware email waitlist API. Register a project, drop in one API call, and start collecting emails.</p>
  <div class="hero-actions">
    <a href="/docs" class="btn btn-primary">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20"/></svg>
      Read the Docs
    </a>
    <a href="https://github.com/ayushozha/email-waitlist" class="btn btn-secondary" target="_blank">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/></svg>
      View Source
    </a>
  </div>

  <div class="code-preview">
    <div class="code-tabs">
      <div class="code-tab active" onclick="switchTab(event, 'tab-fetch')">JavaScript</div>
      <div class="code-tab" onclick="switchTab(event, 'tab-curl')">cURL</div>
      <div class="code-tab" onclick="switchTab(event, 'tab-html')">HTML Form</div>
    </div>
    <div id="tab-fetch" class="code-panel active">
      <div class="code-block"><pre><code><span class="kw">const</span> res = <span class="kw">await</span> <span class="fn">fetch</span>(<span class="str">'https://emailwaitlist.ayushojha.com/api/v1/subscribe'</span>, {
  method: <span class="str">'POST'</span>,
  headers: {
    <span class="str">'Content-Type'</span>: <span class="str">'application/json'</span>,
    <span class="str">'X-API-Key'</span>: <span class="str">'wl_your_project_key'</span>
  },
  body: <span class="fn">JSON.stringify</span>({ email: <span class="str">'user@example.com'</span> })
});

<span class="kw">const</span> data = <span class="kw">await</span> res.<span class="fn">json</span>();
<span class="cm">// { "message": "Successfully joined the waitlist!" }</span></code></pre></div>
    </div>
    <div id="tab-curl" class="code-panel">
      <div class="code-block"><pre><code>curl -X POST https://emailwaitlist.ayushojha.com/api/v1/subscribe \
  -H <span class="str">"Content-Type: application/json"</span> \
  -H <span class="str">"X-API-Key: wl_your_project_key"</span> \
  -d <span class="str">'{"email":"user@example.com"}'</span>

<span class="cm"># Response:</span>
<span class="cm"># {"message":"Successfully joined the waitlist!","subscriber":{...}}</span></code></pre></div>
    </div>
    <div id="tab-html" class="code-panel">
      <div class="code-block"><pre><code><span class="kw">&lt;form</span> id=<span class="str">"waitlist"</span><span class="kw">&gt;</span>
  <span class="kw">&lt;input</span> type=<span class="str">"email"</span> id=<span class="str">"email"</span> placeholder=<span class="str">"you@example.com"</span> required <span class="kw">/&gt;</span>
  <span class="kw">&lt;button</span> type=<span class="str">"submit"</span><span class="kw">&gt;</span>Join Waitlist<span class="kw">&lt;/button&gt;</span>
<span class="kw">&lt;/form&gt;</span>

<span class="kw">&lt;script&gt;</span>
document.<span class="fn">getElementById</span>(<span class="str">'waitlist'</span>).<span class="fn">addEventListener</span>(<span class="str">'submit'</span>, <span class="kw">async</span> (e) => {
  e.<span class="fn">preventDefault</span>();
  <span class="kw">const</span> res = <span class="kw">await</span> <span class="fn">fetch</span>(<span class="str">'https://emailwaitlist.ayushojha.com/api/v1/subscribe'</span>, {
    method: <span class="str">'POST'</span>,
    headers: { <span class="str">'Content-Type'</span>: <span class="str">'application/json'</span>, <span class="str">'X-API-Key'</span>: <span class="str">'wl_...'</span> },
    body: <span class="fn">JSON.stringify</span>({ email: document.<span class="fn">getElementById</span>(<span class="str">'email'</span>).value })
  });
});
<span class="kw">&lt;/script&gt;</span></code></pre></div>
    </div>
  </div>
</section>

<section class="features">
  <h2>Built for developers</h2>
  <p class="sub">Everything you need for email collection, nothing you don't.</p>
  <div class="features-grid">
    <div class="feature-card">
      <div class="feature-icon fi-blue">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
      </div>
      <h3>Multi-tenant</h3>
      <p>Each project gets isolated API keys and subscriber lists. One instance serves all your apps.</p>
    </div>
    <div class="feature-card">
      <div class="feature-icon fi-green">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
      </div>
      <h3>Rate limited</h3>
      <p>Built-in per-IP rate limiting on subscribe endpoints prevents abuse without extra infrastructure.</p>
    </div>
    <div class="feature-card">
      <div class="feature-icon fi-purple">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>
      </div>
      <h3>CORS-aware</h3>
      <p>Per-project allowed origins. Works seamlessly from any frontend framework or static site.</p>
    </div>
    <div class="feature-card">
      <div class="feature-icon fi-amber">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
      </div>
      <h3>CSV export</h3>
      <p>Export your entire subscriber list as CSV with one API call. Ready for Mailchimp, Resend, or a spreadsheet.</p>
    </div>
    <div class="feature-card">
      <div class="feature-icon fi-red">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"/><path d="m19 9-5 5-4-4-3 3"/></svg>
      </div>
      <h3>Built-in stats</h3>
      <p>Subscriber counts by day, week, month. Track growth trends without a separate analytics tool.</p>
    </div>
    <div class="feature-card">
      <div class="feature-icon fi-cyan">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
      </div>
      <h3>Custom metadata</h3>
      <p>Attach arbitrary JSON to each subscriber &mdash; name, referral source, plan interest, or anything else.</p>
    </div>
  </div>
</section>

<section class="how">
  <h2>Three steps to integrate</h2>
  <p class="sub">Add email collection to any project in under 5 minutes.</p>
  <div class="steps">
    <div class="step">
      <div class="step-num">1</div>
      <h3>Register your project</h3>
      <p>Create a project with the admin key. You'll get a unique <code>wl_</code> API key.</p>
      <pre><code>curl -X POST /api/v1/projects \
  -H "X-Admin-Key: $ADMIN_KEY" \
  -d '{"name":"My App","slug":"my-app",
       "allowed_origins":["https://myapp.com"]}'</code></pre>
    </div>
    <div class="step">
      <div class="step-num">2</div>
      <h3>Add the subscribe call</h3>
      <p>One POST request from your frontend collects the email.</p>
      <pre><code>fetch('/api/v1/subscribe', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'wl_your_key'
  },
  body: JSON.stringify({
    email: 'user@example.com'
  })
})</code></pre>
    </div>
    <div class="step">
      <div class="step-num">3</div>
      <h3>View &amp; export</h3>
      <p>List subscribers, check stats, or export as CSV anytime.</p>
      <pre><code><span style="color:#888;"># Stats</span>
GET /api/v1/stats

<span style="color:#888;"># List</span>
GET /api/v1/subscribers?limit=50

<span style="color:#888;"># Export</span>
GET /api/v1/subscribers/export</code></pre>
    </div>
  </div>
</section>

<section class="endpoints">
  <h2>API at a glance</h2>
  <p class="sub">Full documentation available at <a href="/docs">/docs</a></p>
  <table class="ep-table">
    <thead>
      <tr><th>Method</th><th>Endpoint</th><th>Description</th><th>Auth</th></tr>
    </thead>
    <tbody>
      <tr>
        <td><span class="method-badge mb-post">POST</span></td>
        <td><code>/api/v1/subscribe</code></td>
        <td>Collect an email address</td>
        <td><span class="auth-badge ab-api">API Key</span></td>
      </tr>
      <tr>
        <td><span class="method-badge mb-get">GET</span></td>
        <td><code>/api/v1/subscribers</code></td>
        <td>List subscribers (paginated)</td>
        <td><span class="auth-badge ab-api">API Key</span></td>
      </tr>
      <tr>
        <td><span class="method-badge mb-get">GET</span></td>
        <td><code>/api/v1/subscribers/export</code></td>
        <td>Export subscribers as CSV</td>
        <td><span class="auth-badge ab-api">API Key</span></td>
      </tr>
      <tr>
        <td><span class="method-badge mb-delete">DELETE</span></td>
        <td><code>/api/v1/subscribers/{email}</code></td>
        <td>Remove a subscriber</td>
        <td><span class="auth-badge ab-api">API Key</span></td>
      </tr>
      <tr>
        <td><span class="method-badge mb-get">GET</span></td>
        <td><code>/api/v1/stats</code></td>
        <td>Subscriber growth stats</td>
        <td><span class="auth-badge ab-api">API Key</span></td>
      </tr>
      <tr>
        <td><span class="method-badge mb-post">POST</span></td>
        <td><code>/api/v1/projects</code></td>
        <td>Create a new project</td>
        <td><span class="auth-badge ab-admin">Admin</span></td>
      </tr>
      <tr>
        <td><span class="method-badge mb-get">GET</span></td>
        <td><code>/api/v1/projects</code></td>
        <td>List all projects</td>
        <td><span class="auth-badge ab-admin">Admin</span></td>
      </tr>
      <tr>
        <td><span class="method-badge mb-get">GET</span></td>
        <td><code>/health</code></td>
        <td>Health check</td>
        <td><span class="auth-badge ab-none">None</span></td>
      </tr>
    </tbody>
  </table>
</section>

<section class="cta">
  <div class="cta-box">
    <h2>Ready to start collecting emails?</h2>
    <p>Read the full API documentation with integration examples for React, vanilla JS, and more.</p>
    <a href="/docs" class="btn btn-primary">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20"/></svg>
      View Full Docs
    </a>
  </div>
</section>

<footer>
  <p>Built by <a href="https://ayushojha.com" target="_blank">Ayush Ojha</a></p>
  <a href="https://github.com/ayushozha/email-waitlist" target="_blank">GitHub</a>
</footer>

<script>
function switchTab(e, id) {
  document.querySelectorAll('.code-tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('.code-panel').forEach(p => p.classList.remove('active'));
  e.target.classList.add('active');
  document.getElementById(id).classList.add('active');
}
</script>

</body>
</html>`
