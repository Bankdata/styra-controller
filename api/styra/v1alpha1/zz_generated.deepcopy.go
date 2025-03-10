//go:build !ignore_autogenerated

/*
Copyright (C) 2023 Bankdata (bankdata@bankdata.dk)

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitRepo) DeepCopyInto(out *GitRepo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitRepo.
func (in *GitRepo) DeepCopy() *GitRepo {
	if in == nil {
		return nil
	}
	out := new(GitRepo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Library) DeepCopyInto(out *Library) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Library.
func (in *Library) DeepCopy() *Library {
	if in == nil {
		return nil
	}
	out := new(Library)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Library) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LibraryDatasource) DeepCopyInto(out *LibraryDatasource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LibraryDatasource.
func (in *LibraryDatasource) DeepCopy() *LibraryDatasource {
	if in == nil {
		return nil
	}
	out := new(LibraryDatasource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LibraryList) DeepCopyInto(out *LibraryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Library, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LibraryList.
func (in *LibraryList) DeepCopy() *LibraryList {
	if in == nil {
		return nil
	}
	out := new(LibraryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LibraryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LibrarySecretRef) DeepCopyInto(out *LibrarySecretRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LibrarySecretRef.
func (in *LibrarySecretRef) DeepCopy() *LibrarySecretRef {
	if in == nil {
		return nil
	}
	out := new(LibrarySecretRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LibrarySpec) DeepCopyInto(out *LibrarySpec) {
	*out = *in
	if in.Subjects != nil {
		in, out := &in.Subjects, &out.Subjects
		*out = make([]LibrarySubject, len(*in))
		copy(*out, *in)
	}
	if in.SourceControl != nil {
		in, out := &in.SourceControl, &out.SourceControl
		*out = new(SourceControl)
		(*in).DeepCopyInto(*out)
	}
	if in.Datasources != nil {
		in, out := &in.Datasources, &out.Datasources
		*out = make([]LibraryDatasource, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LibrarySpec.
func (in *LibrarySpec) DeepCopy() *LibrarySpec {
	if in == nil {
		return nil
	}
	out := new(LibrarySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LibraryStatus) DeepCopyInto(out *LibraryStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LibraryStatus.
func (in *LibraryStatus) DeepCopy() *LibraryStatus {
	if in == nil {
		return nil
	}
	out := new(LibraryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LibrarySubject) DeepCopyInto(out *LibrarySubject) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LibrarySubject.
func (in *LibrarySubject) DeepCopy() *LibrarySubject {
	if in == nil {
		return nil
	}
	out := new(LibrarySubject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceControl) DeepCopyInto(out *SourceControl) {
	*out = *in
	if in.LibraryOrigin != nil {
		in, out := &in.LibraryOrigin, &out.LibraryOrigin
		*out = new(GitRepo)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceControl.
func (in *SourceControl) DeepCopy() *SourceControl {
	if in == nil {
		return nil
	}
	out := new(SourceControl)
	in.DeepCopyInto(out)
	return out
}
