package main

import (
	"fmt"
	"os"

	"devia/core"
)

func cmdRadix(args []string) {
	fs := newFlagSet("radix")
	from := fs.Int("from", 0, "source base (0 = auto-detect via 0x/0o/0b prefix, else decimal)")
	parseArgs(fs, args)

	text, err := readInput(fs.Arg(0))
	if err != nil {
		usageError(err.Error())
	}
	result, err := core.ConvertRadix(text, *from)
	if err != nil {
		fail(err)
	}
	if jsonMode {
		printResult(result)
		return
	}
	fmt.Printf("bin: %s\noct: %s\ndec: %s\nhex: %s\n", result.Bin, result.Oct, result.Dec, result.Hex)
	os.Exit(ExitOK)
}

func cmdRegex(args []string) {
	if len(args) == 0 {
		usageError("regex requires a subcommand: test|replace")
	}
	sub := args[0]
	rest := args[1:]

	switch sub {
	case "test":
		fs := newFlagSet("regex test")
		flags := fs.String("flags", "", "regex flags: any of i,m,s,U")
		pattern := fs.String("pattern", "", "regular expression pattern (required)")
		parseArgs(fs, rest)
		if *pattern == "" {
			usageError("regex test requires --pattern")
		}
		text, err := readInput(fs.Arg(0))
		if err != nil {
			usageError(err.Error())
		}
		matches, err := core.RegexTest(*pattern, *flags, text)
		if err != nil {
			fail(err)
		}
		if jsonMode {
			printResult(matches)
			return
		}
		fmt.Printf("%d match(es)\n", len(matches))
		for _, m := range matches {
			fmt.Printf("  [%d:%d] %q\n", m.Start, m.End, m.Match)
		}
		os.Exit(ExitOK)

	case "replace":
		fs := newFlagSet("regex replace")
		flags := fs.String("flags", "", "regex flags: any of i,m,s,U")
		pattern := fs.String("pattern", "", "regular expression pattern (required)")
		replacement := fs.String("with", "", "replacement text ($1-style group refs supported)")
		parseArgs(fs, rest)
		if *pattern == "" {
			usageError("regex replace requires --pattern")
		}
		text, err := readInput(fs.Arg(0))
		if err != nil {
			usageError(err.Error())
		}
		result, err := core.RegexReplace(*pattern, *flags, text, *replacement)
		if err != nil {
			fail(err)
		}
		printResult(result)

	default:
		usageError("unknown regex subcommand: " + sub + " (want test|replace)")
	}
}

func cmdDiff(args []string) {
	fs := newFlagSet("diff")
	fileA := fs.String("a", "", "path to the first file")
	fileB := fs.String("b", "", "path to the second file")
	parseArgs(fs, args)

	var textA, textB string
	switch {
	case *fileA != "" && *fileB != "":
		ba, err := os.ReadFile(*fileA)
		if err != nil {
			fail(readFileErr(err, *fileA))
		}
		bb, err := os.ReadFile(*fileB)
		if err != nil {
			fail(readFileErr(err, *fileB))
		}
		textA, textB = string(ba), string(bb)
	case fs.NArg() >= 2:
		textA, textB = fs.Arg(0), fs.Arg(1)
	default:
		usageError("diff requires either --a/--b files or two positional text arguments")
	}

	lines, err := core.DiffText(textA, textB)
	if err != nil {
		fail(err)
	}
	if jsonMode {
		printResult(lines)
		return
	}
	fmt.Print(core.FormatDiff(lines))
	os.Exit(ExitOK)
}

func cmdCert(args []string) {
	if len(args) == 0 {
		usageError("cert requires a subcommand: decode")
	}
	sub := args[0]
	rest := args[1:]
	if sub != "decode" {
		usageError("unknown cert subcommand: " + sub + " (want decode)")
	}

	fs := newFlagSet("cert decode")
	parseArgs(fs, rest)

	var info *core.CertInfo
	var err error
	if path := fs.Arg(0); path != "" {
		info, err = core.DecodeCertificateFile(path)
	} else {
		text, rerr := readInput("")
		if rerr != nil {
			usageError("cert decode requires a file path, or PEM data piped via stdin")
		}
		info, err = core.DecodeCertificate([]byte(text))
	}
	if err != nil {
		fail(err)
	}
	if jsonMode {
		printResult(info)
		return
	}
	fmt.Print(core.FormatCertInfo(info))
	os.Exit(ExitOK)
}
