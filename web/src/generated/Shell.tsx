// Generated from screens/Shell.aui; edit the AUI source.
import { Alert, Button, Card, Heading, Link, Page, Row, Section, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function Shell({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="workspace">
        {data.settingsPage && (
          <Row variant="settings-header">
            <Link href="#view=inbox">← Tasks</Link>
            <Spacer />
            <Button variant="text" onClick={() => onAction("apps")}>Apps</Button>
          </Row>
        )}
        <Stack variant="content">
          {data.error && (
            <Alert tone="error">{data.error}</Alert>
          )}
          {data.notice && (
            <Alert tone="success">{data.notice}</Alert>
          )}
          {data.busy && (
            <Row variant="saving">
              <Spinner />
              <Text>{data.busy}</Text>
            </Row>
          )}
          <Section id="screen-outlet" />
        </Stack>
        {data.appsOpen && (
          <Card variant="dialog" aria-label="Workspace apps">
            <Heading level={2}>Your Workspace apps</Heading>
            {data.appLinks.map((item, i) => (
              <Link key={i} href={item.href}>{item.title}</Link>
            ))}
            <Button variant="secondary" onClick={() => onAction("closeApps")}>Close</Button>
          </Card>
        )}
      </Stack>
    </Page>
  )
}
