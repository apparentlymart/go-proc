package proc

import (
	"context"
	"fmt"
	"runtime/trace"
)

// TraceRegion creates a process implementation that runs the given
// implementation in the context of a Region from the runtime/trace package.
func TraceRegion(regionType string, impl Impl) Impl {
	return func(ctx context.Context) error {
		region := trace.StartRegion(ctx, regionType)
		err := impl(ctx)
		region.End()
		return err
	}
}

// Task wraps a process implementation and gives it two new behaviors:
//
// Firstly, it creates a new Task from the runtime/trace package and runs
// the given implementation in its context.
//
// Secondly, if the process returns any error then it will be wrapped in
// a TaskError before returning it, allowing a caller that is using one
// of the combinator functions from this package to recognize the type
// of task that failed.
func Task(taskType string, impl Impl) Impl {
	return func(ctx context.Context) error {
		cctx, task := trace.NewTask(ctx, taskType)
		defer task.End()
		err := impl(cctx)
		if err != nil {
			return TaskError{taskType, err}
		}
		return nil
	}
}

// TaskError is an error type that wraps another error and annotates it with
// a task type name. This error type is returned from process implementations
// created with function Task.
type TaskError struct {
	TaskType string
	Err      error
}

func (err TaskError) Error() string {
	return fmt.Sprintf("%s: %s", err.TaskType, err.Err)
}

// Cause returns the error that was returned by the task's underlying
// process implementation.
func (err TaskError) Cause() error {
	return err.Err
}
