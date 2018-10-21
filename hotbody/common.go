package hotbody

import (
	"flag"
	"fmt"
	"os"
)

func PrintFlagsError(flagName string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid '%s'; %v\n\n", flagName, err)
	}

	flag.Usage()

	os.Exit(1)
}

func PrintError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n\n", err)
	}

	flag.Usage()

	os.Exit(1)
}
