package main

import (
	"fmt"
	"os"

	"devia/core"
)

func cmdText(args []string) {
	if len(args) == 0 {
		usageError("text requires a mode, e.g. camel|snake|kebab|upper|lower|sentence|title|pascal|constant|cobol|train|alternating|inverse")
	}
	mode := args[0]
	rest := args[1:]

	fs := newFlagSet("text " + mode)
	parseArgs(fs, rest)

	text, err := readInput(fs.Arg(0))
	if err != nil {
		usageError(err.Error())
	}
	result, err := core.TextTransform(mode, text)
	if err != nil {
		fail(err)
	}
	printResult(result)
}

func cmdUUID(args []string) {
	fs := newFlagSet("uuid")
	count := fs.Int("count", 1, "number of UUIDs to generate")
	upper := fs.Bool("upper", false, "output uppercase")
	parseArgs(fs, args)

	ids, err := core.NewUUIDs(*count, *upper)
	if err != nil {
		fail(err)
	}
	if *count <= 1 {
		printResult(ids[0])
		return
	}
	if jsonMode {
		printResult(ids)
		return
	}
	for _, id := range ids {
		fmt.Println(id)
	}
	os.Exit(ExitOK)
}

func cmdLorem(args []string) {
	fs := newFlagSet("lorem")
	kind := fs.String("type", "paragraph", "word|sentence|paragraph")
	count := fs.Int("count", 1, "how many words/sentences/paragraphs")
	classic := fs.Bool("classic", false, "always start with 'Lorem ipsum dolor sit amet...'")
	parseArgs(fs, args)

	result, err := core.LoremGenerate(*kind, *count, *classic)
	if err != nil {
		fail(err)
	}
	printResult(result)
}
