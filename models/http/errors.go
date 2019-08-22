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
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"net/http"
)

var HttpReponseMap = map[models.Code]int{
	models.KindDatabaseError:           http.StatusInternalServerError,
	models.KindServerError:             http.StatusInternalServerError,
	models.KindCommunicationError:      http.StatusInternalServerError,
	models.KindEntityDoesNotExistError: http.StatusNotFound,
	models.KindEntityStateError:        http.StatusConflict,
	models.KindLimitExceeded:           http.StatusRequestEntityTooLarge,
}

func ToHttpResponse(e error, w http.ResponseWriter) {
	// TODO(Anthony) handle situations where `e` does not have a Kind specified.
	kind := models.Kind(e)

	statusCode, ok := HttpReponseMap[kind]
	if !ok {
		// Treat the error as it were a 500 since we cannot determine the category.
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(statusCode)

	}
}

func JsonDecoder(e interface{}) ([]byte, error) {
	return json.Marshal(e)
}
