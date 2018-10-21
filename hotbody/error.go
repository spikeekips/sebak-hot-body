package hotbody

import "fmt"

type ErrorStopRunning struct {
	msg string
}

func NewErrorStopRunning(msg string, a ...interface{}) *ErrorStopRunning {
	return &ErrorStopRunning{msg: fmt.Sprintf(msg, a...)}
}

func (e *ErrorStopRunning) Error() string {
	return e.msg
}
