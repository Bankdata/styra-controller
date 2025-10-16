# CustomResourceDefinition Design

This document describes the design of the custom resource definitions that the
ocp-controller manages.

The custom resources managed by the ocp-controller are:

* `System`
* `Library`

## System  

The `System` custom resource definition (CRD) declaratively defines a desired
bundle in OPA Control Plane (Before a System in Styra DAS). It provides options for configuring the name of the bundle, requirements/datasources, decision mappings(deprecated, only used in Styra DAS), git settings, and access control as a list of users and/or SSO claims (deprecated, only used in Styra DAS).

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
  enableDeltaBundles: true
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

The git credentials which OPA Control Plane will need for fetching policy are specified
by referencing to a credential ID in the controller config `opaControlPlane.gitCredentials.id` and `opaControlPlane.gitCredentials.repoPrefix`.
[controller configuration documentation](configuration.md#default-git-credentials).

## Library

The `Library` custom resource definition (CRD) declaratively defines a desired library in OPA Control Plane (Before Styra DAS). It provides options for configuring the name of the library, a description of it, permissions, git settings, and datasources. Note, a library is just a source in OPA Control Plane.

```yaml
apiVersion: styra.bankdata.dk/v1alpha1
kind: Library
metadata:
  name: my-library
spec:
  name: mylibrary
  description: my library
  sourceControl:
    libraryOrigin:
      url: https://github.com/Bankdata/styra-controller.git
      reference: refs/heads/master
      commit: f37cc9d87251921cbe49349235d9b5305c833769
      path: rego/path
  datasources:
    - path: seconds/datasource
      description: this is the second datasource 
  subjects:
    - kind: user
      name: user1@mail.dk
    - kind: group
      name: mygroup
```

The content of the library is what is found in the folder `<path>/libraries/<library-name>`. 
There is therefore a tight coupling between the library name and the path to the library in the git repository. The library name is also used as the name of the library in OPA Control Plane.
With the above example, the content of the library would be the files found at 
`https://github.com/Bankdata/styra-controller/tree/master/rego/path/libraries/mylibrary` together with the datasource.