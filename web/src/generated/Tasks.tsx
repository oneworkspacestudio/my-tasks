// Generated from screens/Tasks.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Field, Grid, Heading, Icon, Input, Link, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"

export function Tasks({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row>
        <Heading level={1}>{data.title}</Heading>
        <Spacer />
        <Button variant="secondary" onClick={() => onAction("manage")}>Lists & sharing</Button>
        <Button onClick={() => onAction("newTask")}>+ Add task</Button>
        <Button variant="secondary" onClick={() => onAction("board")}>{data.boardLabel}</Button>
        <Button variant="secondary" onClick={() => onAction("export")}>Export ICS</Button>
      </Row>
      <Row>
        <Input name="quickTitle" placeholder="Add a task and press Enter…" onClick={() => onAction("quickAdd")} />
        <Button onClick={() => onAction("quickAdd")}>Add</Button>
      </Row>
      <Row>
        <Input name="query" onClick={() => onAction("search")} type="search" placeholder="Search title, description or tags" />
        <Button variant="secondary" onClick={() => onAction("search")}>Search</Button>
        <Button variant="secondary" onClick={() => onAction("saveView")}>Save view</Button>
        <Select name="bulkAction" options={data.bulkActions} />
        <Select name="bulkList" options={data.listOptions} />
        <Button variant="secondary" onClick={() => onAction("bulk")}>Apply to selected</Button>
      </Row>
      {data.empty && (
        <Card variant="empty">
          <Icon name="ShieldCheck" />
          <Heading level={2}>A little clarity for your day</Heading>
          <Text>Add your first task, plan a date or share a list with your team.</Text>
          <Button onClick={() => onAction("newTask")}>Create a task</Button>
        </Card>
      )}
      <Grid columns={data.columnCount}>
        {data.columns.map((item, i) => (
          <Card key={i}>
            <Heading level={2}>{item.name}</Heading>
            {item.tasks.map((item, i) => (
              <Row key={i} variant="task" item={item.id} draggable={true}>
                <Button variant="secondary" item={item.id} onClick={() => onAction("select")} aria-label="Select $item.title" aria-pressed={item.isSelected}>{item.selection}</Button>
                <Button variant="secondary" item={item.id} onClick={() => onAction("toggle")} aria-label={item.toggleLabel} aria-pressed={item.isDone}>{item.check}</Button>
                <Stack gap="xs">
                  <Link href={item.href}>{item.title}</Link>
                  <Text variant="caption">{item.detail}</Text>
                </Stack>
                <Button variant="secondary" item={item.id} onClick={() => onAction("star")} aria-label="Star $item.title" aria-pressed={item.starred}>{item.star}</Button>
                <Button variant="secondary" item={item.id} onClick={() => onAction("edit")}>Edit</Button>
              </Row>
            ))}
          </Card>
        ))}
      </Grid>
      {data.hasMore && (
        <Button variant="secondary" onClick={() => onAction("more")}>Load more tasks</Button>
      )}
      {data.editing && (
        <Card variant="dialog" aria-label="Task editor">
          <Heading level={2}>{data.editorTitle}</Heading>
          {data.editorError && (
            <Alert tone="error">{data.editorError}</Alert>
          )}
          {data.editorBusy && (
            <Text>{data.editorBusy}</Text>
          )}
          <Field label="Title"><Input name="title" autoFocus={true} placeholder="What needs doing?" /></Field>
          <Field label="List"><Select name="listId" options={data.listOptions} /></Field>
          <Field label="Description (Markdown)"><Input name="description" multiline={true} /></Field>
          <Section id="description-preview" />
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
        <Card>
          <Heading level={2}>Lists & sharing</Heading>
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
