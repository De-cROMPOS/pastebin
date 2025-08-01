package errorhandler

import "errors"

var (
    ErrRecordNotFound = errors.New("record not found")
)