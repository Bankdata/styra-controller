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
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/pkg/ocp"
	"github.com/bankdata/styra-controller/pkg/styra"
)

type bearer struct {
	TokenPath string `yaml:"token_path"`
}

type credentials struct {
	Bearer bearer `yaml:"bearer"`
}

// type s3credentials struct {
// 	S3Signing s3signing `yaml:"s3_signing"`
// }

// type s3signing struct {
// 	S3EnvironmentCredentials map[string]interface{} `yaml:"environment_credentials"`
// }

type authz struct {
	Service  string `yaml:"service"`
	Resource string `yaml:"resource"`
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

// type s3service struct {
// 	Name          string        `yaml:"name"`
// 	URL           string        `yaml:"url"`
// 	S3Credentials s3credentials `yaml:"credentials"`
// }

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

// OpaConfigMap represents the structure of the OPA configuration file
type OpaConfigMap struct {
	Services     []service    `yaml:"services"`
	Labels       labels       `yaml:"labels"`
	Discovery    discovery    `yaml:"discovery"`
	DecisionLogs DecisionLogs `yaml:"decision_logs,omitempty"`
}

// OcpOpaConfigMap represents the structure of the OPA configuration file for OCP
type OcpOpaConfigMap struct {
	Services             []*configv2alpha2.OPAServiceConfig `yaml:"services"`
	Bundles              bundle                             `yaml:"bundles,omitempty"`
	DecisionLogs         DecisionLogs                       `yaml:"decision_logs,omitempty"`
	PersistenceDirectory string                             `yaml:"persistence_directory,omitempty"`
	Labels               labelsOCP                          `yaml:"labels,omitempty"`
}

// OpaConfToK8sOPAConfigMapforOCP creates a ConfigMap for the OPA.
// It configures OPA to fetch bundle from MinIO.
// OpaConfToK8sOPAConfigMapforOCP merges the information given as input into a ConfigMap for OPA
func OpaConfToK8sOPAConfigMapforOCP(
	opaconf ocp.OPAConfig,
	opaDefaultConfig configv2alpha2.OPAConfig,
	customConfig map[string]interface{},
) (corev1.ConfigMap, error) {

	var services []*configv2alpha2.OPAServiceConfig

	if opaconf.BundleService != nil {
		services = append(services, opaconf.BundleService)
	}
	if opaconf.LogService != nil {
		services = append(services, opaconf.LogService)
	}

	fmt.Println("Services in OCP OPA ConfigMap:")
	for _, svc := range services {
		fmt.Printf("- Name: %s, URL: %s\n", svc.Name, svc.URL)
		fmt.Println("  Credentials:", svc.Credentials,
			" ResponseHeaderTimeoutSeconds:", svc.ResponseHeaderTimeoutSeconds)
	}

	fmt.Println("{}", opaconf.DecisionLogReporting)

	ocpOpaConfigMap := OcpOpaConfigMap{
		Bundles: bundle{
			Authz: authz{
				Service:  opaconf.BundleService.Name,
				Resource: opaconf.BundleResource,
			},
		},
		Services: services,
		Labels: labelsOCP{
			UniqueName: opaconf.UniqueName,
			Namespace:  opaconf.Namespace,
		},
		DecisionLogs: DecisionLogs{
			ServiceName:  opaconf.LogService.Name,
			ResourcePath: "/logs",
			Reporting: &DecisionLogReporting{
				MaxDelaySeconds:      opaconf.DecisionLogReporting.MaxDelaySeconds,
				MinDelaySeconds:      opaconf.DecisionLogReporting.MinDelaySeconds,
				UploadSizeLimitBytes: opaconf.DecisionLogReporting.UploadSizeLimitBytes,
			},
		},
	}

	fmt.Println(ocpOpaConfigMap.DecisionLogs)

	if opaDefaultConfig.PersistBundle {
		ocpOpaConfigMap.Bundles.Authz.Persist = opaDefaultConfig.PersistBundle
		ocpOpaConfigMap.PersistenceDirectory = opaDefaultConfig.PersistBundleDirectory
	}

	if opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		ocpOpaConfigMap.DecisionLogs = DecisionLogs{
			RequestContext: requestContext{
				HTTP: http{
					Headers: opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers,
				},
			},
		}
	}

	opaConfigMapMapStringInterface, err := opaConfigMapToMap(ocpOpaConfigMap)
	if err != nil {
		return corev1.ConfigMap{}, err
	}

	fmt.Println(opaConfigMapMapStringInterface)
	fmt.Println(customConfig)

	merged := mergeMaps(opaConfigMapMapStringInterface, customConfig)

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

// OpaConfToK8sOPAConfigMap creates a corev1.ConfigMap for the OPA based on the
// configuration from Styra. The configmap configures the OPA to communicate to
// an SLP.
func OpaConfToK8sOPAConfigMap(
	opaconf styra.OPAConfig,
	slpURL string,
	opaDefaultConfig configv2alpha2.OPAConfig,
	customConfig map[string]interface{},
) (corev1.ConfigMap, error) {

	opaConfigMap := OpaConfigMap{
		Services: []service{{
			Name: "styra",
			URL:  slpURL,
		}},
		Labels: labels{
			SystemID:   opaconf.SystemID,
			SystemType: opaconf.SystemType,
		},
		Discovery: discovery{
			Name:    "discovery",
			Service: "styra",
		},
	}

	if opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		opaConfigMap.DecisionLogs = DecisionLogs{
			RequestContext: requestContext{
				HTTP: http{
					Headers: opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers,
				},
			},
		}
	}

	opaConfigMapMapStringInterface, err := opaConfigMapToMap(opaConfigMap)

	if err != nil {
		return corev1.ConfigMap{}, err
	}

	merged := mergeMaps(opaConfigMapMapStringInterface, customConfig)

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

// OpaConfToK8sSLPConfigMap creates a ConfigMap for the SLP based on the
// configuration from Styra.
func OpaConfToK8sSLPConfigMap(opaconf styra.OPAConfig) (corev1.ConfigMap, error) {
	type Bearer struct {
		TokenPath string `yaml:"token_path"`
	}

	type Credentials struct {
		Bearer Bearer `yaml:"bearer"`
	}

	type Service struct {
		Name        string      `yaml:"name"`
		URL         string      `yaml:"url"`
		Credentials Credentials `yaml:"credentials"`
	}

	type Labels struct {
		SystemID   string `yaml:"system-id"`
		SystemType string `yaml:"system-type"`
	}

	type Discovery struct {
		Name     string `yaml:"name"`
		Resource string `yaml:"resource"`
		Service  string `yaml:"service"`
	}

	type SLPConfigMap struct {
		Services  []Service `yaml:"services"`
		Labels    Labels    `yaml:"labels"`
		Discovery Discovery `yaml:"discovery"`
	}

	slpConfigMap := SLPConfigMap{
		Services: []Service{{
			Name: "styra",
			URL:  opaconf.HostURL,
			Credentials: Credentials{
				Bearer: Bearer{
					TokenPath: "/etc/slp/auth/token",
				},
			},
		}},
		Labels: Labels{
			SystemID:   opaconf.SystemID,
			SystemType: opaconf.SystemType,
		},
		Discovery: Discovery{
			Name:     "discovery",
			Resource: fmt.Sprintf("/systems/%s/discovery", opaconf.SystemID),
			Service:  "styra",
		},
	}

	res, err := yaml.Marshal(&slpConfigMap)
	if err != nil {
		return corev1.ConfigMap{}, errors.Wrap(err, "could not marshal opa config data")
	}

	var cm corev1.ConfigMap
	cm.Data = map[string]string{
		"slp.yaml": string(res),
	}

	return cm, nil
}

// OpaConfToK8sOPAConfigMapNoSLP creates a ConfigMap for the OPA based on the
// configuration from Styra. The ConfigMap configures the OPA to communicate
// directly to Styra and not via an SLP.
func OpaConfToK8sOPAConfigMapNoSLP(
	opaconf styra.OPAConfig,
	opaDefaultConfig configv2alpha2.OPAConfig,
	customConfig map[string]interface{},
) (corev1.ConfigMap, error) {

	opaConfigMap := OpaConfigMap{
		Services: []service{{
			Name: "styra",
			URL:  opaconf.HostURL,
			Credentials: credentials{
				Bearer: bearer{
					TokenPath: "/etc/opa/auth/token",
				},
			},
		},
			{
				Name: "styra-bundles",
				URL:  fmt.Sprintf("%s/bundles", opaconf.HostURL),
				Credentials: credentials{
					Bearer: bearer{
						TokenPath: "/etc/opa/auth/token",
					},
				},
			},
		},
		Labels: labels{
			SystemID:   opaconf.SystemID,
			SystemType: opaconf.SystemType,
		},
		Discovery: discovery{
			Name:    "discovery",
			Prefix:  fmt.Sprintf("/systems/%s", opaconf.SystemID),
			Service: "styra",
		},
	}

	if opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		opaConfigMap.DecisionLogs = DecisionLogs{
			RequestContext: requestContext{
				HTTP: http{
					Headers: opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers,
				},
			},
		}
	}

	opaConfigMapMapStringInterface, err := opaConfigMapToMap(opaConfigMap)

	if err != nil {
		return corev1.ConfigMap{}, err
	}

	merged := mergeMaps(opaConfigMapMapStringInterface, customConfig)

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
	// TODO: input should be of type opaConfigMap. Difference between s3opaConfigMap and opaConfigMap
	// should be handled elsewhere. Marhsalling and unmarshalling back to map[string]interface{} is not optimal.
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

// mergeMaps recursively merges two map[string]interface{} variables
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
