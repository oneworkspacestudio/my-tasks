// Generated from screens/Tasks.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Field, Grid, Heading, Icon, Input, Link, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"

export function Tasks({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="tasks-header">
        <Heading level={1}>Tasks</Heading>
        <Button variant="avatar" onClick={() => onAction("account")} aria-label="Account menu">{data.accountInitial}</Button>
      </Row>
      {data.account && (
        <Card variant="account-menu">
          <Text>{data.accountName}</Text>
          {data.workspace && (
            <Button variant="text" onClick={() => onAction("apps")}>Apps</Button>
          )}
          <Link href="#settings">Settings & backup</Link>
          <Button variant="text" onClick={() => onAction("theme")}>Toggle color theme</Button>
          <Button variant="text" onClick={() => onAction("logout")}>Sign out</Button>
        </Card>
      )}
      <Row variant="list-tabs">
        {data.tabs.map((item, i) => (
          <Link key={i} href={item.href} variant={item.variant}>
            {item.star ? (
              <>
                <Icon name="Star" size={23} />
                <Text variant="sr-only">Starred</Text>
              </>
            ) : (
              <Text>{item.title}</Text>
            )}
          </Link>
        ))}
      </Row>
      <Stack variant="task-content">
        <Grid columns={data.columnCount}>
          {data.columns.map((item, i) => (
            <Card key={i} variant="task-list">
              <Row variant="list-heading">
                <Heading level={2}>{item.name}</Heading>
                <Spacer />
                <Button variant="icon" onClick={() => onAction("sort")} aria-label={data.sortAria}><Icon name="ArrowDownUp" size={23} /></Button>
                <Button variant="icon" onClick={() => onAction("menu")} aria-label="List options"><Icon name="MoreVertical" size={23} /></Button>
              </Row>
              {data.empty && (
                <Stack variant="empty-tasks">
                  <Icon name="CheckCircle2" size={56} />
                  <Heading level={3}>{data.emptyTitle}</Heading>
                  <Text>{data.emptyText}</Text>
                </Stack>
              )}
              {item.tasks.map((item, i) => (
                <Row key={i} variant={item.rowVariant} item={item.id} draggable={true}>
                  <Button variant="complete" item={item.id} onClick={() => onAction("toggle")} aria-label={item.toggleLabel} aria-pressed={item.isDone}>
                    {item.isDone && (
                      <Icon name="Check" size={20} />
                    )}
                  </Button>
                  <Stack variant="task-copy" gap="xs">
                    <Link variant="task-title" href={item.href}>{item.title}</Link>
                    {item.hasDescription && (
                      <Text variant="task-description">{item.description}</Text>
                    )}
                    {item.hasDetail && (
                      <Text variant="task-detail">{item.detail}</Text>
                    )}
                  </Stack>
                  <Button variant={item.starVariant} item={item.id} onClick={() => onAction("star")} aria-label={item.starLabel} aria-pressed={item.starred}><Icon name="Star" size={23} /></Button>
                </Row>
              ))}
            </Card>
          ))}
        </Grid>
        {data.hasCompleted && (
          <Card variant="completed">
            <Button variant="completed-heading" onClick={() => onAction("completed")} aria-label={data.completedLabel} aria-expanded={data.completedOpen}>
              <Text>{data.completedLabel}</Text>
              <Spacer />
              <Icon name={data.completedIcon} size={24} />
            </Button>
            {data.completedOpen && (
              <Stack variant="completed-tasks">
                {data.completed.map((item, i) => (
                  <Row key={i} variant={item.rowVariant} item={item.id} draggable={true}>
                    <Button variant="complete" item={item.id} onClick={() => onAction("toggle")} aria-label={item.toggleLabel} aria-pressed={item.isDone}>
                      {item.isDone && (
                        <Icon name="Check" size={20} />
                      )}
                    </Button>
                    <Stack variant="task-copy" gap="xs">
                      <Link variant="task-title" href={item.href}>{item.title}</Link>
                      {item.hasDescription && (
                        <Text variant="task-description">{item.description}</Text>
                      )}
                      {item.hasDetail && (
                        <Text variant="task-detail">{item.detail}</Text>
                      )}
                    </Stack>
                    <Button variant={item.starVariant} item={item.id} onClick={() => onAction("star")} aria-label={item.starLabel} aria-pressed={item.starred}><Icon name="Star" size={23} /></Button>
                  </Row>
                ))}
              </Stack>
            )}
          </Card>
        )}
        {data.hasMore && (
          <Button variant="text" onClick={() => onAction("more")}>Load more tasks</Button>
        )}
        {data.menu && (
          <Card variant="dialog" aria-label="List options">
            <Row>
              <Heading level={2}>List options</Heading>
              <Spacer />
              <Button variant="icon" onClick={() => onAction("menu")} aria-label="Close list options"><Icon name="X" /></Button>
            </Row>
            <Input name="quickTitle" placeholder="Add a task and press Enter…" onClick={() => onAction("quickAdd")} />
            <Row>
              <Input name="query" onClick={() => onAction("search")} type="search" placeholder="Search title, description or tags" />
              <Button variant="text" onClick={() => onAction("search")}>Search</Button>
            </Row>
            <Row>
              <Button variant="text" onClick={() => onAction("manage")}>Lists & sharing</Button>
              <Button variant="text" onClick={() => onAction("board")}>{data.boardLabel}</Button>
              <Button variant="text" onClick={() => onAction("export")}>Export ICS</Button>
              <Button variant="text" onClick={() => onAction("saveView")}>Save view</Button>
            </Row>
            <Row>
              <Select name="bulkAction" options={data.bulkActions} />
              <Select name="bulkList" options={data.listOptions} />
              <Button variant="text" onClick={() => onAction("bulk")}>Apply to selected</Button>
            </Row>
            <Row>
              <Link href="#view=today">Today</Link>
              <Link href="#view=upcoming">Upcoming</Link>
              <Link href="#view=planned">Planned</Link>
              <Link href="#view=overdue">Overdue</Link>
              <Link href="#view=shared">Shared</Link>
            </Row>
            {data.columns.map((item, i) => (
              <Stack key={i}>
                {item.tasks.map((item, i) => (
                  <Button key={i} variant="text" item={item.id} onClick={() => onAction("select")} aria-pressed={item.isSelected}>Select {item.title} {item.selection}</Button>
                ))}
              </Stack>
            ))}
          </Card>
        )}
        <Button variant="fab" onClick={() => onAction("newTask")} aria-label="Create a task"><Icon name="Plus" size={28} /></Button>
      </Stack>
      {data.editing && (
        <Card variant="dialog" aria-label="Task editor">
          <Heading level={2}>{data.editorTitle}</Heading>
          {data.editorError && (
            <Alert tone="error">{data.editorError}</Alert>
          )}
          {data.editorBusy && (
            <Text>{data.editorBusy}</Text>
          )}
          <Field label="Title"><Input name="title" autoFocus={true} placeholder="New task" /></Field>
          <Field label="List"><Select name="listId" options={data.listOptions} /></Field>
          <Field label="Details"><Input name="description" multiline={true} /></Field>
          <Section id="description-preview" />
          <Button variant="text" onClick={() => onAction("advanced")}>More options</Button>
          {data.advanced && (
            <>
              <Field label="Notes"><Input name="notes" multiline={true} /></Field>
              <Grid columns={3}>
                <Field label="Due date & time"><Input name="dueLocal" type="datetime-local" /></Field>
                <Field label="IANA timezone"><Input name="timezone" placeholder="Europe/London" /></Field>
                <Field label="Priority"><Select name="priority" options={data.priorityOptions} /></Field>
              </Grid>
              <Grid columns={3}>
                <Field label="Status"><Select name="status" options={data.statusOptions} /></Field>
                <Field label="Tags, separated by commas"><Input name="tags" /></Field>
                <Field label="Position"><Input name="position" type="number" /></Field>
              </Grid>
              <Grid columns={2}>
                <Field label="Reminder"><Input name="reminderLocal" type="datetime-local" /></Field>
                <Field label="Snooze until"><Input name="snoozeLocal" type="datetime-local" /></Field>
              </Grid>
              <Field label="Recurrence (new task only; e.g. FREQ=WEEKLY;BYDAY=MO,WE)"><Input name="rule" /></Field>
              <Field label="Recurrence exceptions (YYYY-MM-DD, comma separated)"><Input name="exceptions" /></Field>
              <Field label="Parent task ID (optional, new subtasks only)"><Input name="parentId" /></Field>
              <Field label="Calendar event link (explicit link only)"><Input name="calendarLink" placeholder="/calendar/…" /></Field>
              <Field label="Notes document link (explicit link only)"><Input name="notesLink" placeholder="/notes/…" /></Field>
            </>
          )}
          <Row>
            <Button onClick={() => onAction("saveTask")}>Save task</Button>
            <Button variant="secondary" onClick={() => onAction("closeEditor")}>Close</Button>
            {data.existingTask && (
              <Button variant="danger" onClick={() => onAction("deleteTask")}>Delete task</Button>
            )}
          </Row>
        </Card>
      )}
      {data.managing && (
        <Card variant="dialog" aria-label="Lists & sharing">
          <Row>
            <Heading level={2}>Lists & sharing</Heading>
            <Spacer />
            <Button variant="text" onClick={() => onAction("manage")}>Close</Button>
          </Row>
          <Row>
            <Input name="listName" placeholder="New list name" />
            <Button onClick={() => onAction("createList")}>Create list</Button>
          </Row>
          <Text>Sharing uses the member's immutable MyAccounts ID, shown in their app Settings. Members can be viewers or editors.</Text>
          <Row>
            <Select name="shareList" options={data.allListOptions} />
            <Input name="memberSubject" placeholder="MyAccounts ID" />
            <Select name="memberRole" options={data.memberRoles} />
            <Button variant="secondary" onClick={() => onAction("share")}>Share list</Button>
            <Button variant="secondary" onClick={() => onAction("members")}>Show members</Button>
          </Row>
          <Row>
            <Input name="renameList" placeholder="New name for selected list" />
            <Button variant="secondary" onClick={() => onAction("renameList")}>Rename list</Button>
            <Button variant="danger" onClick={() => onAction("deleteList")}>Delete selected list</Button>
          </Row>
          {data.members.map((item, i) => (
            <Row key={i}>
              <Text>{item.id}</Text>
              <Badge>{item.role}</Badge>
              <Button variant="secondary" item={item.id} onClick={() => onAction("unshare")}>Remove access</Button>
            </Row>
          ))}
        </Card>
      )}
      {data.hasNotifications && (
        <Card>
          <Heading level={2}>Reminders</Heading>
          <Text variant="caption">In-app delivery is always available. External notification services are not configured.</Text>
          {data.notifications.map((item, i) => (
            <Row key={i}>
              <Link href={item.href}>{item.title}</Link>
              <Text variant="caption">{item.at}</Text>
              <Button variant="secondary" item={item.id} onClick={() => onAction("read")}>Mark read</Button>
            </Row>
          ))}
        </Card>
      )}
      {data.hasSavedViews && (
        <Card>
          <Heading level={2}>Saved views</Heading>
          {data.savedViews.map((item, i) => (
            <Row key={i}>
              <Link href={item.href}>{item.name}</Link>
              <Button variant="secondary" item={item.id} onClick={() => onAction("deleteView")}>Remove</Button>
            </Row>
          ))}
        </Card>
      )}
    </Page>
  )
}
