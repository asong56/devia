package cli

import (
	"os"

	"devia/internal/core"
)

func cmdBase64(args []string) {
	if len(args) == 0 {
		usageError("base64 requires a subcommand: encode|decode")
	}
	sub := args[0]
	rest := args[1:]

	fs := newFlagSet("base64 " + sub)
	file := fs.String("file", "", "read input from this file (binary-safe — use for images)")
	out := fs.String("out", "", "decode: write raw bytes to this file instead of stdout")
	dataURI := fs.Bool("data-uri", false, "encode: wrap output as a data: URI (detects common image types)")
	urlSafe := fs.Bool("url", false, "use the URL-safe base64 alphabet")
	parseArgs(fs, rest)

	switch sub {
	case "encode":
		if *file != "" {
			result, err := core.Base64EncodeFile(*file, *dataURI)
			if err != nil {
				fail(err)
			}
			printResult(result)
			return
		}
		text, err := readInput(fs.Arg(0))
		if err != nil {
			usageError(err.Error())
		}
		printResult(core.Base64EncodeText(text, *urlSafe))

	case "decode":
		text, err := readInput(fs.Arg(0))
		if err != nil {
			usageError(err.Error())
		}
		if *out != "" {
			if err := core.Base64DecodeToFile(text, *out); err != nil {
				fail(err)
			}
			printResult("written to " + *out)
			return
		}
		result, err := core.Base64DecodeText(text, *urlSafe)
		if err != nil {
			fail(err)
		}
		printResult(result)

	default:
		usageError("unknown base64 subcommand: " + sub + " (want encode|decode)")
	}
}

func cmdJSON(args []string) {
	if len(args) == 0 {
		usageError("json requires a subcommand: format|minify|validate")
	}
	sub := args[0]
	rest := args[1:]

	fs := newFlagSet("json " + sub)
	indent := fs.String("indent", "  ", "indent string used by format")
	file := fs.String("file", "", "read JSON from this file instead of an argument/stdin")
	parseArgs(fs, rest)

	var text string
	if *file != "" {
		b, err := os.ReadFile(*file)
		if err != nil {
			fail(readFileErr(err, *file))
		}
		text = string(b)
	} else {
		t, err := readInput(fs.Arg(0))
		if err != nil {
			usageError(err.Error())
		}
		text = t
	}

	switch sub {
	case "format":
		result, err := core.JSONFormat(text, *indent)
		if err != nil {
			fail(err)
		}
		printResult(result)
	case "minify":
		result, err := core.JSONMinify(text)
		if err != nil {
			fail(err)
		}
		printResult(result)
	case "validate":
		if err := core.JSONValidate(text); err != nil {
			fail(err)
		}
		printResult("valid")
	default:
		usageError("unknown json subcommand: " + sub + " (want format|minify|validate)")
	}
}

func cmdEscape(args []string, isUnescape bool) {
	verb := "escape"
	if isUnescape {
		verb = "unescape"
	}
	if len(args) == 0 {
		usageError(verb + " requires a mode: json|url|url-path|html|unicode")
	}
	mode := args[0]
	rest := args[1:]

	fs := newFlagSet(verb + " " + mode)
	parseArgs(fs, rest)

	text, err := readInput(fs.Arg(0))
	if err != nil {
		usageError(err.Error())
	}

	var result string
	switch mode {
	case "json":
		if isUnescape {
			result, err = core.UnescapeJSON(text)
		} else {
			result, err = core.EscapeJSON(text)
		}
	case "url":
		if isUnescape {
			result, err = core.UnescapeURL(text, false)
		} else {
			result = core.EscapeURL(text, false)
		}
	case "url-path":
		if isUnescape {
			result, err = core.UnescapeURL(text, true)
		} else {
			result = core.EscapeURL(text, true)
		}
	case "html":
		if isUnescape {
			result = core.UnescapeHTML(text)
		} else {
			result = core.EscapeHTML(text)
		}
	case "unicode":
		if isUnescape {
			result, err = core.UnescapeUnicode(text)
		} else {
			result = core.EscapeUnicode(text)
		}
	default:
		usageError("unknown " + verb + " mode: " + mode + " (want json|url|url-path|html|unicode)")
	}
	if err != nil {
		fail(err)
	}
	printResult(result)
}
