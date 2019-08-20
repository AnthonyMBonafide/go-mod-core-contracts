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
	"net/http"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var HttpReponseMap = map[string]int{
	models.KindDatabaseError:           http.StatusInternalServerError,
	models.KindServerError:             http.StatusInternalServerError,
	models.KindCommunicationError:      http.StatusInternalServerError,
	models.KindEntityDoesNotExistError: http.StatusNotFound,
	models.KindEntityStateError:        http.StatusConflict,
	models.KindLimitExceeded:           http.StatusRequestEntityTooLarge,
}

func ToHttpResponse(e models.EdgexError, w http.ResponseWriter, decoder func(interface{}) ([]byte, error)) error {
	// TODO(Anthony) handle situations where `e` does not have a Kind specified.
	b, err := decoder(e)
	statusCode, ok := HttpReponseMap[e.Kind()]
	if !ok || err != nil {
		// Treat the error as it were a 500 since we cannot determine the category.
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(statusCode)

	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func JsonDecoder(e interface{}) ([]byte, error) {
	return json.Marshal(e)
}

// FromServiceClientError constructs a *CommonEdgexError from a *ErrServiceClient.
func FromServiceClientError(esc *types.ErrServiceClient) *models.CommonEdgexError {
	body := strings.Split(esc.Error(), "-")

	var e models.CommonEdgexError
	err := json.Unmarshal([]byte(body[1]), &e)
	if err != nil {
		return models.NewCommonEdgexError([]string{"FromServiceClientError"}, models.KindServerError, err.Error())
	}
	return &e

}
