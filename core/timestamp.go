package core

import (
	"strconv"
	"time"
)

const DefaultTimeFormat = "2006-01-02 15:04:05"

func resolveLocation(tz string) (*time.Location, error) {
	switch tz {
	case "", "UTC", "utc":
		return time.UTC, nil
	case "Local", "local":
		return time.Local, nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, NewInputError("unknown timezone: " + tz)
	}
	return loc, nil
}

func NowUnix() int64 { return time.Now().Unix() }

// UnixToDate formats a Unix timestamp (seconds) in the given timezone
// ("UTC", "Local", or an IANA name like "Asia/Shanghai") using a Go
// reference-time layout string.
func UnixToDate(unixSeconds int64, tz, format string) (string, error) {
	loc, err := resolveLocation(tz)
	if err != nil {
		return "", err
	}
	if format == "" {
		format = DefaultTimeFormat
	}
	return time.Unix(unixSeconds, 0).In(loc).Format(format), nil
}

// DateToUnix parses a date string in the given timezone and layout,
// returning its Unix timestamp in seconds.
func DateToUnix(dateStr, tz, format string) (int64, error) {
	loc, err := resolveLocation(tz)
	if err != nil {
		return 0, err
	}
	if format == "" {
		format = DefaultTimeFormat
	}
	t, err := time.ParseInLocation(format, dateStr, loc)
	if err != nil {
		return 0, NewInputError("cannot parse date with layout " + format + ": " + err.Error())
	}
	return t.Unix(), nil
}

func ParseUnixArg(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, NewInputError("invalid unix timestamp: " + s)
	}
	return n, nil
}
