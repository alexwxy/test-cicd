/*
Copyright 2023 KDP(Kubernetes Data Platform).

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

package v1alpha1

import (
	"bpaas-core-operator/api/bdc/common"
	"bpaas-core-operator/api/bdc/condition"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type QuotaSpec struct {
	CPU     resource.Quantity `json:"cpu"`
	Memory  resource.Quantity `json:"memory"`
	Storage resource.Quantity `json:"storage,omitempty"`
}
type ResourceQuotaSpec struct {
	Limits   *QuotaSpec `json:"limits"`
	Requests *QuotaSpec `json:"requests"`
}

type LimitRangeSpec struct {
	Max            *QuotaSpec `json:"max"`
	Min            *QuotaSpec `json:"min"`
	Default        *QuotaSpec `json:"default"`
	DefaultRequest *QuotaSpec `json:"defaultRequest"`
}

// ResourceControlSpec defines the desired state of ResourceControl
type ResourceControlSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ResourceQuota *ResourceQuotaSpec `json:"resourceQuota,omitempty"`
	LimitRange    *LimitRangeSpec    `json:"limitRange,omitempty"`
}

// ResourceControlStatus defines the observed state of ResourceControl
type ResourceControlStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status"`
	// ConditionedStatus reflects the observed status of a resource
	condition.ConditionedStatus `json:",inline"`
	SchemaConfigMapRef          string `json:"schemaConfigMapRef"`
	// AppliedResources record the resources that apply.
	AppliedResources []common.BDCObjectReference `json:"appliedResources,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ResourceControl is the Schema for the resourcecontrols API
type ResourceControl struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceControlSpec   `json:"spec,omitempty"`
	Status ResourceControlStatus `json:"status,omitempty"`
}

// SetConditions set condition for ResourceControl
func (cd *ResourceControl) SetConditions(c ...condition.Condition) {
	cd.Status.SetConditions(c...)
}

// GetCondition gets condition from ResourceControl
func (cd *ResourceControl) GetCondition(conditionType condition.ConditionType) condition.Condition {
	return cd.Status.GetCondition(conditionType)
}

//+kubebuilder:object:root=true

// ResourceControlList contains a list of ResourceControl
type ResourceControlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceControl `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceControl{}, &ResourceControlList{})
}
