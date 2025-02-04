// +build !ignore_autogenerated

/*
Copyright 2019 Kazumichi Yamamoto.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License..
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha2

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/errors"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIEndpoint) DeepCopyInto(out *APIEndpoint) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIEndpoint.
func (in *APIEndpoint) DeepCopy() *APIEndpoint {
	if in == nil {
		return nil
	}
	out := new(APIEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Filter) DeepCopyInto(out *Filter) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Filter.
func (in *Filter) DeepCopy() *Filter {
	if in == nil {
		return nil
	}
	out := new(Filter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudCluster) DeepCopyInto(out *SakuraCloudCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudCluster.
func (in *SakuraCloudCluster) DeepCopy() *SakuraCloudCluster {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SakuraCloudCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudClusterList) DeepCopyInto(out *SakuraCloudClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SakuraCloudCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudClusterList.
func (in *SakuraCloudClusterList) DeepCopy() *SakuraCloudClusterList {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SakuraCloudClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudClusterSpec) DeepCopyInto(out *SakuraCloudClusterSpec) {
	*out = *in
	out.CloudProviderConfiguration = in.CloudProviderConfiguration
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudClusterSpec.
func (in *SakuraCloudClusterSpec) DeepCopy() *SakuraCloudClusterSpec {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudClusterStatus) DeepCopyInto(out *SakuraCloudClusterStatus) {
	*out = *in
	if in.APIEndpoints != nil {
		in, out := &in.APIEndpoints, &out.APIEndpoints
		*out = make([]APIEndpoint, len(*in))
		copy(*out, *in)
	}
	if in.ErrorReason != nil {
		in, out := &in.ErrorReason, &out.ErrorReason
		*out = new(errors.ClusterStatusError)
		**out = **in
	}
	if in.ErrorMessage != nil {
		in, out := &in.ErrorMessage, &out.ErrorMessage
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudClusterStatus.
func (in *SakuraCloudClusterStatus) DeepCopy() *SakuraCloudClusterStatus {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudMachine) DeepCopyInto(out *SakuraCloudMachine) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudMachine.
func (in *SakuraCloudMachine) DeepCopy() *SakuraCloudMachine {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudMachine)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SakuraCloudMachine) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudMachineList) DeepCopyInto(out *SakuraCloudMachineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SakuraCloudMachine, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudMachineList.
func (in *SakuraCloudMachineList) DeepCopy() *SakuraCloudMachineList {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudMachineList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SakuraCloudMachineList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudMachineSpec) DeepCopyInto(out *SakuraCloudMachineSpec) {
	*out = *in
	if in.ProviderID != nil {
		in, out := &in.ProviderID, &out.ProviderID
		*out = new(string)
		**out = **in
	}
	if in.MachineRef != nil {
		in, out := &in.MachineRef, &out.MachineRef
		*out = new(SakuraCloudResourceReference)
		(*in).DeepCopyInto(*out)
	}
	in.SourceArchive.DeepCopyInto(&out.SourceArchive)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudMachineSpec.
func (in *SakuraCloudMachineSpec) DeepCopy() *SakuraCloudMachineSpec {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudMachineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudMachineStatus) DeepCopyInto(out *SakuraCloudMachineStatus) {
	*out = *in
	if in.Addresses != nil {
		in, out := &in.Addresses, &out.Addresses
		*out = make([]v1.NodeAddress, len(*in))
		copy(*out, *in)
	}
	if in.SourceArchive != nil {
		in, out := &in.SourceArchive, &out.SourceArchive
		*out = new(SourceArchiveInfo)
		**out = **in
	}
	if in.ErrorReason != nil {
		in, out := &in.ErrorReason, &out.ErrorReason
		*out = new(errors.MachineStatusError)
		**out = **in
	}
	if in.ErrorMessage != nil {
		in, out := &in.ErrorMessage, &out.ErrorMessage
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudMachineStatus.
func (in *SakuraCloudMachineStatus) DeepCopy() *SakuraCloudMachineStatus {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudMachineStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudMachineTemplate) DeepCopyInto(out *SakuraCloudMachineTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudMachineTemplate.
func (in *SakuraCloudMachineTemplate) DeepCopy() *SakuraCloudMachineTemplate {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudMachineTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SakuraCloudMachineTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudMachineTemplateList) DeepCopyInto(out *SakuraCloudMachineTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SakuraCloudMachineTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudMachineTemplateList.
func (in *SakuraCloudMachineTemplateList) DeepCopy() *SakuraCloudMachineTemplateList {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudMachineTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SakuraCloudMachineTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudMachineTemplateResource) DeepCopyInto(out *SakuraCloudMachineTemplateResource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudMachineTemplateResource.
func (in *SakuraCloudMachineTemplateResource) DeepCopy() *SakuraCloudMachineTemplateResource {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudMachineTemplateResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudMachineTemplateSpec) DeepCopyInto(out *SakuraCloudMachineTemplateSpec) {
	*out = *in
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudMachineTemplateSpec.
func (in *SakuraCloudMachineTemplateSpec) DeepCopy() *SakuraCloudMachineTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudMachineTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudProviderConfig) DeepCopyInto(out *SakuraCloudProviderConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudProviderConfig.
func (in *SakuraCloudProviderConfig) DeepCopy() *SakuraCloudProviderConfig {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudProviderConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SakuraCloudResourceReference) DeepCopyInto(out *SakuraCloudResourceReference) {
	*out = *in
	if in.ID != nil {
		in, out := &in.ID, &out.ID
		*out = new(string)
		**out = **in
	}
	if in.Filters != nil {
		in, out := &in.Filters, &out.Filters
		*out = make([]Filter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SakuraCloudResourceReference.
func (in *SakuraCloudResourceReference) DeepCopy() *SakuraCloudResourceReference {
	if in == nil {
		return nil
	}
	out := new(SakuraCloudResourceReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceArchiveInfo) DeepCopyInto(out *SourceArchiveInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceArchiveInfo.
func (in *SourceArchiveInfo) DeepCopy() *SourceArchiveInfo {
	if in == nil {
		return nil
	}
	out := new(SourceArchiveInfo)
	in.DeepCopyInto(out)
	return out
}
