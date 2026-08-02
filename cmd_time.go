package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"devia/core"
)

func cmdTimestamp(args []string) {
	if len(args) == 0 {
		usageError("timestamp requires a subcommand: now|to-date|from-date")
	}
	sub := args[0]
	rest := args[1:]

	switch sub {
	case "now":
		fs := newFlagSet("timestamp now")
		parseArgs(fs, rest)
		printResult(strconv.FormatInt(core.NowUnix(), 10))

	case "to-date":
		fs := newFlagSet("timestamp to-date")
		tz := fs.String("tz", "UTC", "timezone: UTC|Local|IANA name (e.g. Asia/Shanghai)")
		format := fs.String("format", core.DefaultTimeFormat, "Go reference-time layout")
		parseArgs(fs, rest)
		arg, err := readInput(fs.Arg(0))
		if err != nil {
			usageError(err.Error())
		}
		unix, err := core.ParseUnixArg(arg)
		if err != nil {
			fail(err)
		}
		result, err := core.UnixToDate(unix, *tz, *format)
		if err != nil {
			fail(err)
		}
		printResult(result)

	case "from-date":
		fs := newFlagSet("timestamp from-date")
		tz := fs.String("tz", "UTC", "timezone: UTC|Local|IANA name")
		format := fs.String("format", core.DefaultTimeFormat, "Go reference-time layout")
		parseArgs(fs, rest)
		arg, err := readInput(fs.Arg(0))
		if err != nil {
			usageError(err.Error())
		}
		unix, err := core.DateToUnix(arg, *tz, *format)
		if err != nil {
			fail(err)
		}
		printResult(strconv.FormatInt(unix, 10))

	default:
		usageError("unknown timestamp subcommand: " + sub + " (want now|to-date|from-date)")
	}
}

func cmdCron(args []string) {
	fs := newFlagSet("cron")
	next := fs.Int("next", 5, "how many upcoming run times to show")
	parseArgs(fs, args)

	expr, err := readInput(fs.Arg(0))
	if err != nil {
		usageError(err.Error())
	}
	spec, err := core.ParseCron(expr)
	if err != nil {
		fail(err)
	}
	times, err := spec.Next(time.Now(), *next)
	if err != nil {
		fail(err)
	}

	if jsonMode {
		out := make([]string, len(times))
		for i, t := range times {
			out[i] = t.Format("2006-01-02 15:04:05 Mon")
		}
		printResult(map[string]interface{}{
			"description": core.DescribeCron(expr),
			"next":        out,
		})
		return
	}
	fmt.Println(core.DescribeCron(expr))
	for _, t := range times {
		fmt.Println(t.Format("2006-01-02 15:04:05 Mon"))
	}
	os.Exit(ExitOK)
}
