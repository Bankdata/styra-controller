# Configuration of the Styra Controller
This document describes the different configuration options for the Styra Controller. The configuration options are defined in `api/config/v2alpha2/projectconfig_types.go`. The configuration options are assigned in `config/default/config.yaml`. For the ease of reference the configuration options are listed here:

* `controllerClass`         
* `deletionProtectionDefault`
* `enableDeltaBundlesDefault`
* `readOnly`
* `disableCRDWebhooks`
* `enableMigrations`
* `gitCredentials`
* `logLevel`
* `leaderElection`
* `notificationWebhooks`
* `sentry`
* `sso`
* `styra`
* `systemPrefix`
* `systemSuffix`
* `systemUserRoles`

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
The Styra Controller exposes the standard go and controller runtime metrics. In addition, the controller exposes the `controller_system_status_ready` metric that counts the amount of Systems whose status are Ready.

### Notification Webhooks
#### System Datasources
Currently the controller can register a custom notification webhook that will POST the system ID and datasource ID to a URL when a system's datasource is created or updated. The webhook is implemented in `internal/webhook`. The URL is configured by setting `notificationWebhooks.systemDatasourceChanged` and the data is formatted like this:

```json
{
  "systemId": "system ID", 
  "datasourceId": "datasource ID"
}
```
#### Library Datasources
The controller also can register a custom notification webhook that will POST the datasource ID of a Library Datasource to a URL when a library's datasource is created or updated. The webhook is implemented in `internal/webhook`. The URL is configured by setting `notificationWebhooks.libraryDatasourceChanged` and the data is formatted like this:

```json
{
  "datasourceID": "datasource ID"
}
```


## RBAC
Access to Styra can be given based on emails and SSO claims. The access rights given to the users are defined in `systemUserRoles`. For giving access based on SSO claims set the `sso.identityProvider` to the SSO providor used to login to Styra. Which claim in the JWT to give access upon is define by setting `sso.jwtGroupsClaim`. As an example, assume `systemUserRoles` is `[SystemViewer, SystemInstall]`, `sso.identityProvider` is AzureAD, and `sso.jwtGroupsClaim` is `companies`. Then if `.spec.subjects` are:

```yaml
spec:
  subjects:
  - name: user@users.com
  - kind: group
    name: mycompany
```
Then the user with the `user@users.com` email has access to the system with the access rights defined above. If Styra has been configured with SSO login then users will have access to the system with the same rights if the token from the identity provider has a claim called `companies` that contains `mycompany`. **NOTE:** The `kind: group` element has no functional consequences it is just to distinquish between users and SSO claims. 

## Default Git credentials
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

## Multiple instances of the Styra Controller
This section describes how to handle the scenario where multiple controller instances are running in the same cluster and/or are hooked up to the same Styra DAS instance. This is useful when testing a new version of the controller.

### ControllerClass
The controller can be configured to only reconcile resources that has the `styra-controller/class` label set to a specific value. This is configured by setting `controllerClass`. For example, if `controllerClass` is set to `dev` the controller will only reconcile resources with the `styra-controller/class: dev` label. And as default, when no `controllerClass` is configured for the controller, the controller will only reconcile resources that do not have the `styra-controller/class` label. 

### Disabling webhooks
Only one controller per cluster should have webhooks (default and validating) enabled. Therefore, when running multiple controllers in the same cluster set `disableCRDWebhooks` on one of them. Usually, it is the least stable version of the controller that has webhooks disabled.

### Configure prefix and suffix on Systems
The controller can be configured to add a prefix and a suffix to the Systems names when created in Styra. This is achieved by setting `systemPrefix` and `systemSuffix`. 

## Delete Protection
Custom Resources can have delete protection, which means that they will not be deleted by the controller in Styra. The default can be configured by setting `deletionProtectionDefault`.

## Delta Bundles
Styra Systems can have enable Delta Bundles, which means Styra will upload the change between two bundles to the SLP/OPA rather than uploading the entire bundle. 
The default can be configured by setting `enableDeltaBundlesDefault`.
This is recommended to be set to true if all opas are version 0.44.0 or higher.

## Read Only
Styra Systems can be read-only, meaning they cannot be changed in the Styra GUI. This can be configured by setting `readOnly`.

## EnableMigrations
An annotation that allows configuring Systems in Kubernetes to link to a specific system in Styra. The ID that the system in Kubernetes should link to is configured by setting `styra-contoller/migration-id: [styra system id]` annotation on Kubernetes system resource. Should only be set while migrating. 

## Leader Election
If multiple instances of the controller are running together, leader election can be configured by setting `leaderElection.leaseDuration`, `leaderElection.renewDeadline`, `leaderElection.retryPeriod`.