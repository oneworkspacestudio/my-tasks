package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"github.com/teambition/rrule-go"
	"strings"
	"time"
)

// Only bounded calendar-scale rules are accepted; secondly/minutely expansion is
// deliberately excluded. DTSTART and the IANA location are stored independently.
func ruleFor(rule, start, zone string) (*rrule.RRule, error) {
	if len(rule) > 400 || strings.ContainsAny(rule, "\r\n") {
		return nil, errors.New("Enter one calendar recurrence rule")
	}
	allowed := map[string]bool{"FREQ": true, "INTERVAL": true, "COUNT": true, "UNTIL": true, "BYDAY": true, "BYMONTHDAY": true, "BYMONTH": true, "WKST": true}
	for _, part := range strings.Split(strings.TrimPrefix(rule, "RRULE:"), ";") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 || !allowed[kv[0]] || len(strings.Split(kv[1], ",")) > 31 {
			return nil, errors.New("Unsupported recurrence option")
		}
	}
	loc, e := time.LoadLocation(zone)
	if e != nil {
		return nil, e
	}
	at, e := time.Parse(time.RFC3339, start)
	if e != nil {
		return nil, errors.New("Recurring tasks need a due date")
	}
	o, e := rrule.StrToROptionInLocation(strings.TrimPrefix(rule, "RRULE:"), loc)
	if e != nil {
		return nil, errors.New("Check the recurrence rule")
	}
	if o.Freq != rrule.DAILY && o.Freq != rrule.WEEKLY && o.Freq != rrule.MONTHLY && o.Freq != rrule.YEARLY {
		return nil, errors.New("Use daily, weekly, monthly or yearly recurrence")
	}
	if o.Interval > 1000 || o.Count > 10000 {
		return nil, errors.New("Recurrence interval or count is too large")
	}
	o.Dtstart = at.In(loc)
	return rrule.NewRRule(*o)
}
func createRule(tx *sql.Tx, t *Task) error {
	if t.ParentID != "" {
		return errors.New("Add recurrence to the parent task")
	}
	if _, e := ruleFor(t.Rule, t.Due, t.Timezone); e != nil {
		return e
	}
	if len(t.Exceptions) > 366 {
		return errors.New("Too many recurrence exceptions")
	}
	for _, d := range t.Exceptions {
		if _, e := time.Parse("2006-01-02", d); e != nil {
			return errors.New("Exceptions use YYYY-MM-DD")
		}
	}
	at, _ := time.Parse(time.RFC3339, t.Due)
	loc, _ := time.LoadLocation(t.Timezone)
	for _, exception := range t.Exceptions {
		if exception == at.In(loc).Format("2006-01-02") {
			return errors.New("The first due date is excluded; choose the first included occurrence")
		}
	}
	t.RuleID = apphost.ID()
	ex, _ := json.Marshal(t.Exceptions)
	template, _ := json.Marshal(t)
	_, e := tx.Exec("INSERT INTO task_rules VALUES(?,?,?,?,?,?,?)", t.RuleID, t.ListID, t.Rule, t.Due, t.Timezone, string(ex), string(template))
	return e
}
func nextOccurrence(tx *sql.Tx, t Task) error {
	var rule, start, zone, exceptions, template string
	if e := tx.QueryRow("SELECT rule,start,timezone,exceptions,template FROM task_rules WHERE id=?", t.RuleID).Scan(&rule, &start, &zone, &exceptions, &template); e != nil {
		return e
	}
	rr, e := ruleFor(rule, start, zone)
	if e != nil {
		return e
	}
	var ex []string
	json.Unmarshal([]byte(exceptions), &ex)
	skip := map[string]bool{}
	for _, v := range ex {
		skip[v] = true
	}
	at, _ := time.Parse(time.RFC3339, t.Due)
	loc, _ := time.LoadLocation(zone)
	original, _ := time.Parse(time.RFC3339, start)
	wall := original.In(loc)
	for i := 0; i < 500; i++ {
		n := rr.After(at, false)
		if n.IsZero() {
			return nil
		}
		at = n
		local := n.In(loc)
		if skip[local.Format("2006-01-02")] || local.Hour() != wall.Hour() || local.Minute() != wall.Minute() {
			continue
		}
		var next Task
		if e = json.Unmarshal([]byte(template), &next); e != nil {
			return e
		}
		next.ID = apphost.ID()
		next.RuleID = t.RuleID
		next.GeneratedBy = t.ID
		next.Due = n.UTC().Format(time.RFC3339Nano)
		next.Status = "todo"
		next.Previous = "todo"
		next.Version = 1
		next.Updated = now()
		next.Snooze = ""
		if next.Reminder != "" {
			rem, e := time.Parse(time.RFC3339, next.Reminder)
			if e == nil {
				next.Reminder = n.Add(rem.Sub(original)).UTC().Format(time.RFC3339Nano)
			}
		}
		var exists int
		if e = tx.QueryRow("SELECT COUNT(*) FROM tasks WHERE rule_id=? AND due=?", next.RuleID, next.Due).Scan(&exists); e != nil {
			return e
		}
		if exists > 0 {
			return nil
		}
		return insert(tx, next)
	}
	return errors.New("No occurrence found within the bounded recurrence window")
}
