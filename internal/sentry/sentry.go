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

// Package sentry contains a reconciler middleware which sends errors to
// Sentry.
package sentry

import (
	"context"
	"errors"
	"strings"

	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/getsentry/sentry-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type sentryReconciler struct {
	next reconcile.Reconciler
}

// Reconcile implements reconciler.Reconcile for the sentry middleware.
func (r *sentryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result, err := r.next.Reconcile(ctx, req)

	if sentry.CurrentHub().Client() != nil {
		if err != nil {
			hub := sentry.CurrentHub().Clone()
			var styraerror *httperror.HTTPError
			if errors.As(err, &styraerror) {
				hub.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetContext("Styra Client", map[string]interface{}{
						"body":       styraerror.Body,
						"statuscode": styraerror.StatusCode,
					})
				})
			}

			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTags(map[string]string{
					"namespace": req.Namespace,
					"name":      req.Name,
				})
			})
			if !isUserError(err.Error()) {
				hub.CaptureException(err)
			}
		}
	}
	return result, err
}

// Decorate applies the sentry middleware to the given reconcile.Reconciler.
func Decorate(r reconcile.Reconciler) reconcile.Reconciler {
	return &sentryReconciler{next: r}
}

func isUserError(msg string) bool {
	uniqueGitConfig := "the combination of url, branch, commit-sha and path must be unique across all git repos"
	couldNotFindCredentialsSecret := "Could not find credentials Secret"

	return strings.Contains(msg, uniqueGitConfig) ||
		strings.Contains(msg, couldNotFindCredentialsSecret)
}
