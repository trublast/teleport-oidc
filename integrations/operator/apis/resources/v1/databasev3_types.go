/*
Copyright 2026 Flant JSC

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/integrations/operator/apis/resources"
)

func init() {
	SchemeBuilder.Register(&TeleportDatabaseV3{}, &TeleportDatabaseV3List{})
}

// TeleportDatabaseV3Spec defines the desired state of TeleportDatabaseV3
type TeleportDatabaseV3Spec types.DatabaseSpecV3

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TeleportDatabaseV3 is the Schema for the databasesv3 API
type TeleportDatabaseV3 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TeleportDatabaseV3Spec   `json:"spec,omitempty"`
	Status TeleportDatabaseV3Status `json:"status,omitempty"`
}

// TeleportDatabaseV3Status defines the observed state of TeleportDatabaseV3
type TeleportDatabaseV3Status struct {
	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// +optional
	TeleportResourceID int64 `json:"teleportResourceID,omitempty"`
}

//+kubebuilder:object:root=true

// TeleportDatabaseV3List contains a list of TeleportDatabaseV3
type TeleportDatabaseV3List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TeleportDatabaseV3 `json:"items"`
}

func (r TeleportDatabaseV3) ToTeleport() types.Database {
	return &types.DatabaseV3{
		Kind:    types.KindDatabase,
		Version: types.V3,
		Metadata: types.Metadata{
			Name:        r.Name,
			Labels:      r.Labels,
			Description: r.Annotations[resources.DescriptionKey],
		},
		Spec: types.DatabaseSpecV3(r.Spec),
	}
}

// Marshal serializes a spec into binary data.
func (spec *TeleportDatabaseV3Spec) Marshal() ([]byte, error) {
	return (*types.DatabaseSpecV3)(spec).Marshal()
}

// Unmarshal deserializes a spec from binary data.
func (spec *TeleportDatabaseV3Spec) Unmarshal(data []byte) error {
	return (*types.DatabaseSpecV3)(spec).Unmarshal(data)
}

// DeepCopyInto deep-copies one database spec into another.
// Required to satisfy runtime.Object interface.
func (spec *TeleportDatabaseV3Spec) DeepCopyInto(out *TeleportDatabaseV3Spec) {
	data, err := spec.Marshal()
	if err != nil {
		panic(err)
	}
	*out = TeleportDatabaseV3Spec{}
	if err = out.Unmarshal(data); err != nil {
		panic(err)
	}
}

// StatusConditions returns a pointer to Status.Conditions slice.
func (r *TeleportDatabaseV3) StatusConditions() *[]metav1.Condition {
	return &r.Status.Conditions
}
