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

## Final publication and both-PC delivery complete

Public Tasks4 source71a822e6bb920fd0e27de1dd2b2e15d9cbed31d2; executable SHA256
dc82d585b0316dd4513f41bdd5e212fba0ea2f967c026de8cd98633fac1e5958.
All13 release assets/digests and anonymous EXE/manifest/channel verified. Suite
promotion0926501 changes only Tasks pin/default; prior records/withholdings remain.
Public3 and4 tags/assets are immutable. Workspace manager92 is unchanged.

Final clean4 build: frontend/AUI contracts, Go tests/vet and exact-native Chrome
task/list/account flows pass. The shared I-drive dependency reads stalled during
repeat checks; stopped only task-owned invocations. Verified31 local module ZIPs
against committed go.sum and staged them in the named NTFS cache; complete final
release checks then passed. JS/CSS/Roboto bytes are identical to accepted3.

Both normal protected Tasks-only updates completed4. Primary independent audit
confirms exact EXE, identity/credentials/config/paths/app selection and all other
app contracts/package/executable bytes retained. All JS/CSS/fonts served match the
tested package. Cinema exceeded90-second startup readiness after the full restart;
normal RetryInstalledApps restored it without restarting healthy apps. All27
services verified locally and through primary public Chrome health.
Secondary completed in2m16s; all25 other displayed versions retained, all27 public
services running, Tasks4 runtime, plain Tasks title and expected hashed JS/CSS seen.
Its progress page needed a read-only reload after restart, never a repeated apply.

Synthetic phone/small-dark/desktop visuals and flows use only fixture records.
Deployed browser visual review was at desktop size; the browser viewport override
did not change screenshot dimensions, so native Playwright captures are the mobile
evidence. No owner functional fixtures, physical-phone/clean-host/reboot claim.
Newest recovery/full snapshots retained by normal updater. Fixture29718 stopped;
all named NTFS source/dependency/test data removed after process/path/reparse checks.
No current task work remains. Artifacts/receipts remain in I tooling; do not rerun
applies or rewrite public tags. Future updates must preserve concurrent app work.
