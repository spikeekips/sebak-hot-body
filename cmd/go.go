package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	logging "github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
	"github.com/spikeekips/sebak-hot-body/hotbody"
	"github.com/stellar/go/keypair"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/node"
)

var (
	goCmd *cobra.Command
)

func init() {
	goCmd = &cobra.Command{
		Use:   "go <secret seed>",
		Short: "Run hot-body",
		Run: func(c *cobra.Command, args []string) {
			parseGoFlags(args)

			runGo()
		},
	}

	var err error
	var currentDirectory string
	if currentDirectory, err = os.Getwd(); err != nil {
		hotbody.PrintError(goCmd, err)
	}
	if currentDirectory, err = filepath.Abs(currentDirectory); err != nil {
		hotbody.PrintError(goCmd, err)
	}

	now := time.Now().Format("20060102150405")
	flagResultOutput = currentDirectory + fmt.Sprintf("/hot-body-result-%s.log", now)
	flagLog = currentDirectory + fmt.Sprintf("/hot-body-%s.log", now)

	goCmd.Flags().StringVar(&flagSEBAKEndpoint, "sebak", flagSEBAKEndpoint, "sebak endpoint")
	goCmd.Flags().StringVar(&flagLogLevel, "log-level", flagLogLevel, "log level, {crit, error, warn, info, debug}")
	goCmd.Flags().StringVar(&flagLogFormat, "log-format", flagLogFormat, "log format, {terminal, json}")
	goCmd.Flags().StringVar(&flagLog, "log", flagLog, "set log file")
	goCmd.Flags().IntVar(&flagConcurrentTransaction, "concurrent", flagConcurrentTransaction, "number of transactions, they will be sent concurrently")
	goCmd.Flags().StringVar(&flagRequestTimeout, "request-timeout", flagRequestTimeout, "timeout for requests")
	goCmd.Flags().StringVar(&flagConfirmDuration, "confirm-duration", flagConfirmDuration, "duration for checking transaction confirmed")
	goCmd.Flags().StringVar(&flagTimeout, "timeout", flagTimeout, "timeout for running")
	goCmd.Flags().IntVar(&flagOperations, "operations", flagOperations, "number of operations in one transaction")
	goCmd.Flags().StringVar(&flagResultOutput, "result-output", flagResultOutput, "result output file")

	rootCmd.AddCommand(goCmd)
}

func parseGoFlags(args []string) {
	var err error

	if len(args) < 1 {
		hotbody.PrintError(goCmd, fmt.Errorf("<secret seed> is missing"))
	}
	if parsedKP, err := keypair.Parse(args[0]); err != nil {
		hotbody.PrintError(goCmd, fmt.Errorf("invalid <secret seed>: %v", err))
	} else {
		var ok bool
		if kp, ok = parsedKP.(*keypair.Full); !ok {
			hotbody.PrintError(goCmd, fmt.Errorf("invalid <secret seed>: not secret seed"))
		}
	}

	if p, err := common.ParseEndpoint(flagSEBAKEndpoint); err != nil {
		hotbody.PrintFlagsError(goCmd, "--sebak", err)
	} else {
		sebakEndpoint = p
		flagSEBAKEndpoint = sebakEndpoint.String()
	}
	if flagConcurrentTransaction < 1 {
		hotbody.PrintFlagsError(goCmd, "--concurrent", errors.New("at least bigger than 0"))
	}
	if flagOperations < 1 {
		hotbody.PrintFlagsError(goCmd, "--operations", errors.New("at least bigger than 0"))
	}
	if len(flagRequestTimeout) < 1 {
		hotbody.PrintFlagsError(goCmd, "--request-timeout", errors.New("must be given"))
	} else if requestTimeout, err = time.ParseDuration(flagRequestTimeout); err != nil {
		hotbody.PrintFlagsError(goCmd, "--request-timeout", err)
	}
	if len(flagConfirmDuration) < 1 {
		hotbody.PrintFlagsError(goCmd, "--confirm-duration", errors.New("must be given"))
	} else if confirmDuration, err = time.ParseDuration(flagConfirmDuration); err != nil {
		hotbody.PrintFlagsError(goCmd, "--confirm-duration", err)
	}
	if len(flagTimeout) < 1 {
		hotbody.PrintFlagsError(goCmd, "--timeout", errors.New("must be given"))
	} else if timeout, err = time.ParseDuration(flagTimeout); err != nil {
		hotbody.PrintFlagsError(goCmd, "--timeout", err)
	}

	if logLevel, err = logging.LvlFromString(flagLogLevel); err != nil {
		hotbody.PrintFlagsError(goCmd, "--log-level", err)
	}

	var logFormatter logging.Format
	switch flagLogFormat {
	case "terminal":
		logFormatter = logging.TerminalFormat()
	case "json":
		logFormatter = common.JsonFormatEx(false, true)
	default:
		hotbody.PrintFlagsError(goCmd, "--log-format", fmt.Errorf("'%s'", flagLogFormat))
	}

	logHandler := logging.StreamHandler(os.Stdout, logFormatter)
	if len(flagLog) > 0 {
		if logHandler, err = logging.FileHandler(flagLog, logFormatter); err != nil {
			hotbody.PrintFlagsError(goCmd, "--log", err)
		}
	}

	log.SetHandler(logging.LvlFilterHandler(logLevel, logging.CallerFileHandler(logHandler)))
	hotbody.SetLogging(logLevel, logHandler)

	parsedFlags := []interface{}{}
	parsedFlags = append(parsedFlags, "\n\tsebak", flagSEBAKEndpoint)
	parsedFlags = append(parsedFlags, "\n\tlog-level", flagLogLevel)
	parsedFlags = append(parsedFlags, "\n\tlog-format", flagLogFormat)
	parsedFlags = append(parsedFlags, "\n\tlog", flagLog)
	parsedFlags = append(parsedFlags, "\n\tt", flagConcurrentTransaction)
	parsedFlags = append(parsedFlags, "\n\trequest-timeout", flagRequestTimeout)
	parsedFlags = append(parsedFlags, "\n\ttimeout", flagTimeout)
	parsedFlags = append(parsedFlags, "\n\tconfirm-duration", flagConfirmDuration)
	parsedFlags = append(parsedFlags, "\n\tresult-output", flagResultOutput)
	parsedFlags = append(parsedFlags, "\n\toperations", flagOperations)
	parsedFlags = append(parsedFlags, "\n", "")

	log.Debug("parsed flags:", parsedFlags...)
}

func runGo() {
	var err error

	// request node info to sebak
	var client *hotbody.HTTP2Client

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("User-Agent", "sebak-hot-body/v1.0")

	client, err = hotbody.NewHTTP2Client(
		requestTimeout,
		(*url.URL)(sebakEndpoint),
		headers,
	)
	if err != nil {
		hotbody.PrintError(goCmd, fmt.Errorf("failed to create HTTP2Client: %v", err))
	}
	client.Transport().MaxIdleConnsPerHost = flagConcurrentTransaction

	var b []byte
	if b, err = client.Get("/", nil); err != nil {
		hotbody.PrintFlagsError(goCmd, "--sebak", err)
	}

	if nodeInfo, err = node.NewNodeInfoFromJSON(b); err != nil {
		hotbody.PrintError(goCmd, fmt.Errorf("failed to parse node info response: %v", err))
	}
	log.Debug("sebak info", "sebak", sebakEndpoint.String())
	log.Debug(fmt.Sprintf(
		`================================================================================
%s
================================================================================
`, b))

	hotterConfig := hotbody.HotterConfig{
		Node:            nodeInfo,
		T:               flagConcurrentTransaction,
		KP:              kp,
		InitAccount:     kp.Address(),
		Timeout:         timeout,
		RequestTimeout:  requestTimeout,
		ConfirmDuration: confirmDuration,
		ResultOutput:    flagResultOutput,
		Operations:      flagOperations,
	}

	var hotter *hotbody.Hotter
	hotter, err = hotbody.NewHotter(hotterConfig, client)
	if err != nil {
		hotbody.PrintError(goCmd, fmt.Errorf("something wrong: %v", err))
	}

	if _, err := hotter.GetAccount(kp.Address(), true); err != nil {
		hotbody.PrintError(goCmd, fmt.Errorf("account of <secret seed> not found"))
	}

	if err := hotter.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "end with error: %v\n", err)
		os.Exit(1)
	}

	log.Debug("hot-body ended")
	os.Exit(0)
}
