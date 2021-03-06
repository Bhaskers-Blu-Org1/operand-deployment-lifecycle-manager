//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OperandConfigSpec defines the desired state of OperandConfig
// +k8s:openapi-gen=true
type OperandConfigSpec struct {
	// Services is a list of configuration of service
	// +optional
	// +listType=set
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Services []ConfigService `json:"services,omitempty"`
}

// ConfigService defines the configuration of the service
type ConfigService struct {
	// Name is the subscription name
	Name string `json:"name"`
	// Spec is the configuration map of custom resource
	Spec map[string]runtime.RawExtension `json:"spec"`
	// State is a flag to enable or disable service
	State string `json:"state,omitempty"`
}

// OperandConfigStatus defines the observed state of OperandConfig
// +k8s:openapi-gen=true
type OperandConfigStatus struct {
	// Phase describes the overall phase of operands in the OperandConfig
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +optional
	Phase ServicePhase `json:"phase,omitempty"`
	// ServiceStatus defines all the status of a operator
	// +optional
	ServiceStatus map[string]CrStatus `json:"serviceStatus,omitempty"`
}

// CrStatus defines the status of the custom resource
type CrStatus struct {
	CrStatus map[string]ServicePhase `json:"customResourceStatus,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperandConfig is the Schema for the operandconfigs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=operandconfigs,shortName=opcon,scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=.metadata.creationTimestamp
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=.status.phase,description="Current Phase"
// +kubebuilder:printcolumn:name="Created At",type=string,JSONPath=.metadata.creationTimestamp
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="OperandConfig"
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Deployment,v1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`ReplicaSet,v1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Service,v1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Pod,v1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Configmap,v1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Installplan,v1alpha1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Catalogsource,v1alpha1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Clusterserviceversion,v1alpha1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Operatorgroup,v1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Subscription,v1alpha1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Operandconfig,v1alpha1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Operandrequest,v1alpha1,""`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Operandregistry,v1alpha1,""`
type OperandConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperandConfigSpec   `json:"spec,omitempty"`
	Status OperandConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperandConfigList contains a list of OperandConfig
type OperandConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperandConfig `json:"items"`
}

// ServicePhase defines the service status
type ServicePhase string

// Service status
const (
	ServiceReady   ServicePhase = "Ready for Deployment"
	ServiceRunning ServicePhase = "Running"
	ServiceFailed  ServicePhase = "Failed"
	ServiceInit    ServicePhase = "Initialized"
	ServiceNone    ServicePhase = ""
)

func init() {
	SchemeBuilder.Register(&OperandConfig{}, &OperandConfigList{})
}

// GetService obtain the service definition with the operand name
func (r *OperandConfig) GetService(operandName string) *ConfigService {
	for _, s := range r.Spec.Services {
		if s.Name == operandName {
			return &s
		}
	}
	return nil
}

//InitConfigStatus OperandConfig status
func (r *OperandConfig) InitConfigStatus() {
	if (reflect.DeepEqual(r.Status, OperandConfigStatus{})) {
		r.Status.Phase = ServiceInit
	}
}

//InitConfigServiceStatus service status in the OperandConfig instance
func (r *OperandConfig) InitConfigServiceStatus() {
	r.Status.ServiceStatus = make(map[string]CrStatus)

	for _, operator := range r.Spec.Services {
		r.Status.ServiceStatus[operator.Name] = CrStatus{CrStatus: make(map[string]ServicePhase)}
		for service := range operator.Spec {
			r.Status.ServiceStatus[operator.Name].CrStatus[service] = ServiceReady
		}
	}
	r.UpdateOperandPhase()
}

// UpdateOperandPhase sets the current Phase status
func (r *OperandConfig) UpdateOperandPhase() {
	operandStatusStat := struct {
		readyNum   int
		runningNum int
		failedNum  int
	}{
		readyNum:   0,
		runningNum: 0,
		failedNum:  0,
	}
	for _, operator := range r.Status.ServiceStatus {
		for _, service := range operator.CrStatus {
			switch service {
			case ServiceReady:
				operandStatusStat.readyNum++
			case ServiceRunning:
				operandStatusStat.runningNum++
			case ServiceFailed:
				operandStatusStat.failedNum++
			}
		}
	}
	if operandStatusStat.failedNum > 0 {
		r.Status.Phase = ServiceFailed
	} else if operandStatusStat.runningNum > 0 {
		r.Status.Phase = ServiceRunning
	} else if operandStatusStat.readyNum > 0 {
		r.Status.Phase = ServiceReady
	} else {
		r.Status.Phase = ServiceNone
	}
}
