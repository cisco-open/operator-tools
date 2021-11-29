//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by controller-gen. DO NOT EDIT.

package helm

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvConfigMap) DeepCopyInto(out *EnvConfigMap) {
	*out = *in
	in.ConfigMapKeyRef.DeepCopyInto(&out.ConfigMapKeyRef)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvConfigMap.
func (in *EnvConfigMap) DeepCopy() *EnvConfigMap {
	if in == nil {
		return nil
	}
	out := new(EnvConfigMap)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvResourceField) DeepCopyInto(out *EnvResourceField) {
	*out = *in
	in.ResourceFieldRef.DeepCopyInto(&out.ResourceFieldRef)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvResourceField.
func (in *EnvResourceField) DeepCopy() *EnvResourceField {
	if in == nil {
		return nil
	}
	out := new(EnvResourceField)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvSecret) DeepCopyInto(out *EnvSecret) {
	*out = *in
	in.SecretKeyRef.DeepCopyInto(&out.SecretKeyRef)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvSecret.
func (in *EnvSecret) DeepCopy() *EnvSecret {
	if in == nil {
		return nil
	}
	out := new(EnvSecret)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvironmentVariables) DeepCopyInto(out *EnvironmentVariables) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.EnvSecrets != nil {
		in, out := &in.EnvSecrets, &out.EnvSecrets
		*out = make([]EnvSecret, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EnvResourceField != nil {
		in, out := &in.EnvResourceField, &out.EnvResourceField
		*out = make([]EnvResourceField, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EnvConfigMap != nil {
		in, out := &in.EnvConfigMap, &out.EnvConfigMap
		*out = make([]EnvConfigMap, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvironmentVariables.
func (in *EnvironmentVariables) DeepCopy() *EnvironmentVariables {
	if in == nil {
		return nil
	}
	out := new(EnvironmentVariables)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Image) DeepCopyInto(out *Image) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Image.
func (in *Image) DeepCopy() *Image {
	if in == nil {
		return nil
	}
	out := new(Image)
	in.DeepCopyInto(out)
	return out
}
