# Research: Pangolin Ingress Controller MVP

**Feature**: 001-pic-mvp  
**Date**: 2025-11-30

## Research Tasks

### 1. Controller-Runtime Best Practices

**Decision**: Use controller-runtime v0.17+ with standard Reconciler pattern

**Rationale**:
- Industry standard for Kubernetes controllers
- Provides leader election, metrics, health probes out of the box
- Well-documented with extensive community support
- Used by Kubebuilder and Operator SDK

**Alternatives Considered**:
- Raw client-go: Too low-level, requires implementing leader election, caching manually
- Operator SDK: Adds unnecessary scaffolding for a simple controller

### 2. Host Splitting Algorithm

**Decision**: Use `golang.org/x/net/publicsuffix` for domain extraction

**Rationale**:
- Maintained by Go team
- Handles all TLDs including complex ones (co.uk, com.au)
- Updated regularly with new TLDs
- Zero external dependencies

**Alternatives Considered**:
- Manual parsing: Fragile, doesn't handle edge cases
- Third-party libraries: Additional dependency risk

**Implementation**:
```go
import "golang.org/x/net/publicsuffix"

func SplitHost(host string) (subdomain, domain string, err error) {
    domain, err = publicsuffix.EffectiveTLDPlusOne(host)
    if err != nil {
        return "", "", err
    }
    if host == domain {
        return "", domain, nil
    }
    subdomain = strings.TrimSuffix(host, "."+domain)
    return subdomain, domain, nil
}
```

### 3. PangolinResource CRD Schema

**Decision**: Mirror pangolin-operator's CRD structure with minimal required fields

**Rationale**:
- Ensures compatibility with existing operator
- Only include fields PIC needs to set
- Status fields are read-only (set by operator)

**Key Fields**:
```yaml
spec:
  enabled: bool
  protocol: string  # "http" or "https"
  tunnelRef:
    name: string    # Reference to PangolinTunnel
  httpConfig:
    domainName: string
    subdomain: string
  target:
    ip: string      # Service FQDN
    port: int32
    method: string  # "http"
```

### 4. Deterministic Naming Strategy

**Decision**: Use pattern `pic-<namespace>-<ingress>-<hash>`

**Rationale**:
- Predictable for debugging
- Hash prevents collisions for same ingress name across namespaces
- Prefix `pic-` identifies PIC-managed resources

**Implementation**:
```go
func GenerateName(namespace, ingressName, host string) string {
    hash := sha256.Sum256([]byte(namespace + "/" + ingressName + "/" + host))
    shortHash := hex.EncodeToString(hash[:4])
    return fmt.Sprintf("pic-%s-%s-%s", namespace, ingressName, shortHash)
}
```

### 5. Ownership and Garbage Collection

**Decision**: Use Kubernetes ownerReferences for cascading deletion

**Rationale**:
- Built-in Kubernetes mechanism
- Automatic cleanup on Ingress deletion
- No need for finalizers
- controller-runtime handles this via `ctrl.SetControllerReference`

**Implementation**:
```go
if err := ctrl.SetControllerReference(ingress, pangolinResource, r.Scheme); err != nil {
    return ctrl.Result{}, err
}
```

### 6. Event Emission Strategy

**Decision**: Use controller-runtime's EventRecorder

**Rationale**:
- Standard Kubernetes pattern
- Events appear in `kubectl describe ingress`
- Actionable for operators

**Events**:
| Event | Type | Reason | Message |
|-------|------|--------|---------|
| Resource created | Normal | Created | Created PangolinResource {name} |
| Resource updated | Normal | Updated | Updated PangolinResource {name} |
| Tunnel not found | Warning | TunnelNotFound | Tunnel {name} not found |
| Invalid host | Warning | InvalidHost | Host {host} cannot be processed |

### 7. Configuration Loading

**Decision**: Environment variables with sensible defaults

**Rationale**:
- 12-factor app compliance
- Easy to configure via Kubernetes Deployment
- No config files to mount

**Variables**:
```
PIC_DEFAULT_TUNNEL_NAME     # Default: "default"
PIC_TUNNEL_CLASS_MAPPING    # Default: ""
PIC_BACKEND_SCHEME          # Default: "http"
PIC_RESYNC_PERIOD           # Default: "5m"
PIC_LOG_LEVEL               # Default: "info"
PIC_WATCH_NAMESPACES        # Default: "" (all)
```

## Resolved Clarifications

All technical decisions made based on constitution requirements and best practices. No outstanding clarifications.
