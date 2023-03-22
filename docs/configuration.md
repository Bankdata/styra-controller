# Configuration of the Styra Controller
This document describes the different configuration options for the Styra Controller. The configuration options are defined in `api/config/v1/projectconfig_types.go`. The configuration options are assigned in `config/default/config.yaml`. For the ease of reference the configuration options are listed here:

* `StyraToken`               
* `StyraAddress`           
* `StyraSystemUserRoles`    
* `StyraSystemPrefix`       
* `StyraSystemSuffix`        
* `LogLevel`                 
* `SentryDSN`                
* `SentryDebug`              
* `Environment`              
* `SentryHTTPSProxy`         
* `ControllerClass`         
* `WebhooksDisabled`         
* `DatasourceWebhookAddress` 
* `IdentityProvider`         
* `JwtGroupClaim`            
* `GitCredentials`

## Observability
### Local Logging
The controllers logs are written to stdout. If the logs should be persistent an external system should be configured to scrape and store the logs. The controllers log level is configured by setting `logLevel` to an integer.

| Level | Meaning |
|-------|---------|
| -1    | Debug   | 
| 0     | Info    | 
| 1     | Warn    |
| 2     | Error   | 

### Logging to Sentry
To configure Sentry there exists four configuration options: `sentryDSN`, `environment`, `sentryDebug`, and `sentryHTTPSProxy`. `sentryDSN` is the DSN to the Sentry instance. `environment` specifies the Sentry environment that the log should be categorized under in Sentry. `sentryDebug` toggles whether information sendt to Sentry should also be sent to stdout. If Sentry can only be reaches through a proxy set `sentryHTTPSProxy` to the proxy URL.

In `internal/sentry` is a sentry reconciler that wraps the other reconcilers. The sentry reconciler simply calls the reconcilers. If the reconcilers return an error and Sentry has been configured, the sentry reconciler will send the error to Sentry. 

### Metrics
The Styra Controller exposes the standard go and controller runtime metrics. In addition, the controller exposes the `controller_system_status_ready` metric that counts the amount of Systems whose status are Ready.

### Notification Webhook
Currently the controller can register a custom notification webhook that will POST the system ID and datasource ID to an URL when a systems datasource is created or updated. The webhook is implemented in `internal/webhook`. The URL is configured by setting `datasourceWebhookAddress` and the data is formatted like this:

```json
{
  "systemId": "system ID", 
  "datasourceId": "datasource ID"
}
```

## RBAC
Access to Styra can be given based on emails and SSO claims. The access rights given to the users are defined in `styraSystemUserRoles`. For giving access based on SSO claims set the `identityProvider` to the SSO providor used to login to Styra. Which claim in the JWT to give access upon is define by setting `jwtGroupClaim`. As an example, assume `styraSystemUserRoles` is `[SystemViewer, SystemInstall]`, `identityProvider` is AzureAD, and `jwtGroupClaim` is `companies`. Then if `.spec.subjects` are:

```yaml
spec:
  subjects:
  - name: user@users.com
  - kind: group
    name: mycompany
```
Then the user with the `user@users.com` email has access to the system with the access rights defined above. If Styra has been configured with SSO login then users will have access to the system with the same rights if the token from the identity provider has a claim called `companies` that contains `mycompany`. **NOTE:** The `kind: group` element has no functional consequences it is just to distinquish between users and SSO claims. 

## Default Git credentials
Styra needs a set of credentials for fetching the Systems policies the Git repository. The controller can be configured with a set of Git credentials for different domains. This is done by setting the `GitCredentials`. For example, if `GitCredentials` is: 

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
Only one controller per cluster should have webhooks (default and validating) enabled. Therefore, when running multiple controllers in the same cluster set `webhooksDisabled` on one of them. Usually, it is the least stable version of the controller that has webhooks disabled.

### Configure prefix and suffix on Systems
The controller can be configured to add a prefix and a suffix to the Systems names when created in Styra. This is achieved by setting `styraSystemPrefix` and `styraSystemSuffix`. 

