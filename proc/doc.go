// Package proc defines a lightweight primitive called a "process" and
// provides functions that compose processes to allow common concurrency
// patterns to be expressed conveniently.
//
// A "process" is, fundamentally, just the time spent executing a function
// with a particular signature. A process runs in a Context, blocks until
// it completes (or is cancelled), and may return an error.
//
// The implementation of a process is a function of type proc.Impl, though
// implementors of processes need not import this package just to define
// such a function.
//
// As well as some helpers that act as combinators for processors, this
// package also contains helpers to easily use processes in conjunction with
// the Task and Region concepts from the runtime/trace package.
package proc
