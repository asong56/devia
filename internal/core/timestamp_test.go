package core

import "testing"

func TestUnixToDate_UTC(t *testing.T) {
	got, err := UnixToDate(1735689600, "UTC", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "2025-01-01 00:00:00"
	if got != want {
		t.Errorf("UnixToDate = %q, want %q", got, want)
	}
}

func TestUnixToDate_NamedTimezone(t *testing.T) {
	// Asia/Shanghai is UTC+8 with no DST — 1735689600 is exactly
	// 2025-01-01 00:00:00 UTC, so it should be 08:00:00 local.
	got, err := UnixToDate(1735689600, "Asia/Shanghai", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "2025-01-01 08:00:00"
	if got != want {
		t.Errorf("UnixToDate(Asia/Shanghai) = %q, want %q", got, want)
	}
}

func TestUnixToDate_EmptyTZDefaultsToUTC(t *testing.T) {
	withEmpty, err := UnixToDate(1735689600, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	withUTC, err := UnixToDate(1735689600, "UTC", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if withEmpty != withUTC {
		t.Errorf("empty tz = %q, should equal explicit UTC %q", withEmpty, withUTC)
	}
}

func TestUnixToDate_CustomFormat(t *testing.T) {
	// 1735689600 is 2025-01-01 00:00:00 UTC.
	got, err := UnixToDate(1735689600, "UTC", "2006/01/02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "2025/01/01"
	if got != want {
		t.Errorf("UnixToDate with custom format = %q, want %q", got, want)
	}
}

func TestUnixToDate_UnknownTimezone(t *testing.T) {
	_, err := UnixToDate(1735689600, "Not/A_Real_Zone", "")
	if err == nil {
		t.Fatal("expected an error for an unknown timezone, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestDateToUnix_UTC(t *testing.T) {
	got, err := DateToUnix("2025-01-01 00:00:00", "UTC", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1735689600 {
		t.Errorf("DateToUnix = %d, want %d", got, 1735689600)
	}
}

func TestDateToUnix_NamedTimezone(t *testing.T) {
	got, err := DateToUnix("2025-01-01 08:00:00", "Asia/Shanghai", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1735689600 {
		t.Errorf("DateToUnix(Asia/Shanghai 08:00) = %d, want %d (same instant as UTC midnight)", got, 1735689600)
	}
}

func TestUnixToDate_DateToUnix_RoundTrip(t *testing.T) {
	original := int64(1700000000)
	dateStr, err := UnixToDate(original, "UTC", "")
	if err != nil {
		t.Fatalf("UnixToDate error: %v", err)
	}
	back, err := DateToUnix(dateStr, "UTC", "")
	if err != nil {
		t.Fatalf("DateToUnix error: %v", err)
	}
	if back != original {
		t.Errorf("round trip = %d, want %d", back, original)
	}
}

func TestDateToUnix_MismatchedFormat(t *testing.T) {
	_, err := DateToUnix("not a date at all", "UTC", "")
	if err == nil {
		t.Fatal("expected an error for a date string that doesn't match the layout, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestDateToUnix_UnknownTimezone(t *testing.T) {
	_, err := DateToUnix("2025-01-01 00:00:00", "Definitely/Not_A_Zone", "")
	if err == nil {
		t.Fatal("expected an error for an unknown timezone, got nil")
	}
}

func TestParseUnixArg_Valid(t *testing.T) {
	n, err := ParseUnixArg("1735689600")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1735689600 {
		t.Errorf("ParseUnixArg = %d, want %d", n, 1735689600)
	}
}

func TestParseUnixArg_Negative(t *testing.T) {
	// Negative Unix timestamps (pre-1970) are legitimate and must be
	// accepted, not rejected as "invalid".
	n, err := ParseUnixArg("-86400")
	if err != nil {
		t.Fatalf("unexpected error for a valid negative timestamp: %v", err)
	}
	if n != -86400 {
		t.Errorf("ParseUnixArg(-86400) = %d, want -86400", n)
	}
}

func TestParseUnixArg_Invalid(t *testing.T) {
	_, err := ParseUnixArg("not-a-number")
	if err == nil {
		t.Fatal("expected an error for non-numeric input, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestParseUnixArg_RejectsFloats(t *testing.T) {
	// A Unix timestamp is documented as whole seconds — "1735689600.5"
	// should be rejected, not silently truncated, since silent
	// truncation of user input is exactly the kind of surprise this
	// tool tries to avoid elsewhere (see the CLI craftsmanship notes
	// on explicit errors over guessing).
	_, err := ParseUnixArg("1735689600.5")
	if err == nil {
		t.Fatal("expected an error for a fractional timestamp, got nil")
	}
}
