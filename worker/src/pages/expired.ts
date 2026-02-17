export function expiredPage(id: string, expiredAt: string): string {
  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Expired - Share</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #0a0a0a;
      color: #e5e5e5;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .container {
      text-align: center;
      max-width: 480px;
      padding: 2rem;
    }
    .code {
      font-size: 6rem;
      font-weight: 700;
      color: #525252;
      line-height: 1;
      margin-bottom: 1rem;
    }
    h1 {
      font-size: 1.5rem;
      font-weight: 600;
      margin-bottom: 0.75rem;
    }
    p {
      color: #a3a3a3;
      line-height: 1.6;
      margin-bottom: 1.5rem;
    }
    .id-ref {
      font-family: "SF Mono", Monaco, "Cascadia Code", monospace;
      background: #1a1a1a;
      border: 1px solid #262626;
      border-radius: 6px;
      padding: 0.25rem 0.5rem;
      font-size: 0.875rem;
      color: #737373;
    }
    .expired-time {
      display: block;
      margin-top: 0.5rem;
      font-size: 0.8rem;
      color: #525252;
    }
    .link {
      display: inline-block;
      margin-top: 1rem;
      color: #525252;
      font-size: 0.8rem;
      text-decoration: none;
    }
    .link:hover { color: #737373; }
  </style>
</head>
<body>
  <div class="container">
    <div class="code">410</div>
    <h1>Share expired</h1>
    <p>
      The share <span class="id-ref">${escapeHtml(id)}</span> has expired
      and is no longer available.
      <span class="expired-time">Expired on ${escapeHtml(expiredAt)}</span>
    </p>
    <a class="link" href="mailto:abuse@techops.services">Report abuse</a>
  </div>
</body>
</html>`;
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}
