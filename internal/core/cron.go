package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronSpec is a parsed 5-field (minute hour dom month dow) or 6-field
// (second minute hour dom month dow) cron expression.
type CronSpec struct {
	hasSeconds bool
	seconds    map[int]bool
	minutes    map[int]bool
	hours      map[int]bool
	doms       map[int]bool // day of month, 1-31
	months     map[int]bool // 1-12
	dows       map[int]bool // day of week, 0-6, Sunday = 0
	domStar    bool
	dowStar    bool
}

// ParseCron parses a standard 5 or 6 field cron expression. Supports
// "*", lists ("1,2,3"), ranges ("1-5"), and steps ("*/15", "1-30/5").
func ParseCron(expr string) (*CronSpec, error) {
	fields := strings.Fields(strings.TrimSpace(expr))
	spec := &CronSpec{}

	var secF, minF, hourF, domF, monF, dowF string
	switch len(fields) {
	case 5:
		minF, hourF, domF, monF, dowF = fields[0], fields[1], fields[2], fields[3], fields[4]
	case 6:
		spec.hasSeconds = true
		secF, minF, hourF, domF, monF, dowF = fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]
	default:
		return nil, NewInputError("cron expression must have 5 fields (min hour dom mon dow) or 6 (sec min hour dom mon dow)")
	}

	var err error
	if spec.hasSeconds {
		if spec.seconds, err = parseCronField(secF, 0, 59); err != nil {
			return nil, err
		}
	} else {
		spec.seconds = map[int]bool{0: true}
	}
	if spec.minutes, err = parseCronField(minF, 0, 59); err != nil {
		return nil, err
	}
	if spec.hours, err = parseCronField(hourF, 0, 23); err != nil {
		return nil, err
	}
	spec.domStar = domF == "*" || domF == "?"
	if spec.doms, err = parseCronField(domF, 1, 31); err != nil {
		return nil, err
	}
	if spec.months, err = parseCronField(monF, 1, 12); err != nil {
		return nil, err
	}
	spec.dowStar = dowF == "*" || dowF == "?"
	if spec.dows, err = parseCronField(dowF, 0, 6); err != nil {
		return nil, err
	}
	return spec, nil
}

func parseCronField(field string, min, max int) (map[int]bool, error) {
	result := make(map[int]bool)
	if field == "*" || field == "?" {
		for i := min; i <= max; i++ {
			result[i] = true
		}
		return result, nil
	}
	for _, part := range strings.Split(field, ",") {
		step := 1
		rangePart := part
		if idx := strings.Index(part, "/"); idx != -1 {
			rangePart = part[:idx]
			s, err := strconv.Atoi(part[idx+1:])
			if err != nil || s < 1 {
				return nil, NewInputError("invalid step in cron field: " + part)
			}
			step = s
		}

		lo, hi := min, max
		if rangePart != "*" {
			if idx := strings.Index(rangePart, "-"); idx != -1 {
				a, err1 := strconv.Atoi(rangePart[:idx])
				b, err2 := strconv.Atoi(rangePart[idx+1:])
				if err1 != nil || err2 != nil || a < min || b > max || a > b {
					return nil, NewInputError("invalid range in cron field: " + part)
				}
				lo, hi = a, b
			} else {
				v, err := strconv.Atoi(rangePart)
				if err != nil || v < min || v > max {
					return nil, NewInputError("invalid value in cron field: " + part)
				}
				lo, hi = v, v
			}
		}
		for i := lo; i <= hi; i += step {
			result[i] = true
		}
	}
	if len(result) == 0 {
		return nil, NewInputError("empty cron field: " + field)
	}
	return result, nil
}

func (s *CronSpec) matches(t time.Time) bool {
	if !s.months[int(t.Month())] {
		return false
	}
	domMatch := s.doms[t.Day()]
	dowMatch := s.dows[int(t.Weekday())]
	// Standard cron semantics: if BOTH day-of-month and day-of-week are
	// restricted, a match on either is enough (OR). If only one is
	// restricted, it must match (the other is "*" and always true).
	if !s.domStar && !s.dowStar {
		if !domMatch && !dowMatch {
			return false
		}
	} else if !domMatch || !dowMatch {
		return false
	}
	if !s.hours[t.Hour()] {
		return false
	}
	if !s.minutes[t.Minute()] {
		return false
	}
	if s.hasSeconds && !s.seconds[t.Second()] {
		return false
	}
	return true
}

// Next returns the next n scheduled times strictly after "from". Uses
// simple forward stepping (by second if the expression has a seconds
// field, otherwise by minute) with a ~4 year safety cap so a
// never-matching expression (e.g. Feb 30) fails fast instead of
// looping forever.
func (s *CronSpec) Next(from time.Time, n int) ([]time.Time, error) {
	step := time.Minute
	t := from.Truncate(time.Minute).Add(time.Minute)
	maxIters := 4 * 366 * 24 * 60
	if s.hasSeconds {
		step = time.Second
		t = from.Truncate(time.Second).Add(time.Second)
		maxIters = 4 * 366 * 24 * 60 * 60
	}

	out := make([]time.Time, 0, n)
	for i := 0; i < maxIters && len(out) < n; i++ {
		if s.matches(t) {
			out = append(out, t)
		}
		t = t.Add(step)
	}
	if len(out) < n {
		return out, NewInputError("could not find enough matching times within the search window")
	}
	return out, nil
}

// DescribeCron gives a short best-effort human description.
func DescribeCron(expr string) string {
	fields := strings.Fields(strings.TrimSpace(expr))
	allStar := true
	for _, f := range fields {
		if f != "*" {
			allStar = false
			break
		}
	}
	if allStar {
		if len(fields) == 6 {
			return "every second"
		}
		return "every minute"
	}
	return fmt.Sprintf("custom schedule (%s)", expr)
}
