package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/apcera/termtables"
	"github.com/spf13/cobra"
	"github.com/spikeekips/sebak-hot-body/hotbody"

	"boscoin.io/sebak/lib/common"
)

var (
	resultCmd    *cobra.Command
	resultOutput *os.File
	started      time.Time
	ended        time.Time
)

func init() {
	resultCmd = &cobra.Command{
		Use:   "result <result output>",
		Short: "Parse result",
		Run: func(c *cobra.Command, args []string) {
			parseResultFlags(args)

			runResult()
		},
	}

	resultCmd.Flags().StringVar(&flagLogLevel, "log-level", flagLogLevel, "log level, {crit, error, warn, info, debug}")
	resultCmd.Flags().StringVar(&flagLogFormat, "log-format", flagLogFormat, "log format, {terminal, json}")
	resultCmd.Flags().StringVar(&flagLog, "log", flagLog, "set log file")

	rootCmd.AddCommand(resultCmd)
}

func parseResultFlags(args []string) {
	if len(args) < 1 {
		hotbody.PrintError(resultCmd, fmt.Errorf("<result output> is missing"))
	}
	flagResultOutput = args[0]

	var err error
	if resultOutput, err = os.Open(flagResultOutput); err != nil {
		hotbody.PrintError(resultCmd, fmt.Errorf("failed to open <result output>; %v", err))
	}

	parseLogging(resultCmd)

	parsedFlags := []interface{}{}
	parsedFlags = append(parsedFlags, "\n\tresult-log", flagResultOutput)
	parsedFlags = append(parsedFlags, "\n\tlog-level", flagLogLevel)
	parsedFlags = append(parsedFlags, "\n\tlog-format", flagLogFormat)
	parsedFlags = append(parsedFlags, "\n\tlog", flagLog)
	parsedFlags = append(parsedFlags, "\n", "")

	log.Debug("parsed flags:", parsedFlags...)
}

func loadLine(l string) (record hotbody.Record, err error) {
	var d map[string]interface{}
	if err = json.Unmarshal([]byte(l), &d); err != nil {
		return
	}

	if _, found := d["type"]; !found {
		err = fmt.Errorf("found invalid format")
		return
	}

	recordType := d["type"].(string)
	switch recordType {
	case "started":
		started, _ = common.ParseISO8601(d["time"].(string))
		return
	case "ended":
		ended, _ = common.ParseISO8601(d["time"].(string))
		return
	case "config":
		var b []byte
		if b, err = json.Marshal(d["config"]); err != nil {
			return
		}
		var hotterConfig hotbody.HotterConfig
		if err = json.Unmarshal(b, &hotterConfig); err != nil {
			return
		}

		record = hotterConfig
	case "create-accounts":
		var createAccounts hotbody.RecordCreateAccounts
		if err = json.Unmarshal([]byte(l), &createAccounts); err != nil {
			return
		}

		record = createAccounts
	case "payment":
		var payment hotbody.RecordPayment
		if err = json.Unmarshal([]byte(l), &payment); err != nil {
			return
		}

		record = payment
	default:
		err = fmt.Errorf("unknown type found: %v", recordType)
		return
	}

	return
}

func runResult() {
	defer resultOutput.Close()

	var err error

	sc := bufio.NewScanner(resultOutput)
	sc.Split(bufio.ScanLines)

	var config hotbody.HotterConfig

	sc.Scan()
	headLine := sc.Text()

	var record hotbody.Record
	if record, err = loadLine(headLine); err != nil {
		hotbody.PrintError(resultCmd, fmt.Errorf("something wrong to read <result output>; %v; %v", err, headLine))
	} else {
		config = record.(hotbody.HotterConfig)
	}
	log.Debug("config loaded", "config", config)

	log.Debug("trying to load record")
	var records []hotbody.Record
	for sc.Scan() {
		s := sc.Text()

		if record, err = loadLine(s); err != nil {
			hotbody.PrintError(resultCmd, fmt.Errorf("something wrong to read <result output>; %v; %v", err, s))
		} else if record == nil {
			continue
		}

		records = append(records, record)
	}
	log.Debug("records loaded", "count", len(records))

	if err = sc.Err(); err != nil {
		hotbody.PrintError(resultCmd, fmt.Errorf("something wrong to read <result output>; %v", err))
	}

	var maxElapsedTime float64
	var minElapsedTime float64 = -1

	var countError int
	errorTypes := map[hotbody.RecordErrorType]int{}
	for _, r := range records {
		es := float64(r.GetElapsed())
		maxElapsedTime = math.Max(maxElapsedTime, es)
		if minElapsedTime < 0 {
			minElapsedTime = es
		} else {
			minElapsedTime = math.Min(minElapsedTime, es)
		}

		if r.GetError() == nil {
			continue
		}
		countError++
		errorTypes[r.GetErrorType()]++
	}

	alignKey := func(s string) string {
		return fmt.Sprintf("% 20s", s)
	}

	alignValue := func(v interface{}) string {
		s := fmt.Sprintf("%v", v)
		switch v.(type) {
		case float64:
			s = fmt.Sprintf("%15.10f", v)
		}

		return fmt.Sprintf("% 30s", s)
	}

	alignHead := func(s string) string {
		return fmt.Sprintf("* %-10s", s)
	}

	formatAddress := func(s string) string {
		return fmt.Sprintf("%s...%s", s[:13], s[len(s)-13:])
	}

	var table *termtables.Table

	{
		table = termtables.CreateTable()
		table.AddRow(alignHead("config"), alignKey("testing time"), alignValue(config.Timeout))
		table.AddRow("", alignKey("concurrent requests"), alignValue(config.T))
		table.AddRow("", alignKey("initial account"), alignValue(formatAddress(config.InitAccount)))
		table.AddRow("", alignKey("request timeout"), alignValue(config.RequestTimeout))
		table.AddRow("", alignKey("confirm duration"), alignValue(config.ConfirmDuration))
		table.AddRow("", alignKey("operations"), alignValue(config.Operations))
	}

	{
		table.AddSeparator()
		table.AddRow(alignHead("network"), alignKey("network id"), alignValue(config.Node.Policy.NetworkID))
		table.AddRow("", alignKey("initial balance"), alignValue(config.Node.Policy.InitialBalance))
		table.AddRow("", alignKey("block time"), alignValue(config.Node.Policy.BlockTime))
		table.AddRow("", alignKey("base reserve"), alignValue(config.Node.Policy.BaseReserve))
		table.AddRow("", alignKey("base fee"), alignValue(config.Node.Policy.BaseFee))
	}

	{
		table.AddSeparator()
		table.AddRow(alignHead("node"), alignKey("endpoint"), alignValue(config.Node.Node.Endpoint))
		table.AddRow("", alignKey("address"), alignValue(formatAddress(config.Node.Node.Address)))
		table.AddRow("", alignKey("state"), alignValue(config.Node.Node.State))
		table.AddRow("", alignKey("block height"), alignValue(config.Node.Block.Height))
		table.AddRow("", alignKey("block hash"), alignValue(formatAddress(config.Node.Block.Hash)))
		table.AddRow("", alignKey("block totaltxs"), alignValue(config.Node.Block.TotalTxs))
	}

	{
		table.AddSeparator()
		table.AddRow(alignHead("time"), alignKey("started"), alignValue(FormatISO8601(started)))
		table.AddRow("", alignKey("ended"), alignValue(FormatISO8601(ended)))
	}

	{
		table.AddSeparator()
		table.AddRow(alignHead("result"), alignKey("# requests"), alignValue(len(records)))
		table.AddRow("", alignKey("# operations"), alignValue(len(records)*config.Operations))
		table.AddRow(
			"",
			alignKey("error rates"),
			alignValue(
				fmt.Sprintf(
					"%2.5f％ (%d/%d)",
					float64(countError)/float64(len(records))*100,
					countError,
					len(records),
				),
			),
		)
		table.AddRow("", alignKey("max elapsed time"), alignValue(maxElapsedTime/float64(10000000000)))
		table.AddRow("", alignKey("min elapsed time"), alignValue(minElapsedTime/float64(10000000000)))

		lastTime := records[len(records)-1].GetTime()

		totalSeconds := lastTime.Sub(started).Seconds()

		ops := float64((len(records))*config.Operations) / float64(totalSeconds)
		table.AddRow("", alignKey("expected OPS"), alignValue(int(ops)))
		ops = float64((len(records)-countError)*config.Operations) / float64(totalSeconds)
		table.AddRow("", alignKey("real OPS"), alignValue(int(ops)))
	}

	{
		table.AddSeparator()
		if countError < 1 {
			table.AddRow(alignHead("error"), alignKey("no error"), "")
		} else {
			var c int
			for errorType, errorCount := range errorTypes {
				h := ""
				if c == 0 {
					h = alignHead("error")
				}
				c++
				table.AddRow(
					h,
					alignKey(string(errorType)),
					alignValue(
						fmt.Sprintf(
							"%d | % 10s",
							errorCount,
							fmt.Sprintf(
								"%.5f％",
								float64(errorCount)/float64(countError)*100,
							),
						),
					),
				)
			}
		}
	}
	fmt.Fprintf(os.Stdout, table.Render())

	os.Exit(0)
}
