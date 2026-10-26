# Google Tasks reference layout

User requests screenshot-matched My Tasks UI, push/release and installation on both
Workspace PCs. Repository-history loaded; baseline954c7e3/public preview2 clean.
Independent original backend/apphost and its immutable identity/data contracts
remain. No Google synchronization or Gmail integration claim.

AUI screens implement peach mobile canvas, centered title/avatar, horizontal list
tabs, rounded task cards, circle/star actions, expandable Completed card, add FAB
and bottom-sheet editor. Local Roboto font5.3.0 bundled with license. Extra tools
remain in list/account dialogs; Inbox is a display alias My Tasks, not a DB rename.
More options retains recurrence/timezone/sharing/priority and other existing fields.
Fast list-switch regression found that a stale visible row could receive a click
as the new list replaced it. Render only the matching loaded filter and discard
stale load revisions; the complete browser flows pass after correction.
Nested AUI If/For needs an explicit Stack wrapper for correct compiler output.

Windows Go tests/vet and frontend contracts/build pass. First Go tests correctly
refused exFAT temp; rerun used dedicated NTFS cache. Native Chrome synthetic checks
cover real create/edit/star/complete/reopen/reload/list/new-list/account isolation;
384x776/320x640 dark/1440x1000 captures inspected. Owner data never a fixture.
Source/build artifacts on I:/Projects/tooling/tasks-google-20261026; temporary
NTFS build/fixture root C:/Projects/.cache/tasks-google-20261026. Frontend dependencies
on exFAT were slow; clean NTFS npm ci/build worked. Track fixture PID receipt and
stop/clean only task-owned paths after publication.

Next: final clean source packaging of preview3; verify asset SHA256/source tag,
publish and promote public channel in dedicated workspace-tasks-ui worktree.
Both owners currently Workspace92; use normal backed-up tasks-only apply and verify
served assets/all services/version retention. Do not overwrite concurrent Maps or
other independent updates. Public releases are immutable; Actions remain disabled.

## Preview3 installed; final title correction

Preview3 source85f035b/public assets and channelb6f88cc were verified; normal
Tasks-only updates completed on BOTH PCs. Primary exact byte/config/identity
receipt passed, all27 services and all JS/CSS/fonts matched. Secondary normal
update completed in2m24s, all27 public services running, Tasks3 runtime and other
25 displayed versions retained. Its stale progress page needed one read-only
reload after restart; no second apply was submitted.

Live inspection found an introduced double-encoded separator in the browser tab
title. Preparing immutable preview4 with plain Tasks title and updated version
metadata, preserving the already accepted layout/behavior. Do not claim final
preview4 publication/installation until its receipts exist. Old3 remains immutable.
