package core

import (
	"testing"
	"time"
)

func TestUnixToDateEpoch(t *testing.T) {
	got, err := UnixToDate(0, "UTC", "")
	if err != nil {
		t.Fatal(err)
	}
	want := "1970-01-01 00:00:00"
	if got != want {
		t.Errorf("UnixToDate(0, UTC) = %q, want %q", got, want)
	}
}

func TestUnixToDateCustomFormat(t *testing.T) {
	got, err := UnixToDate(0, "UTC", "2006-01-02")
	if err != nil {
		t.Fatal(err)
	}
	if got != "1970-01-01" {
		t.Errorf("UnixToDate with custom format = %q, want %q", got, "1970-01-01")
	}
}

func TestUnixToDateUnknownTimezone(t *testing.T) {
	_, err := UnixToDate(0, "Not/A_Real_Zone", "")
	if err == nil {
		t.Fatal("expected an error for an unknown timezone")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("unknown timezone should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestDateToUnixRoundTrip(t *testing.T) {
	original := int64(1_700_000_000) // an arbitrary, fixed Unix timestamp
	dateStr, err := UnixToDate(original, "UTC", DefaultTimeFormat)
	if err != nil {
		t.Fatal(err)
	}
	back, err := DateToUnix(dateStr, "UTC", DefaultTimeFormat)
	if err != nil {
		t.Fatal(err)
	}
	if back != original {
		t.Errorf("round trip failed: got %d, want %d", back, original)
	}
}

func TestDateToUnixInvalidFormat(t *testing.T) {
	_, err := DateToUnix("not a date", "UTC", DefaultTimeFormat)
	if err == nil {
		t.Fatal("expected an error for a date string that doesn't match the layout")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got code %d", CodeOf(err))
	}
}

func TestParseUnixArg(t *testing.T) {
	got, err := ParseUnixArg("1700000000")
	if err != nil {
		t.Fatal(err)
	}
	if got != 1700000000 {
		t.Errorf("ParseUnixArg = %d, want 1700000000", got)
	}

	_, err = ParseUnixArg("not-a-number")
	if err == nil {
		t.Fatal("expected an error for a non-numeric timestamp")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got code %d", CodeOf(err))
	}
}

func TestNowUnixIsCurrentAndMonotonicallySensible(t *testing.T) {
	before := time.Now().Unix()
	got := NowUnix()
	after := time.Now().Unix()

	if got < before || got > after {
		t.Errorf("NowUnix() = %d, expected it to be between %d and %d", got, before, after)
	}
}
