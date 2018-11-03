package cmd

import (
	"fmt"
	"os"
	"time"

	"boscoin.io/sebak/lib/common"
	logging "github.com/inconshreveable/log15"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/spikeekips/sebak-hot-body/hotbody"
)

func printFlagsError(cmd *cobra.Command, flagName string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid '%s'; %v\n\n", flagName, err)
	}

	cmd.Help()

	os.Exit(1)
}

func printError(cmd *cobra.Command, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n\n", err)
	}

	cmd.Help()

	os.Exit(1)
}

func FormatISO8601(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000000000")
}

func setLogging() {
	var err error
	if logLevel, err = logging.LvlFromString(flagLogLevel); err != nil {
		printFlagsError(goCmd, "--log-level", err)
	}

	var logFormatter logging.Format
	switch flagLogFormat {
	case "terminal":
		if isatty.IsTerminal(os.Stdout.Fd()) && len(flagLog) < 1 {
			logFormatter = logging.TerminalFormat()
		} else {
			logFormatter = logging.LogfmtFormat()
		}
	case "json":
		logFormatter = common.JsonFormatEx(false, true)
	default:
		printFlagsError(goCmd, "--log-format", fmt.Errorf("'%s'", flagLogFormat))
	}

	logHandler := logging.StreamHandler(os.Stdout, logFormatter)
	if len(flagLog) > 0 {
		if logHandler, err = logging.FileHandler(flagLog, logFormatter); err != nil {
			printFlagsError(goCmd, "--log", err)
		}
	}

	log.SetHandler(logging.LvlFilterHandler(logLevel, logging.CallerFileHandler(logHandler)))
	hotbody.SetLogging(logLevel, logHandler)
}
