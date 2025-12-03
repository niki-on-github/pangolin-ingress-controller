# Data Model: SSO & Multi-Path Support

**Feature**: 002-sso-multipath-support  
**Date**: 2025-12-03

## Entity Changes

### PangolinResource CRD (PIC)

**File**: `internal/pangolincrd/types.go`

```go
// HTTPConfig - UPDATED
type HTTPConfig struct {
    DomainName  string `json:"domainName,omitempty"`
    Subdomain   string `json:"subdomain,omitempty"`
    SSO         bool   `json:"sso"`         // NEW - already added
    BlockAccess bool   `json:"blockAccess"` // NEW - already added
}

// Target - UPDATED  
type Target struct {
    IP            string `json:"ip,omitempty"`
    Port          int32  `json:"port,omitempty"`
    Method        string `json:"method,omitempty"`
    Path          string `json:"path,omitempty"`          // NEW
    PathMatchType string `json:"pathMatchType,omitempty"` // NEW: exact, prefix, regex
    Priority      int32  `json:"priority,omitempty"`      // NEW: 1-1000
}

// PangolinResourceSpec - UPDATED
type PangolinResourceSpec struct {
    Name       string      `json:"name,omitempty"`
    Enabled    bool        `json:"enabled,omitempty"`
    Protocol   string      `json:"protocol,omitempty"`
    TunnelRef  TunnelRef   `json:"tunnelRef,omitempty"`
    HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
    Targets    []Target    `json:"targets,omitempty"` // CHANGED: single Target → []Target
}
```

### PangolinResource CRD (Operator)

**File**: `api/v1alpha1/pangolinresource_types.go`

```go
// HTTPConfig - mirrors PIC
type HTTPConfig struct {
    Subdomain   string `json:"subdomain"`
    DomainID    string `json:"domainId,omitempty"`
    DomainName  string `json:"domainName,omitempty"`
    SSO         bool   `json:"sso"`         // already added
    BlockAccess bool   `json:"blockAccess"` // already added
}

// TargetConfig - UPDATED
type TargetConfig struct {
    IP            string `json:"ip"`
    Port          int32  `json:"port"`
    Method        string `json:"method,omitempty"`
    Path          string `json:"path,omitempty"`          // NEW
    PathMatchType string `json:"pathMatchType,omitempty"` // NEW
    Priority      int32  `json:"priority,omitempty"`      // NEW
}

// PangolinResourceSpec - UPDATED
type PangolinResourceSpec struct {
    Name       string        `json:"name,omitempty"`
    Enabled    bool          `json:"enabled"`
    Protocol   string        `json:"protocol"`
    TunnelRef  LocalObjectReference `json:"tunnelRef"`
    HTTPConfig *HTTPConfig   `json:"httpConfig,omitempty"`
    Targets    []TargetConfig `json:"targets,omitempty"` // CHANGED: single → array
}
```

### Pangolin API Types (Operator)

**File**: `pkg/pangolin/types.go`

```go
// ResourceCreateSpec - unchanged (SSO not supported at create time)
type ResourceCreateSpec struct {
    Name        string `json:"name"`
    HTTP        bool   `json:"http"`
    Protocol    string `json:"protocol"`
    Subdomain   string `json:"subdomain,omitempty"`
    DomainID    string `json:"domainId,omitempty"`
    ProxyPort   int32  `json:"proxyPort,omitempty"`
    EnableProxy bool   `json:"enableProxy,omitempty"`
}

// ResourceUpdateSpec - NEW
type ResourceUpdateSpec struct {
    SSO         *bool `json:"sso,omitempty"`
    BlockAccess *bool `json:"blockAccess,omitempty"`
}

// TargetCreateSpec - UPDATED
type TargetCreateSpec struct {
    IP            string `json:"ip"`
    Port          int32  `json:"port"`
    Method        string `json:"method,omitempty"`
    SiteID        int    `json:"siteId"`
    Enabled       bool   `json:"enabled"`
    Path          string `json:"path,omitempty"`          // NEW
    PathMatchType string `json:"pathMatchType,omitempty"` // NEW
    Priority      int32  `json:"priority,omitempty"`      // NEW
}
```

## Validation Rules

### SSO Fields
- `sso`: boolean, defaults to `false`
- `blockAccess`: boolean, defaults to `false`, only meaningful when `sso: true`

### Path Fields
- `path`: optional string, e.g., `/api`, `/web`
- `pathMatchType`: enum `["exact", "prefix", "regex"]`, defaults to `prefix`
- `priority`: int 1-1000, defaults to 100, higher = more specific match first

### Target Array
- At least one target required
- Each target can have unique path configuration
- Priority determines matching order for overlapping paths

## State Transitions

### SSO State Machine

```
┌─────────────┐                    ┌─────────────┐
│  SSO: off   │◄──── disable ─────│  SSO: on    │
│ BlockAccess │                    │ BlockAccess │
│    off      │──── enable ───────►│   on/off    │
└─────────────┘                    └─────────────┘
     │                                   │
     ▼                                   ▼
  Public                           Protected
  Access                           (auth required if blockAccess)
```

### Target Lifecycle

```
Ingress Created
     │
     ▼
Parse paths[] ──► For each path:
     │                 │
     │                 ▼
     │           Create Target with:
     │           - IP: service FQDN
     │           - Port: service port
     │           - Path: ingress path
     │           - PathMatchType: from pathType
     │           - Priority: based on specificity
     │
     ▼
PangolinResource.Spec.Targets = [target1, target2, ...]
```

## Migration Notes

### Breaking Change: Target → Targets

The change from single `Target` to `[]Targets` is a breaking change for the CRD schema.

**Migration Strategy**:
1. Update CRD with new schema (operator regenerates CRDs)
2. Existing resources with single target continue to work (single-element array)
3. New resources can specify multiple targets
4. Reconciliation updates existing resources to array format
