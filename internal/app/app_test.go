package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type client struct {
	h      *apphost.Host
	cookie *http.Cookie
	s      apphost.Session
}

func (c *client) call(method, path string, body any) *httptest.ResponseRecorder {
	var b bytes.Buffer
	if body != nil {
		json.NewEncoder(&b).Encode(body)
	}
	r := httptest.NewRequest(method, c.h.Config.Origin+path, &b)
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("Origin", c.h.Config.Origin)
	r.Header.Set("X-CSRF-Token", c.s.CSRF)
	if c.cookie != nil {
		r.AddCookie(c.cookie)
	}
	w := httptest.NewRecorder()
	c.h.Handler().ServeHTTP(w, r)
	return w
}
func fixture(t *testing.T) (*App, *client, *client) {
	t.Helper()
	h, e := apphost.New(apphost.Config{ID: "tasks", Name: "My Tasks", Data: filepath.Join(t.TempDir(), "data"), Origin: "http://localhost:29008", Listen: "127.0.0.1:29008"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { h.Close() })
	a, e := New(h)
	if e != nil {
		t.Fatal(e)
	}
	if e = h.Bootstrap("owner@example.test", "fixture-password"); e != nil {
		t.Fatal(e)
	}
	owner := &client{h: h}
	login(t, owner, "owner@example.test")
	w := owner.call("POST", "/api/local-users", map[string]string{"email": "member@example.test", "password": "fixture-password"})
	expect(t, w, 201)
	member := &client{h: h}
	login(t, member, "member@example.test")
	return a, owner, member
}
func login(t *testing.T, c *client, email string) {
	w := c.call("POST", "/api/login", map[string]string{"email": email, "password": "fixture-password"})
	expect(t, w, 200)
	c.cookie = w.Result().Cookies()[0]
	w = c.call("GET", "/api/me", nil)
	expect(t, w, 200)
	if e := json.Unmarshal(w.Body.Bytes(), &c.s); e != nil {
		t.Fatal(e)
	}
}
func expect(t *testing.T, w *httptest.ResponseRecorder, n int) {
	t.Helper()
	if w.Code != n {
		t.Fatalf("status=%d want=%d body=%s", w.Code, n, w.Body)
	}
}
func decodeTask(t *testing.T, w *httptest.ResponseRecorder) Task {
	t.Helper()
	var task Task
	if e := json.Unmarshal(w.Body.Bytes(), &task); e != nil {
		t.Fatal(e)
	}
	return task
}
func TestIsolationSharingRevocationAndOptimisticEdits(t *testing.T) {
	_, o, m := fixture(t)
	w := o.call("POST", "/api/tasks", Task{Title: "Private <script>alert(1)</script>", Timezone: "Europe/London"})
	expect(t, w, 201)
	task := decodeTask(t, w)
	expect(t, m.call("GET", "/api/tasks/"+task.ID, nil), 404)
	expect(t, m.call("POST", "/api/tasks/bulk", map[string]any{"ids": []string{task.ID}, "action": "done"}), 403)
	expect(t, o.call("PUT", "/api/lists/"+task.ListID+"/members", map[string]string{"subject": m.s.User.Subject, "role": "viewer"}), 200)
	expect(t, m.call("GET", "/api/tasks/"+task.ID, nil), 200)
	task.Title = "Attempt"
	expect(t, m.call("PUT", "/api/tasks/"+task.ID, task), 403)
	expect(t, o.call("PUT", "/api/lists/"+task.ListID+"/members", map[string]string{"subject": m.s.User.Subject, "role": "editor"}), 200)
	expect(t, m.call("PUT", "/api/tasks/"+task.ID, task), 200)
	expect(t, o.call("PUT", "/api/tasks/"+task.ID, task), 409)
	expect(t, o.call("DELETE", "/api/lists/"+task.ListID+"/members/"+m.s.User.ID, nil), 200)
	expect(t, m.call("GET", "/api/tasks/"+task.ID, nil), 404)
	w = m.call("GET", "/api/tasks?q=Private", nil)
	expect(t, w, 200)
	if strings.Contains(w.Body.String(), "Private") {
		t.Fatal("revoked search leak")
	}
}
func TestRecurrenceDSTExceptionsCompletionUndoAndReminderRetry(t *testing.T) {
	a, o, _ := fixture(t)
	w := o.call("POST", "/api/tasks", Task{Title: "DST fixture", Timezone: "Europe/London", Due: "2026-10-24T09:00:00+01:00", Rule: "FREQ=DAILY;COUNT=5", Exceptions: []string{"2026-10-26"}, Reminder: "2026-10-24T08:50:00+01:00"})
	expect(t, w, 201)
	task := decodeTask(t, w)
	task.Status = "done"
	expect(t, o.call("PUT", "/api/tasks/"+task.ID, task), 200)
	var due, id string
	if e := a.H.DB.QueryRow("SELECT id,due FROM tasks WHERE generated_by=?", task.ID).Scan(&id, &due); e != nil {
		t.Fatal(e)
	}
	if due != "2026-10-25T09:00:00Z" {
		t.Fatal("DST wall time changed", due)
	}
	expect(t, o.call("POST", "/api/tasks/bulk", map[string]any{"ids": []string{task.ID}, "action": "undo"}), 200)
	var n int
	a.H.DB.QueryRow("SELECT count(*) FROM tasks WHERE generated_by=?", task.ID).Scan(&n)
	if n != 0 {
		t.Fatal("undo left occurrence")
	}
	expect(t, o.call("POST", "/api/tasks/bulk", map[string]any{"ids": []string{task.ID, task.ID}, "action": "done"}), 200)
	a.H.DB.QueryRow("SELECT count(*) FROM tasks WHERE generated_by=?", task.ID).Scan(&n)
	if n != 1 {
		t.Fatal("not idempotent")
	}
	attempts := 0
	a.Deliver = func(context.Context, Notification) error {
		attempts++
		if attempts == 1 {
			return errors.New("fixture unavailable")
		}
		return nil
	}
	when := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	if e := a.Reminders(context.Background(), when); e != nil {
		t.Fatal(e)
	}
	if e := a.Reminders(context.Background(), when.Add(2*time.Minute)); e != nil {
		t.Fatal(e)
	}
	if e := a.Reminders(context.Background(), when.Add(3*time.Minute)); e != nil {
		t.Fatal(e)
	}
	a.H.DB.QueryRow("SELECT count(*) FROM task_notifications").Scan(&n)
	if n != 1 || attempts != 2 {
		t.Fatal("reminder duplicate/retry", n, attempts)
	}
}
func TestRecurrenceValidationPaginationAndSafeICS(t *testing.T) {
	_, o, _ := fixture(t)
	expect(t, o.call("POST", "/api/tasks", Task{Title: "Bad", Due: "2026-10-01T09:00:00Z", Rule: "FREQ=SECONDLY"}), 400)
	expect(t, o.call("POST", "/api/tasks", Task{Title: "Bad link", CalendarLink: "https://evil.test/calendar/"}), 400)
	w := o.call("POST", "/api/tasks", Task{Title: "Title\r\nBEGIN:VEVENT", Description: "Comma, semicolon; slash\\", Due: time.Now().Add(time.Hour).UTC().Format(time.RFC3339), Timezone: "UTC"})
	expect(t, w, 201)
	w = o.call("GET", "/api/export.ics", nil)
	expect(t, w, 200)
	if strings.Contains(w.Body.String(), "\r\nBEGIN:VEVENT") {
		t.Fatal("ICS injection")
	}
	if !strings.Contains(w.Body.String(), "\\nBEGIN:VEVENT") {
		t.Fatal("ICS not escaped")
	}
	for _, s := range []string{"FREQ=DAILY;BYSECOND=1", "FREQ=DAILY;COUNT=10001"} {
		if _, e := ruleFor(s, "2026-01-01T09:00:00Z", "UTC"); e == nil {
			t.Fatal("unbounded rule", s)
		}
	}
}

func TestLocalTimeDSTGapAndFold(t *testing.T) {
	if _, e := resolveLocal("2026-03-29T01:30", "Europe/London", ""); e == nil {
		t.Fatal("missing spring time accepted")
	}
	first, e := resolveLocal("2026-10-25T01:30", "Europe/London", "")
	if e != nil {
		t.Fatal(e)
	}
	second, e := resolveLocal("2026-10-25T01:30", "Europe/London", "later")
	if e != nil || second.Sub(first) != time.Hour {
		t.Fatal("ambiguous autumn hour", first, second, e)
	}
	if got, e := resolveLocal("2026-10-25T09:00", "Europe/London", ""); e != nil || got.Format(time.RFC3339) != "2026-10-25T09:00:00Z" {
		t.Fatal(got, e)
	}
}
