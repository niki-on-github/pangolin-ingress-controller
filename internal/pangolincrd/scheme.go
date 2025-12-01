package pangolincrd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// GroupName is the API group for Pangolin CRDs.
	GroupName = "tunnel.pangolin.io"

	// Version is the API version for Pangolin CRDs.
	Version = "v1alpha1"
)

var (
	// GroupVersion is the group version for Pangolin CRDs.
	GroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

	// SchemeBuilder is used to add types to the scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the Pangolin types to the scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// addKnownTypes adds the Pangolin types to the scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&PangolinResource{},
		&PangolinResourceList{},
		&PangolinTunnel{},
		&PangolinTunnelList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}

// Resource returns the GroupResource for a given resource name.
func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}
