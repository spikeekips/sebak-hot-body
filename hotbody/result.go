package hotbody

import (
	"encoding/json"
	"fmt"
	"os"
)

type Result struct {
	config HotterConfig
	output *os.File
}

func NewResult(config HotterConfig) (result *Result, err error) {
	var b []byte
	if b, err = json.Marshal(map[string]interface{}{"type": "config", "config": config}); err != nil {
		return
	}

	var output *os.File
	if output, err = os.Create(config.ResultOutput); err != nil {
		return
	}

	if _, err = fmt.Fprintln(output, string(b)); err != nil {
		return
	}

	result = &Result{
		config: config,
		output: output,
	}
	return
}

func (r *Result) Write(t string, args ...interface{}) {
	if len(args)%2 == 1 {
		panic(fmt.Errorf("invalid pair of args"))
	}

	d := map[string]interface{}{
		"type": t,
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

	b, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	if _, err := fmt.Fprintln(r.output, string(b)); err != nil {
		panic(err)
	}
}
