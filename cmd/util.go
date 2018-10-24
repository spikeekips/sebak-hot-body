package cmd

import (
	"fmt"
	"os"
	"time"

	logging "github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
	"github.com/spikeekips/sebak-hot-body/hotbody"

	"boscoin.io/sebak/lib/common"
)

func parseLogging(c *cobra.Command) {
	var err error
	if logLevel, err = logging.LvlFromString(flagLogLevel); err != nil {
		hotbody.PrintFlagsError(c, "--log-level", err)
	}

	var logFormatter logging.Format
	switch flagLogFormat {
	case "terminal":
		logFormatter = logging.TerminalFormat()
	case "json":
		logFormatter = common.JsonFormatEx(false, true)
	default:
		hotbody.PrintFlagsError(c, "--log-format", fmt.Errorf("'%s'", flagLogFormat))
	}

	logHandler := logging.StreamHandler(os.Stdout, logFormatter)
	if len(flagLog) > 0 {
		if logHandler, err = logging.FileHandler(flagLog, logFormatter); err != nil {
			hotbody.PrintFlagsError(c, "--log", err)
		}
	}

	log.SetHandler(logging.LvlFilterHandler(logLevel, logging.CallerFileHandler(logHandler)))
	hotbody.SetLogging(logLevel, logHandler)
}

func FormatISO8601(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000000000")
}
