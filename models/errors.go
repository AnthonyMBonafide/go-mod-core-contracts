/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package models

import (
	"errors"
	"fmt"
	"runtime"
)

// ErrContractInvalid is a specific error type for handling model validation failures. Type checking within
// the calling application will facilitate more explicit error handling whereby it's clear that validation
// has failed as opposed to something unexpected happening.
type ErrContractInvalid struct {
	errMsg string
}

// NewErrContractInvalid returns an instance of the error interface with ErrContractInvalid as its implementation.
func NewErrContractInvalid(message string) error {
	return ErrContractInvalid{errMsg: message}
}

// Error fulfills the error interface and returns an error message assembled from the state of ErrContractInvalid.
func (e ErrContractInvalid) Error() string {
	return e.errMsg
}

// Code a categorical identifier used to give high-level insight as to the error type.
type Code string

const (
	// Constant Kind identifiers which can be used to label and group errors.
	KindUnknown                 Code = "Unknown"
	KindDatabaseError           Code = "Database"
	KindCommunicationError      Code = "Communication"
	KindEntityDoesNotExistError Code = "NotFound"
	KindEntityStateError        Code = "InvalidState"
	KindServerError             Code = "Unknown/Unexpected"
	KindLimitExceeded           Code = "LimitExceeded"
)

// EdgexError provides an abstraction for all internal EdgeX errors.
// This exists so that we can use this type in our method signatures and return nil which will fit better with the way
// the Go builtin errors are normally handled.
type EdgexError interface {
	// Error obtains the error message associated with the error.
	Error() string
}

// TODO(Anthony) update to work with the sem-structured logging currently in place.
// CommonEdgexError generalizes an error structure which can be used for any type of EdgeX error.
type CommonEdgexError struct {
	op string
	// Category contains information regarding the high level error type.
	Kind Code `json:"kind"`
	// Message contains detailed information about the error.
	Message string `json:"message"`
	// err is a nested error which is used to form a chain of errors for better context.
	err error
}

// Kind determines the Kind associated with an error by inspecting the chain of errors. The top-most matching Kind is
// returned or KinkUnknown if no Kind can be determined.
func Kind(err error) Code {
	var e CommonEdgexError
	if !errors.As(err, &e) {
		return KindUnknown
	}

	return e.Kind
}

// Error creates an error message taking all nested and wrapped errors into account.
func (ce CommonEdgexError) Error() string {
	if ce.err == nil {
		return ce.Message
	}

	// ce.err.Error functionality gets the error message of the nested error and which will handle both CommonEdgexError
	// types and Go standard errors(both wrapped and non-wrapped).
	return ce.op + " " + ce.Message + ": " + ce.err.Error()
}

// Unwrap retrieves the next nested error in the wrapped error chain.
// This is used by the new wrapping and unwrapping features available in Go 1.13 and aids in traversing the error chain
// of wrapped errors.
func (ce CommonEdgexError) Unwrap() error {
	return ce.err
}

// Is determines if an error is of type CommonEdgexError.
// This is used by the new wrapping and unwrapping features available in Go 1.13 and aids the errors.Is function when
// determining is an error or any error in the wrapped chain contains an error of a particular type.
func (ce CommonEdgexError) Is(err error) bool {
	switch err.(type) {
	case CommonEdgexError:
		return true
	default:
		return false

	}
}

// NewCommonEdgexError creates a new CommonEdgexError with the information provided.
func NewCommonEdgexError(kind Code, message string, wrappedError error) CommonEdgexError {
	return CommonEdgexError{
		Kind:    kind,
		op:      addCallerInformation(),
		Message: message,
		err:     wrappedError,
	}
}

// addCallerInformation generates information about the caller function. This function skips the caller which has
// invoked this function, but rather introspects the calling function 3 frames below this frame in the call stack. This
// function is a helper function which eliminates the need for the 'op' field in the `CommonEdgexError` type and
// providing an 'op' string when creating an 'CommonEdgexError'
func addCallerInformation() string {
	pc := make([]uintptr, 10)
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	// TODO(Anthony) Come up with a better structure as this is too long
	return fmt.Sprintf("[%s]-%s(line %d)\n", file, f.Name(), line)
}
