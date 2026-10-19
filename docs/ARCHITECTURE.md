# ADR001: My Tasks

Decision: original Apache-2.0 app at oneworkspacestudio/my-tasks. SQLite tasks,
lists, role memberships, recurrence rules and materialised occurrences are separate
records. Use rrule-go (MIT) for RFC5545 rules, embedded IANA tzdata, explicit
exception dates and DST regression tests. Unique occurrence/reminder keys prevent
duplicates. Only current list members can read tasks or notifications. Owner,
editor and viewer roles apply at every query/mutation. Rich description is safe
Markdown, never arbitrary HTML. Reminders have durable retry state and in-app
notifications; there is no central notification delivery service today, nor a
supported background user-mail submission contract. Do not invent one. Calendar
and Notes links are explicit deep links only, with no remote writes. ICS export
is scoped to the signed-in user's accessible tasks; no CalDAV compatibility claim.
Vikunja is AGPL-3.0: no Vikunja source is reused or compatibility asserted.

Shared decision: native Go server, embedded AUI-generated React/TypeScript UI,
SQLite (modernc, BSD-3-Clause), protected per-app NTFS data folders and one EXE.
Reuse the existing One Workspace Apache-2.0 platform, OIDC, branding and supervisor
conventions through a public pkg/apphost package in the actual suite repository.
Canonicalise the suite Go module path to its existing organization repository;
there are no supported external consumers of its internal packages. Each original
app has a dedicated organization repository, rather than a misleading fork.
OIDC uses code+PKCE, state cookie, nonce, issuer/audience verification and immutable
issuer/subject keys. Session/CSRF secrets stay on the server. Local bootstrap is
loopback-only and explicitly disabled in managed mode. No GitHub Actions: user
instruction overrides the brief's Actions-template suggestion; run gates locally.
