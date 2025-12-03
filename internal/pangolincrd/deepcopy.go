package pangolincrd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copies the receiver into out.
func (in *PangolinResource) DeepCopyInto(out *PangolinResource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a deep copy of PangolinResource.
func (in *PangolinResource) DeepCopy() *PangolinResource {
	if in == nil {
		return nil
	}
	out := new(PangolinResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject returns a deep copy as runtime.Object.
func (in *PangolinResource) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto copies the receiver into out.
func (in *PangolinResourceSpec) DeepCopyInto(out *PangolinResourceSpec) {
	*out = *in
	out.TunnelRef = in.TunnelRef
	if in.HTTPConfig != nil {
		out.HTTPConfig = new(HTTPConfig)
		*out.HTTPConfig = *in.HTTPConfig
	}
	if in.Targets != nil {
		out.Targets = make([]Target, len(in.Targets))
		copy(out.Targets, in.Targets)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *PangolinResourceStatus) DeepCopyInto(out *PangolinResourceStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
	if in.LastSyncTime != nil {
		out.LastSyncTime = in.LastSyncTime.DeepCopy()
	}
}

// DeepCopyInto copies the receiver into out.
func (in *PangolinResourceList) DeepCopyInto(out *PangolinResourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]PangolinResource, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// DeepCopy creates a deep copy of PangolinResourceList.
func (in *PangolinResourceList) DeepCopy() *PangolinResourceList {
	if in == nil {
		return nil
	}
	out := new(PangolinResourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject returns a deep copy as runtime.Object.
func (in *PangolinResourceList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto copies the receiver into out.
func (in *PangolinTunnel) DeepCopyInto(out *PangolinTunnel) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a deep copy of PangolinTunnel.
func (in *PangolinTunnel) DeepCopy() *PangolinTunnel {
	if in == nil {
		return nil
	}
	out := new(PangolinTunnel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject returns a deep copy as runtime.Object.
func (in *PangolinTunnel) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto copies the receiver into out.
func (in *PangolinTunnelStatus) DeepCopyInto(out *PangolinTunnelStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

// DeepCopyInto copies the receiver into out.
func (in *PangolinTunnelList) DeepCopyInto(out *PangolinTunnelList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]PangolinTunnel, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// DeepCopy creates a deep copy of PangolinTunnelList.
func (in *PangolinTunnelList) DeepCopy() *PangolinTunnelList {
	if in == nil {
		return nil
	}
	out := new(PangolinTunnelList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject returns a deep copy as runtime.Object.
func (in *PangolinTunnelList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}
