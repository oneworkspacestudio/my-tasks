package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

func (a *App) bulk(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs    []string `json:"ids"`
		Action string   `json:"action"`
		List   string   `json:"list"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	if len(in.IDs) == 0 || len(in.IDs) > 100 {
		apphost.Problem(w, 400, "Select 1–100 tasks")
		return
	}
	switch in.Action {
	case "done", "undo", "delete", "star", "move", "reorder":
	default:
		apphost.Problem(w, 400, "Unknown bulk action")
		return
	}
	u := apphost.Current(r).User.ID
	if in.Action == "move" && !canWrite(a.role(u, in.List)) {
		apphost.Problem(w, 403, "Destination is not editable")
		return
	}
	items := []Task{}
	seen := map[string]bool{}
	for _, id := range in.IDs {
		if seen[id] {
			continue
		}
		seen[id] = true
		t, role, e := a.task(u, id)
		if e != nil || !canWrite(role) {
			apphost.Problem(w, 403, "One or more tasks are no longer editable")
			return
		}
		if in.Action == "move" && (t.RuleID != "" || t.ParentID != "") {
			apphost.Problem(w, 400, "Move a non-recurring parent task; its subtasks move with it")
			return
		}
		items = append(items, t)
	}
	tx, e := a.H.DB.Begin()
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer tx.Rollback()
	for i, t := range items {
		if !writableTx(tx, u, t.ListID) || in.Action == "move" && !writableTx(tx, u, in.List) {
			apphost.Problem(w, 403, "A list is no longer editable")
			return
		}
		switch in.Action {
		case "delete":
			_, e = tx.Exec("DELETE FROM tasks WHERE id=? OR parent_id=?", t.ID, t.ID)
		case "move":
			_, e = tx.Exec("UPDATE tasks SET list_id=?,version=version+1,updated=? WHERE id=? OR parent_id=?", in.List, now(), t.ID, t.ID)
		case "star":
			_, e = tx.Exec("UPDATE tasks SET starred=NOT starred,version=version+1,updated=? WHERE id=?", now(), t.ID)
		case "reorder":
			_, e = tx.Exec("UPDATE tasks SET position=?,version=version+1,updated=? WHERE id=?", float64(i), now(), t.ID)
		case "done":
			if t.Status != "done" {
				_, e = tx.Exec("UPDATE tasks SET previous_status=status,status='done',version=version+1,updated=? WHERE id=?", now(), t.ID)
				if e == nil && t.RuleID != "" {
					e = nextOccurrence(tx, t)
				}
			}
		case "undo":
			if t.Status == "done" {
				_, e = tx.Exec("DELETE FROM tasks WHERE generated_by=? AND version=1 AND status='todo'", t.ID)
				if e == nil {
					_, e = tx.Exec("UPDATE tasks SET status=previous_status,version=version+1,updated=? WHERE id=?", now(), t.ID)
				}
			}
		}
		if e != nil {
			a.H.DBError(w, e)
			return
		}
	}
	if e = tx.Commit(); e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]int{"changed": len(items)})
}
func (a *App) notifications(w http.ResponseWriter, r *http.Request) {
	u := apphost.Current(r).User.ID
	rows, e := a.H.DB.Query("SELECT n.id,n.task_id,n.at,n.read,t.title FROM task_notifications n JOIN tasks t ON t.id=n.task_id JOIN task_lists l ON l.id=t.list_id WHERE n.user_id=? AND "+access+" ORDER BY n.at DESC LIMIT 100", u, u, u)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []Notification{}
	for rows.Next() {
		var n Notification
		if e = rows.Scan(&n.ID, &n.TaskID, &n.At, &n.Read, &n.Title); e != nil {
			a.H.DBError(w, e)
			return
		}
		out = append(out, n)
	}
	apphost.JSON(w, 200, out)
}
func (a *App) readNotification(w http.ResponseWriter, r *http.Request) {
	_, e := a.H.DB.Exec("UPDATE task_notifications SET read=1 WHERE id=? AND user_id=?", r.PathValue("id"), apphost.Current(r).User.ID)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) Reminders(ctx context.Context, at time.Time) error {
	rows, e := a.H.DB.QueryContext(ctx, "SELECT t.id,t.title,CASE WHEN snooze<>'' THEN snooze ELSE reminder END,l.owner FROM tasks t JOIN task_lists l ON l.id=t.list_id WHERE t.status<>'done' AND (t.reminder<>'' OR t.snooze<>'')")
	if e != nil {
		return e
	}
	type due struct{ task, title, when, owner string }
	items := []due{}
	for rows.Next() {
		var d due
		if e = rows.Scan(&d.task, &d.title, &d.when, &d.owner); e != nil {
			rows.Close()
			return e
		}
		when, e := time.Parse(time.RFC3339, d.when)
		if e == nil && !when.After(at) {
			items = append(items, d)
		}
		if len(items) >= 1000 {
			break
		}
	}
	rows.Close()
	for _, d := range items {
		users := []string{d.owner}
		rs, e := a.H.DB.QueryContext(ctx, "SELECT user_id FROM task_members WHERE list_id=(SELECT list_id FROM tasks WHERE id=?)", d.task)
		if e != nil {
			return e
		}
		for rs.Next() {
			var u string
			rs.Scan(&u)
			users = append(users, u)
		}
		rs.Close()
		for _, u := range users {
			id := apphost.Hash(d.task + "|" + u + "|" + d.when)
			_, e = a.H.DB.ExecContext(ctx, "INSERT OR IGNORE INTO task_delivery(id,task_id,user_id,at)VALUES(?,?,?,?)", id, d.task, u, d.when)
			if e != nil {
				return e
			}
			var state, next string
			var attempts int
			if e = a.H.DB.QueryRowContext(ctx, "SELECT state,next_attempt,attempts FROM task_delivery WHERE id=?", id).Scan(&state, &next, &attempts); e != nil {
				return e
			}
			if state == "sent" {
				continue
			}
			if n, e := time.Parse(time.RFC3339, next); e == nil && n.After(at) {
				continue
			}
			n := Notification{ID: id, TaskID: d.task, UserID: u, At: d.when, Title: d.title}
			_, e = a.H.DB.ExecContext(ctx, "INSERT OR IGNORE INTO task_notifications(id,task_id,user_id,at)VALUES(?,?,?,?)", id, d.task, u, d.when)
			if e != nil {
				return e
			}
			if a.Deliver != nil {
				e = a.Deliver(ctx, n)
			}
			if e != nil {
				_, e = a.H.DB.ExecContext(ctx, "UPDATE task_delivery SET attempts=attempts+1,next_attempt=? WHERE id=?", at.Add(time.Duration(min(attempts+1, 60))*time.Minute).UTC().Format(time.RFC3339Nano), id)
			} else {
				_, e = a.H.DB.ExecContext(ctx, "UPDATE task_delivery SET state='sent',attempts=attempts+1 WHERE id=?", id)
			}
			if e != nil {
				return e
			}
		}
	}
	return nil
}

type View struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Filter map[string]string `json:"filter"`
}

func (a *App) views(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT id,name,filter FROM task_views WHERE user_id=? ORDER BY name", apphost.Current(r).User.ID)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []View{}
	for rows.Next() {
		var v View
		var f string
		if e = rows.Scan(&v.ID, &v.Name, &f); e != nil {
			a.H.DBError(w, e)
			return
		}
		json.Unmarshal([]byte(f), &v.Filter)
		out = append(out, v)
	}
	apphost.JSON(w, 200, out)
}
func (a *App) saveView(w http.ResponseWriter, r *http.Request) {
	var v View
	if !apphost.Decode(w, r, &v) {
		return
	}
	if strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 || len(v.Filter) > 5 {
		apphost.Problem(w, 400, "Name this saved view")
		return
	}
	for k, val := range v.Filter {
		if !(k == "q" || k == "view" || k == "list" || k == "tag") || len(val) > 200 {
			apphost.Problem(w, 400, "Unsupported view filter")
			return
		}
	}
	v.ID = apphost.ID()
	f, _ := json.Marshal(v.Filter)
	_, e := a.H.DB.Exec("INSERT INTO task_views VALUES(?,?,?,?)", v.ID, apphost.Current(r).User.ID, v.Name, string(f))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 201, v)
}
func (a *App) deleteView(w http.ResponseWriter, r *http.Request) {
	_, e := a.H.DB.Exec("DELETE FROM task_views WHERE id=? AND user_id=?", r.PathValue("id"), apphost.Current(r).User.ID)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func icsEscape(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\r", "", "\n", "\\n", ";", "\\;", ",", "\\,").Replace(s)
}
func fold(s string) string {
	out := ""
	limit := 75
	for len(s) > limit {
		n := limit
		for !utf8.RuneStart(s[n]) {
			n--
		}
		out += s[:n] + "\r\n "
		s = s[n:]
		limit = 74
	}
	return out + s + "\r\n"
}
func (a *App) exportICS(w http.ResponseWriter, r *http.Request) {
	u := apphost.Current(r).User.ID
	rows, e := a.H.DB.Query("SELECT "+cols+" FROM tasks t JOIN task_lists l ON l.id=t.list_id WHERE "+access+" ORDER BY t.id LIMIT 10001", u, u)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	var lines []string
	count := 0
	for rows.Next() {
		t, e := scanTask(rows)
		if e != nil {
			rows.Close()
			a.H.DBError(w, e)
			return
		}
		count++
		if count > 10000 {
			rows.Close()
			apphost.Problem(w, 413, "Export up to 10,000 tasks at once")
			return
		}
		lines = append(lines, "BEGIN:VTODO", "UID:"+t.ID+"@mytasks.local", "DTSTAMP:"+time.Now().UTC().Format("20060102T150405Z"), "SUMMARY:"+icsEscape(t.Title), "DESCRIPTION:"+icsEscape(t.Description))
		if due, e := time.Parse(time.RFC3339, t.Due); e == nil {
			lines = append(lines, "DUE:"+due.UTC().Format("20060102T150405Z"))
		}
		status := "NEEDS-ACTION"
		if t.Status == "done" {
			status = "COMPLETED"
		}
		lines = append(lines, "STATUS:"+status, "END:VTODO")
	}
	rows.Close()
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="tasks.ics"`)
	fmt.Fprint(w, "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//One Workspace//My Tasks//EN\r\n")
	for _, line := range lines {
		fmt.Fprint(w, fold(line))
	}
	fmt.Fprint(w, "END:VCALENDAR\r\n")
}

var _ = sql.ErrNoRows
