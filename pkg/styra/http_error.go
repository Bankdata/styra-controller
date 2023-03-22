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
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// HTTPError represents an error that occurred when interacting with the Styra
// API.
type HTTPError struct {
	StatusCode int
	Body       string
	Message    string `yaml:"message,omitempty"`
}

// Error implements the error interface.
func (styraerror *HTTPError) Error() string {
	return fmt.Sprintf("styra: unexpected statuscode: %d, body: %s", styraerror.StatusCode, styraerror.Body)
}

// NewHTTPError creates a new HTTPError based on the statuscode and body from a
// failed call to the Styra API.
func NewHTTPError(statuscode int, body string) error {
	err := &HTTPError{
		StatusCode: statuscode,
		Body:       body,
	}

	if err := yaml.Unmarshal([]byte(body), &err); err != nil {
		return errors.Wrap(err, "could not unmarshal error body")
	}

	return errors.WithStack(err)
}
