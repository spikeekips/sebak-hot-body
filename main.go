package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	logging "github.com/inconshreveable/log15"
	"github.com/spikeekips/sebak-hot-body/hotbody"
	"github.com/stellar/go/keypair"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/node"
)

const (
	defaultLogLevel              logging.Lvl = logging.LvlInfo
	defaultLogFormat             string      = "terminal"
	defaultConcurrentTransaction int         = 10
	defaultRequestTimeout        string      = "30s"
	defaultConfirmDuration       string      = "60s"
	defaultTimeout               string      = "1m"
)

var (
	flagSEBAKEndpoint         string = "http://127.0.0.1:12345"
	flagLogLevel              string = defaultLogLevel.String()
	flagLogFormat             string = defaultLogFormat
	flagLog                   string
	flagConcurrentTransaction int    = defaultConcurrentTransaction
	flagRequestTimeout        string = defaultRequestTimeout
	flagConfirmDuration       string = defaultConfirmDuration
	flagTimeout               string = defaultTimeout
	flagResultOutput          string
)

var (
	sebakEndpoint   *common.Endpoint
	logLevel        logging.Lvl
	log             logging.Logger = logging.New("module", "hot-body")
	nodeInfo        node.NodeInfo
	kp              *keypair.Full
	timeout         time.Duration
	requestTimeout  time.Duration
	confirmDuration time.Duration
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <secret seed>\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	var err error
	var currentDirectory string
	if currentDirectory, err = os.Getwd(); err != nil {
		hotbody.PrintError(err)
	}
	if currentDirectory, err = filepath.Abs(currentDirectory); err != nil {
		hotbody.PrintError(err)
	}

	flag.StringVar(&flagSEBAKEndpoint, "sebak", flagSEBAKEndpoint, "sebak endpoint")
	flag.StringVar(&flagLogLevel, "log-level", flagLogLevel, "log level, {crit, error, warn, info, debug}")
	flag.StringVar(&flagLogFormat, "log-format", flagLogFormat, "log format, {terminal, json}")
	flag.StringVar(&flagLog, "log", flagLog, "set log file")
	flag.IntVar(&flagConcurrentTransaction, "t", flagConcurrentTransaction, "number of transactions, they will be sent concurrently")
	flag.StringVar(&flagRequestTimeout, "request-timeout", flagRequestTimeout, "timeout for requests")
	flag.StringVar(&flagConfirmDuration, "confirm-duration", flagConfirmDuration, "duration for checking transaction confirmed")
	flag.StringVar(&flagTimeout, "timeout", flagTimeout, "timeout for running")

	outputFile := currentDirectory + fmt.Sprintf(
		"/result-%s.log",
		time.Now().Format("20060102-150405.000000000Z0700"),
	)
	flag.StringVar(&flagResultOutput, "result-output", outputFile, "result output file")

	flag.Parse()
	if flag.NArg() < 1 {
		hotbody.PrintError(fmt.Errorf("<secret seed> is missing"))
	}
	if parsedKP, err := keypair.Parse(flag.Arg(0)); err != nil {
		hotbody.PrintError(fmt.Errorf("invalid <secret seed>: %v", err))
	} else {
		var ok bool
		if kp, ok = parsedKP.(*keypair.Full); !ok {
			hotbody.PrintError(fmt.Errorf("invalid <secret seed>: not secret seed"))
		}
	}

	if p, err := common.ParseEndpoint(flagSEBAKEndpoint); err != nil {
		hotbody.PrintFlagsError("--sebak", err)
	} else {
		sebakEndpoint = p
		flagSEBAKEndpoint = sebakEndpoint.String()
	}
	if flagConcurrentTransaction < 1 {
		hotbody.PrintFlagsError("--sebak", errors.New("at least bigger than 0"))
	}
	if len(flagRequestTimeout) < 1 {
		hotbody.PrintFlagsError("--request-timeout", errors.New("must be given"))
	} else if requestTimeout, err = time.ParseDuration(flagRequestTimeout); err != nil {
		hotbody.PrintFlagsError("--request-timeout", err)
	}
	if len(flagConfirmDuration) < 1 {
		hotbody.PrintFlagsError("--confirm-duration", errors.New("must be given"))
	} else if confirmDuration, err = time.ParseDuration(flagConfirmDuration); err != nil {
		hotbody.PrintFlagsError("--confirm-duration", err)
	}
	if len(flagTimeout) < 1 {
		hotbody.PrintFlagsError("--timeout", errors.New("must be given"))
	} else if timeout, err = time.ParseDuration(flagTimeout); err != nil {
		hotbody.PrintFlagsError("--timeout", err)
	}

	if logLevel, err = logging.LvlFromString(flagLogLevel); err != nil {
		hotbody.PrintFlagsError("--log-level", err)
	}

	var logFormatter logging.Format
	switch flagLogFormat {
	case "terminal":
		logFormatter = logging.TerminalFormat()
	case "json":
		logFormatter = common.JsonFormatEx(false, true)
	default:
		hotbody.PrintFlagsError("--log-format", fmt.Errorf("'%s'", flagLogFormat))
	}

	logHandler := logging.StreamHandler(os.Stdout, logFormatter)
	if len(flagLog) > 0 {
		if logHandler, err = logging.FileHandler(flagLog, logFormatter); err != nil {
			hotbody.PrintFlagsError("--log", err)
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
	parsedFlags = append(parsedFlags, "\n", "")

	log.Debug("parsed flags:", parsedFlags...)
}

func main() {
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
		hotbody.PrintError(fmt.Errorf("failed to create HTTP2Client: %v", err))
	}
	client.Transport().MaxIdleConnsPerHost = flagConcurrentTransaction

	var b []byte
	if b, err = client.Get("/", nil); err != nil {
		hotbody.PrintFlagsError("--sebak", err)
	}

	if nodeInfo, err = node.NewNodeInfoFromJSON(b); err != nil {
		hotbody.PrintError(fmt.Errorf("failed to parse node info response: %v", err))
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
	}

	var hotter *hotbody.Hotter
	hotter, err = hotbody.NewHotter(hotterConfig, client)
	if err != nil {
		hotbody.PrintError(fmt.Errorf("something wrong: %v", err))
	}

	if _, err := hotter.GetAccount(kp.Address(), true); err != nil {
		hotbody.PrintError(fmt.Errorf("account of <secret seed> not found"))
	}

	if err := hotter.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "end with error: %v\n", err)
		os.Exit(1)
	}

	log.Debug("hot-body ended")
	os.Exit(0)
}
