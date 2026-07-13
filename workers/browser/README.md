# Browser Worker

Planned Playwright worker for browser smoke tests.

Responsibilities:

- Visit the configured frontend URL.
- Use test account credentials when configured.
- Capture screenshots and traces when enabled.
- Collect console errors and failed network requests.

All navigation and network activity must respect project allowed hosts.
