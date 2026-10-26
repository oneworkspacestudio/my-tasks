# My Tasks 0.1.0-preview.4

Preview4 corrects the browser tab title to Tasks after a live preview3 check
found an incorrectly encoded separator. Task data and app contracts are unchanged.

The mobile task screen follows the supplied Google Tasks reference: peach canvas,
centered Tasks title, horizontal list tabs, warm rounded task cards, circular
completion controls, stars, a collapsible Completed card and a bottom add button.
Roboto is bundled locally. Phone task creation opens a bottom sheet; additional
scheduling, recurrence, priority and integration fields remain in More options.

List options contain search, sharing, board view, saved views, bulk actions and
ICS export. Account settings and Workspace Apps are under the avatar. Fast list switching hides stale rows until the selected list is ready. Existing
lists, tasks, permissions, identity, recurrence and backups retain their existing
storage contracts. Inbox is displayed as My Tasks without renaming stored data.
This remains an original One Workspace application with its own backend; Google
account synchronization and Gmail task attachments are not provided.

Built locally for Windows x64 without GitHub Actions. TypeScript/Vite/AUI contracts,
Go tests/vet and synthetic native Chrome task/account flows were checked. Phone,
small dark phone and desktop screenshots use synthetic records. Preview remains
unsigned; no physical-phone, clean-host, reboot or independent audit claim.
