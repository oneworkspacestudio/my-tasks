package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

func writableTx(tx *sql.Tx, user, list string) bool {
	var owner string
	var role sql.NullString
	if tx.QueryRow("SELECT owner,(SELECT role FROM task_members WHERE list_id=l.id AND user_id=?) FROM task_lists l WHERE id=?", user, list).Scan(&owner, &role) != nil {
		return false
	}
	return owner == user || role.String == "editor"
}
func reviseRecurrence(tx *sql.Tx, t Task) error {
	if t.RuleID == "" {
		if t.Rule != "" {
			return errors.New("Create a new recurring task to start a series")
		}
		return nil
	}
	var rule string
	if e := tx.QueryRow("SELECT rule FROM task_rules WHERE id=?", t.RuleID).Scan(&rule); e != nil {
		return e
	}
	if t.Rule != rule {
		return errors.New("This occurrence belongs to an existing series. Keep its recurrence rule, or create a new recurring task")
	}
	if len(t.Exceptions) > 366 {
		return errors.New("Too many exception dates")
	}
	for _, v := range t.Exceptions {
		if _, e := time.Parse("2006-01-02", v); e != nil {
			return errors.New("Exception dates use YYYY-MM-DD")
		}
	}
	raw, _ := json.Marshal(t.Exceptions)
	_, e := tx.Exec("UPDATE task_rules SET exceptions=? WHERE id=?", string(raw), t.RuleID)
	return e
}
