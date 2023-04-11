<p>Packages:</p>
<ul>
<li>
<a href="#styra.bankdata.dk%2fv1alpha1">styra.bankdata.dk/v1alpha1</a>
</li>
</ul>
<h2 id="styra.bankdata.dk/v1alpha1">styra.bankdata.dk/v1alpha1</h2>
<div>
<p>Package v1alpha1 contains API Schema definitions for the styra v1alpha1 API
group.</p>
</div>
Resource Types:
<ul></ul>
<h3 id="styra.bankdata.dk/v1alpha1.GlobalDatasource">GlobalDatasource
</h3>
<div>
<p>GlobalDatasource is a resource used for creating global datasources in
Styra. These datasources are available across systems and can be used for
shared data. GlobalDatasource can also be used to create libraries by using
the GlobalDatasourceCategoryGitRego category.</p>
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
<a href="#styra.bankdata.dk/v1alpha1.GlobalDatasourceSpec">
GlobalDatasourceSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>category</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.GlobalDatasourceCategory">
GlobalDatasourceCategory
</a>
</em>
</td>
<td>
<p>Category is the datasource category. For more information about
supported categories see the available GlobalDatasourceCategory
constants in the package.</p>
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
<p>Description describes the datasource.</p>
</td>
</tr>
<tr>
<td>
<code>enabled</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enabled toggles whether or not the datasource should be enabled.</p>
</td>
</tr>
<tr>
<td>
<code>pollingInterval</code><br/>
<em>
string
</em>
</td>
<td>
<p>PollingInterval sets the interval for when the datasource should be refreshed.</p>
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
<p>Commit is a commit SHA for the git/xx datasources. If <code>Reference</code> and this
is set, this takes precedence.</p>
</td>
</tr>
<tr>
<td>
<code>credentialsSecretRef</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.GlobalDatasourceSecretRef">
GlobalDatasourceSecretRef
</a>
</em>
</td>
<td>
<p>CredentialsSecretRef references a secret with keys <code>name</code> and <code>secret</code>
which will be used for authentication.</p>
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
<p>Reference is a git reference for the git/xx datasources.</p>
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
<p>URL is used in http and git/xx datasources.</p>
</td>
</tr>
<tr>
<td>
<code>deletionProtection</code><br/>
<em>
bool
</em>
</td>
<td>
<p>DeletionProtection skips deleting the datasource in Styra when the
<code>GlobalDatasource</code> resource is deleted.</p>
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
<p>Path is the path in git in git/xx datasources.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.GlobalDatasourceStatus">
GlobalDatasourceStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1alpha1.GlobalDatasourceCategory">GlobalDatasourceCategory
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.GlobalDatasourceSpec">GlobalDatasourceSpec</a>)
</p>
<div>
<p>GlobalDatasourceCategory represents a datasource category.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;git/rego&#34;</p></td>
<td><p>GlobalDatasourceCategoryGitRego represents the git/rego datasource category.</p>
</td>
</tr></tbody>
</table>
<h3 id="styra.bankdata.dk/v1alpha1.GlobalDatasourceSecretRef">GlobalDatasourceSecretRef
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.GlobalDatasourceSpec">GlobalDatasourceSpec</a>)
</p>
<div>
<p>GlobalDatasourceSecretRef represents a reference to a secret.</p>
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
<code>namespace</code><br/>
<em>
string
</em>
</td>
<td>
<p>Namespace is the namespace where the secret resides.</p>
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
<p>Name is the name of the secret.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1alpha1.GlobalDatasourceSpec">GlobalDatasourceSpec
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.GlobalDatasource">GlobalDatasource</a>)
</p>
<div>
<p>GlobalDatasourceSpec is the specification of the GlobalDatasource.</p>
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
<code>category</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.GlobalDatasourceCategory">
GlobalDatasourceCategory
</a>
</em>
</td>
<td>
<p>Category is the datasource category. For more information about
supported categories see the available GlobalDatasourceCategory
constants in the package.</p>
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
<p>Description describes the datasource.</p>
</td>
</tr>
<tr>
<td>
<code>enabled</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enabled toggles whether or not the datasource should be enabled.</p>
</td>
</tr>
<tr>
<td>
<code>pollingInterval</code><br/>
<em>
string
</em>
</td>
<td>
<p>PollingInterval sets the interval for when the datasource should be refreshed.</p>
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
<p>Commit is a commit SHA for the git/xx datasources. If <code>Reference</code> and this
is set, this takes precedence.</p>
</td>
</tr>
<tr>
<td>
<code>credentialsSecretRef</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.GlobalDatasourceSecretRef">
GlobalDatasourceSecretRef
</a>
</em>
</td>
<td>
<p>CredentialsSecretRef references a secret with keys <code>name</code> and <code>secret</code>
which will be used for authentication.</p>
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
<p>Reference is a git reference for the git/xx datasources.</p>
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
<p>URL is used in http and git/xx datasources.</p>
</td>
</tr>
<tr>
<td>
<code>deletionProtection</code><br/>
<em>
bool
</em>
</td>
<td>
<p>DeletionProtection skips deleting the datasource in Styra when the
<code>GlobalDatasource</code> resource is deleted.</p>
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
<p>Path is the path in git in git/xx datasources.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1alpha1.GlobalDatasourceStatus">GlobalDatasourceStatus
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.GlobalDatasource">GlobalDatasource</a>)
</p>
<div>
<p>GlobalDatasourceStatus holds the status of the GlobalDatasource resource.</p>
</div>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>ef394ca</code>.
</em></p>
