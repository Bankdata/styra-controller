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
	"path"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// SystemSpec is the specification of the System resource.
type SystemSpec struct {
	// DeletionProtection disables deletion of the system in Styra, when the
	// System resource is deleted.
	DeletionProtection *bool `json:"deletionProtection,omitempty"`

	// EnableDeltaBundles decides whether DeltaBundles are enabled
	EnableDeltaBundles *bool `json:"enableDeltaBundles,omitempty"`

	// Subjects is the list of subjects which should have access to the system.
	Subjects []Subject `json:"subjects,omitempty"`

	// DecisionMappings holds the list of decision mappings for the system.
	DecisionMappings []DecisionMapping `json:"decisionMappings,omitempty"`

	// Datasources represents a list of Styra datasources to be mounted in the
	// system.
	Datasources []Datasource `json:"datasources,omitempty"`

	// DiscoveryOverrides is an opa config which will take precedence over the
	// configuration supplied by Styra discovery API. Configuration set here
	// will be merged with the configuration supplied by the discovery API.
	DiscoveryOverrides *DiscoveryOverrides `json:"discoveryOverrides,omitempty"`

	SourceControl *SourceControl `json:"sourceControl,omitempty"`
	LocalPlane    *LocalPlane    `json:"localPlane,omitempty"`

	// CustomOPAConfig allows the owner of a System resource to set custom features
	// without having to extend the Controller
	CustomOPAConfig *runtime.RawExtension `json:"customOPAConfig,omitempty"`
}

// DiscoveryOverrides specifies system specific overrides for the configuration
// served from the Styra OPA Discovery API
type DiscoveryOverrides struct {
	Status             *OPAConfigStatus             `json:"status"`
	DistributedTracing *OPAConfigDistributedTracing `json:"distributed_tracing,omitempty"`
}

// OPAConfigStatus configures the `status` key in the OPA configuration
type OPAConfigStatus struct {
	Prometheus bool `json:"prometheus"`
}

// OPAConfigDistributedTracing configures the `distributed_tracing` key in the
// OPA configuration.
type OPAConfigDistributedTracing struct {
	Type             string `json:"type,omitempty"`
	Address          string `json:"address,omitempty"`
	ServiceName      string `json:"service_name,omitempty"`
	SamplePercentage int    `json:"sample_percentage,omitempty"`
	//+kubebuilder:validation:Enum=off;tls;mtls
	Encryption        string `json:"encryption,omitempty"`
	AllowInsecureTLS  bool   `json:"allow_insecure_tls,omitempty"`
	TLSCACertFile     string `json:"tls_ca_cert_file,omitempty"`
	TLSCertFile       string `json:"tls_cert_file,omitempty"`
	TLSPrivateKeyFile string `json:"tls_private_key_file,omitempty"`
}

// LocalPlane specifies how the Styra Local Plane should be configured. This is
// used to generate Secret and ConfigMap for the SLP to consume.
type LocalPlane struct {
	// Name is the hostname of the SLP service.
	Name string `json:"name"`
}

// SubjectKind represents a kind of a subject.
type SubjectKind string

const (
	// SubjectKindUser is the subject kind user.
	SubjectKindUser SubjectKind = "user"

	// SubjectKindGroup is the subject kind group.
	SubjectKindGroup SubjectKind = "group"
)

// Subject represents a subject which has been granted access to the system.
// The subject is assigned the roles set in the controller configuration file.
type Subject struct {
	// Kind is the SubjectKind of the subject.
	//+kubebuilder:validation:Enum=user;group
	Kind SubjectKind `json:"kind,omitempty"`

	// Name is the name of the subject. The meaning of this field depends on the
	// SubjectKind.
	Name string `json:"name"`
}

// IsUser returns whether or not the kind of the subject is a user.
func (subject Subject) IsUser() bool {
	return subject.Kind == SubjectKindUser || subject.Kind == ""
}

// DecisionMapping specifies how a system decision mapping should be
// configured. This allows configuration of when a decision is considered
// allowed or not. It also provides the ability to show additional columns in
// Styra.
type DecisionMapping struct {
	// Name is the name of the decision mapping.
	//+kubebuilder:validation:Optional
	Name string `json:"name"`

	// Columns holds a list of ColumnMapping for the decision mapping.
	Columns []ColumnMapping `json:"columns,omitempty"`

	//+kubebuilder:validation:Optional
	Reason ReasonMapping `json:"reason,omitempty"`

	Allowed *AllowedMapping `json:"allowed,omitempty"`
}

// AllowedMapping specifies how to determine if a decision is allowed or not.
type AllowedMapping struct {
	// Expected is the value we expect to be set in the Path in order to consider
	// the decision allowed.
	Expected *Expected `json:"expected,omitempty"`

	// Negated negates the expectation.
	//+kubebuilder:validation:Optional
	Negated bool `json:"negated,omitempty"`

	// Path is the path to the value which we check our expectation against.
	Path string `json:"path"`
}

// Expected represents an expected value. When using this type only one of the
// fields should be set.
type Expected struct {
	// String holds a pointer to a string if the Expected value represents a
	// string.
	//+kubebuilder:validation:Optional
	String *string `json:"string,omitempty"`

	// Boolean holds a pointer to a bool if the Expected value represents a
	// bool.
	//+kubebuilder:validation:Optional
	Boolean *bool `json:"boolean,omitempty"`

	// Integer holds a pointer to an int if the Expected value represents an int.
	//+kubebuilder:validation:Optional
	Integer *int `json:"integer,omitempty"`
}

// Value returns the value of an Expected type. It is either a string, boolean,
// or an integer.
func (e Expected) Value() interface{} {
	switch {
	case e.String != nil && e.Boolean == nil && e.Integer == nil:
		return *e.String
	case e.String == nil && e.Boolean != nil && e.Integer == nil:
		return *e.Boolean
	case e.String == nil && e.Boolean == nil && e.Integer != nil:
		return *e.Integer
	default:
		return true
	}
}

// ColumnMapping specifies how a value in the decision result should be mapped
// to a column in the Styra decision log.
type ColumnMapping struct {
	// Key is the name of the column as shown in the decision log.
	Key string `json:"key"`

	// Path is where in the decision result the value for the column is found.
	Path string `json:"path"`
}

// ReasonMapping specifies where the reason of the decision can be found.
type ReasonMapping struct {
	// Path is the path to where the reason is found in the decision result.
	Path string `json:"path,omitempty"`
}

// SourceControl holds SourceControl configuration.
type SourceControl struct {
	Origin GitRepo `json:"origin"`
}

// GitRepo specifies the configuration for how to pull policy from git.
type GitRepo struct {
	// CredentialsSecretName is a reference to an existing secret which holds git
	// credentials. This secret should have the keys `name` and `secret`. The
	// `name` key should contain the http basic auth username and the `secret`
	// key should contain the http basic auth password.
	CredentialsSecretName string `json:"credentialsSecretName,omitempty"`

	// Path is the path in the git repo where the policies are located.
	Path string `json:"path,omitempty"`

	// Reference is used to point to a tag or branch. This will be ignored if
	// `Commit` is specified.
	Reference string `json:"reference,omitempty"`

	// Commit is used to point to a specific commit SHA. This takes precedence
	// over `Reference` if both are specified.
	Commit string `json:"commit,omitempty"`

	// URL is the URL of the git repo.
	URL string `json:"url"`
}

// Datasource represents a Styra datasource to be mounted in the system.
type Datasource struct {
	// Path is the path within the system where the datasource should reside.
	Path string `json:"path"`

	// Description is a description of the datasource
	Description string `json:"description,omitempty"`
}

// SystemStatus defines the observed state of System.
type SystemStatus struct {
	// ID is the system ID in Styra.
	ID string `json:"id,omitempty"`

	// Ready is true when the system is created and in sync.
	Ready bool `json:"ready"`

	// Phase is the current state of syncing the system.
	//+kubebuilder:default=Pending
	//+kubebuilder:validation:Enum=Pending;Failed;Created
	Phase SystemPhase `json:"phase,omitempty"`

	// Failure message holds a message when Phase is Failed.
	FailureMessage string `json:"failureMessage,omitempty"`

	// Conditions holds a list of Condition which describes the state of the
	// System.
	Conditions []Condition `json:"conditions,omitempty"`
}

// SystemPhase is a status phase of the System.
type SystemPhase string

const (
	// SystemPhasePending is a SystemPhase used when the System has not yet been
	// reconciled.
	SystemPhasePending SystemPhase = "Pending"

	// SystemPhaseFailed is a SystemPhase used when the System failed to
	// reconcile.
	SystemPhaseFailed SystemPhase = "Failed"

	// SystemPhaseCreated is a SystemPhase used when the System is fully
	// reconciled.
	SystemPhaseCreated SystemPhase = "Created"
)

// Condition represents a System condition.
type Condition struct {
	// Type is the ConditionType of the Condition.
	Type ConditionType `json:"type"`

	// Status is the status of the Condition.
	Status metav1.ConditionStatus `json:"status"`

	// LastProbeTime is a timestamp for the last time the condition was checked.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`

	// LastTransitionTime is a timestamp for the last time that the condition
	// changed state.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// ConditionType is a System Condition type.
type ConditionType string

const (
	// ConditionTypeCreatedInStyra is a ConditionType used when the system has
	// been created in Styra.
	ConditionTypeCreatedInStyra ConditionType = "CreatedInStyra"

	// ConditionTypeGitCredentialsUpdated is a ConditionType used when git
	// credentials are updated in Styra.
	ConditionTypeGitCredentialsUpdated ConditionType = "GitCredentialsUpdated"

	// ConditionTypeSubjectsUpdated is a ConditionType used when the subjects of
	// the System are updated in Styra.
	ConditionTypeSubjectsUpdated ConditionType = "SubjectsUpdated"

	// ConditionTypeDatasourcesUpdated is a ConditionType used when
	// the datasources of the System are updated in Styra.
	ConditionTypeDatasourcesUpdated ConditionType = "DatasourcesUpdated"

	// ConditionTypeOPATokenUpdated is a ConditionType used when
	// the secret with the Styra token has been updated in the cluster.
	ConditionTypeOPATokenUpdated ConditionType = "OPATokenUpdated"

	// ConditionTypeOPAConfigMapUpdated is a ConditionType used when
	// the ConfigMap for the OPA are updated in the cluster.
	ConditionTypeOPAConfigMapUpdated ConditionType = "OPAConfigMapUpdated"

	// ConditionTypeOPAUpToDate is a ConditionType used to say whether
	// the OPA is up to date or needs to be restarted.
	ConditionTypeOPAUpToDate ConditionType = "OPAUpToDate"

	// ConditionTypeSLPConfigMapUpdated is a ConditionType used when
	// the ConfigMap for the SLP are updated in the cluster.
	ConditionTypeSLPConfigMapUpdated ConditionType = "SLPConfigMapUpdated"

	// ConditionTypeSLPUpToDate is a ConditionType used to say whether
	// the SLP is up to date or needs to be restarted.
	ConditionTypeSLPUpToDate ConditionType = "SLPUpToDate"

	// ConditionTypeSystemConfigUpdated is a ConditionType used when
	// the configuration of the System are updated in Styra.
	ConditionTypeSystemConfigUpdated ConditionType = "SystemConfigUpdated"

	// ConditionTypeRequirementsUpdated is a ConditionType used when
	// the requirements of for the System's bundle is updated in OCP.
	ConditionTypeRequirementsUpdated ConditionType = "RequirementsUpdated"

	// ConditionTypeSystemSourceUpdated is a ConditionType used when
	// the source for the System is updated in OCP.
	ConditionTypeSystemSourceUpdated ConditionType = "SystemSourceUpdated"

	// ConditionTypeSystemBundleUpdated is a ConditionType used when
	// the bundle for the System is updated in OCP.
	ConditionTypeSystemBundleUpdated ConditionType = "SystemBundleUpdated"

	// ConditionTypeOPASecretUpdated is a ConditionType used when
	// the OPA secret for the System is updated in the cluster.
	ConditionTypeOPASecretUpdated ConditionType = "OPASecretUpdated"
)

// EventType is a type of event which can be emitted by the System controller.
type EventType string

const (
	// EventErrorSetFinalizer is an EventType used when the controller fails to set
	// the finalizer on the System resource.
	EventErrorSetFinalizer EventType = "ErrorSetFinalizer"

	// EventErrorDeleteSystemInStyra is an EventType used when the controller fails
	// to delete the System in Styra.
	EventErrorDeleteSystemInStyra EventType = "ErrorDeleteSystemInStyra"

	// EventErrorRemovingFinalizer is an EventType used when the controller fails to
	// remove the finalizer from the System resource.
	EventErrorRemovingFinalizer EventType = "ErrorRemovingFinalizer"

	// EventErrorFetchOPAConfig is an EventType used when the controller fails to fetch
	// the OPA configuration from Styra.
	EventErrorFetchOPAConfig EventType = "ErrorFetchOPAConfig"

	// EventErrorUpdateStatus is an EventType used when the controller fails to update
	// the status of the System resource.
	EventErrorUpdateStatus EventType = "ErrorUpdateStatus"

	// EventErrorRestartSLPs is an EventType used when the controller fails to restart the SLPs.
	EventErrorRestartSLPs EventType = "ErrorRestartSLPs"

	// EventErrorPhaseToCreated is an EventType used when the controller fails to set the
	// phase of the System resource to Created.
	EventErrorPhaseToCreated EventType = "ErrorPhaseToCreated"

	// EventErrorGetStatefulSet is an EventType used when the controller fails to get the StatefulSet
	// for the SLP used by the System.
	EventErrorGetStatefulSet EventType = "ErrorGetStatefulSet"

	// EventErrorStatefulSetNotFound is an EventType used when a system with 'localPlane' enabled but which
	// does not have a StatefulSet created for the SLP.
	EventErrorStatefulSetNotFound EventType = "ErrorStatefulSetNotFound"

	// EventErrorPatchStatefulSet is an EventType used when the controller fails to patch the StatefulSet
	// for the SLP used by the System.
	EventErrorPatchStatefulSet EventType = "ErrorPatchStatefulSet"

	// EventErrorFetchSystemFromStyra is an EventType used when the controller fails to fetch the System from Styra.
	EventErrorFetchSystemFromStyra EventType = "ErrorFetchSystemFromStyra"

	// EventErrorCreateSystemInStyra is an EventType used when the controller fails to create the System in Styra.
	EventErrorCreateSystemInStyra EventType = "ErrorCreateSystemInStyra"

	// EventErrorCredentialsSecretNotFound is an EventType used when the controller gets a 404 when fetching
	// secret referenced by the System resource under Spec.SourceControl.Origin.CredentialsSecretName.
	EventErrorCredentialsSecretNotFound EventType = "ErrorCredentialsSecretNotFound"

	// EventErrorCredentialsSecretCouldNotFetch is an EventType used when the controller fails to fetch the
	// secret referenced by the System resource under Spec.SourceControl.Origin.CredentialsSecretName.
	EventErrorCredentialsSecretCouldNotFetch EventType = "ErrorCredentialsSecretCouldNotFetch"

	// EventErrorCreateUpdateSecret is an EventType used when the controller fails to create or update the
	// Styra secret containing the Git credentials used to access Git.
	EventErrorCreateUpdateSecret EventType = "ErrorCreateUpdateSecret"

	// EventErrorDeleteDefaultPolicy is an EventType used when the controller fails to delete the default policy
	// in the System in Styra.
	EventErrorDeleteDefaultPolicy EventType = "ErrorDeleteDefaultPolicy"

	// EventErrorReconcileID is an EventType used when the controller fails to reconcile the ID for the System.
	EventErrorReconcileID EventType = "ErrorReconcileID"

	// EventErrorGetUsersFromStyra is an EventType used when the controller fails to get the users
	// for a system in Styra.
	EventErrorGetUsersFromStyra EventType = "ErrorGetUsersFromStyra"

	// EventErrorCreateInvitation is an EventType used when the controller fails to create an invitation
	// for a new user in Styra.
	EventErrorCreateInvitation EventType = "ErrorCreateInvitation"

	// EventErrorGetSystemRolebindings is an EventType used when the controller fails to get the rolebindings
	// for a system in Styra.
	EventErrorGetSystemRolebindings EventType = "ErrorGetSystemRolebindings"

	// EventErrorCreateRolebinding is an EventType used when the controller fails to create a rolebinding
	// for a user in Styra.
	EventErrorCreateRolebinding EventType = "ErrorCreateRolebinding"

	// EventErrorUpdateRolebinding is an EventType used when the controller fails to update a rolebinding
	// for a user in Styra.
	EventErrorUpdateRolebinding EventType = "ErrorUpdateRolebinding"

	// EventErrorUpsertDatasource is an EventType used when the controller fails to upsert a datasource in Styra.
	EventErrorUpsertDatasource EventType = "ErrorUpsertDatasource"

	// EventErrorCallWebhook is an EventType used when the controller fails to call the datasource changed webhook.
	EventErrorCallWebhook EventType = "ErrorCallWebhook"

	// EventErrorDeleteDatasource is an EventType used when the controller fails to delete a datasource in Styra.
	EventErrorDeleteDatasource EventType = "ErrorDeleteDatasource"

	// EventErrorOPATokenSecretNoToken is an EventType used when the controller creates the OPA token Secret
	// but the downloaded OPA config from Styra does not contain a token.
	EventErrorOPATokenSecretNoToken EventType = "ErrorOPATokenSecretNoToken"

	// EventErrorOwnerRefOPATokenSecret is an EventType used when the controller fails to set the owner reference
	// on the OPA token secret.
	EventErrorOwnerRefOPATokenSecret EventType = "ErrorOwnerRefOPATokenSecret"

	// EventErrorCreateOPATokenSecret is an EventType used when the controller fails to create the OPA token Secret.
	EventErrorCreateOPATokenSecret EventType = "ErrorCreateOPATokenSecret"

	// EventErrorFetchOPATokenSecret is an EventType used when the controller fails to fetch the OPA token Secret.
	EventErrorFetchOPATokenSecret EventType = "ErrorFetchOPATokenSecret"

	// EventErrorSecretNotOwnedByController is an EventType used when the controller tries to update a Secret
	// that is not owned by the controller.
	EventErrorSecretNotOwnedByController EventType = "ErrorSecretNotOwnedByController"

	// EventErrorUpdateOPATokenSecret is an EventType used when the controller fails to update the OPA token Secret.
	EventErrorUpdateOPATokenSecret EventType = "ErrorUpdateOPATokenSecret"

	// EventErrorConvertOPAConf is an EventType used when the controller fails to convert the OPA config from Styra
	// to a ConfigMap for the OPA or the SLP.
	EventErrorConvertOPAConf EventType = "ErrorConvertOPAConfig"

	// EventErrorCreateOPAConfigMap is an EventType used when the controller fails to create the OPA ConfigMap.
	EventErrorCreateOPAConfigMap EventType = "ErrorCreateOPAConfigMap"

	// EventErrorFetchOPAConfigMap is an EventType used when the controller fails to fetch the OPA ConfigMap.
	EventErrorFetchOPAConfigMap EventType = "ErrorFetchOPAConfigMap"

	// EventErrorOwnerRefOPAConfigMap is an EventType used when the controller fails to set the owner reference
	// on the OPA config map.
	EventErrorOwnerRefOPAConfigMap EventType = "ErrorOwnerRefOPAConfigMap"

	// EventErrorConfigMapNotOwnedByController is an EventType used when the controller tries to update a ConfigMap
	// that is not owned by the controller.
	EventErrorConfigMapNotOwnedByController EventType = "ErrorConfigMapNotOwnedByController"

	// EventErrorUpdateOPAConfigMap is an EventType used when the controller fails to update the OPA ConfigMap.
	EventErrorUpdateOPAConfigMap EventType = "ErrorUpdateOPAConfigMap"

	// EventErrorUpdateOPASecret is an EventType used when the controller fails to update the OPA ConfigMap.
	EventErrorUpdateOPASecret EventType = "ErrorUpdateOPASecret"

	// EventErrorOwnerRefSLPConfigMap is an EventType used when the controller fails to set the owner reference
	// on the SLP ConfigMap.
	EventErrorOwnerRefSLPConfigMap EventType = "ErrorOwnerRefSLPConfigMap"

	// EventErrorCreateSLPConfigMap is an EventType used when the controller fails to create the SLP ConfigMap.
	EventErrorCreateSLPConfigMap EventType = "ErrorCreateSLPConfigMap"

	// EventErrorFetchSLPConfigMap is an EventType used when the controller fails to fetch the SLP ConfigMap.
	EventErrorFetchSLPConfigMap EventType = "ErrorFetchSLPConfigMap"

	// EventErrorUpdateSLPConfigmap is an EventType used when the controller fails to update the SLP ConfigMap.
	EventErrorUpdateSLPConfigmap EventType = "ErrorUpdateSLPConfigMap"

	// EventErrorUpdateSystem is an EventType used when the controller fails to update the System in Styra.
	EventErrorUpdateSystem EventType = "ErrorUpdateSystem"

	// EventErrorUpdateSource is an EventType used when the controller fails to update the Source in OCP.
	EventErrorUpdateSource EventType = "ErrorUpdateSource"

	// EventErrorUpdateBundle is an EventType used when the controller fails to update the Source in OCP.
	EventErrorUpdateBundle EventType = "ErrorUpdateBundle"

	// EventErrorDeleteBundleInOCP is an EventType used when the controller fails
	// to delete the System's Bundle in OCP.
	EventErrorDeleteBundleInOCP EventType = "ErrorDeleteBundleInOCP"

	// EventErrorDeleteSourceInOCP is an EventType used when the controller fails
	// to delete the System's Source in OCP.
	EventErrorDeleteSourceInOCP EventType = "ErrorDeleteSourceInOCP"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// System is the Schema for the Systems API.
type System struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the System resource.
	Spec SystemSpec `json:"spec,omitempty"`

	// Status is the status of the System resource.
	Status SystemStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SystemList represents a list of System resources.
type SystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []System `json:"items"`
}

func init() {
	SchemeBuilder.Register(&System{}, &SystemList{})
}

// SetCondition updates the matching condition under the System's status field.
func (s *System) SetCondition(conditionType ConditionType, status metav1.ConditionStatus) {
	s.setCondition(time.Now, conditionType, status)
}

// GetCondition gets the matching condition under the System's status field.
func (s *System) GetCondition(conditionType ConditionType) *metav1.ConditionStatus {
	for _, con := range s.Status.Conditions {
		if con.Type == conditionType {
			return &con.Status
		}
	}
	return nil
}

func (s *System) setCondition(timeNow func() time.Time, conditionType ConditionType, status metav1.ConditionStatus) {
	now := metav1.NewTime(timeNow())

	for i, con := range s.Status.Conditions {
		if con.Type != conditionType {
			continue
		}
		if con.Status != status {
			con.LastTransitionTime = now
			con.Status = status
		}
		con.LastProbeTime = now
		s.Status.Conditions[i] = con
		return
	}

	s.Status.Conditions = append(s.Status.Conditions, Condition{
		LastProbeTime:      now,
		LastTransitionTime: now,
		Status:             status,
		Type:               conditionType,
	})
}

// DisplayName returns the System's name with a prefix and suffix.
func (s *System) DisplayName(prefix, suffix string) string {
	return path.Join(prefix, s.Namespace, s.Name, suffix)
}

// OCPUniqueName returns the System's name with a prefix and suffix.
func (s *System) OCPUniqueName(prefix, suffix string) string {
	return strings.ReplaceAll(path.Join(prefix, s.Namespace, s.Name, suffix), "/", "-")
}

// GitSecretID returns the Styra internal ID of the Git Secret used by the
// System.
func (s *System) GitSecretID() string {
	return path.Join("systems", s.Status.ID, "git")
}
