package main

import (
	"fmt"
	"os"

	"devia/core"
)

func cmdHash(args []string) {
	fs := newFlagSet("hash")
	algo := fs.String("algo", "sha256", "hash algorithm: md5|sha1|sha256|sha512")
	hmacKey := fs.String("hmac", "", "compute an HMAC using this key instead of a plain hash")
	useB64 := fs.Bool("base64", false, "output base64 instead of hex")
	file := fs.String("file", "", "hash a file instead of text (streamed, not loaded into memory)")
	parseArgs(fs, args)

	if *file != "" {
		out, err := core.HashFile(*algo, *file, *useB64)
		if err != nil {
			fail(err)
		}
		printResult(out)
		return
	}

	text, err := readInput(fs.Arg(0))
	if err != nil {
		usageError(err.Error())
	}
	out, err := core.HashText(*algo, text, *hmacKey, *useB64)
	if err != nil {
		fail(err)
	}
	printResult(out)
}

func cmdChecksum(args []string) {
	fs := newFlagSet("checksum")
	algo := fs.String("algo", "sha256", "hash algorithm: md5|sha1|sha256|sha512")
	compare := fs.String("compare", "", "expected checksum to compare against")
	parseArgs(fs, args)

	path := fs.Arg(0)
	if path == "" {
		usageError("checksum requires a file path")
	}

	if *compare != "" {
		actual, match, err := core.CompareChecksum(*algo, path, *compare)
		if err != nil {
			fail(err)
		}
		if jsonMode {
			printResult(map[string]interface{}{"checksum": actual, "match": match})
			return
		}
		fmt.Println(actual)
		if match {
			fmt.Println("match")
			os.Exit(ExitOK)
		}
		fmt.Println("mismatch")
		os.Exit(ExitError)
	}

	out, err := core.HashFile(*algo, path, false)
	if err != nil {
		fail(err)
	}
	printResult(out)
}
