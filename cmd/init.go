package cmd

import (
	"os"
	"time"

	logging "github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
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
	defaultOperations            int         = 1
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
	flagOperations            int = defaultOperations
	flagBrief                 bool
)

var (
	sebakEndpoints  []*common.Endpoint
	logLevel        logging.Lvl
	log             logging.Logger = logging.New("module", "hot-body")
	nodeInfo        node.NodeInfo
	kp              *keypair.Full
	timeout         time.Duration
	requestTimeout  time.Duration
	confirmDuration time.Duration
)

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "sebak-hot-body",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			c.Usage()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printFlagsError(rootCmd, "", err)
	}
}

func SetArgs(s []string) {
	rootCmd.SetArgs(s)
}
