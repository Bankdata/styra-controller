# Configuration of the ocp-controller
This document describes the different configuration options for the ocp-controller. The configuration options are defined in `api/config/v2alpha2/projectconfig_types.go`. The configuration options are assigned in `config/default/config.yaml`. For the ease of reference the configuration options are listed here:

## Multi-file configuration

The `--config` flag can be specified multiple times. When multiple config files are provided, they are deep-merged in order: the first file serves as the base and each subsequent file is an overlay whose values take precedence. Fields not present in an overlay file are preserved from the base.

This is designed to separate non-secret configuration (in a ConfigMap) from secret values (in a Secret):

```
--config=/etc/styra-controller/config.yaml \
--config=/etc/styra-controller-secrets/config-secrets.yaml
```

**Base config** (`config.yaml` in a ConfigMap) — contains all non-secret settings like addresses, log levels, system prefix/suffix, OPA config, etc.

**Secrets overlay** (`config-secrets.yaml` in a Secret) — contains only sensitive fields like API tokens, passwords, S3 keys, Sentry DSN, and TLS material. Only the fields you want to override need to be present; everything else is inherited from the base.

Example secrets overlay:
```yaml
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  token: my-ocp-token
userCredentialHandler:
  s3:
    accessKeyID: my-access-key
    secretAccessKey: my-secret-key
```

Single-file usage (`--config=config.yaml`) remains fully supported and behaves identically to previous versions.

#### Supported configuration options for OPA Control Plane
* `opaControlPlaneConfig`
* `systemPrefix`
* `systemSuffix`
* `logLevel`
* `leaderElection`
* `sentry`
* `controllerClass`         
* `deletionProtectionDefault`
* `disableCRDWebhooks`

## Observability

### Local Logging

The controllers logs are written to stdout. If the logs should be persistent an
external system should be configured to scrape and store the logs. The
verbosity of the controller logs is configured by setting `logLevel` to an
integer. A log level above 0 should only be set for debugging purposes.

### Logging to Sentry
To configure Sentry there exists four configuration options: `sentry.dsn`, `sentry.environment`, `sentry.debug`, and `sentry.httpsProxy`. `sentry.dsn` is the DSN to the Sentry instance. `sentry.environment` specifies the Sentry environment that the log should be categorized under in Sentry. `sentry.debug` toggles whether information sendt to Sentry should also be sent to stdout. If Sentry can only be reaches through a proxy set `sentry.httpsProxy` to the proxy URL.

In `internal/sentry` is a sentry reconciler that wraps the other reconcilers. The sentry reconciler simply calls the reconcilers. If the reconcilers return an error and Sentry has been configured, the sentry reconciler will send the error to Sentry. 

### Metrics
The ocp-controller exposes the standard go and controller runtime metrics. In addition, the controller exposes:
- `controller_system_status_ready` metric that counts the amount of Systems whose status are Ready.
  "datasourceID": "datasource ID"
spec:
  subjects:
  - name: user@users.com
  - kind: group
    name: mycompany
```
Then the user with the `user@users.com` email has access to the system with the access rights defined above. If Styra has been configured with SSO login then users will have access to the system with the same rights if the token from the identity provider has a claim called `companies` that contains `mycompany`. **NOTE:** The `kind: group` element has no functional consequences it is just to distinquish between users and SSO claims. 

## Default Git credentials (DEPRECATED)
Styra needs a set of credentials for fetching the Systems policies the Git repository. The controller can be configured with a set of Git credentials for different domains. This is done by setting the `gitCredentials`. For example, if `gitCredentials` is: 

```yaml
gitCredentials: 
  - user: "default mydomain user"
    password: "secret password"
    repoPrefix: https://mydomain.com
  - user: "default myotherdomain user"
    password: "secret password"
    repoPrefix: https://myotherdomain.com
```
then the first set of credentials will be used for systems that fetch policies from `mydomain.com` and the second set of credentials for systems that fetch policies from `myotherdomain.com`. 

## Multiple instances of the ocp-controller
This section describes how to handle the scenario where multiple controller instances are running in the same cluster and/or are hooked up to the same OCP instance. This is useful in order to be able to identify which controller created a system.

### ControllerClass
The controller can be configured to only reconcile resources that has the `styra-controller/class` label set to a specific value. This is configured by setting `controllerClass`. For example, if `controllerClass` is set to `dev` the controller will only reconcile resources with the `styra-controller/class: dev` label. And as default, when no `controllerClass` is configured for the controller, the controller will only reconcile resources that do not have the `styra-controller/class` label. 

### Disabling webhooks
Only one controller per cluster should have webhooks (default and validating) enabled. Therefore, when running multiple controllers in the same cluster set `disableCRDWebhooks` on one of them. Usually, it is the least stable version of the controller that has webhooks disabled.

### Configure prefix and suffix on Systems
The controller can be configured to add a prefix and a suffix to managed system names. This is achieved by setting `systemPrefix` and `systemSuffix`. 

## Delete Protection
Custom Resources can have delete protection, which means that backing OCP resources will not be deleted by the controller. The default can be configured by setting `deletionProtectionDefault`.

## Leader Election
If multiple instances of the controller are running together, leader election can be configured by setting the `leaderElection.leaseDuration`, `leaderElection.renewDeadline`, `leaderElection.retryPeriod`.
