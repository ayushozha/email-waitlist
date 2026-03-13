package handler

import "net/http"

func DocsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(docsHTML))
}

const docsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Email Waitlist API</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0a0a0a; color: #e0e0e0; line-height: 1.6; }
  .container { max-width: 900px; margin: 0 auto; padding: 2rem 1.5rem; }
  h1 { font-size: 2rem; color: #fff; margin-bottom: 0.25rem; }
  .subtitle { color: #888; margin-bottom: 2.5rem; font-size: 1.05rem; }
  h2 { font-size: 1.1rem; color: #fff; margin: 2.5rem 0 1rem; padding-bottom: 0.5rem; border-bottom: 1px solid #222; }
  .endpoint { background: #111; border: 1px solid #222; border-radius: 8px; margin-bottom: 1rem; overflow: hidden; }
  .endpoint-header { display: flex; align-items: center; gap: 0.75rem; padding: 0.85rem 1rem; cursor: pointer; user-select: none; }
  .endpoint-header:hover { background: #161616; }
  .method { font-size: 0.75rem; font-weight: 700; padding: 0.2rem 0.55rem; border-radius: 4px; font-family: monospace; letter-spacing: 0.5px; }
  .method-get { background: #0d3320; color: #34d399; }
  .method-post { background: #1e2a4a; color: #60a5fa; }
  .method-delete { background: #3b1111; color: #f87171; }
  .path { font-family: monospace; font-size: 0.9rem; color: #ccc; }
  .desc { color: #888; font-size: 0.85rem; margin-left: auto; }
  .auth { font-size: 0.7rem; padding: 0.15rem 0.45rem; border-radius: 3px; font-family: monospace; }
  .auth-api { background: #2a1f0d; color: #fbbf24; }
  .auth-admin { background: #2d0d2e; color: #c084fc; }
  .auth-none { background: #1a1a1a; color: #666; }
  .endpoint-body { display: none; padding: 1rem; border-top: 1px solid #222; background: #0d0d0d; }
  .endpoint.open .endpoint-body { display: block; }
  .endpoint.open .arrow { transform: rotate(90deg); }
  .arrow { color: #555; transition: transform 0.15s; font-size: 0.8rem; }
  pre { background: #161616; border: 1px solid #222; border-radius: 6px; padding: 0.85rem 1rem; overflow-x: auto; font-size: 0.82rem; line-height: 1.55; margin: 0.5rem 0; }
  code { font-family: 'SF Mono', 'Fira Code', monospace; }
  .label { font-size: 0.8rem; color: #888; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; margin: 0.75rem 0 0.35rem; }
  .label:first-child { margin-top: 0; }
  table { width: 100%; border-collapse: collapse; font-size: 0.82rem; margin: 0.5rem 0; }
  th { text-align: left; color: #888; font-weight: 600; padding: 0.4rem 0.6rem; border-bottom: 1px solid #222; }
  td { padding: 0.4rem 0.6rem; border-bottom: 1px solid #1a1a1a; }
  td code { color: #60a5fa; font-size: 0.8rem; }
  .required { color: #f87171; font-size: 0.7rem; }
  .optional { color: #555; font-size: 0.7rem; }
  .quickstart { background: #111; border: 1px solid #222; border-radius: 8px; padding: 1.25rem; margin-bottom: 2rem; }
  .quickstart h3 { color: #fff; font-size: 0.95rem; margin-bottom: 0.75rem; }
  .quickstart p { color: #888; font-size: 0.85rem; margin-bottom: 0.5rem; }
  .response-code { color: #34d399; }
  .step { display: flex; gap: 0.75rem; margin-bottom: 1rem; }
  .step-num { background: #1e2a4a; color: #60a5fa; font-weight: 700; font-size: 0.75rem; width: 1.5rem; height: 1.5rem; border-radius: 50%; display: flex; align-items: center; justify-content: center; flex-shrink: 0; margin-top: 0.15rem; }
  .step-content { flex: 1; }
  .step-content p { color: #ccc; font-size: 0.85rem; margin-bottom: 0.35rem; }
  footer { text-align: center; color: #444; font-size: 0.8rem; margin-top: 3rem; padding-top: 1.5rem; border-top: 1px solid #1a1a1a; }
</style>
</head>
<body>
<div class="container">
  <h1>Email Waitlist API</h1>
  <p class="subtitle">Plug-and-play email collection microservice. Multi-tenant, rate-limited, CORS-aware.</p>

  <div class="quickstart">
    <h3>Quick Start</h3>
    <div class="step">
      <div class="step-num">1</div>
      <div class="step-content">
        <p>Create a project (admin only, one-time)</p>
        <pre><code>curl -X POST https://emailwaitlist.ayushojha.com/api/v1/projects \
  -H "X-Admin-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name":"My App","slug":"my-app","allowed_origins":["https://myapp.com"]}'</code></pre>
      </div>
    </div>
    <div class="step">
      <div class="step-num">2</div>
      <div class="step-content">
        <p>Collect emails from your frontend</p>
        <pre><code>fetch('https://emailwaitlist.ayushojha.com/api/v1/subscribe', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'wl_your_project_api_key'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    metadata: { name: 'Jane', source: 'landing-page' }
  })
})</code></pre>
      </div>
    </div>
  </div>

  <h2>Authentication</h2>
  <p style="color:#999;font-size:0.85rem;margin-bottom:0.5rem;">Two auth mechanisms — use the right header for each endpoint type:</p>
  <table>
    <tr><th>Header</th><th>Value</th><th>Used for</th></tr>
    <tr><td><code>X-API-Key</code></td><td>Project API key (<code>wl_...</code>)</td><td>All project-scoped endpoints</td></tr>
    <tr><td><code>X-Admin-Key</code></td><td>Server admin key</td><td>Project management endpoints</td></tr>
  </table>

  <h2>Endpoints</h2>

  <!-- POST /api/v1/subscribe -->
  <div class="endpoint">
    <div class="endpoint-header" onclick="this.parentElement.classList.toggle('open')">
      <span class="arrow">&#9654;</span>
      <span class="method method-post">POST</span>
      <span class="path">/api/v1/subscribe</span>
      <span class="desc">Collect an email</span>
      <span class="auth auth-api">API Key</span>
    </div>
    <div class="endpoint-body">
      <div class="label">Request Body</div>
      <table>
        <tr><th>Field</th><th>Type</th><th></th><th>Description</th></tr>
        <tr><td><code>email</code></td><td>string</td><td><span class="required">required</span></td><td>Email address (max 320 chars)</td></tr>
        <tr><td><code>metadata</code></td><td>object</td><td><span class="optional">optional</span></td><td>Arbitrary JSON (max 4KB)</td></tr>
      </table>
      <div class="label">Example</div>
      <pre><code>curl -X POST https://emailwaitlist.ayushojha.com/api/v1/subscribe \
  -H "Content-Type: application/json" \
  -H "X-API-Key: wl_abc123" \
  -d '{"email":"user@example.com","metadata":{"name":"Jane"}}'</code></pre>
      <div class="label">Response <span class="response-code">201</span></div>
      <pre><code>{
  "message": "Successfully joined the waitlist!",
  "subscriber": {
    "id": "uuid",
    "project_id": "uuid",
    "email": "user@example.com",
    "metadata": {"name": "Jane"},
    "subscribed_at": "2026-03-12T10:00:00Z"
  }
}</code></pre>
      <div class="label">Errors</div>
      <table>
        <tr><th>Code</th><th>Reason</th></tr>
        <tr><td><code>400</code></td><td>Invalid email or request body</td></tr>
        <tr><td><code>401</code></td><td>Missing or invalid API key</td></tr>
        <tr><td><code>409</code></td><td>Email already subscribed to this project</td></tr>
        <tr><td><code>429</code></td><td>Rate limit exceeded (30 req/min/IP)</td></tr>
      </table>
    </div>
  </div>

  <!-- GET /api/v1/subscribers -->
  <div class="endpoint">
    <div class="endpoint-header" onclick="this.parentElement.classList.toggle('open')">
      <span class="arrow">&#9654;</span>
      <span class="method method-get">GET</span>
      <span class="path">/api/v1/subscribers</span>
      <span class="desc">List emails (paginated)</span>
      <span class="auth auth-api">API Key</span>
    </div>
    <div class="endpoint-body">
      <div class="label">Query Parameters</div>
      <table>
        <tr><th>Param</th><th>Type</th><th>Default</th><th>Description</th></tr>
        <tr><td><code>limit</code></td><td>int</td><td>50</td><td>Results per page (max 500)</td></tr>
        <tr><td><code>offset</code></td><td>int</td><td>0</td><td>Skip N results</td></tr>
      </table>
      <div class="label">Example</div>
      <pre><code>curl https://emailwaitlist.ayushojha.com/api/v1/subscribers?limit=20&offset=0 \
  -H "X-API-Key: wl_abc123"</code></pre>
      <div class="label">Response <span class="response-code">200</span></div>
      <pre><code>{
  "subscribers": [...],
  "total": 142,
  "limit": 20,
  "offset": 0
}</code></pre>
    </div>
  </div>

  <!-- GET /api/v1/subscribers/export -->
  <div class="endpoint">
    <div class="endpoint-header" onclick="this.parentElement.classList.toggle('open')">
      <span class="arrow">&#9654;</span>
      <span class="method method-get">GET</span>
      <span class="path">/api/v1/subscribers/export</span>
      <span class="desc">CSV download</span>
      <span class="auth auth-api">API Key</span>
    </div>
    <div class="endpoint-body">
      <div class="label">Example</div>
      <pre><code>curl https://emailwaitlist.ayushojha.com/api/v1/subscribers/export \
  -H "X-API-Key: wl_abc123" \
  -o subscribers.csv</code></pre>
      <div class="label">Response</div>
      <p style="color:#999;font-size:0.85rem;">Returns a CSV file with columns: <code>email</code>, <code>metadata</code>, <code>subscribed_at</code></p>
    </div>
  </div>

  <!-- DELETE /api/v1/subscribers/{email} -->
  <div class="endpoint">
    <div class="endpoint-header" onclick="this.parentElement.classList.toggle('open')">
      <span class="arrow">&#9654;</span>
      <span class="method method-delete">DELETE</span>
      <span class="path">/api/v1/subscribers/{email}</span>
      <span class="desc">Unsubscribe</span>
      <span class="auth auth-api">API Key</span>
    </div>
    <div class="endpoint-body">
      <div class="label">Example</div>
      <pre><code>curl -X DELETE https://emailwaitlist.ayushojha.com/api/v1/subscribers/user@example.com \
  -H "X-API-Key: wl_abc123"</code></pre>
      <div class="label">Response <span class="response-code">200</span></div>
      <pre><code>{"message": "subscriber removed"}</code></pre>
      <div class="label">Errors</div>
      <table>
        <tr><th>Code</th><th>Reason</th></tr>
        <tr><td><code>404</code></td><td>Email not found in this project</td></tr>
      </table>
    </div>
  </div>

  <!-- GET /api/v1/stats -->
  <div class="endpoint">
    <div class="endpoint-header" onclick="this.parentElement.classList.toggle('open')">
      <span class="arrow">&#9654;</span>
      <span class="method method-get">GET</span>
      <span class="path">/api/v1/stats</span>
      <span class="desc">Dashboard stats</span>
      <span class="auth auth-api">API Key</span>
    </div>
    <div class="endpoint-body">
      <div class="label">Example</div>
      <pre><code>curl https://emailwaitlist.ayushojha.com/api/v1/stats \
  -H "X-API-Key: wl_abc123"</code></pre>
      <div class="label">Response <span class="response-code">200</span></div>
      <pre><code>{
  "total": 142,
  "today": 8,
  "this_week": 34,
  "this_month": 89,
  "by_day": [
    {"date": "2026-03-10", "count": 12},
    {"date": "2026-03-11", "count": 14},
    {"date": "2026-03-12", "count": 8}
  ]
}</code></pre>
    </div>
  </div>

  <!-- POST /api/v1/projects -->
  <div class="endpoint">
    <div class="endpoint-header" onclick="this.parentElement.classList.toggle('open')">
      <span class="arrow">&#9654;</span>
      <span class="method method-post">POST</span>
      <span class="path">/api/v1/projects</span>
      <span class="desc">Create project</span>
      <span class="auth auth-admin">Admin</span>
    </div>
    <div class="endpoint-body">
      <div class="label">Request Body</div>
      <table>
        <tr><th>Field</th><th>Type</th><th></th><th>Description</th></tr>
        <tr><td><code>name</code></td><td>string</td><td><span class="required">required</span></td><td>Display name</td></tr>
        <tr><td><code>slug</code></td><td>string</td><td><span class="required">required</span></td><td>URL-safe identifier (lowercase, hyphens)</td></tr>
        <tr><td><code>allowed_origins</code></td><td>string[]</td><td><span class="optional">optional</span></td><td>CORS origins. Empty = allow all. Use <code>["*"]</code> for wildcard.</td></tr>
      </table>
      <div class="label">Example</div>
      <pre><code>curl -X POST https://emailwaitlist.ayushojha.com/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: YOUR_ADMIN_KEY" \
  -d '{"name":"My App","slug":"my-app","allowed_origins":["https://myapp.com"]}'</code></pre>
      <div class="label">Response <span class="response-code">201</span></div>
      <pre><code>{
  "message": "Project created. Save your API key — it won't be shown again.",
  "project": {
    "id": "uuid",
    "name": "My App",
    "slug": "my-app",
    "api_key": "wl_abc123...",
    "allowed_origins": ["https://myapp.com"],
    "created_at": "2026-03-12T10:00:00Z"
  }
}</code></pre>
    </div>
  </div>

  <!-- GET /api/v1/projects -->
  <div class="endpoint">
    <div class="endpoint-header" onclick="this.parentElement.classList.toggle('open')">
      <span class="arrow">&#9654;</span>
      <span class="method method-get">GET</span>
      <span class="path">/api/v1/projects</span>
      <span class="desc">List projects</span>
      <span class="auth auth-admin">Admin</span>
    </div>
    <div class="endpoint-body">
      <div class="label">Example</div>
      <pre><code>curl https://emailwaitlist.ayushojha.com/api/v1/projects \
  -H "X-Admin-Key: YOUR_ADMIN_KEY"</code></pre>
      <div class="label">Response <span class="response-code">200</span></div>
      <pre><code>{"projects": [...]}</code></pre>
    </div>
  </div>

  <!-- GET /health -->
  <div class="endpoint">
    <div class="endpoint-header" onclick="this.parentElement.classList.toggle('open')">
      <span class="arrow">&#9654;</span>
      <span class="method method-get">GET</span>
      <span class="path">/health</span>
      <span class="desc">Health check</span>
      <span class="auth auth-none">None</span>
    </div>
    <div class="endpoint-body">
      <div class="label">Response <span class="response-code">200</span></div>
      <pre><code>{"status": "ok"}</code></pre>
    </div>
  </div>

  <h2>Integration Guide</h2>
  <p style="color:#999;font-size:0.85rem;margin-bottom:1.25rem;">Follow these three steps to add email collection to any website.</p>

  <div class="quickstart">
    <h3>Step 1 &mdash; Register your site</h3>
    <p>Create a project to get an API key. Run this once per website (requires the admin key).</p>
    <pre><code>curl -X POST https://emailwaitlist.ayushojha.com/api/v1/projects \
  -H "X-Admin-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Website",
    "slug": "my-website",
    "allowed_origins": ["https://mywebsite.com"]
  }'</code></pre>
    <p>Save the <code>api_key</code> from the response &mdash; it starts with <code>wl_</code> and won't be shown again.</p>
  </div>

  <div class="quickstart">
    <h3>Step 2 &mdash; Add the form to your frontend</h3>
    <p>Drop this into any page. Works with React, Vue, Svelte, plain HTML &mdash; anything that can call <code>fetch</code>.</p>
    <div class="label">React / Next.js</div>
    <pre><code>async function subscribe(email, name) {
  const res = await fetch('https://emailwaitlist.ayushojha.com/api/v1/subscribe', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'wl_your_project_api_key'
    },
    body: JSON.stringify({
      email,
      metadata: { name, source: 'landing-page' }
    })
  });

  const data = await res.json();

  if (res.ok) {
    // Success &mdash; show confirmation
    return { success: true, message: data.message };
  } else if (res.status === 409) {
    // Already subscribed
    return { success: false, message: 'You're already on the list!' };
  } else if (res.status === 429) {
    // Rate limited
    return { success: false, message: 'Too many requests. Try again shortly.' };
  } else {
    return { success: false, message: data.error || 'Something went wrong.' };
  }
}</code></pre>
    <div class="label">Plain HTML + Vanilla JS</div>
    <pre><code>&lt;form id="waitlist-form"&gt;
  &lt;input type="email" id="wl-email" placeholder="you@example.com" required /&gt;
  &lt;button type="submit"&gt;Join Waitlist&lt;/button&gt;
  &lt;p id="wl-msg"&gt;&lt;/p&gt;
&lt;/form&gt;

&lt;script&gt;
document.getElementById('waitlist-form').addEventListener('submit', async (e) =&gt; {
  e.preventDefault();
  const email = document.getElementById('wl-email').value;
  const msg = document.getElementById('wl-msg');

  const res = await fetch('https://emailwaitlist.ayushojha.com/api/v1/subscribe', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'wl_your_project_api_key'
    },
    body: JSON.stringify({ email })
  });

  const data = await res.json();
  msg.textContent = res.ok ? data.message : data.error;
});
&lt;/script&gt;</code></pre>
  </div>

  <div class="quickstart">
    <h3>Step 3 &mdash; Manage your subscribers</h3>
    <p>Use these endpoints with the same API key to view, export, or remove subscribers.</p>
    <table>
      <tr><th>Action</th><th>Request</th></tr>
      <tr><td>List subscribers</td><td><code>GET /api/v1/subscribers?limit=50&amp;offset=0</code></td></tr>
      <tr><td>Export CSV</td><td><code>GET /api/v1/subscribers/export</code></td></tr>
      <tr><td>Unsubscribe</td><td><code>DELETE /api/v1/subscribers/{email}</code></td></tr>
      <tr><td>Dashboard stats</td><td><code>GET /api/v1/stats</code></td></tr>
    </table>
    <p style="margin-top:0.75rem;">All requests require the <code>X-API-Key</code> header.</p>
  </div>

  <div class="quickstart">
    <h3>Response Handling Cheatsheet</h3>
    <table>
      <tr><th>Status</th><th>Meaning</th><th>What to show the user</th></tr>
      <tr><td><code>201</code></td><td>Subscribed</td><td>Success message</td></tr>
      <tr><td><code>400</code></td><td>Bad email</td><td>"Please enter a valid email"</td></tr>
      <tr><td><code>409</code></td><td>Duplicate</td><td>"You're already on the waitlist"</td></tr>
      <tr><td><code>429</code></td><td>Rate limited</td><td>"Please try again in a minute"</td></tr>
      <tr><td><code>401</code></td><td>Bad API key</td><td>Check your X-API-Key header</td></tr>
    </table>
  </div>

  <h2>Rate Limiting</h2>
  <p style="color:#999;font-size:0.85rem;">The <code>POST /api/v1/subscribe</code> endpoint is rate-limited to <strong>30 requests per minute per IP</strong>. When exceeded, the API returns <code>429 Too Many Requests</code>. Other endpoints are not rate-limited.</p>

  <h2>CORS</h2>
  <p style="color:#999;font-size:0.85rem;">Each project can define <code>allowed_origins</code> to restrict which domains can call the API from browsers. If no origins are set, all origins are allowed. The API handles <code>OPTIONS</code> preflight requests automatically.</p>

  <footer>Email Waitlist API &mdash; Built by Ayush Ojha</footer>
</div>
</body>
</html>`
