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
	"fmt"
	"strings"
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

const (
	// TODO(Anthony) should this be a Kind as it can leak harmful information to the client about the system?
	KindDatabaseError           = "Database"
	KindCommunicationError      = "Communication"
	KindEntityDoesNotExistError = "NotFound"
	KindEntityStateError        = "InvalidState"
	KindServerError             = "Unknown/Unexpected"
	KindLimitExceeded           = "LimitExceeded"
)

// TODO(Anthony) see if we can leverage the error wrapping functionality in Go 1.13
// TODO(Anthony) see how we can easily incorporate this into EdgeX, possibly update pre-existing error types so that
//   the implement these methods and can be returned as an EdgexError type.
// TODO(Anthony) update to work with the sem-structured logging currently in place.
// EdgexError provides functionality to which all internal EdgeX errors must adhere.
type EdgexError interface {
	// AddOps pre-appends an operation to the ordered list of operations affected by this error.
	AddOps(ops ...string)
	// Op constructs a string representation of the operation(s) which the error was either created or observed.
	Op() string
	// Kind obtains the category of the error which can be used to determine how to handle this error at a higher level.
	Kind() string
	// Error obtains the error message associated with the error.
	Error() string
	// String creates a string representation of the error to be used for logging in human readable format.
	String() string
}

// CommonEdgexError generalizes an error structure which can be used for any type of EdgeX error.
type CommonEdgexError struct {
	// ops contains a FIFO list of operations affected by the error.
	ops []string // This is omitted from marshaling for security reasons. We do not want to inform the client about any implementation details.
	// Category contains information regarding the high level error type.
	Category string `json:"category"`
	// Message contains detailed information about the error. Security sensitive information should NOT be contained
	// here as the EdgexError struct is meant to absract the details of the error in a generalized fashion. This means
	// that EdgexErrors which are passed to calling functions could be either logged or marshaled and sent to the
	// client.
	Message string `json:"message"`
}

func (ce *CommonEdgexError) AddOps(ops ...string) {
	if ce.ops == nil {
		ce.ops = ops
		return
	}

	ce.ops = append(ce.ops, ops...)
}

func (ce CommonEdgexError) Op() string {
	return fmt.Sprintf("[%s]", strings.Join(ce.ops, " -> "))
}

func (ce CommonEdgexError) Kind() string {
	return ce.Category
}

func (ce CommonEdgexError) Error() string {
	return ce.Message
}

func (ce CommonEdgexError) String() string {
	return fmt.Sprintf("Operations: %s \n Category: %s \n Message: %s", ce.Op(), ce.Category, ce.Message)
}

// NewCommonEdgexError creates a new CommonEdgexError with the information provided.
func NewCommonEdgexError(ops []string, kind string, message string) *CommonEdgexError {
	return &CommonEdgexError{
		ops:      ops,
		Category: kind,
		Message:  message,
	}
}
