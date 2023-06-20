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

// Package config provides utilities for reading configfiles
package config

import (
	"os"

	v1 "github.com/bankdata/styra-controller/api/config/v1"
	"github.com/bankdata/styra-controller/api/config/v2alpha1"
	"github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	healthProbeBindAddress = ":8081"
	metricsBindAddress     = ":8080"
	leaderElectionID       = "5d272013.bankdata.dk"
	webhookPort            = 9443
)

// Load loads controller configuration from the given file using the types
// registered in the scheme.
func Load(file string, scheme *runtime.Scheme) (*v2alpha2.ProjectConfig, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "could not read config file")
	}
	return deserialize(bs, scheme)
}

// OptionsFromConfig creates a manager.Options based on a configuration file
func OptionsFromConfig(cfg *v2alpha2.ProjectConfig) manager.Options {
	o := manager.Options{
		HealthProbeBindAddress: healthProbeBindAddress,
		MetricsBindAddress:     metricsBindAddress,
		WebhookServer:          webhook.NewServer(webhook.Options{Port: webhookPort}),
	}

	if cfg.LeaderElection != nil {
		o.LeaderElection = true
		o.LeaseDuration = &cfg.LeaderElection.LeaseDuration
		o.RenewDeadline = &cfg.LeaderElection.RenewDeadline
		o.RetryPeriod = &cfg.LeaderElection.RetryPeriod
		o.LeaderElectionID = leaderElectionID
	}

	return o
}

func deserialize(data []byte, scheme *runtime.Scheme) (*v2alpha2.ProjectConfig, error) {
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
	_, gvk, err := decoder.Decode(data, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode config")
	}

	if gvk.Group != v2alpha2.GroupVersion.Group {
		return nil, errors.New("unsupported api group")
	}

	if gvk.Kind != "ProjectConfig" {
		return nil, errors.New("unsupported api kind")
	}

	cfg := &v2alpha2.ProjectConfig{}

	switch gvk.Version {
	case v2alpha2.GroupVersion.Version:
		if _, _, err := decoder.Decode(data, nil, cfg); err != nil {
			return nil, errors.Wrap(err, "could not decode into kind")
		}
	case v2alpha1.GroupVersion.Version:
		var v2cfg v2alpha1.ProjectConfig
		if _, _, err := decoder.Decode(data, nil, &v2cfg); err != nil {
			return nil, errors.Wrap(err, "could not decode into kind")
		}
		cfg = v2cfg.ToV2Alpha2()
	case v1.GroupVersion.Version:
		var v1cfg v1.ProjectConfig
		if _, _, err := decoder.Decode(data, nil, &v1cfg); err != nil {
			return nil, errors.Wrap(err, "could not decode into kind")
		}
		cfg = v1cfg.ToV2Alpha1().ToV2Alpha2()
	default:
		return nil, errors.New("unsupported api version")
	}

	return cfg, nil
}
