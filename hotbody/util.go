package hotbody

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ElapsedTime(s time.Time) string {
	return fmt.Sprintf("%.10f", float64(time.Now().Sub(s).Nanoseconds())/float64(1000000000))
}

func ParseRecordElapsedTime(s string) (int64, error) {
	return strconv.ParseInt(strings.Replace(s, ".", "", 1), 10, 64)
}

/*
# reset-by-peer
{
    "Err": {
        "Addr": {
            "IP": "127.0.0.1",
            "Port": 12345,
            "Zone": ""
        },
        "Err": {
            "Err": 54,
            "Syscall": "read"
        },
        "Net": "tcp",
        "Op": "read",
        "Source": {
            "IP": "127.0.0.1",
            "Port": 61321,
            "Zone": ""
        }
    },
    "Op": "Get",
    "URL": "http://127.0.0.1:12345/api/v1/accounts/GAUSKC4GYKVNXSTVGZSZ6R3NEFNM7ZCEL5RIZXNOJ3Z2PQE67XNCQRCO"
}
*/
func ParseRecordErrorHTTPProblem(e map[string]interface{}) RecordErrorType {
	switch e["type"].(string) {
	case "https://boscoin.io/sebak/error/134":
		return RecordErrorTxDoesNotExist
	case "https://boscoin.io/sebak/error/139":
		return RecordErrorSameSourceFound
	}

	return RecordErrorUnknown
}

func ParseRecordErrorNetError(e map[string]interface{}) RecordErrorType {
	var found bool
	var m map[string]interface{}
	if m, found = e["Err"].(map[string]interface{}); found {
		if _, found = m["Syscall"]; found {
			var code interface{}
			if code, found = m["Err"]; !found {
				return RecordErrorUnknown
			} else if code.(float64) == float64(54) {
				return RecordErrorECONNRESET
			}
		}
	}

	return RecordErrorNetworkError
}

func ParseRecordError(e map[string]interface{}) RecordErrorType {
	var found bool
	var m map[string]interface{}

	{
		var v interface{}
		if v, found = e["code"]; found {
			code := v.(float64)
			if code != 163 {
				return RecordErrorUnknown
			}

			if v, found = e["data"]; !found {
				return RecordErrorUnknown
			}

			b := v.(map[string]interface{})["body"]
			if b == nil {
				return RecordErrorUnknown
			}

			if err := json.Unmarshal([]byte(b.(string)), &m); err != nil {
				return RecordErrorUnknown
			}

			return ParseRecordErrorHTTPProblem(m)
		}
	}

	{
		if m, found = e["Err"].(map[string]interface{}); found {
			return ParseRecordErrorNetError(m)
		}
	}

	return RecordErrorUnknown
}
