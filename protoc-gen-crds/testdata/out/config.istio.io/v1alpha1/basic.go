//
// GENERATED CODE -- DO NOT EDIT
//
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Basic struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec BasicSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BasicList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Ingress.
	Items []Basic `json:"items" protobuf:"bytes,2,rep,name=items"`
}


type BasicSpec struct { 
	FieldInt32 int `json:"fieldInt32,omitempty"`
	FieldString string `json:"fieldString,omitempty"`
	FieldRepeatedString []string `json:"fieldRepeatedString,omitempty"`
	FieldMessage *InnerMessage `json:"fieldMessage,omitempty"`
	FieldRepeatedMessage []*InnerMessage `json:"fieldRepeatedMessage,omitempty"`
}


type InnerMessage struct { 
	FieldBool bool `json:"fieldBool,omitempty"`
}

