<p>Packages:</p>
<ul>
<li>
<a href="#styra.bankdata.dk%2fv1beta1">styra.bankdata.dk/v1beta1</a>
</li>
</ul>
<h2 id="styra.bankdata.dk/v1beta1">styra.bankdata.dk/v1beta1</h2>
<div>
<p>Package v1beta1 contains API Schema definitions for the styra v1beta1 API
group.</p>
</div>
Resource Types:
<ul></ul>
<h3 id="styra.bankdata.dk/v1beta1.AllowedMapping">AllowedMapping
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.DecisionMapping">DecisionMapping</a>)
</p>
<div>
<p>AllowedMapping specifies how to determine if a decision is allowed or not.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>expected</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.Expected">
Expected
</a>
</em>
</td>
<td>
<p>Expected is the value we expect to be set in the Path in order to consider
the decision allowed.</p>
</td>
</tr>
<tr>
<td>
<code>negated</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Negated negates the expectation.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path is the path to the value which we check our expectation against.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.ColumnMapping">ColumnMapping
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.DecisionMapping">DecisionMapping</a>)
</p>
<div>
<p>ColumnMapping specifies how a value in the decision result should be mapped
to a column in the Styra decision log.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<p>Key is the name of the column as shown in the decision log.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path is where in the decision result the value for the column is found.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.Condition">Condition
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SystemStatus">SystemStatus</a>)
</p>
<div>
<p>Condition represents a System condition.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.ConditionType">
ConditionType
</a>
</em>
</td>
<td>
<p>Type is the ConditionType of the Condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://v1-20.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#conditionstatus-v1-meta">
k8s.io/apimachinery/pkg/apis/meta/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status is the status of the Condition.</p>
</td>
</tr>
<tr>
<td>
<code>lastProbeTime</code><br/>
<em>
<a href="https://v1-20.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#time-v1-meta">
k8s.io/apimachinery/pkg/apis/meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastProbeTime is a timestamp for the last time the condition was checked.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code><br/>
<em>
<a href="https://v1-20.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#time-v1-meta">
k8s.io/apimachinery/pkg/apis/meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastTransitionTime is a timestamp for the last time that the condition
changed state.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.ConditionType">ConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.Condition">Condition</a>)
</p>
<div>
<p>ConditionType is a System Condition type.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;CreatedInStyra&#34;</p></td>
<td><p>ConditionTypeCreatedInStyra is a ConditionType used when the system has
been created in Styra.</p>
</td>
</tr><tr><td><p>&#34;DatasourcesUpdated&#34;</p></td>
<td><p>ConditionTypeDatasourcesUpdated is a ConditionType used when
the datasources of the System are updated in Styra.</p>
</td>
</tr><tr><td><p>&#34;GitCredentialsUpdated&#34;</p></td>
<td><p>ConditionTypeGitCredentialsUpdated is a ConditionType used when git
credentials are updated in Styra.</p>
</td>
</tr><tr><td><p>&#34;OPAConfigMapUpdated&#34;</p></td>
<td><p>ConditionTypeOPAConfigMapUpdated is a ConditionType used when
the ConfigMap for the OPA are updated in the cluster.</p>
</td>
</tr><tr><td><p>&#34;OPATokenUpdated&#34;</p></td>
<td><p>ConditionTypeOPATokenUpdated is a ConditionType used when
the secret with the Styra token has been updated in the cluster.</p>
</td>
</tr><tr><td><p>&#34;OPAUpToDate&#34;</p></td>
<td><p>ConditionTypeOPAUpToDate is a ConditionType used to say whether
the OPA is up to date or needs to be restarted.</p>
</td>
</tr><tr><td><p>&#34;SLPConfigMapUpdated&#34;</p></td>
<td><p>ConditionTypeSLPConfigMapUpdated is a COnditionType used when
the ConfigMap for the SLP are updated in the cluster.</p>
</td>
</tr><tr><td><p>&#34;SLPUpToDate&#34;</p></td>
<td><p>ConditionTypeSLPUpToDate is a ConditionType used to say whether
the SLP is up to date or needs to be restarted.</p>
</td>
</tr><tr><td><p>&#34;SubjectsUpdated&#34;</p></td>
<td><p>ConditionTypeSubjectsUpdated is a ConditionType used when the subjects of
the System are updated in Styra.</p>
</td>
</tr><tr><td><p>&#34;SystemConfigUpdated&#34;</p></td>
<td><p>ConditionTypeSystemConfigUpdated is a ConditionType used when
the configuration of the System are updated in Styra.</p>
</td>
</tr></tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.Datasource">Datasource
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SystemSpec">SystemSpec</a>)
</p>
<div>
<p>Datasource represents a Styra datasource to be mounted in the system.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path is the path within the system where the datasource should reside.</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<p>Description is a description of the datasource</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.DecisionMapping">DecisionMapping
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SystemSpec">SystemSpec</a>)
</p>
<div>
<p>DecisionMapping specifies how a system decision mapping should be
configured. This allows configuration of when a decision is considered
allowed or not. It also provides the ability to show additional columns in
Styra.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the decision mapping.</p>
</td>
</tr>
<tr>
<td>
<code>columns</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.ColumnMapping">
[]ColumnMapping
</a>
</em>
</td>
<td>
<p>Columns holds a list of ColumnMapping for the decision mapping.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.ReasonMapping">
ReasonMapping
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>allowed</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.AllowedMapping">
AllowedMapping
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.DiscoveryOverrides">DiscoveryOverrides
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SystemSpec">SystemSpec</a>)
</p>
<div>
<p>DiscoveryOverrides specifies system specific overrides for the configuration
served from the Styra OPA Discovery API</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.OPAConfigStatus">
OPAConfigStatus
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>distributed_tracing</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.OPAConfigDistributedTracing">
OPAConfigDistributedTracing
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.Expected">Expected
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.AllowedMapping">AllowedMapping</a>)
</p>
<div>
<p>Expected represents an expected value. When using this type only one of the
fields should be set.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>string</code><br/>
<em>
string
</em>
</td>
<td>
<p>String holds a pointer to a string if the Expected value represents a
string.</p>
</td>
</tr>
<tr>
<td>
<code>boolean</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Boolean holds a pointer to a bool if the Expected value represents a
bool.</p>
</td>
</tr>
<tr>
<td>
<code>integer</code><br/>
<em>
int
</em>
</td>
<td>
<p>Integer holds a pointer to an int if the Expected value represents an int.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.GitRepo">GitRepo
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SourceControl">SourceControl</a>)
</p>
<div>
<p>GitRepo specifies the configuration for how to pull policy from git.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>credentialsSecretName</code><br/>
<em>
string
</em>
</td>
<td>
<p>CredentialsSecretName is a reference to an existing secret which holds git
credentials. This secret should have the keys <code>name</code> and <code>secret</code>. The
<code>name</code> key should contain the http basic auth username and the <code>secret</code>
key should contain the http basic auth password.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path is the path in the git repo where the policies are located.</p>
</td>
</tr>
<tr>
<td>
<code>reference</code><br/>
<em>
string
</em>
</td>
<td>
<p>Reference is used to point to a tag or branch. This will be ignored if
<code>Commit</code> is specified.</p>
</td>
</tr>
<tr>
<td>
<code>commit</code><br/>
<em>
string
</em>
</td>
<td>
<p>Commit is used to point to a specific commit SHA. This takes precedence
over <code>Reference</code> if both are specified.</p>
</td>
</tr>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<p>URL is the URL of the git repo.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.LocalPlane">LocalPlane
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SystemSpec">SystemSpec</a>)
</p>
<div>
<p>LocalPlane specifies how the Styra Local Plane should be configured. This is
used to generate Secret and ConfigMap for the SLP to consume.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name is the hostname of the SLP service.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.OPAConfigDistributedTracing">OPAConfigDistributedTracing
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.DiscoveryOverrides">DiscoveryOverrides</a>)
</p>
<div>
<p>OPAConfigDistributedTracing configures the <code>distributed_tracing</code> key in the
OPA configuration.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>address</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>service_name</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>sample_percentage</code><br/>
<em>
int
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>encryption</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>allow_insecure_tls</code><br/>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tls_ca_cert_file</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tls_cert_file</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tls_private_key_file</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.OPAConfigStatus">OPAConfigStatus
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.DiscoveryOverrides">DiscoveryOverrides</a>)
</p>
<div>
<p>OPAConfigStatus configures the <code>status</code> key in the OPA configuration</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>prometheus</code><br/>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.ReasonMapping">ReasonMapping
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.DecisionMapping">DecisionMapping</a>)
</p>
<div>
<p>ReasonMapping specifies where the reason of the decision can be found.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path is the path to where the reason is found in the decision result.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.SourceControl">SourceControl
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SystemSpec">SystemSpec</a>)
</p>
<div>
<p>SourceControl holds SourceControl configuration.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>origin</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.GitRepo">
GitRepo
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.Subject">Subject
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SystemSpec">SystemSpec</a>)
</p>
<div>
<p>Subject represents a subject which has been granted access to the system.
The subject is assigned the roles set in the controller configuration file.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>kind</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.SubjectKind">
SubjectKind
</a>
</em>
</td>
<td>
<p>Kind is the SubjectKind of the subject.</p>
</td>
</tr>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the subject. The meaning of this field depends on the
SubjectKind.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.SubjectKind">SubjectKind
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.Subject">Subject</a>)
</p>
<div>
<p>SubjectKind represents a kind of a subject.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;group&#34;</p></td>
<td><p>SubjectKindGroup is the subject kind group.</p>
</td>
</tr><tr><td><p>&#34;user&#34;</p></td>
<td><p>SubjectKindUser is the subject kind user.</p>
</td>
</tr></tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.System">System
</h3>
<div>
<p>System is the Schema for the Systems API.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-20.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">
k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.SystemSpec">
SystemSpec
</a>
</em>
</td>
<td>
<p>Spec is the specification of the System resource.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>deletionProtection</code><br/>
<em>
bool
</em>
</td>
<td>
<p>DeletionProtection disables deletion of the system in Styra, when the
System resource is deleted.</p>
</td>
</tr>
<tr>
<td>
<code>enableDeltaBundles</code><br/>
<em>
bool
</em>
</td>
<td>
<p>EnableDeltaBundles decides whether DeltaBundles are enabled</p>
</td>
</tr>
<tr>
<td>
<code>subjects</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.Subject">
[]Subject
</a>
</em>
</td>
<td>
<p>Subjects is the list of subjects which should have access to the system.</p>
</td>
</tr>
<tr>
<td>
<code>decisionMappings</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.DecisionMapping">
[]DecisionMapping
</a>
</em>
</td>
<td>
<p>DecisionMappings holds the list of decision mappings for the system.</p>
</td>
</tr>
<tr>
<td>
<code>datasources</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.Datasource">
[]Datasource
</a>
</em>
</td>
<td>
<p>Datasources represents a list of Styra datasources to be mounted in the
system.</p>
</td>
</tr>
<tr>
<td>
<code>discoveryOverrides</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.DiscoveryOverrides">
DiscoveryOverrides
</a>
</em>
</td>
<td>
<p>DiscoveryOverrides is an opa config which will take precedence over the
configuration supplied by Styra discovery API. Configuration set here
will be merged with the configuration supplied by the discovery API.</p>
</td>
</tr>
<tr>
<td>
<code>sourceControl</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.SourceControl">
SourceControl
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>localPlane</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.LocalPlane">
LocalPlane
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>customOPAConfig</code><br/>
<em>
k8s.io/apimachinery/pkg/runtime.RawExtension
</em>
</td>
<td>
<p>CustomOPAConfig allows the owner of a System resource to set custom features
without having to extend the Controller</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.SystemStatus">
SystemStatus
</a>
</em>
</td>
<td>
<p>Status is the status of the System resource.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.SystemPhase">SystemPhase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.SystemStatus">SystemStatus</a>)
</p>
<div>
<p>SystemPhase is a status phase of the System.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Created&#34;</p></td>
<td><p>SystemPhaseCreated is a SystemPhase used when the System is fully
reconciled.</p>
</td>
</tr><tr><td><p>&#34;Failed&#34;</p></td>
<td><p>SystemPhaseFailed is a SystemPhase used when the System failed to
reconcile.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>SystemPhasePending is a SystemPhase used when the System has not yet been
reconciled.</p>
</td>
</tr></tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.SystemSpec">SystemSpec
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.System">System</a>)
</p>
<div>
<p>SystemSpec is the specification of the System resource.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>deletionProtection</code><br/>
<em>
bool
</em>
</td>
<td>
<p>DeletionProtection disables deletion of the system in Styra, when the
System resource is deleted.</p>
</td>
</tr>
<tr>
<td>
<code>enableDeltaBundles</code><br/>
<em>
bool
</em>
</td>
<td>
<p>EnableDeltaBundles decides whether DeltaBundles are enabled</p>
</td>
</tr>
<tr>
<td>
<code>subjects</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.Subject">
[]Subject
</a>
</em>
</td>
<td>
<p>Subjects is the list of subjects which should have access to the system.</p>
</td>
</tr>
<tr>
<td>
<code>decisionMappings</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.DecisionMapping">
[]DecisionMapping
</a>
</em>
</td>
<td>
<p>DecisionMappings holds the list of decision mappings for the system.</p>
</td>
</tr>
<tr>
<td>
<code>datasources</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.Datasource">
[]Datasource
</a>
</em>
</td>
<td>
<p>Datasources represents a list of Styra datasources to be mounted in the
system.</p>
</td>
</tr>
<tr>
<td>
<code>discoveryOverrides</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.DiscoveryOverrides">
DiscoveryOverrides
</a>
</em>
</td>
<td>
<p>DiscoveryOverrides is an opa config which will take precedence over the
configuration supplied by Styra discovery API. Configuration set here
will be merged with the configuration supplied by the discovery API.</p>
</td>
</tr>
<tr>
<td>
<code>sourceControl</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.SourceControl">
SourceControl
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>localPlane</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.LocalPlane">
LocalPlane
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>customOPAConfig</code><br/>
<em>
k8s.io/apimachinery/pkg/runtime.RawExtension
</em>
</td>
<td>
<p>CustomOPAConfig allows the owner of a System resource to set custom features
without having to extend the Controller</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1beta1.SystemStatus">SystemStatus
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1beta1.System">System</a>)
</p>
<div>
<p>SystemStatus defines the observed state of System.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>id</code><br/>
<em>
string
</em>
</td>
<td>
<p>ID is the system ID in Styra.</p>
</td>
</tr>
<tr>
<td>
<code>ready</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Ready is true when the system is created and in sync.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.SystemPhase">
SystemPhase
</a>
</em>
</td>
<td>
<p>Phase is the current state of syncing the system.</p>
</td>
</tr>
<tr>
<td>
<code>failureMessage</code><br/>
<em>
string
</em>
</td>
<td>
<p>Failure message holds a message when Phase is Failed.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#styra.bankdata.dk/v1beta1.Condition">
[]Condition
</a>
</em>
</td>
<td>
<p>Conditions holds a list of Condition which describes the state of the
System.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>f1136ff2</code>.
</em></p>
