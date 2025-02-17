package v2alpha1

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=dbd;dbdef;dbdefinition
type DatabaseClusterDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseClusterDefinitionSpec   `json:"spec,omitempty"`
	Status DatabaseClusterDefinitionStatus `json:"status,omitempty"`
}

type DatabaseClusterDefinitionSpec struct {
	Definitions Definitions `json:"definitions,omitempty"`
}

type Definitions struct {
	GlobalDefinition GlobalDefinition               `json:"global,omitempty"`
	Components       map[string]ComponentDefinition `json:"components,omitempty"`
}

type GlobalDefinition struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +k8s:conversion-gen=false
	OpenAPIV3Schema *apiextensionsv1.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`
}

type ComponentDefinition struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +k8s:conversion-gen=false
	OpenAPIV3Schema *apiextensionsv1.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`
	Defaults        *ComponentPodSpec                `json:"defaults,omitempty"`
}

type ComponentPodSpec struct {
	Annotations                    map[string]string              `json:"annotations,omitempty"`
	Labels                         map[string]string              `json:"labels,omitempty"`
	Container                      *corev1.Container              `json:"container,omitempty"`
	Sidecars                       []corev1.Container             `json:"sidecars,omitempty"`
	Volumes                        []corev1.Volume                `json:"volumes,omitempty"`
	ServiceAccountName             string                         `json:"serviceAccountName,omitempty"`
	ImagePullSecrets               []corev1.LocalObjectReference  `json:"imagePullSecrets,omitempty"`
	AdditionalVolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"additionalVolumeClaimTemplates,omitempty"`
}

type DatabaseClusterDefinitionStatus struct{}

// DatabaseClusterList contains a list of DatabaseCluster.
//
// +kubebuilder:object:root=true
type DatabaseClusterDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseClusterDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseClusterDefinition{}, &DatabaseClusterDefinitionList{})
}
