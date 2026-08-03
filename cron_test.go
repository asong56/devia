package core

import (
	"testing"
	"time"
)

func mustParseCron(t *testing.T, expr string) *CronSpec {
	t.Helper()
	spec, err := ParseCron(expr)
	if err != nil {
		t.Fatalf("ParseCron(%q): %v", expr, err)
	}
	return spec
}

func TestParseCronFieldCounts(t *testing.T) {
	if _, err := ParseCron("* * * *"); err == nil {
		t.Error("expected an error for a 4-field expression")
	}
	if _, err := ParseCron("* * * * * * *"); err == nil {
		t.Error("expected an error for a 7-field expression")
	}
	if _, err := ParseCron("* * * * *"); err != nil {
		t.Errorf("5-field expression should be valid: %v", err)
	}
	if _, err := ParseCron("* * * * * *"); err != nil {
		t.Errorf("6-field expression should be valid: %v", err)
	}
}

func TestParseCronInvalidRange(t *testing.T) {
	_, err := ParseCron("99 * * * *") // minute 99 is out of range (0-59)
	if err == nil {
		t.Fatal("expected an error for an out-of-range field")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got code %d", CodeOf(err))
	}
}

func TestCronNextEveryQuarterHour(t *testing.T) {
	spec := mustParseCron(t, "*/15 * * * *")
	from := time.Date(2024, 1, 1, 0, 5, 0, 0, time.UTC)

	times, err := spec.Next(from, 4)
	if err != nil {
		t.Fatal(err)
	}

	want := []time.Time{
		time.Date(2024, 1, 1, 0, 15, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 30, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 45, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
	}
	if len(times) != len(want) {
		t.Fatalf("got %d times, want %d", len(times), len(want))
	}
	for i, got := range times {
		if !got.Equal(want[i]) {
			t.Errorf("time[%d] = %s, want %s", i, got, want[i])
		}
	}
}

func TestCronNextStrictlyAfterFrom(t *testing.T) {
	// "from" lands exactly on a matching minute; the result must still
	// be strictly after it, not include it.
	spec := mustParseCron(t, "0 * * * *") // top of every hour
	from := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

	times, err := spec.Next(from, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 6, 1, 11, 0, 0, 0, time.UTC)
	if !times[0].Equal(want) {
		t.Errorf("Next() = %s, want %s (strictly after the start time)", times[0], want)
	}
}

func TestCronDomDowOrSemantics(t *testing.T) {
	// When both day-of-month and day-of-week are restricted, standard
	// cron semantics say either matching is enough (OR), not both (AND).
	// Day-of-week is numeric only (0-6, Sunday=0), so Monday is "1".
	spec := mustParseCron(t, "0 0 1 * 1") // midnight on the 1st OR any Monday

	// Jan 8 2024 is a Monday, but not the 1st of the month.
	monday := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	if !spec.matches(monday) {
		t.Error("a Monday that isn't the 1st should still match under OR semantics")
	}

	// Feb 1 2024 is a Thursday, not the 1st field's day-of-week, but is
	// the 1st of the month.
	firstOfMonth := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	if !spec.matches(firstOfMonth) {
		t.Error("the 1st of the month should still match under OR semantics, even on a non-Monday")
	}

	// A Tuesday the 2nd matches neither condition.
	neither := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if spec.matches(neither) {
		t.Error("a day matching neither dom nor dow should not match")
	}
}

func TestCronNextExhaustsSafetyCapGracefully(t *testing.T) {
	// February 30th never exists, so this must fail with an error
	// instead of hanging forever.
	spec := mustParseCron(t, "0 0 30 2 *")
	_, err := spec.Next(time.Now(), 1)
	if err == nil {
		t.Fatal("expected an error for a cron expression that can never match")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got code %d", CodeOf(err))
	}
}

func TestDescribeCron(t *testing.T) {
	if got := DescribeCron("* * * * *"); got != "every minute" {
		t.Errorf("DescribeCron(5-field all-star) = %q, want %q", got, "every minute")
	}
	if got := DescribeCron("* * * * * *"); got != "every second" {
		t.Errorf("DescribeCron(6-field all-star) = %q, want %q", got, "every second")
	}
	if got := DescribeCron("*/15 * * * *"); got == "every minute" {
		t.Error("a non-trivial expression should not be described as 'every minute'")
	}
}
