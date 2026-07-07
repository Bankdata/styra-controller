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

// Package k8sconv contains helpers related to converting data to Kubernetes
// resources.
package k8sconv

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/pkg/ocp"
)

type bearer struct {
	TokenPath string `yaml:"token_path"`
}

type credentials struct {
	Bearer bearer `yaml:"bearer"`
}

type authz struct {
	Service  string `yaml:"service,omitempty"`
	Resource string `yaml:"resource,omitempty"`
	Persist  bool   `yaml:"persist,omitempty"`
}

type bundle struct {
	Authz authz `yaml:"authz"`
}

type service struct {
	Name        string      `yaml:"name"`
	URL         string      `yaml:"url"`
	Credentials credentials `yaml:"credentials,omitempty"`
}

type labels struct {
	SystemID   string `yaml:"system-id"`
	SystemType string `yaml:"system-type"`
}

type labelsOCP struct {
	UniqueName string `yaml:"unique-name"`
	Namespace  string `yaml:"namespace"`
}

type discovery struct {
	Name    string `yaml:"name"`
	Prefix  string `yaml:"prefix,omitempty"`
	Service string `yaml:"service"`
}

type http struct {
	Headers []string `yaml:"headers"`
}

type requestContext struct {
	HTTP http `yaml:"http"`
}

// DecisionLogs contains configuration for decision logs
type DecisionLogs struct {
	RequestContext requestContext        `json:"request_context,omitempty" yaml:"request_context,omitempty"`
	ServiceName    string                `json:"service,omitempty" yaml:"service,omitempty"`
	ResourcePath   string                `json:"resource_path,omitempty" yaml:"resource_path,omitempty"`
	Reporting      *DecisionLogReporting `json:"reporting,omitempty" yaml:"reporting,omitempty"`
}

// DecisionLogReporting contains configuration for decision log reporting
type DecisionLogReporting struct {
	MaxDelaySeconds      int `json:"max_delay_seconds,omitempty" yaml:"max_delay_seconds,omitempty"`
	MinDelaySeconds      int `json:"min_delay_seconds,omitempty" yaml:"min_delay_seconds,omitempty"`
	UploadSizeLimitBytes int `json:"upload_size_limit_bytes,omitempty" yaml:"upload_size_limit_bytes,omitempty"`
}

// OPAConfigMap represents the structure of the OPA configuration file
type OPAConfigMap struct {
	Services     []service    `yaml:"services"`
	Labels       labels       `yaml:"labels"`
	Discovery    discovery    `yaml:"discovery"`
	DecisionLogs DecisionLogs `yaml:"decision_logs,omitempty"`
}

// OcpOPAConfigMap represents the structure of the OPA configuration file for OCP
type OcpOPAConfigMap struct {
	Services             []*ocp.OPAServiceConfig `yaml:"services"` //nolint:staticcheck
	Bundles              bundle                  `yaml:"bundles,omitempty"`
	DecisionLogs         DecisionLogs            `yaml:"decision_logs,omitempty"`
	PersistenceDirectory string                  `yaml:"persistence_directory,omitempty"`
	Labels               labelsOCP               `yaml:"labels,omitempty"`
	Server               Serverconfig            `yaml:"server,omitempty"`
	Status               StatusConfig            `yaml:"status,omitempty"`
}

// StatusConfig represents the status configuration for OPA
type StatusConfig struct {
	Prometheus bool `yaml:"prometheus,omitempty"`
}

// Serverconfig represents the server configuration for OPA
type Serverconfig struct {
	Metrics Metricsconfig `yaml:"metrics,omitempty"`
}

// Metricsconfig represents the metrics configuration for OPA
type Metricsconfig struct {
	Prometheus PrometheusMetricsConfig `yaml:"prom,omitempty"`
}

// PrometheusMetricsConfig represents the Prometheus metrics configuration for OPA
type PrometheusMetricsConfig struct {
	HTTP HTTPMetricsConfig `yaml:"http_request_duration_seconds,omitempty"`
}

// HTTPMetricsConfig represents the HTTP metrics configuration for OPA
type HTTPMetricsConfig struct {
	Buckets []float64 `yaml:"buckets,omitempty"`
}

// OPAConfToK8sOPAConfigMapforOCP creates a ConfigMap for the OPA.
// It configures OPA to fetch bundle from MinIO.
// OPAConfToK8sOPAConfigMapforOCP merges the information given as input into a ConfigMap for OPA
func OPAConfToK8sOPAConfigMapforOCP( //nolint:staticcheck
	opaconf ocp.OPAConfig,
	legacyControllerOpaConfig configv2alpha2.OPAConfig, //nolint:staticcheck
	controllerOpaConfig map[string]interface{},
	legacySystemCustomOpaConfig map[string]interface{},
	systemOpaConfig map[string]interface{},
	_ logr.Logger,
	legacyBundleService *ocp.OPAServiceConfig, //nolint:staticcheck
	legacyLogService *ocp.OPAServiceConfig, //nolint:staticcheck
	legacyDecisionLogReporting configv2alpha2.DecisionLogReporting, //nolint:staticcheck
) (corev1.ConfigMap, error) {
	var services []*ocp.OPAServiceConfig //nolint:staticcheck

	if legacyBundleService != nil {
		services = append(services, legacyBundleService)
	}

	if legacyLogService != nil {
		services = append(services, legacyLogService)
	}

	ocpOPAConfigMap := OcpOPAConfigMap{
		Services: services,
		Labels: labelsOCP{
			UniqueName: opaconf.UniqueName,
			Namespace:  opaconf.Namespace,
		},
		Bundles: bundle{
			Authz: authz{
				Resource: opaconf.BundleResource,
			},
		},
	}

	if legacyBundleService != nil {
		ocpOPAConfigMap.Bundles.Authz.Service = legacyBundleService.Name
	}

	if legacyLogService != nil {
		ocpOPAConfigMap.DecisionLogs = DecisionLogs{
			ServiceName:  legacyLogService.Name,
			ResourcePath: "/logs",
			Reporting: &DecisionLogReporting{
				MaxDelaySeconds:      legacyDecisionLogReporting.MaxDelaySeconds,
				MinDelaySeconds:      legacyDecisionLogReporting.MinDelaySeconds,
				UploadSizeLimitBytes: legacyDecisionLogReporting.UploadSizeLimitBytes,
			},
		}
	}

	if legacyControllerOpaConfig.Metrics.Prometheus.HTTP.Buckets != nil {
		ocpOPAConfigMap.Server = Serverconfig{
			Metrics: Metricsconfig{
				Prometheus: PrometheusMetricsConfig{
					HTTP: HTTPMetricsConfig{
						Buckets: legacyControllerOpaConfig.Metrics.Prometheus.HTTP.Buckets,
					},
				},
			},
		}
		ocpOPAConfigMap.Status = StatusConfig{
			Prometheus: true,
		}
	}

	if legacyControllerOpaConfig.PersistBundle {
		ocpOPAConfigMap.Bundles.Authz.Persist = legacyControllerOpaConfig.PersistBundle
		ocpOPAConfigMap.PersistenceDirectory = legacyControllerOpaConfig.PersistBundleDirectory
	}

	if legacyControllerOpaConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		ocpOPAConfigMap.DecisionLogs.RequestContext = requestContext{
			HTTP: http{
				Headers: legacyControllerOpaConfig.DecisionLogs.RequestContext.HTTP.Headers,
			},
		}
	}

	opaConfigMapMapStringInterface, err := opaConfigMapToMap(ocpOPAConfigMap)
	if err != nil {
		return corev1.ConfigMap{}, err
	}

	merged := mergeMaps(opaConfigMapMapStringInterface, controllerOpaConfig)
	merged = mergeMaps(merged, legacySystemCustomOpaConfig)
	merged = mergeMaps(merged, systemOpaConfig)

	res, err := yaml.Marshal(&merged)
	if err != nil {
		return corev1.ConfigMap{}, errors.Wrap(err, "Could not marshal configmap data")
	}

	var cm corev1.ConfigMap
	cm.Data = map[string]string{
		"opa-conf.yaml": string(res),
	}

	return cm, nil
}

func opaConfigMapToMap(cm interface{}) (map[string]interface{}, error) {
	res, err := yaml.Marshal(&cm)
	if err != nil {
		return nil, errors.Wrap(err, "Could not marshal configmap data")
	}

	var opaConfigMapMapStringInterface map[string]interface{}

	err = yaml.Unmarshal(res, &opaConfigMapMapStringInterface)
	if err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal configmap data to map[string]interface{}")
	}

	return opaConfigMapMapStringInterface, nil
}

// mergeMaps recursively merges two map[string]interface{} variables. map2 takes precedence
// over map1 in case of key conflicts.
func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	// TODO: some times, yaml structs have a name as a key and the value under it
	// but other times, it is a list, where 'name' is one of the fields.
	// This function does not handle that case yet.
	mergedMap := make(map[string]interface{})

	// Copy all key-value pairs from map1 to mergedMap
	for key, value := range map1 {
		mergedMap[key] = value
	}

	// Copy all key-value pairs from map2 to mergedMap
	// Overwrite rule:
	// - If key(from map2) is absent in mergedMap: take map2's value.
	// - If key(from map2) exists in mergedMap, but either cannot normalize to map[string]interface{}: overwrite.
	// - If both normalize to maps: recurse (preserving existing nested keys and applying overrides).
	for key, value := range map2 {
		if existingValue, ok := mergedMap[key]; ok {
			// Attempt to normalize both existing and new values to map[string]interface{} before deciding overwrite
			existingMap, existingIsMap := normalizeToStringMap(existingValue)
			valueMap, valueIsMap := normalizeToStringMap(value)
			if existingIsMap && valueIsMap {
				mergedMap[key] = mergeMaps(existingMap, valueMap)
				continue
			}
		}
		mergedMap[key] = value
	}

	return mergedMap
}

// normalizeToStringMap converts supported map types (map[string]interface{} or map[interface{}]interface{})
// into map[string]interface{} recursively. Returns the normalized map and a bool indicating success.
func normalizeToStringMap(in interface{}) (map[string]interface{}, bool) {
	switch m := in.(type) {
	case map[string]interface{}:
		// Need to recursively normalize nested maps that may still be map[interface{}]interface{}
		res := make(map[string]interface{}, len(m))
		for k, v := range m {
			if nested, ok := normalizeToStringMap(v); ok {
				res[k] = nested
			} else {
				res[k] = v
			}
		}
		return res, true
	case map[interface{}]interface{}:
		res := make(map[string]interface{}, len(m))
		for k, v := range m {
			ks, ok := k.(string)
			if !ok {
				// Skip non-string keys; YAML object keys for our use-case should be strings
				continue
			}
			if nested, ok := normalizeToStringMap(v); ok {
				res[ks] = nested
			} else {
				res[ks] = v
			}
		}
		return res, true
	default:
		return nil, false
	}
}
