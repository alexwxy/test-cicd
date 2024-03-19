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
	"bpaas-core-operator/api/bdc/condition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomSettingSpec defines the desired state of CustomSetting
type CustomSettingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name string            `json:"name"`
	Data map[string]string `json:"data"`
}

// CustomSettingStatus defines the observed state of CustomSetting
type CustomSettingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status"`
	// ConditionedStatus reflects the observed status of a resource
	condition.ConditionedStatus `json:",inline"`
	SchemaConfigMapRef          string `json:"schemaConfigMapRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// CustomSetting is the Schema for the customsettings API
type CustomSetting struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomSettingSpec   `json:"spec"`
	Status CustomSettingStatus `json:"status,omitempty"`
}

// SetConditions set condition for CustomSetting
func (cd *CustomSetting) SetConditions(c ...condition.Condition) {
	cd.Status.SetConditions(c...)
}

// GetCondition gets condition from CustomSetting
func (cd *CustomSetting) GetCondition(conditionType condition.ConditionType) condition.Condition {
	return cd.Status.GetCondition(conditionType)
}

//+kubebuilder:object:root=true

// CustomSettingList contains a list of CustomSetting
type CustomSettingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomSetting `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomSetting{}, &CustomSettingList{})
}
