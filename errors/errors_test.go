package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {

	msg := "hello"
	err := New(msg)

	custErr, ok := err.(*container)
	if !ok {
		t.Error("Type assertion of custom error failed.")
	}

	if custErr.err == nil {
		t.Error("Call to New didn't create an error.")
	}
	if err.Error() != msg {
		t.Error("Call to New didn't create an error with the supplied message.")
	}
	if len(custErr.comments) > 0 {
		t.Error("Call to New resulted in err with comments.")
	}
	if len(custErr.stack) == 0 {
		t.Error("Call to New didn't produce a stack.")
	}
}

func TestNewF(t *testing.T) {

	msg := "hello %s"
	arg := "awooo"
	err := NewF(msg, arg)

	if err.Error() != "hello awooo" {
		t.Error("Error message incorrectly formatted.")
	}
}

func TestComment(t *testing.T) {

	msg := "hello"
	com := "yoo"

	cases := []struct {
		err     error
		wantNil bool
	}{
		{Comment(nil, com), true},
		{Comment(errors.New(msg), com), false}, // Adding comment to standard error.
		{Comment(New(msg), com), false},        // Adding comment to custom error.
	}

	for _, c := range cases {

		err := c.err

		if err == nil {
			if c.wantNil {
				continue
			}
			t.Error("Unexpected nil error.")
		}

		custErr, ok := err.(*container)
		if !ok {
			t.Error("Type assertion of custom error failed.")
		}

		if len(custErr.stack) == 0 {
			t.Error("No stack.")
		}
		if len(custErr.comments) != 1 {
			t.Error("Incorrect number of comments.")
		}
		if custErr.comments[0] != com {
			t.Error("Incorrect comment.")
		}
		if err.Error() != "yoo: hello" {
			t.Error("Incorrect error string.")
		}
	}
}

func TestCommentF(t *testing.T) {

	msg := "hello"
	com := "yoo %s"
	arg := "awooo"
	err := CommentF(errors.New(msg), com, arg)

	if err.Error() != "yoo awooo: hello" {
		t.Error("Error message incorrectly formatted.")
	}
}

func TestStack(t *testing.T) {

	msg := "hello"
	com := "yoo"

	cases := []struct {
		err     error
		wantNil bool
	}{
		{AddStack(nil), true},                            // Add stack to nil
		{AddStack(errors.New(msg)), false},               // Add stack to standard error.
		{AddStack(New(msg)), false},                      // Add stack to custom error.
		{AddStack(Comment(errors.New(msg), com)), false}, // Add stack to commented standard error.
		{AddStack(Comment(New(msg), com)), false},        // Add stack to commented custom error.
	}

	for _, c := range cases {

		err := c.err

		if err == nil {
			if c.wantNil {
				continue
			}
			t.Error("Unexpected nil error.")
		}

		custErr, ok := err.(*container)
		if !ok {
			t.Error("Type assertion of custom error failed.")
		}

		if len(custErr.stack) == 0 {
			t.Error("No stack.")
		}
		if len(custErr.comments) > 0 && custErr.comments[0] != com {
			t.Error("Incorrect comment.")
		}
		if len(custErr.comments) > 0 && err.Error() != "yoo: hello" {
			t.Error("Incorrect error string.")
		}
		if len(custErr.comments) == 0 && err.Error() != "hello" {
			t.Error("Incorrect error string.")
		}
	}
}

func TestFormat(t *testing.T) {

	fErr := "Error incorrectly formatted."

	errStr := fmt.Sprint(New("hello"))
	if !strings.Contains(errStr, "Error: ") {
		t.Error(fErr)
	}
	if !strings.Contains(errStr, "\n") {
		t.Error(fErr)
	}

	fs := fmt.Sprintf("%s", New("hello"))
	if fs != "hello" {
		t.Error(fErr)
	}

	fq := fmt.Sprintf("%q", New("hello"))
	if fq != `"hello"` {
		t.Error(fErr)
	}
}

func TestCause(t *testing.T) {

	err := errors.New("hello")
	cause := Cause(err)
	if cause != err {
		t.Error("Error returned from Cause not equal to original error.")
	}

	err = Cause(nil)
	if err != nil {
		t.Error("Expected nil return from Cause after passing nil.")
	}

	err = New("hello")
	custErr, ok := err.(*container)
	if !ok {
		t.Error("Type assertion failed for custom error.")
	}

	err = Cause(err)
	if err != custErr.err {
		t.Error("Error returned from Cause not equal to original error.")
	}
}
