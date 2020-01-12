/*

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KeystoneServerSpec defines the desired state of KeystoneServer
type KeystoneServerSpec struct {
	Image    string                    `json:"image,omitempty"`
	Release  string                    `json:"release,omitempty"`
	Replicas int32                     `json:"replicas,omitempty"`
	Config   map[string]KyestoneConfig `json:"config,omitempty"`
}

// KyestoneConfig configuration parameters for keystone
type KyestoneConfig map[string]string

// KeystoneServerStatus defines the observed state of KeystoneServer
type KeystoneServerStatus struct {
	Ready bool `json:"ready,omitempty"`
}

// +kubebuilder:object:root=true

// KeystoneServer is the Schema for the keystoneservers API
type KeystoneServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeystoneServerSpec   `json:"spec,omitempty"`
	Status KeystoneServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KeystoneServerList contains a list of KeystoneServer
type KeystoneServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeystoneServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeystoneServer{}, &KeystoneServerList{})
}
