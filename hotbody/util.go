package hotbody

import (
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
func ParseRecordError(e map[string]interface{}) RecordErrorType {
	var m map[string]interface{}
	var found bool
	if m, found = e["Err"].(map[string]interface{}); found {
		if m, found = m["Err"].(map[string]interface{}); found {
			if _, found = m["Syscall"]; found {
				var code interface{}
				if code, found = m["Err"]; !found {
					return RecordErrorUnknown
				} else if code.(float64) == float64(54) {
					return RecordErrorECONNRESET
				}
			}
		}
	}

	return RecordErrorUnknown
}
