/*
Copyright (C) 2025 Bankdata (bankdata@bankdata.dk)

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

// Package errors contains errors.
package errors

import (
	"github.com/pkg/errors"

	"github.com/bankdata/styra-controller/api/styra/v1beta1"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ReconcilerErr defines an error that occurs during reconciliation.
type ReconcilerErr struct {
	err           error
	Event         string
	ConditionType string
}

// New returns a new RenconcilerErr.
func New(msg string) *ReconcilerErr {
	return Wrap(errors.New(msg), "")
}

// Wrap wraps an error as a RenconcilerErr.
func Wrap(err error, msg string) *ReconcilerErr {
	return &ReconcilerErr{
		err: errors.Wrap(err, msg),
	}
}

// WithEvent adds event metadata to the ReconcilerErr.
func (err *ReconcilerErr) WithEvent(event v1beta1.EventType) *ReconcilerErr {
	err.Event = string(event)
	return err
}

// WithSystemCondition adds condition metadata to the ReconcilerErr.
func (err *ReconcilerErr) WithSystemCondition(contype v1beta1.ConditionType) *ReconcilerErr {
	err.ConditionType = string(contype)
	return err
}

// Error implements the error interface.
func (err *ReconcilerErr) Error() string {
	return err.err.Error()
}

// Cause returns the cause of the error.
func (err *ReconcilerErr) Cause() error {
	return err.err
}

// Unwrap is the same as `Cause()`
func (err *ReconcilerErr) Unwrap() error {
	return err.Cause()
}

// StackTrace implements the stackTracer interface.
func (err *ReconcilerErr) StackTrace() errors.StackTrace {
	var st stackTracer
	if errors.As(err.err, &st) {
		return st.StackTrace()
	}
	return nil
}
