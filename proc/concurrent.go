package proc

import (
	"context"
	"sync"
)

// Concurrent returns a process implementation that runs concurrently
// all of the given process implementations, blocking until they all
// complete, and returning any errors.
//
// If more than one process returns an error, the result is a MultiError
// containing all of them in an undefined order.
func Concurrent(impls ...Impl) Impl {
	return concurrent(false, impls...)
}

// ConcurrentGroup is similar to concurrent but treats the given implementations
// as a cooperating group: if any of the processes returns an error then it
// will signal all of the others to cancel (via their contexts) and wait
// for them to complete before returning.
//
// ConcurrentGroup also cancels process contexts before returning on the
// happy path (no errors), so that all child processes are guaranteed cancelled
// before this function returns. (There may still be goroutines spawned from
// those processes that do not respond to cancellation, though.)
//
// This could be useful for managing a set of long-running processes at the
// top level of a program where the program should bail out quickly if any
// of the processes fail. In that situation, process implementations should
// be designed to recover from problems where possible and only return errors
// if they encounter unrecoverable problems that prevent continued system
// operation.
func ConcurrentGroup(impls ...Impl) Impl {
	return concurrent(true, impls...)
}

func concurrent(group bool, impls ...Impl) Impl {
	return func(ctx context.Context) error {
		chiCtx := ctx
		var cancel func()
		if group {
			chiCtx, cancel = context.WithCancel(ctx)
		}

		var errMut sync.Mutex
		var allErr error
		var wg sync.WaitGroup
		wg.Add(len(impls))
		for _, impl := range impls {
			go func(impl Impl) {
				err := impl(chiCtx)
				if err != nil {
					errMut.Lock()
					allErr = appendErrs(allErr, err)
					errMut.Unlock()
					if cancel != nil {
						cancel()
					}
				}
				wg.Done()
			}(impl)
		}
		wg.Wait()

		// Everything should've finished by the time we get here by
		// definition, but we'll cancel here if we're in group mode
		// just to clean up any in-flight stuff the processes might have
		// left running when they returned.
		if cancel != nil {
			cancel()
		}

		return allErr
	}
}
