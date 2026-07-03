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

package v1beta1

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/bankdata/styra-controller/pkg/opaconfig"
)

// OPAConfigSpec mirrors the official OPA configuration schema.
type OPAConfigSpec struct {
	Services                     *runtime.RawExtension            `json:"services,omitempty"`
	Labels                       map[string]string                `json:"labels,omitempty"`
	Discovery                    *runtime.RawExtension            `json:"discovery,omitempty"`
	Bundle                       *runtime.RawExtension            `json:"bundle,omitempty"`
	Bundles                      *runtime.RawExtension            `json:"bundles,omitempty"`
	DecisionLogs                 *runtime.RawExtension            `json:"decision_logs,omitempty"`
	Status                       *runtime.RawExtension            `json:"status,omitempty"`
	Plugins                      map[string]*runtime.RawExtension `json:"plugins,omitempty"`
	Keys                         *runtime.RawExtension            `json:"keys,omitempty"`
	DefaultDecision              *string                          `json:"default_decision,omitempty"`
	DefaultAuthorizationDecision *string                          `json:"default_authorization_decision,omitempty"`
	Caching                      *runtime.RawExtension            `json:"caching,omitempty"`
	NDBuiltinCache               bool                             `json:"nd_builtin_cache,omitempty"`
	PersistenceDirectory         *string                          `json:"persistence_directory,omitempty"`
	DistributedTracing           *runtime.RawExtension            `json:"distributed_tracing,omitempty"`
	MetricsExport                *runtime.RawExtension            `json:"metrics_export,omitempty"`
	Server                       *OPAServerConfig                 `json:"server,omitempty"`
	Storage                      *OPAStorageConfig                `json:"storage,omitempty"`
}

// OPAServerConfig mirrors OPA's server configuration section.
type OPAServerConfig struct {
	Metrics      *runtime.RawExtension `json:"metrics,omitempty"`
	Encoding     *runtime.RawExtension `json:"encoding,omitempty"`
	Decoding     *runtime.RawExtension `json:"decoding,omitempty"`
	LoggerPlugin *string               `json:"logger_plugin,omitempty"`
}

// OPAStorageConfig mirrors OPA's storage configuration section.
type OPAStorageConfig struct {
	Disk *runtime.RawExtension `json:"disk,omitempty"`
}

// UnmarshalJSON validates the OPA config against the known schema keys before
// decoding it.
func (c *OPAConfigSpec) UnmarshalJSON(data []byte) error {
	if err := opaconfig.ValidateRaw(data); err != nil {
		return err
	}

	type alias OPAConfigSpec
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	*c = OPAConfigSpec(decoded)
	return nil
}
