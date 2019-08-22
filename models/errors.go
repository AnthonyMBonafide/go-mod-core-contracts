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

type Code string

const (
	// TODO(Anthony) should this be a Kind as it can leak harmful information to the client about the system?
	KindUnknown                 Code = "Unknown"
	KindDatabaseError           Code = "Database"
	KindCommunicationError      Code = "Communication"
	KindEntityDoesNotExistError Code = "NotFound"
	KindEntityStateError        Code = "InvalidState"
	KindServerError             Code = "Unknown/Unexpected"
	KindLimitExceeded           Code = "LimitExceeded"
)

// TODO(Anthony) see if we can leverage the error wrapping functionality in Go 1.13
// TODO(Anthony) see how we can easily incorporate this into EdgeX, possibly update pre-existing error types so that
//   the implement these methods and can be returned as an EdgexError type.
// TODO(Anthony) update to work with the sem-structured logging currently in place.

// CommonEdgexError generalizes an error structure which can be used for any type of EdgeX error.
type CommonEdgexError struct {
	// ops contains a FIFO list of operations affected by the error.
	op string // This is omitted from marshaling for security reasons. We do not want to inform the client about any implementation details.
	// Category contains information regarding the high level error type.
	Kind Code `json:"category"`
	// Message contains detailed information about the error. Security sensitive information should NOT be contained
	// here as the EdgexError struct is meant to absract the details of the error in a generalized fashion. This means
	// that EdgexErrors which are passed to calling functions could be either logged or marshaled and sent to the
	// client.
	Err     error
}

func Ops(ce CommonEdgexError) []string {
	res := []string{ce.op}

	subErr, ok := ce.Err.(CommonEdgexError)
	if !ok {
		return res
	}

	// recursively return Ops as long as we have CommonEdgexErrors to work through
	res = append(res, Ops(subErr)...)

	return res
}

func Kind(err error) Code {
	e, ok := err.(CommonEdgexError)
	if !ok {
		return KindUnknown
	}

	// We want to return the first "Kind" we see that isn't Unknown, because
	// the higher in the stack the Kind was specified the more context we had.
	if e.Kind != KindUnknown {
		return e.Kind
	}

	return Kind(e.Err)
}

func (ce CommonEdgexError) Error() string {
	return ce.Err.Error()
}

// NewCommonEdgexError creates a new CommonEdgexError with the information provided.
func NewCommonEdgexError(args ...interface{}) CommonEdgexError {
	e := CommonEdgexError{Kind: KindUnknown}
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			e.op = arg
		case Code:
			e.Kind = arg
		case error:
			e.Err = arg
		default:
			panic("bad call to E")
		}
	}

	return e
}
