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
	"context"
	"net/http"
	"path"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrlpred "sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	configv2alpha1 "github.com/bankdata/styra-controller/api/config/v2alpha1"
	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	ctrlerr "github.com/bankdata/styra-controller/internal/errors"
	"github.com/bankdata/styra-controller/internal/fields"
	"github.com/bankdata/styra-controller/internal/predicate"
	"github.com/bankdata/styra-controller/internal/sentry"
	"github.com/bankdata/styra-controller/pkg/styra"
)

// GlobalDatasourceReconciler reconciles a GlobalDatasource object.
type GlobalDatasourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Styra  styra.ClientInterface
	Config *configv2alpha1.ProjectConfig
}

//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=globaldatasources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=globaldatasources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=globaldatasources/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile implements renconcile.Renconciler and has responsibility of
// ensuring that the current state of the GlobalDatasource resource renconciled
// towards the desired state.
func (r *GlobalDatasourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciliation begins")

	var gds styrav1alpha1.GlobalDatasource
	if err := r.Get(ctx, req.NamespacedName, &gds); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Could not find GlobalDatasource")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrap(err, "Could not get GlobalDatasource")
	}

	if gds.Spec.CredentialsSecretRef != nil {
		log.Info("Reconciling git credentials from kubernetes secret")
		sr := gds.Spec.CredentialsSecretRef
		var s corev1.Secret
		if err := r.Get(ctx, types.NamespacedName{Namespace: sr.Namespace, Name: sr.Name}, &s); err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not get git credential secret")
		}
		if len(s.Data["name"]) == 0 {
			return ctrl.Result{}, ctrlerr.New("Key `name` is required in git credential secret")
		}
		if len(s.Data["secret"]) == 0 {
			return ctrl.Result{}, ctrlerr.New("Key `secret` is required in git credential secret")
		}
		_, err := r.Styra.CreateUpdateSecret(
			ctx,
			path.Join("libraries/global", gds.Name, "git"),
			&styra.CreateUpdateSecretsRequest{
				Name:   string(s.Data["name"]),
				Secret: string(s.Data["secret"]),
			},
		)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not update Styra secret")
		}
	} else {
		log.Info("Reconciling git credentials from default credentials")
		gitCredential := r.Config.GetGitCredentialForRepo(gds.Spec.URL)
		if gitCredential == nil {
			log.Info("Could not find matching credentials", "url", gds.Spec.URL)
		} else {
			_, err := r.Styra.CreateUpdateSecret(
				ctx,
				path.Join("libraries/global", gds.Name, "git"),
				&styra.CreateUpdateSecretsRequest{
					Name:   gitCredential.User,
					Secret: gitCredential.Password,
				},
			)
			if err != nil {
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not update Styra secret")
			}
		}
	}

	log.Info("Reconciling datasource")
	update := false
	dsName := path.Join("global", gds.Name)
	gdsr, err := r.Styra.GetDatasource(ctx, dsName)
	if err != nil {
		var herr *styra.HTTPError
		if errors.As(err, &herr) {
			if herr.StatusCode == http.StatusNotFound {
				update = true
			}
		}
	} else if r.needsUpdate(&gds, gdsr.DatasourceConfig) {
		update = true
	}
	if update {
		log.V(1).Info("UpsertDatasource")
		_, err := r.Styra.UpsertDatasource(ctx, dsName, r.specToUpdate(&gds))
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not upsert datasource")
		}
	}

	log.Info("Reconciliation succeeded")
	return ctrl.Result{}, nil
}

func (r *GlobalDatasourceReconciler) specToUpdate(gds *styrav1alpha1.GlobalDatasource) *styra.UpsertDatasourceRequest {
	if gds == nil {
		return nil
	}
	s := gds.Spec
	req := &styra.UpsertDatasourceRequest{
		Category:    string(s.Category),
		Description: s.Description,
		Commit:      s.Commit,
		Reference:   s.Reference,
		URL:         s.URL,
		Path:        s.Path,
	}
	if s.Enabled == nil || *s.Enabled {
		req.Enabled = true
	}

	credentials := path.Join("libraries/global", gds.Name, "git")
	if s.CredentialsSecretRef != nil {
		req.Credentials = credentials
	} else if r.Config.GetGitCredentialForRepo(gds.Spec.URL) != nil {
		req.Credentials = credentials
	}

	return req
}

func (r *GlobalDatasourceReconciler) needsUpdate(gds *styrav1alpha1.GlobalDatasource, dc *styra.DatasourceConfig) bool {
	if gds == nil {
		return false
	}
	s := gds.Spec
	if dc == nil ||
		string(s.Category) != dc.Category ||
		s.Description != dc.Description ||
		s.Commit != dc.Commit ||
		// treat nil as true because it is the default in the webhook
		(s.Enabled == nil || *s.Enabled) != dc.Enabled ||
		s.Reference != dc.Reference ||
		s.URL != dc.URL ||
		s.Path != dc.Path {
		return true
	}

	if dc.Credentials != path.Join("libraries/global", gds.Name, "git") {
		if s.CredentialsSecretRef != nil {
			return true
		}
		if r.Config.GetGitCredentialForRepo(gds.Spec.URL) != nil {
			return true
		}
	}
	return false
}

// SetupWithManager registers the the GlobalDatasource controller with the
// Manager.
func (r *GlobalDatasourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := setupFieldIndexer(mgr.GetFieldIndexer()); err != nil {
		return err
	}

	p, err := predicate.ControllerClass(r.Config.ControllerClass)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&styrav1alpha1.GlobalDatasource{}, builder.WithPredicates(p)).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.findGlobalDatasourcesForSecret),
			builder.WithPredicates(ctrlpred.ResourceVersionChangedPredicate{}),
		).
		Complete(sentry.Decorate(r))
}

func setupFieldIndexer(indexer client.FieldIndexer) error {
	ctx := context.Background()

	if err := indexer.IndexField(
		ctx,
		&styrav1alpha1.GlobalDatasource{},
		fields.GlobalDatasourceSecretRefNamespace,
		func(o client.Object) []string {
			if gds, ok := o.(*styrav1alpha1.GlobalDatasource); ok {
				if gds.Spec.CredentialsSecretRef != nil {
					return []string{gds.Spec.CredentialsSecretRef.Namespace}
				}
			}
			return nil
		},
	); err != nil {
		return errors.Wrapf(err, "could not setup index for %s", fields.GlobalDatasourceSecretRefNamespace)
	}

	if err := indexer.IndexField(
		ctx,
		&styrav1alpha1.GlobalDatasource{},
		fields.GlobalDatasourceSecretRefName,
		func(o client.Object) []string {
			if gds, ok := o.(*styrav1alpha1.GlobalDatasource); ok {
				if gds.Spec.CredentialsSecretRef != nil {
					return []string{gds.Spec.CredentialsSecretRef.Name}
				}
			}
			return nil
		},
	); err != nil {
		return errors.Wrapf(err, "could not setup index for %s", fields.GlobalDatasourceSecretRefName)
	}

	return nil
}

func (r *GlobalDatasourceReconciler) findGlobalDatasourcesForSecret(secret client.Object) []reconcile.Request {
	var gdsl styrav1alpha1.GlobalDatasourceList
	if err := r.List(context.Background(), &gdsl, &client.ListOptions{
		FieldSelector: fields.GlobalDatasourceCredentialsSecretRefFieldSelector(secret.GetNamespace(), secret.GetName()),
	}); err != nil {
		return nil
	}

	reqs := make([]reconcile.Request, len(gdsl.Items))
	for i, gds := range gdsl.Items {
		reqs[i] = reconcile.Request{NamespacedName: types.NamespacedName{Name: gds.Name}}
	}
	return reqs
}
