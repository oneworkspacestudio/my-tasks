package app

import (
	"errors"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"sort"
	"time"
)

func resolveLocal(value, zone, fold string) (time.Time, error) {
	wall, e := time.Parse("2006-01-02T15:04", value)
	if e != nil {
		return time.Time{}, errors.New("Choose a date and time")
	}
	loc, e := time.LoadLocation(zone)
	if e != nil {
		return time.Time{}, errors.New("Choose an IANA timezone")
	}
	offsets := map[int]bool{}
	for h := -48; h <= 48; h += 3 {
		_, o := wall.Add(time.Duration(h) * time.Hour).In(loc).Zone()
		offsets[o] = true
	}
	candidates := []time.Time{}
	for o := range offsets {
		candidate := wall.Add(-time.Duration(o) * time.Second)
		if candidate.In(loc).Format("2006-01-02T15:04") == value {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		return time.Time{}, errors.New("This local time does not exist because the clocks change. Choose another time")
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Before(candidates[j]) })
	if fold == "later" {
		return candidates[len(candidates)-1], nil
	}
	return candidates[0], nil
}
func (a *App) resolveTime(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	at, e := resolveLocal(q.Get("local"), q.Get("zone"), q.Get("fold"))
	if e != nil {
		apphost.Problem(w, 400, e.Error())
		return
	}
	apphost.JSON(w, 200, map[string]string{"instant": at.UTC().Format(time.RFC3339)})
}
