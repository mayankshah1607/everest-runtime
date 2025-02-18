package v2alpha1

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=db;dbc;dbcluster
type DatabaseCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseClusterSpec   `json:"spec,omitempty"`
	Status DatabaseClusterStatus `json:"status,omitempty"`
}

type DatabaseClusterSpec struct {
	Plugin string `json:"plugin,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Global     *runtime.RawExtension `json:"global,omitempty"`
	Components []ComponentSpec       `json:"components,omitempty"`
}

func (db *DatabaseCluster) GetComponentsOfType(t string) []ComponentSpec {
	var result []ComponentSpec
	for _, c := range db.Spec.Components {
		if c.Type == t {
			result = append(result, c)
		}
	}
	return result
}

type DatabaseClusterPhase string

const (
	DatabaseClusterPhaseCreating DatabaseClusterPhase = "Creating"
	DatabaseClusterPhaseRunning  DatabaseClusterPhase = "Running"
	DatabaseClusterPhaseFailed   DatabaseClusterPhase = "Failed"
	DatabaseClusterPhaseDeleting DatabaseClusterPhase = "Deleting"
)

type DatabaseClusterStatus struct {
	// Phase of the database cluster.
	Phase DatabaseClusterPhase `json:"phase,omitempty"`
	// ConnectionURL is the URL to connect to the database cluster.
	ConnectionURL string `json:"connectionURL,omitempty"`
	// CredentialSecretRef is a reference to the secret containing the credentials.
	// This Secret contains the keys `username` and `password`.
	CredentialSecretRef corev1.LocalObjectReference `json:"credentialSecretRef,omitempty"`
	// Components is the status of the components in the database cluster.
	Components []ComponentStatus `json:"components,omitempty"`
	// TODO: more fields
}

const (
	StateReady      = "Ready"
	StateInProgress = "InProgress"
	StateError      = "Error"
)

type ComponentStatus struct {
	Pods  []corev1.LocalObjectReference `json:"pods,omitempty"`
	Total *int32                        `json:"total,omitempty"`
	Ready *int32                        `json:"ready,omitempty"`
	State string                        `json:"state,omitempty"`
}

type CustomOptions map[string]json.RawMessage

type ComponentSpec struct {
	// Name of the component.
	Name string `json:"name,omitempty"`
	// Type of the component from DatabaseClusterDefinition.
	Type string `json:"type,omitempty"`
	// Version of the component from ComponentVersions.
	Version string `json:"version,omitempty"`
	// Image specifies an override for the image to use.
	// When unspecified, it is autmatically set from the ComponentVersions
	// based on the Version specified.
	// +optional
	Image string `json:"image,omitempty"`
	// Storage requirements for this component.
	// For stateless components, this is an optional field.
	// +optional
	Storage *Storage `json:"storage,omitempty"`
	// Resources requirements for this component.
	// +optional
	Resources *Resources `json:"resources,omitempty"`
	// Config specifies the component specific configuration.
	// +optional
	Config *Config `json:"config,omitempty"`
	// Replicas specifies the number of replicas for this component.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Shards specifies the number of shards for this component.
	// +optional
	Shards *int32 `json:"shards,omitempty"`

	// TODO: TLS settings

	// +kubebuilder:pruning:PreserveUnknownFields
	// CustomSpec provides an API for customising this component.
	// The API is defined in the DatabaseCluserDefinition.
	CustomSpec *runtime.RawExtension `json:"customSpec,omitempty"`

	// Internal field, this is not provided in the CRD.
	// This field is automatically populated during runtime.
	// This contains the combined information about the component pod
	// from the defaults in the DatabaseClusterDefinition.
	PodSpec *ComponentPodSpec `json:"-,omitempty"`
}

type Config struct {
	SecretRef    corev1.LocalObjectReference `json:"secretRef,omitempty"`
	ConfigMapRef corev1.LocalObjectReference `json:"configMapRef,omitempty"`
	Key          string                      `json:"key,omitempty"`
}

type Storage struct {
	Size         resource.Quantity `json:"size,omitempty"`
	StorageClass *string           `json:"storageClass,omitempty"`
}

type Resources struct {
	CPU    resource.Quantity `json:"cpu,omitempty"`
	Memory resource.Quantity `json:"memory,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseClusterList contains a list of DatabaseCluster.
type DatabaseClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseCluster{}, &DatabaseClusterList{})
}
