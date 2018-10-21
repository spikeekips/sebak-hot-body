package hotbody

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func PrintFlagsError(cmd *cobra.Command, flagName string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid '%s'; %v\n\n", flagName, err)
	}

	cmd.Help()

	os.Exit(1)
}

func PrintError(cmd *cobra.Command, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n\n", err)
	}

	cmd.Help()

	os.Exit(1)
}
