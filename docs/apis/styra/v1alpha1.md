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
<h3 id="styra.bankdata.dk/v1alpha1.GitRepo">GitRepo
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.SourceControl">SourceControl</a>)
</p>
<div>
<p>GitRepo defines the Git configurations a library can be defined by</p>
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
<h3 id="styra.bankdata.dk/v1alpha1.Library">Library
</h3>
<div>
<p>Library is the Schema for the libraries API</p>
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
<a href="#styra.bankdata.dk/v1alpha1.LibrarySpec">
LibrarySpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name is the name the Library will have in Styra DAS</p>
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
<p>Description is the description of the Library</p>
</td>
</tr>
<tr>
<td>
<code>subjects</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.LibrarySubject">
[]LibrarySubject
</a>
</em>
</td>
<td>
<p>Subjects is the list of subjects which should have access to the system.</p>
</td>
</tr>
<tr>
<td>
<code>sourceControl</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.SourceControl">
SourceControl
</a>
</em>
</td>
<td>
<p>SourceControl is the sourcecontrol configuration for the Library</p>
</td>
</tr>
<tr>
<td>
<code>datasources</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.LibraryDatasource">
[]LibraryDatasource
</a>
</em>
</td>
<td>
<p>Datasources is the list of datasources in the Library</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.LibraryStatus">
LibraryStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1alpha1.LibraryDatasource">LibraryDatasource
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.LibrarySpec">LibrarySpec</a>)
</p>
<div>
<p>LibraryDatasource contains metadata of a datasource, stored in a library</p>
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
<h3 id="styra.bankdata.dk/v1alpha1.LibrarySecretRef">LibrarySecretRef
</h3>
<div>
<p>LibrarySecretRef defines how to access a k8s secret for the library.</p>
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
<h3 id="styra.bankdata.dk/v1alpha1.LibrarySpec">LibrarySpec
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.Library">Library</a>)
</p>
<div>
<p>LibrarySpec defines the desired state of Library</p>
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
<p>Name is the name the Library will have in Styra DAS</p>
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
<p>Description is the description of the Library</p>
</td>
</tr>
<tr>
<td>
<code>subjects</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.LibrarySubject">
[]LibrarySubject
</a>
</em>
</td>
<td>
<p>Subjects is the list of subjects which should have access to the system.</p>
</td>
</tr>
<tr>
<td>
<code>sourceControl</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.SourceControl">
SourceControl
</a>
</em>
</td>
<td>
<p>SourceControl is the sourcecontrol configuration for the Library</p>
</td>
</tr>
<tr>
<td>
<code>datasources</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.LibraryDatasource">
[]LibraryDatasource
</a>
</em>
</td>
<td>
<p>Datasources is the list of datasources in the Library</p>
</td>
</tr>
</tbody>
</table>
<h3 id="styra.bankdata.dk/v1alpha1.LibraryStatus">LibraryStatus
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.Library">Library</a>)
</p>
<div>
<p>LibraryStatus defines the observed state of Library</p>
</div>
<h3 id="styra.bankdata.dk/v1alpha1.LibrarySubject">LibrarySubject
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.LibrarySpec">LibrarySpec</a>)
</p>
<div>
<p>LibrarySubject represents a subject which has been granted access to the Library.
The subject is assigned to the LibraryViewer role.</p>
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
<a href="#styra.bankdata.dk/v1alpha1.LibrarySubjectKind">
LibrarySubjectKind
</a>
</em>
</td>
<td>
<p>Kind is the LibrarySubjectKind of the subject.</p>
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
<h3 id="styra.bankdata.dk/v1alpha1.LibrarySubjectKind">LibrarySubjectKind
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.LibrarySubject">LibrarySubject</a>)
</p>
<div>
<p>LibrarySubjectKind represents a kind of a subject.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;group&#34;</p></td>
<td><p>LibrarySubjectKindGroup is the subject kind group.</p>
</td>
</tr><tr><td><p>&#34;user&#34;</p></td>
<td><p>LibrarySubjectKindUser is the subject kind user.</p>
</td>
</tr></tbody>
</table>
<h3 id="styra.bankdata.dk/v1alpha1.SourceControl">SourceControl
</h3>
<p>
(<em>Appears on:</em><a href="#styra.bankdata.dk/v1alpha1.LibrarySpec">LibrarySpec</a>)
</p>
<div>
<p>SourceControl is a struct from styra where we only use a single field
but kept for clarity when comparing to the API</p>
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
<code>libraryOrigin</code><br/>
<em>
<a href="#styra.bankdata.dk/v1alpha1.GitRepo">
GitRepo
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>7c22e6b</code>.
</em></p>
