package hotbody

import (
	"fmt"
	"time"
)

func ElapsedTime(s time.Time) string {
	return fmt.Sprintf("%.10f", float64(time.Now().Sub(s).Nanoseconds())/float64(1000000000))
}
