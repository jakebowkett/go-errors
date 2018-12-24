/*
Package errors adds context to errors while preserving
the original.

The package shadows the standard library's errors.New and
returns a custom error that has its own stack trace beginning
at the call site.

	err := errors.New("whoops")
	fmt.Print(err)

The above will output something like this:

	Error: whoops
	  │
	  ├─ (package.function)
	  │     C:/path/to/package/of/origin/file.go:12
	  │
	  ├─ (package.function)
	  │     C:/path/to/package/of/caller/file.go:36
	  │
	  └─ (package.function)
	        C:/path/to/package/of/callers/caller/file.go:36

Existing errors can be prefixed with additional context by calling Prefix.
If the error supplied was created by this package its message will be
prefixed with the supplied string. If the error is from the standard
library or some other source it will be prefixed and have a stack added to
it.

	err := errors.New("whoops")
	err = errors.Prefix(err, "oh no")
	fmt.Print(err)

Output:

	Error: oh no: whoops
	  │
	  ├─ (package.function)
	  │     C:/path/to/package/of/origin/file.go:12
	  │
	  ├─ (package.function)
	  │     C:/path/to/package/of/caller/file.go:36
	  │
	  └─ (package.function)
	        C:/path/to/package/of/callers/caller/file.go:36


You can pass errors from the standard library and other sources to
functions in this package and they will be preserved and retrievable:

	// Let's assume this produces an error.
	f, err := ioutil.ReadFile(path)
	err = errors.AddStack(err)

	// Prints the error and stack trace as in above examples.
	fmt.Print(err)

	// Retrieve and print only the original error.
	err = errors.Cause(err)
	fmt.Print(err) // prints "file does not exist"

Less verbose forms of printing are possible:

	err := errors.New("whoops")
	fmt.Print("%s")
	fmt.Print("%q")

These will respectively print:

	Error: whoops
	"Error: whoops"

If the error passed to a function is nil it will return nil. Errors
that already have a stack will not have it replaced by calling AddStack
or Prefix on them.

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
	prefixes []string
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
Prefix takes an error and annotates it with prefix to
give more context, It also adds a stack trace from the point
it was called if one doesn't already exist. The original
error message is preserved and can be retrieved with Cause.
Returns nil if err is nil.
*/
func Prefix(err error, prefix string) error {
	return addPrefix(err, prefix)
}

/*
PrefixF is the same as Prefix and formats the prefix
according to format.
*/
func PrefixF(err error, format string, a ...interface{}) error {
	return addPrefix(err, fmt.Sprintf(format, a...))
}

func addPrefix(err error, prefix string) error {

	if err == nil {
		return nil
	}

	// Standard error.
	custErr, ok := err.(*container)
	if !ok {
		return &container{
			err:      err,
			prefixes: []string{prefix},
			stack:    stack(3),
		}
	}

	// One of ours.
	custErr.prefixes = append(custErr.prefixes, prefix)
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
annotated with prefixes or a stack. Standard errors are returned
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

/*
Equals returns true if the original error value of err1 and err2
is the same. Equivalent to:

	errors.Cause(err1) == errors.Cause(err2)
*/
func Equals(err1, err2 error) bool {
	return Cause(err1) == Cause(err2)
}

func (e *container) Error() string {
	var s string
	for _, p := range e.prefixes {
		s += p + ": "
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

	pc := make([]uintptr, 16)
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
