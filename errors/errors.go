/*
Package errors allows adding comments and stack traces
to errors while preserving the original error.
*/
package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type container struct {
	err      error
	comments []string
	stack    []frame
}

type frame struct {
	line     int
	file     string
	function string
}

/*
New returns an error that has its own stack trace.
*/
func New(msg string) error {
	return newErr(msg)
}

/*
NewF returns an error that has its own stack trace and is
formatted according to format.
*/
func NewF(format string, a ...interface{}) error {
	return newErr(fmt.Sprintf(format, a...))
}

func newErr(msg string) error {
	return &container{
		err:   errors.New(msg),
		stack: stack(3),
	}
}

/*
Comment takes an error and annotates it with comment to
give more context and adds a stack trace from the point
it was called if one doesn't already exist. The original
error message is preserved and can be retrieved with Cause.
Returns nil if err is nil.
*/
func Comment(err error, comment string) error {
	return addComment(err, comment)
}

/*
CommentF is the same as Comment and formats the comment
according to format.
*/
func CommentF(err error, format string, a ...interface{}) error {
	return addComment(err, fmt.Sprintf(format, a...))
}

func addComment(err error, comment string) error {

	if err == nil {
		return nil
	}

	// Standard error.
	custErr, ok := err.(*container)
	if !ok {
		return &container{
			err:      err,
			comments: []string{comment},
			stack:    stack(3),
		}
	}

	// One of ours.
	custErr.comments = append(custErr.comments, comment)
	return custErr
}

/*
AddStack takes an existing error and gives it a stack trace.
The original error is preserved and can be retrieved with Cause.
Calling AddStack on an error that already has one will do nothing
and return the original error. Returns nil if err is nil.
*/
func AddStack(err error) error {
	if err == nil {
		return nil
	}
	_, ok := err.(*container)
	if !ok {
		return &container{
			err:   err,
			stack: stack(2),
		}
	}
	return err
}

/*
Cause retrieves the original error if it has been previously
annotated with comments or a stack. Standard errors are returned
as-is. Cause returns nil if err is nil.
*/
func Cause(err error) error {
	if err == nil {
		return nil
	}
	custErr, ok := err.(*container)
	if !ok {
		return err
	}
	return custErr.err
}

func (e *container) Error() string {
	var s string
	for _, c := range e.comments {
		s += c + ": "
	}
	return s + e.err.Error()
}

func (e *container) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		fmt.Fprintf(s, "Error: %s\n  │\n", e.Error())
		for i, f := range e.stack {
			start := "├─ "
			fileStart := "│"
			if i == len(e.stack)-1 {
				start = "└─ "
				fileStart = " "
			}
			fmt.Fprintf(s,
				"  %s(%s)\n"+
					"  %s     %s:%d\n"+
					"  %s\n",
				start, f.function, fileStart, f.file, f.line, fileStart)
		}
	case 's':
		fmt.Fprint(s, e.err.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.err.Error())
	}
}

func stack(skip int) []frame {

	pc := make([]uintptr, 32)
	n := runtime.Callers(1, pc)
	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	var stack []frame

	for {

		f, more := frames.Next()

		if strings.Contains(f.File, "runtime/") {
			break
		}

		stack = append(stack, frame{
			file:     f.File,
			line:     f.Line,
			function: f.Function,
		})

		if !more {
			break
		}
	}

	// We skip these frames because they're
	// calls internal to this package.
	if len(stack) < skip {
		return stack[0:0]
	}
	return stack[skip:]
}
