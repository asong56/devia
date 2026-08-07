package core

import (
	"testing"
	"time"
)

func mustUTC(y, mo, d, h, mi, s int) time.Time {
	return time.Date(y, time.Month(mo), d, h, mi, s, 0, time.UTC)
}

func TestParseCron_FieldCount(t *testing.T) {
	if _, err := ParseCron("* * * *"); err == nil { // 4 fields
		t.Error("expected an error for a 4-field expression, got nil")
	}
	if _, err := ParseCron("* * * * *"); err != nil { // 5 fields, valid
		t.Errorf("unexpected error for a valid 5-field expression: %v", err)
	}
	if _, err := ParseCron("* * * * * *"); err != nil { // 6 fields, valid
		t.Errorf("unexpected error for a valid 6-field expression: %v", err)
	}
	if _, err := ParseCron("* * * * * * *"); err == nil { // 7 fields
		t.Error("expected an error for a 7-field expression, got nil")
	}
}

func TestParseCron_OutOfRangeValue(t *testing.T) {
	// Minute field only goes 0-59.
	if _, err := ParseCron("60 * * * *"); err == nil {
		t.Error("expected an error for minute=60 (out of range), got nil")
	}
}

func TestParseCron_InvalidStep(t *testing.T) {
	if _, err := ParseCron("*/0 * * * *"); err == nil {
		t.Error("expected an error for a zero step, got nil")
	}
	if _, err := ParseCron("*/abc * * * *"); err == nil {
		t.Error("expected an error for a non-numeric step, got nil")
	}
}

func TestParseCron_InvertedRange(t *testing.T) {
	// 30-10 is backwards (a > b) and must be rejected, not silently
	// produce an empty or wrapped-around range.
	if _, err := ParseCron("30-10 * * * *"); err == nil {
		t.Error("expected an error for an inverted range (30-10), got nil")
	}
}

func TestCronNext_EveryFifteenMinutes(t *testing.T) {
	spec, err := ParseCron("*/15 * * * *")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	from := mustUTC(2024, 1, 1, 0, 0, 0) // a Monday
	times, err := spec.Next(from, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []time.Time{
		mustUTC(2024, 1, 1, 0, 15, 0),
		mustUTC(2024, 1, 1, 0, 30, 0),
		mustUTC(2024, 1, 1, 0, 45, 0),
		mustUTC(2024, 1, 1, 1, 0, 0),
		mustUTC(2024, 1, 1, 1, 15, 0),
	}
	if len(times) != len(want) {
		t.Fatalf("got %d times, want %d", len(times), len(want))
	}
	for i, wt := range want {
		if !times[i].Equal(wt) {
			t.Errorf("times[%d] = %s, want %s", i, times[i], wt)
		}
	}
}

func TestCronNext_SixFieldWithSeconds(t *testing.T) {
	spec, err := ParseCron("30 * * * * *") // every minute, at :30 seconds
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	from := mustUTC(2024, 1, 1, 0, 0, 0)
	times, err := spec.Next(from, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []time.Time{
		mustUTC(2024, 1, 1, 0, 0, 30),
		mustUTC(2024, 1, 1, 0, 1, 30),
		mustUTC(2024, 1, 1, 0, 2, 30),
	}
	for i, wt := range want {
		if !times[i].Equal(wt) {
			t.Errorf("times[%d] = %s, want %s", i, times[i], wt)
		}
	}
}

func TestCronNext_DomDowOrSemantics(t *testing.T) {
	// When BOTH day-of-month and day-of-week are restricted (neither is
	// "*"), standard cron semantics say a match on EITHER is enough.
	// "0 0 1 * 1" = 00:00 on the 1st of the month OR every Monday.
	spec, err := ParseCron("0 0 1 * 1")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	from := mustUTC(2024, 1, 2, 0, 0, 0) // Tuesday, Jan 2 2024
	times, err := spec.Next(from, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Jan 8 2024 and Jan 15 2024 are both Mondays, and neither is the
	// 1st of the month — proving the match came from the dow branch,
	// not the dom branch.
	want := []time.Time{
		mustUTC(2024, 1, 8, 0, 0, 0),
		mustUTC(2024, 1, 15, 0, 0, 0),
	}
	if len(times) != 2 {
		t.Fatalf("got %d times, want 2: %v", len(times), times)
	}
	for i, wt := range want {
		if !times[i].Equal(wt) {
			t.Errorf("times[%d] = %s, want %s", i, times[i], wt)
		}
	}
}

func TestCronNext_StrictlyAfterFrom(t *testing.T) {
	// If "from" itself is an exact match instant, Next must not return
	// it — only times strictly after "from".
	spec, err := ParseCron("0 0 * * *") // midnight every day
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	from := mustUTC(2024, 1, 1, 0, 0, 0) // itself a match
	times, err := spec.Next(from, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if times[0].Equal(from) {
		t.Error("Next returned the 'from' instant itself; it must only return times strictly after it")
	}
	want := mustUTC(2024, 1, 2, 0, 0, 0)
	if !times[0].Equal(want) {
		t.Errorf("times[0] = %s, want %s", times[0], want)
	}
}

func TestCronNext_NeverMatchingExpressionFailsRatherThanHanging(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow (~2M iteration) safety-cap test in -short mode")
	}
	// February never has a 30th day — this expression can never match
	// any real calendar date, so Next must hit its ~4 year safety cap
	// and return an error instead of looping forever.
	spec, err := ParseCron("0 0 30 2 *")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	from := mustUTC(2024, 1, 1, 0, 0, 0)
	_, err = spec.Next(from, 1)
	if err == nil {
		t.Fatal("expected an error for a never-matching expression, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestDescribeCron(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"* * * * *", "every minute"},
		{"* * * * * *", "every second"},
		{"*/15 * * * *", "custom schedule (*/15 * * * *)"},
	}
	for _, c := range cases {
		got := DescribeCron(c.expr)
		if got != c.want {
			t.Errorf("DescribeCron(%q) = %q, want %q", c.expr, got, c.want)
		}
	}
}
