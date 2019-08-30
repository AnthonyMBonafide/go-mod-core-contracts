/*
 *****************************************************************************
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
 ******************************************************************************
 */

package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var ResponseMap = map[models.Code]int{
	models.KindUnknown:                 http.StatusInternalServerError,
	models.KindDatabaseError:           http.StatusInternalServerError,
	models.KindServerError:             http.StatusInternalServerError,
	models.KindCommunicationError:      http.StatusInternalServerError,
	models.KindEntityDoesNotExistError: http.StatusNotFound,
	models.KindEntityStateError:        http.StatusConflict,
	models.KindLimitExceeded:           http.StatusRequestEntityTooLarge,
}

// ToHttpResponse determines the correct HTTP response code for the given error and responds to the HTTP request.
// The decoder argument is a function which marshals an object to the desired media type(i.e. JSON, CBOR, XML, etc).
// This makes this higher-order function more flexible as it can handle responding to an HTTP request with any content
// type.
func ToHttpResponse(e error, w http.ResponseWriter, decoder func(interface{}) ([]byte, error)) {
	kind := Kind(e)

	var ce models.CommonEdgexError
	ok := errors.As(e, &ce)

	// Not an EdgeX error nor does not contain an EdgeX error in the chain.
	if !ok {
		// Treat the error as it were a 500 since we cannot determine the category.
		w.WriteHeader(http.StatusInternalServerError)
		message, err := decoder(e)
		if err != nil {
			_, _ = w.Write(message)
			return
		} else {
			// TODO(Anthony) make this into a valid content type.
			_, _ = w.Write([]byte("Unknown error"))
		}
	}

	statusCode, ok := ResponseMap[models.Kind(ce)]
	message, err := decoder(ce)
	w.WriteHeader(statusCode)

	if err != nil {
		// TODO(Anthony) make this into a valid content type.
		_, _ = w.Write([]byte("Unable to process error"))
	} else {
		_, _ = w.Write(message)
	}
}

// JsonDecoder marshals an object into a byte slice.
// This is a convenience function which can be used as an argument to the higher-order function 'ToHttpResponse' as a
// decoder.
func JsonDecoder(e interface{}) ([]byte, error) {
	return json.Marshal(e)
}

// FromServiceClientError constructs a *CommonEdgexError from a *ErrServiceClient.
func FromServiceClientError(esc *types.ErrServiceClient) models.EdgexError {
	body := strings.Split(esc.Error(), "-")

	var e models.CommonEdgexError
	err := json.Unmarshal([]byte(body[1]), &e)
	if err != nil {
		return models.NewCommonEdgexError(models.KindServerError, "Client error", err)
	}

	return e
}
