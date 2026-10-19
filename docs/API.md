# API and integration

All paths below are relative to the app origin/subpath. GET /healthz is public health only. GET /api/bootstrap returns non-secret configuration. GET /api/me returns the signed-in account and session CSRF token. Writes require an exact Origin header plus X-CSRF-Token from that session; use same-origin credentials and never log tokens. Owner-only POST /api/backups streams the encrypted archive. Login/OIDC/session endpoints are provided by the public version-pinned One Workspace apphost package.

GET/POST /api/lists; PUT/DELETE /api/lists/{id}; GET/PUT /api/lists/{id}/members; DELETE /api/lists/{id}/members/{user}; GET/POST /api/tasks; GET/PUT/DELETE /api/tasks/{id}; POST /api/tasks/bulk; GET/POST /api/views; DELETE /api/views/{id}; GET /api/notifications; PUT /api/notifications/{id}; GET /api/time; GET /api/export.ics.

User resources are authorised on every request. Missing/inaccessible user records do not expose private titles. Limits and filters use bounded input; unsupported operations fail explicitly. No API grants direct database access to another app. Managed MyAccounts credentials are server-side only. Search and Drive integrations verify that the source's current immutable subject matches the requesting app, and recheck identity after search. Source permissions apply again when opening a link.
