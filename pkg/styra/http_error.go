/*
Copyright (C) 2023 Bankdata (bankdata@bankdata.dk)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package styra

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// HTTPError represents an error that occurred when interacting with the Styra
// API.
type HTTPError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface.
func (styraerror *HTTPError) Error() string {
	return fmt.Sprintf("styra: unexpected statuscode: %d, body: %s", styraerror.StatusCode, styraerror.Body)
}

// NewHTTPError creates a new HTTPError based on the statuscode and body from a
// failed call to the Styra API.
func NewHTTPError(statuscode int, body string) error {
	styraerror := &HTTPError{
		StatusCode: statuscode,
	}

	if isValidJSON(body) {
		styraerror.Body = body
	} else {
		styraerror.Body = "invalid JSON response"
	}

	return errors.WithStack(styraerror)
}

func isValidJSON(data string) bool {
	var out interface{}
	return json.Unmarshal([]byte(data), &out) == nil
}
