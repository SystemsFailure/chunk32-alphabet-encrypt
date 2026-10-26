
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	decodeFlag := flag.Bool(
		"d",
		false,
		"Decode input (chunk32 -d)",
	)

	upperFlag := flag.Bool(
		"u",
		false,
		"Use uppercase characters when encoding (chunk32 -u)",
	)

	checkFlag := flag.Bool(
		"check",
		false,
		"Add a check symbol when encoding (chunk32 -check)",
	)

	noNewlineFlag := flag.Bool(
		"n",
		false,
		"Do not print a trailing newline (chunk32 -d -n)",
	)

	flag.Parse()

	var err error

	if *decodeFlag {
		err = decodeFromStdin(*noNewlineFlag)
	} else {
		err = encodeFromStdin(
			*upperFlag,
			*checkFlag,
			*noNewlineFlag,
		)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "chunk32:", err)
		os.Exit(1)
	}
}
