package proc

import "context"

// Sequence returns a process implementation that runs all of the given
// implementations in sequence, returning early if any of them return an
// error.
//
// If an error is returned, it's the error from the first process that
// failed, and any subsequent implementations were not started at all.
// If no error is returned then all of the processes ran to completion
// successfully.
func Sequence(impls ...Impl) Impl {
	return func(ctx context.Context) error {
		for _, impl := range impls {
			err := impl(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
