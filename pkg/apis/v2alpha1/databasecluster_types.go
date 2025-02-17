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
	Phase               DatabaseClusterPhase        `json:"phase,omitempty"`
	ConnectionURL       string                      `json:"connectionURL,omitempty"`
	CredentialSecretRef corev1.LocalObjectReference `json:"credentialSecretRef,omitempty"`
	// TODO: more fields
}

type CustomOptions map[string]json.RawMessage

type ComponentSpec struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Version string `json:"version,omitempty"`
	// +optional
	Image     string     `json:"image,omitempty"`
	Storage   *Storage   `json:"storage,omitempty"`
	Resources *Resources `json:"resources,omitempty"`
	Config    *Config    `json:"config,omitempty"`
	Replicas  *int32     `json:"replicas,omitempty"`
	Shards    *int32     `json:"shards,omitempty"`
	// TODO: TLS settings

	// +kubebuilder:pruning:PreserveUnknownFields
	CustomSpec *runtime.RawExtension `json:"customSpec,omitempty"`

	// INTERNAL FIELDS: these are not part of the CRD.
	PodSpec *ComponentPodSpec `json:"-,omitempty"`
}

type Config struct {
	SecretRef    corev1.LocalObjectReference `json:"secretRef,omitempty"`
	ConfigMapRef corev1.LocalObjectReference `json:"configMapRef,omitempty"`
	Key          string                      `json:"key,omitempty"`
}

type Storage struct {
	Size         resource.Quantity `json:"size,omitempty"`
	StorageClass string            `json:"storageClass,omitempty"`
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
