package hotbody

import (
	"encoding/json"
	"fmt"
	"os"

	"boscoin.io/sebak/lib/common"
)

type Result struct {
	config HotterConfig
	output *os.File
}

func NewResult(config HotterConfig) (result *Result, err error) {
	var output *os.File
	if output, err = os.Create(config.ResultOutput); err != nil {
		return
	}

	result = &Result{
		config: config,
		output: output,
	}

	result.write(map[string]interface{}{"type": "config", "config": config, "time": common.NowISO8601()})

	return
}

func (r *Result) Close() {
	r.output.Close()
}

func (r *Result) write(o interface{}) {
	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}

	if _, err := fmt.Fprintln(r.output, string(b)); err != nil {
		panic(err)
	}
}

func (r *Result) Write(t string, args ...interface{}) {
	if len(args)%2 == 1 {
		panic(fmt.Errorf("invalid pair of args"))
	}

	d := map[string]interface{}{
		"type": t,
		"time": common.NowISO8601(),
	}

	for i := 0; i < len(args)-1; i = i + 2 {
		k := args[i]
		v := args[i+1]

		var key string
		switch k.(type) {
		case string:
			key = k.(string)
		case int, int64, uint64, float64:
			key = fmt.Sprintf("%v", k)
		default:
			panic(fmt.Errorf("invalid key type found: %T, %v", k, k))
		}

		d[key] = v
	}

	r.write(d)
}
