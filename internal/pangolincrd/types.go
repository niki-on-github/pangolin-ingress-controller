// Package pangolincrd provides Go types for Pangolin CRDs.
package pangolincrd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PangolinResource represents an exposed endpoint in Pangolin.
// This resource is created by PIC and processed by pangolin-operator.
type PangolinResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PangolinResourceSpec   `json:"spec,omitempty"`
	Status PangolinResourceStatus `json:"status,omitempty"`
}

// PangolinResourceSpec defines the desired state of a PangolinResource.
type PangolinResourceSpec struct {
	// Name is the display name for the resource in Pangolin.
	Name string `json:"name,omitempty"`

	// Enabled indicates whether the resource should be active.
	Enabled bool `json:"enabled"`

	// Protocol is the external access protocol ("http" or "https").
	Protocol string `json:"protocol,omitempty"`

	// TunnelRef references the tunnel to use for this resource.
	TunnelRef TunnelRef `json:"tunnelRef,omitempty"`

	// HTTPConfig contains HTTP-specific configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`

	// Target defines the backend service to route to.
	Target *Target `json:"target,omitempty"`
}

// TunnelRef is a reference to a PangolinTunnel.
type TunnelRef struct {
	// Name is the name of the PangolinTunnel resource.
	Name string `json:"name,omitempty"`

	// Namespace is the namespace of the PangolinTunnel resource.
	Namespace string `json:"namespace,omitempty"`
}

// HTTPConfig contains HTTP-specific configuration.
type HTTPConfig struct {
	// DomainName is the domain for the exposed endpoint.
	DomainName string `json:"domainName,omitempty"`

	// Subdomain is the subdomain for the exposed endpoint.
	Subdomain string `json:"subdomain,omitempty"`

	// SSO enables SSO authentication for this resource.
	// +optional
	SSO bool `json:"sso"`

	// BlockAccess blocks access until user is authenticated.
	// Only effective when SSO is enabled.
	// +optional
	BlockAccess bool `json:"blockAccess"`
}

// Target defines the backend service configuration.
type Target struct {
	// IP is the target hostname (typically service FQDN).
	IP string `json:"ip,omitempty"`

	// Port is the target port number.
	Port int32 `json:"port,omitempty"`

	// Method is the backend protocol ("http" or "https").
	Method string `json:"method,omitempty"`
}

// PangolinResourceStatus defines the observed state of PangolinResource.
// This is set by pangolin-operator, read-only for PIC.
type PangolinResourceStatus struct {
	// URL is the public URL where the resource is accessible.
	URL string `json:"url,omitempty"`

	// ResourceID is the Pangolin-side resource identifier.
	ResourceID string `json:"resourceId,omitempty"`

	// Phase indicates the current state: Pending, Ready, Failed.
	Phase string `json:"phase,omitempty"`

	// Conditions provide detailed status information.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastSyncTime is the last time the resource was synced with Pangolin.
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`
}

// +kubebuilder:object:root=true

// PangolinResourceList contains a list of PangolinResource.
type PangolinResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PangolinResource `json:"items"`
}

// +kubebuilder:object:root=true

// PangolinTunnel represents a Pangolin tunnel/site.
// PIC only reads this resource; it is managed by pangolin-operator.
type PangolinTunnel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PangolinTunnelSpec   `json:"spec,omitempty"`
	Status PangolinTunnelStatus `json:"status,omitempty"`
}

// PangolinTunnelSpec defines the desired state of a PangolinTunnel.
type PangolinTunnelSpec struct {
	// SiteID is the Pangolin site identifier.
	SiteID string `json:"siteId,omitempty"`
}

// PangolinTunnelStatus defines the observed state of PangolinTunnel.
type PangolinTunnelStatus struct {
	// Phase indicates the current state: Pending, Ready, Failed.
	Phase string `json:"phase,omitempty"`

	// Conditions provide detailed status information.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// PangolinTunnelList contains a list of PangolinTunnel.
type PangolinTunnelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PangolinTunnel `json:"items"`
}
