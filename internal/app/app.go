package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type App struct {
	H       *apphost.Host
	Deliver func(context.Context, Notification) error
}
type List struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	Owner    string  `json:"owner"`
	Position float64 `json:"position"`
}
type Task struct {
	ID           string   `json:"id"`
	ListID       string   `json:"listId"`
	ParentID     string   `json:"parentId"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Notes        string   `json:"notes"`
	Due          string   `json:"due"`
	Timezone     string   `json:"timezone"`
	Priority     int      `json:"priority"`
	Status       string   `json:"status"`
	Starred      bool     `json:"starred"`
	Tags         []string `json:"tags"`
	Position     float64  `json:"position"`
	Version      int      `json:"version"`
	Updated      string   `json:"updated"`
	RuleID       string   `json:"ruleId"`
	Rule         string   `json:"rule,omitempty"`
	Exceptions   []string `json:"exceptions,omitempty"`
	Reminder     string   `json:"reminder"`
	Snooze       string   `json:"snooze"`
	CalendarLink string   `json:"calendarLink"`
	NotesLink    string   `json:"notesLink"`
	GeneratedBy  string   `json:"generatedBy"`
	Previous     string   `json:"previousStatus"`
}
type Notification struct {
	ID     string `json:"id"`
	TaskID string `json:"taskId"`
	UserID string `json:"userId"`
	At     string `json:"at"`
	Title  string `json:"title"`
	Read   bool   `json:"read"`
}

const schema = `
 CREATE TABLE task_lists(id TEXT PRIMARY KEY,owner TEXT NOT NULL,name TEXT NOT NULL,position REAL NOT NULL DEFAULT 0,created TEXT NOT NULL);
 CREATE TABLE task_members(list_id TEXT NOT NULL REFERENCES task_lists(id) ON DELETE CASCADE,user_id TEXT NOT NULL,role TEXT NOT NULL CHECK(role IN('editor','viewer')),PRIMARY KEY(list_id,user_id));
 CREATE TABLE task_rules(id TEXT PRIMARY KEY,list_id TEXT NOT NULL REFERENCES task_lists(id) ON DELETE CASCADE,rule TEXT NOT NULL,start TEXT NOT NULL,timezone TEXT NOT NULL,exceptions TEXT NOT NULL DEFAULT '[]',template TEXT NOT NULL);
 CREATE TABLE tasks(id TEXT PRIMARY KEY,list_id TEXT NOT NULL REFERENCES task_lists(id) ON DELETE CASCADE,parent_id TEXT NOT NULL DEFAULT '',title TEXT NOT NULL,description TEXT NOT NULL DEFAULT '',notes TEXT NOT NULL DEFAULT '',due TEXT NOT NULL DEFAULT '',timezone TEXT NOT NULL,status TEXT NOT NULL CHECK(status IN('todo','doing','done')),priority INTEGER NOT NULL DEFAULT 0,starred INTEGER NOT NULL DEFAULT 0,tags TEXT NOT NULL DEFAULT '[]',position REAL NOT NULL DEFAULT 0,version INTEGER NOT NULL DEFAULT 1,updated TEXT NOT NULL,rule_id TEXT NOT NULL DEFAULT '',reminder TEXT NOT NULL DEFAULT '',snooze TEXT NOT NULL DEFAULT '',calendar_link TEXT NOT NULL DEFAULT '',notes_link TEXT NOT NULL DEFAULT '',generated_by TEXT NOT NULL DEFAULT '',previous_status TEXT NOT NULL DEFAULT 'todo');
 CREATE UNIQUE INDEX task_occurrence ON tasks(rule_id,due) WHERE rule_id<>'';
 CREATE INDEX task_list_status ON tasks(list_id,status,position);
 CREATE TABLE task_notifications(id TEXT PRIMARY KEY,task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,user_id TEXT NOT NULL,at TEXT NOT NULL,read INTEGER NOT NULL DEFAULT 0);
 CREATE TABLE task_delivery(id TEXT PRIMARY KEY,task_id TEXT NOT NULL,user_id TEXT NOT NULL,at TEXT NOT NULL,attempts INTEGER NOT NULL DEFAULT 0,next_attempt TEXT NOT NULL DEFAULT '',state TEXT NOT NULL DEFAULT 'pending');
 CREATE TABLE task_views(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,name TEXT NOT NULL,filter TEXT NOT NULL);
`
const cols = `t.id,t.list_id,t.parent_id,t.title,t.description,t.notes,t.due,t.timezone,t.priority,t.status,t.starred,t.tags,t.position,t.version,t.updated,t.rule_id,t.reminder,t.snooze,t.calendar_link,t.notes_link,t.generated_by,t.previous_status`
const access = `(l.owner=? OR EXISTS(SELECT 1 FROM task_members m WHERE m.list_id=l.id AND m.user_id=?))`

func New(h *apphost.Host) (*App, error) {
	if e := h.Migrate(1, schema); e != nil {
		return nil, e
	}
	a := &App{H: h}
	m := h.Mux
	m.HandleFunc("GET /api/time", h.Require(a.resolveTime))
	m.HandleFunc("GET /api/lists", h.Require(a.lists))
	m.HandleFunc("POST /api/lists", h.Require(a.createList))
	m.HandleFunc("PUT /api/lists/{id}", h.Require(a.updateList))
	m.HandleFunc("DELETE /api/lists/{id}", h.Require(a.deleteList))
	m.HandleFunc("GET /api/lists/{id}/members", h.Require(a.members))
	m.HandleFunc("PUT /api/lists/{id}/members", h.Require(a.share))
	m.HandleFunc("DELETE /api/lists/{id}/members/{user}", h.Require(a.unshare))
	m.HandleFunc("GET /api/tasks", h.Require(a.tasks))
	m.HandleFunc("POST /api/tasks", h.Require(a.createTask))
	m.HandleFunc("GET /api/tasks/{id}", h.Require(a.getTask))
	m.HandleFunc("PUT /api/tasks/{id}", h.Require(a.updateTask))
	m.HandleFunc("DELETE /api/tasks/{id}", h.Require(a.deleteTask))
	m.HandleFunc("POST /api/tasks/bulk", h.Require(a.bulk))
	m.HandleFunc("GET /api/notifications", h.Require(a.notifications))
	m.HandleFunc("PUT /api/notifications/{id}", h.Require(a.readNotification))
	m.HandleFunc("GET /api/views", h.Require(a.views))
	m.HandleFunc("POST /api/views", h.Require(a.saveView))
	m.HandleFunc("DELETE /api/views/{id}", h.Require(a.deleteView))
	m.HandleFunc("GET /api/export.ics", h.Require(a.exportICS))
	return a, nil
}
func (a *App) Run(ctx context.Context) error {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			a.H.Freeze.RLock()
			e := a.Reminders(ctx, now)
			a.H.Freeze.RUnlock()
			if e != nil {
				a.H.Log.Warn("reminder pass deferred")
			}
		}
	}
}
func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }
func (a *App) inbox(user string) (string, error) {
	id := apphost.Identity("inbox", user)
	_, e := a.H.DB.Exec("INSERT OR IGNORE INTO task_lists(id,owner,name,created)VALUES(?,?,?,?)", id, user, "Inbox", now())
	return id, e
}
func (a *App) role(user, list string) string {
	var owner string
	var role sql.NullString
	e := a.H.DB.QueryRow("SELECT owner,(SELECT role FROM task_members WHERE list_id=l.id AND user_id=?) FROM task_lists l WHERE id=?", user, list).Scan(&owner, &role)
	if e != nil {
		return ""
	}
	if owner == user {
		return "owner"
	}
	return role.String
}
func canWrite(role string) bool { return role == "owner" || role == "editor" }
func (a *App) listCheck(w http.ResponseWriter, r *http.Request, write, owner bool) bool {
	role := a.role(apphost.Current(r).User.ID, r.PathValue("id"))
	if role == "" {
		apphost.Problem(w, 404, "List not found")
		return false
	}
	if write && !canWrite(role) || owner && role != "owner" {
		apphost.Problem(w, 403, "Your list role does not allow this change")
		return false
	}
	return true
}
func (a *App) lists(w http.ResponseWriter, r *http.Request) {
	u := apphost.Current(r).User.ID
	if _, e := a.inbox(u); e != nil {
		a.H.DBError(w, e)
		return
	}
	rows, e := a.H.DB.Query("SELECT l.id,l.name,l.owner,l.position FROM task_lists l WHERE "+access+" ORDER BY position,name", u, u)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	out := []List{}
	for rows.Next() {
		var l List
		if e = rows.Scan(&l.ID, &l.Name, &l.Owner, &l.Position); e != nil {
			rows.Close()
			a.H.DBError(w, e)
			return
		}
		out = append(out, l)
	}
	rows.Close()
	for i := range out {
		out[i].Role = a.role(u, out[i].ID)
	}
	apphost.JSON(w, 200, out)
}
func (a *App) createList(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 100 {
		apphost.Problem(w, 400, "Name the list using up to 100 characters")
		return
	}
	l := List{ID: apphost.ID(), Name: in.Name, Owner: apphost.Current(r).User.ID, Role: "owner"}
	_, e := a.H.DB.Exec("INSERT INTO task_lists(id,owner,name,created)VALUES(?,?,?,?)", l.ID, l.Owner, l.Name, now())
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 201, l)
}
func (a *App) updateList(w http.ResponseWriter, r *http.Request) {
	if !a.listCheck(w, r, true, false) {
		return
	}
	var in struct {
		Name     string  `json:"name"`
		Position float64 `json:"position"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Name) == "" || len(in.Name) > 100 {
		apphost.Problem(w, 400, "Enter a list name")
		return
	}
	_, e := a.H.DB.Exec("UPDATE task_lists SET name=?,position=? WHERE id=?", in.Name, in.Position, r.PathValue("id"))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) deleteList(w http.ResponseWriter, r *http.Request) {
	if !a.listCheck(w, r, true, true) {
		return
	}
	u := apphost.Current(r).User.ID
	if r.PathValue("id") == apphost.Identity("inbox", u) {
		apphost.Problem(w, 400, "Keep Inbox; move or delete its tasks individually")
		return
	}
	_, e := a.H.DB.Exec("DELETE FROM task_lists WHERE id=?", r.PathValue("id"))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) members(w http.ResponseWriter, r *http.Request) {
	if !a.listCheck(w, r, false, false) {
		return
	}
	rows, e := a.H.DB.Query("SELECT user_id,role FROM task_members WHERE list_id=?", r.PathValue("id"))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []map[string]string{}
	for rows.Next() {
		var id, role string
		rows.Scan(&id, &role)
		out = append(out, map[string]string{"id": id, "role": role})
	}
	apphost.JSON(w, 200, out)
}
func (a *App) share(w http.ResponseWriter, r *http.Request) {
	if !a.listCheck(w, r, true, true) {
		return
	}
	var in struct {
		Subject string `json:"subject"`
		Role    string `json:"role"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	if in.Subject == "" || len(in.Subject) > 254 || strings.ContainsAny(in.Subject, "\r\n\x00") || !(in.Role == "editor" || in.Role == "viewer") {
		apphost.Problem(w, 400, "Enter the member's immutable MyAccounts ID and a viewer/editor role")
		return
	}
	u := apphost.Current(r).User
	id := apphost.Identity(u.Issuer, in.Subject)
	if id == u.ID {
		apphost.Problem(w, 400, "You already own this list")
		return
	}
	_, e := a.H.DB.Exec("INSERT INTO task_members VALUES(?,?,?) ON CONFLICT(list_id,user_id)DO UPDATE SET role=excluded.role", r.PathValue("id"), id, in.Role)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	a.H.Audit(u.ID, "share-list", r.PathValue("id"))
	apphost.JSON(w, 200, map[string]string{"id": id, "role": in.Role})
}
func (a *App) unshare(w http.ResponseWriter, r *http.Request) {
	if !a.listCheck(w, r, true, true) {
		return
	}
	_, e := a.H.DB.Exec("DELETE FROM task_members WHERE list_id=? AND user_id=?", r.PathValue("id"), r.PathValue("user"))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func scanTask(row interface{ Scan(...any) error }) (Task, error) {
	var t Task
	var tags string
	e := row.Scan(&t.ID, &t.ListID, &t.ParentID, &t.Title, &t.Description, &t.Notes, &t.Due, &t.Timezone, &t.Priority, &t.Status, &t.Starred, &tags, &t.Position, &t.Version, &t.Updated, &t.RuleID, &t.Reminder, &t.Snooze, &t.CalendarLink, &t.NotesLink, &t.GeneratedBy, &t.Previous)
	json.Unmarshal([]byte(tags), &t.Tags)
	return t, e
}
func (a *App) task(user, id string) (Task, string, error) {
	t, e := scanTask(a.H.DB.QueryRow("SELECT "+cols+" FROM tasks t JOIN task_lists l ON l.id=t.list_id WHERE t.id=? AND "+access, id, user, user))
	if e != nil {
		return t, "", e
	}
	return t, a.role(user, t.ListID), nil
}
func (a *App) tasks(w http.ResponseWriter, r *http.Request) {
	u := apphost.Current(r).User.ID
	q := r.URL.Query()
	if len(q.Get("q")) > 200 {
		apphost.Problem(w, 400, "Search is too long")
		return
	}
	where := access
	args := []any{u, u}
	if list := q.Get("list"); list != "" {
		where += " AND t.list_id=?"
		args = append(args, list)
	}
	view := q.Get("view")
	if view == "completed" {
		where += " AND t.status='done'"
	} else if view != "all"&&q.Get("includeCompleted")!="true" {
		where += " AND t.status<>'done'"
	}
	switch view {
	case "starred":
		where += " AND t.starred=1"
	case "shared":
		where += " AND EXISTS(SELECT 1 FROM task_members m WHERE m.list_id=l.id)"
	case "planned", "today", "upcoming", "overdue":
		where += " AND t.due<>''"
	}
	if search := q.Get("q"); search != "" {
		where += " AND instr(lower(t.title||' '||t.description||' '||t.tags),lower(?))>0"
		args = append(args, search)
	}
	if tag := q.Get("tag"); tag != "" {
		where += " AND EXISTS(SELECT 1 FROM json_each(t.tags)WHERE value=?)"
		args = append(args, tag)
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	if offset < 0 || offset > 100000 {
		offset = 0
	}
	rows, e := a.H.DB.Query("SELECT "+cols+" FROM tasks t JOIN task_lists l ON l.id=t.list_id WHERE "+where+" ORDER BY t.priority DESC,t.position,t.updated DESC", args...)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	out := []Task{}
	for rows.Next() {
		t, e := scanTask(rows)
		if e != nil {
			rows.Close()
			a.H.DBError(w, e)
			return
		}
		if view == "today" || view == "upcoming" || view == "overdue" {
			due, e := time.Parse(time.RFC3339, t.Due)
			if e != nil {
				continue
			}
			loc, e := time.LoadLocation(t.Timezone)
			if e != nil {
				continue
			}
			today := time.Now().In(loc).Format("2006-01-02")
			date := due.In(loc).Format("2006-01-02")
			if view == "today" && date != today || view == "upcoming" && date <= today || view == "overdue" && due.After(time.Now()) {
				continue
			}
		}
		if offset > 0 {
			offset--
			continue
		}
		out = append(out, t)
		if len(out) > 100 {
			break
		}
	}
	rows.Close()
	more := len(out) > 100
	if more {
		out = out[:100]
	}
	apphost.JSON(w, 200, map[string]any{"tasks": out, "hasMore": more})
}
func (a *App) getTask(w http.ResponseWriter, r *http.Request) {
	t, role, e := a.task(apphost.Current(r).User.ID, r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, "Task not found")
		return
	}
	if t.RuleID != "" {
		var ex string
		a.H.DB.QueryRow("SELECT rule,exceptions FROM task_rules WHERE id=?", t.RuleID).Scan(&t.Rule, &ex)
		json.Unmarshal([]byte(ex), &t.Exceptions)
	}
	apphost.JSON(w, 200, map[string]any{"task": t, "role": role})
}
func validTask(t *Task) error {
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" || len(t.Title) > 300 || len(t.Description) > 100000 || len(t.Notes) > 100000 {
		return fmt.Errorf("use a title up to 300 characters and shorter notes")
	}
	if t.Timezone == "" {
		t.Timezone = "UTC"
	}
	if _, e := time.LoadLocation(t.Timezone); e != nil {
		return fmt.Errorf("choose a valid IANA timezone")
	}
	if t.Status == "" {
		t.Status = "todo"
	}
	if t.Status != "todo" && t.Status != "doing" && t.Status != "done" || t.Priority < 0 || t.Priority > 3 || len(t.Tags) > 30 {
		return fmt.Errorf("check task status, priority and tags")
	}
	for _, tag := range t.Tags {
		if len(tag) > 40 {
			return fmt.Errorf("tags must be at most 40 characters")
		}
	}
	for _, date := range []*string{&t.Due, &t.Reminder, &t.Snooze} {
		if *date != "" {
			parsed, e := time.Parse(time.RFC3339, *date)
			if e != nil {
				return fmt.Errorf("dates must include a timezone offset")
			}
			*date = parsed.UTC().Format(time.RFC3339Nano)
		}
	}
	for prefix, link := range map[string]string{"/calendar": t.CalendarLink, "/notes": t.NotesLink} {
		if link == "" {
			continue
		}
		u, e := url.Parse(link)
		if e != nil || u.IsAbs() || u.Host != "" || !strings.HasPrefix(u.Path, prefix+"/") || strings.ContainsAny(link, "\\\r\n") {
			return fmt.Errorf("links must open the selected Workspace app")
		}
	}
	return nil
}
func taskArgs(t Task) []any {
	tags, _ := json.Marshal(t.Tags)
	return []any{t.ID, t.ListID, t.ParentID, t.Title, t.Description, t.Notes, t.Due, t.Timezone, t.Priority, t.Status, t.Starred, string(tags), t.Position, t.Version, t.Updated, t.RuleID, t.Reminder, t.Snooze, t.CalendarLink, t.NotesLink, t.GeneratedBy, t.Previous}
}
func insert(tx *sql.Tx, t Task) error {
	_, e := tx.Exec("INSERT INTO tasks("+strings.ReplaceAll(cols, "t.", "")+")VALUES("+strings.TrimRight(strings.Repeat("?,", 22), ",")+")", taskArgs(t)...)
	return e
}
func (a *App) createTask(w http.ResponseWriter, r *http.Request) {
	var t Task
	if !apphost.Decode(w, r, &t) {
		return
	}
	u := apphost.Current(r).User.ID
	if t.ListID == "" {
		id, e := a.inbox(u)
		if e != nil {
			a.H.DBError(w, e)
			return
		}
		t.ListID = id
	}
	if !canWrite(a.role(u, t.ListID)) {
		apphost.Problem(w, 404, "Editable list not found")
		return
	}
	if e := validTask(&t); e != nil {
		apphost.Problem(w, 400, e.Error())
		return
	}
	if t.ParentID != "" {
		parent, _, e := a.task(u, t.ParentID)
		if e != nil || parent.ListID != t.ListID || parent.ParentID != "" {
			apphost.Problem(w, 400, "Choose a top-level parent in this list")
			return
		}
	}
	t.ID = apphost.ID()
	t.Version = 1
	t.Updated = now()
	t.Previous = "todo"
	t.RuleID = ""
	t.GeneratedBy = ""
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer tx.Rollback()
	if !writableTx(tx, u, t.ListID) {
		apphost.Problem(w, 403, "This list is no longer editable")
		return
	}
	if t.Rule != "" {
		if e = createRule(tx, &t); e != nil {
			apphost.Problem(w, 400, e.Error())
			return
		}
	}
	if e = insert(tx, t); e != nil {
		a.H.DBError(w, e)
		return
	}
	if e = tx.Commit(); e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 201, t)
}
func (a *App) updateTask(w http.ResponseWriter, r *http.Request) {
	u := apphost.Current(r).User.ID
	old, role, e := a.task(u, r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, "Task not found")
		return
	}
	if !canWrite(role) {
		apphost.Problem(w, 403, "This list is view-only")
		return
	}
	var t Task
	if !apphost.Decode(w, r, &t) {
		return
	}
	if e = validTask(&t); e != nil {
		apphost.Problem(w, 400, e.Error())
		return
	}
	if t.Version != old.Version {
		apphost.Problem(w, 409, "This task changed elsewhere. Reload before saving.")
		return
	}
	if t.ListID != old.ListID || t.ParentID != old.ParentID {
		apphost.Problem(w, 400, "Move tasks explicitly instead of changing their parent identifiers")
		return
	}
	t.ID = old.ID
	t.RuleID = old.RuleID
	t.GeneratedBy = old.GeneratedBy
	t.Version++
	t.Updated = now()
	t.Previous = old.Previous
	if old.Status != "done" && t.Status == "done" {
		t.Previous = old.Status
	}
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer tx.Rollback()
	if !writableTx(tx, u, t.ListID) {
		apphost.Problem(w, 403, "This list is no longer editable")
		return
	}
	if e = reviseRecurrence(tx, t); e != nil {
		apphost.Problem(w, 400, e.Error())
		return
	}
	args := taskArgs(t)
	sets := []string{}
	for _, c := range strings.Split(strings.ReplaceAll(cols, "t.", ""), ",") {
		sets = append(sets, c+"=?")
	}
	args = append(args, t.ID, old.Version)
	result, e := tx.Exec("UPDATE tasks SET "+strings.Join(sets, ",")+" WHERE id=? AND version=?", args...)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		apphost.Problem(w, 409, "Task changed; reload it")
		return
	}
	if old.Status != "done" && t.Status == "done" && t.RuleID != "" {
		if e = nextOccurrence(tx, t); e != nil {
			apphost.Problem(w, 400, e.Error())
			return
		}
	}
	if old.Status == "done" && t.Status != "done" {
		if _, e = tx.Exec("DELETE FROM tasks WHERE generated_by=? AND version=1 AND status='todo'", t.ID); e != nil {
			a.H.DBError(w, e)
			return
		}
	}
	if e = tx.Commit(); e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, t)
}
func (a *App) deleteTask(w http.ResponseWriter, r *http.Request) {
	t, role, e := a.task(apphost.Current(r).User.ID, r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, "Task not found")
		return
	}
	if !canWrite(role) {
		apphost.Problem(w, 403, "This list is view-only")
		return
	}
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer tx.Rollback()
	if !writableTx(tx, apphost.Current(r).User.ID, t.ListID) {
		apphost.Problem(w, 403, "This list is no longer editable")
		return
	}
	_, e = tx.Exec("DELETE FROM tasks WHERE id=? OR parent_id=?", t.ID, t.ID)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	if e = tx.Commit(); e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
