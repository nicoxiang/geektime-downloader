package geektime

import "fmt"

// ErrGeekTimeAPIBadCode ...
type ErrGeekTimeAPIBadCode struct {
    Path string
    Code int
    Msg string
}

// Error implements error interface
func (e ErrGeekTimeAPIBadCode) Error() string {
    return fmt.Sprintf("make geektime api call %s failed, code %d, msg %s", e.Path, e.Code, e.Msg)
}