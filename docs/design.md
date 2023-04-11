# CustomResourceDefinition Design

This document describes the design of the custom resource definitions that the
styra-controller manages.

The custom resources managed by the styra-controller are:

* `System`
* `GlobalDatasource`

## System  

The `System` custom resource definition (CRD) declaratively defines a desired
system in Styra DAS. It provides options for configuring the name of the
system, datasources, decision mappings, git settings, and access control as a
list of users and/or SSO claims. It also supports the use of 
[Styra Local Plane](https://docs.styra.com/das/policies/policy-organization/systems/use-styra-local-plane).

```yaml
apiVersion: styra.bankdata.dk/v1beta1
kind: System
metadata:
  name: example-system
  labels:
    app: example-system
spec:
  datasources:
    - path: datasources/example
  decisionMappings:
    - allowed:
        expected:
          boolean: true
        path: result.allowed
      columns:
        - key: extra
          path: input.extra
      name: path/to/example/rule
      reason:
        path: result.reasons
  deletionProtection: true
  localPlane:
    name: styra-local-plane-example
  sourceControl:
    origin:
      commit: commitSHA
      path: path/to/policy/in/git/repo
      url: 'git-repo-url'
  subjects:
    - name: user@user.com
    - kind: group
      name: my-group
```

The git credentials which Styra DAS will need for fetching policy are specified
by referencing a secret by setting
`.spec.sourceControl.origin.credentialsSecretName`. For instance, if you set
`.spec.sourceControler.origin.credentialsSecretName: my-git-credentials` the
styra-controller will look for a secret called `my-git-credentials` in the same
namespace as the `System` resource. The secret is expected to contain a `name`
and a `secret` key. The `name` key should contain the basic auth username and
`secret` should contain the basic auth password.

If you would rather not have to set this for every system, the controller also
supports default credentials which will be used if the `credentialsSecretName`
field is left empty. Read more about this in the 
[controller configuration documentation](configuration.md#default-git-credentials).

## GlobalDatasource
 
The `GlobalDatasource` custom resource definition (CRD) declaratively defines a
desired global datasource in Styra DAS. It provides options for configuring the
name of the datasource and git settings.

Currently, the only supported
datasource category is `git/rego`. When the `git/rego` category is used on a
global datasource, Styra DAS will treat is as a rego library.

```yaml
apiVersion: styra.bankdata.dk/v1alpha1
kind: GlobalDatasource
metadata:
  name: global-datasource-example
spec:
  category: git/rego
  deletionProtection: true
  enabled: true
  name: jwt
  reference: path/to/policy/in/git/repo
  url: 'git-repo-url'
  credentialsSecretRef:
    name: my-git-credentials
    namespace: my-namespace
```

The git credentials which Styra DAS will need for fetching policy are specified
by referencing a secret by setting `.spec.credentialsSecretRef`. For instance,
if you set `.spec.credentialsSecretRef.name: my-git-credentials` and
`.spec.credentialsSecretRef.namespace: my-namespace` the styra-controller will
look for a secret called `my-git-credentials` in the namespace `my-namespace`.
The secret is expected to contain a `name` and a `secret` key.

If you would rather not have to set this for every `GlobalDatasource`, the
controller also supports default credentials which will be used if the
`credentialsSecretName` field is left empty. Read more about this in the
[controller configuration
documentation](configuration.md#default-git-credentials).
