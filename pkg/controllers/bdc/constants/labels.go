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

package constants

const (
	LabelReferredAPIResource = "api-resource.bdc.bdos.io/type"
	// LabelDefinition is the label for definition
	LabelDefinition = "definition.bdc.bdos.io"
	// LabelDefinitionName is the label for definition name
	LabelDefinitionName = "definition.bdc.bdos.io/name"
	LabelBDCOrgName     = "bdc.bdos.io/org"
)

const (
	// FinalizerResourceTracker is the application finalizer for gc
	FinalizerResourceTracker = "bdc.bdos.io/resource-tracker-finalizer"
)
