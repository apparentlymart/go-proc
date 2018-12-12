package proc

import (
	"context"
)

// Impl is the type for a process's implementation function.
//
// Long-running processes should watch for the given Context signalling "done"
// and exit as soon as possible.
//
// A process implementation can return an error if it is unable to complete
// its task successfully. Use normal Go error patterns. Some of the process
// combinators in this package have special behaviors for when a process
// returns an error.
type Impl func(ctx context.Context) error
