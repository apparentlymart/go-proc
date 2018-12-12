package proc

import (
	"fmt"
	"strings"
)

// MultiError is an error implementation that wraps a number of other
// errors that occurred together.
type MultiError []error

func (errs MultiError) Error() string {
	if len(errs) == 0 {
		// Degenerate case; a zero-length MultiError is pointless
		return "no errors"
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}
	var buf strings.Builder
	fmt.Fprintf(&buf, "%d errors:", len(errs))
	for _, err := range errs {
		fmt.Fprintf(&buf, "\n- %s", err.Error())
	}
	return buf.String()
}

func appendErrs(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	var ret error
	for _, err := range errs {
		switch tErr := err.(type) {
		case nil:
			// ignore
		case MultiError:
			if tExist, ok := ret.(MultiError); ok {
				ret = append(tExist, tErr...)
			} else if ret != nil {
				new := make(MultiError, len(tErr)+1)
				new[0] = ret
				copy(new[1:], tErr)
			} else {
				ret = tErr
			}
		default:
			if tExist, ok := ret.(MultiError); ok {
				ret = append(tExist, err)
			} else if ret != nil {
				ret = MultiError{ret, err}
			} else {
				ret = err
			}
		}
	}
	return ret
}
