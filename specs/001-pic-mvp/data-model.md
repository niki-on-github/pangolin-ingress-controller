# Data Model: Pangolin Ingress Controller MVP

**Feature**: 001-pic-mvp  
**Date**: 2025-11-30

## Entities

### 1. Ingress (External - Kubernetes Core)

Standard Kubernetes Ingress resource. PIC reads but does not modify.

**Key Fields Used**:
| Field | Type | Usage |
|-------|------|-------|
| `metadata.name` | string | Identifier for naming |
| `metadata.namespace` | string | Target namespace |
| `metadata.uid` | string | Unique ID for labels |
| `metadata.annotations` | map | Override configuration |
| `spec.ingressClassName` | string | Management decision |
| `spec.rules[].host` | string | Domain extraction |
| `spec.rules[].http.paths[].backend` | object | Service reference |

### 2. PangolinTunnel (External - pangolin-operator)

Represents a Pangolin tunnel/site. PIC reads but does not modify.

**Key Fields Used**:
| Field | Type | Usage |
|-------|------|-------|
| `metadata.name` | string | Tunnel reference |
| `status.phase` | string | Readiness check (optional) |

### 3. PangolinResource (Managed by PIC)

Represents an exposed endpoint in Pangolin. Created/updated/deleted by PIC.

```yaml
apiVersion: tunnel.pangolin.io/v1alpha1
kind: PangolinResource
metadata:
  name: pic-<namespace>-<ingress>-<hash>
  namespace: <same as Ingress>
  labels:
    pic.ingress.k8s.io/uid: "<ingress-uid>"
    pic.ingress.k8s.io/name: "<ingress-name>"
    pic.ingress.k8s.io/namespace: "<namespace>"
  ownerReferences:
    - apiVersion: networking.k8s.io/v1
      kind: Ingress
      name: <ingress-name>
      uid: <ingress-uid>
      controller: true
      blockOwnerDeletion: true
spec:
  enabled: true
  protocol: http
  tunnelRef:
    name: <tunnel-name>
  httpConfig:
    domainName: "<extracted-domain>"
    subdomain: "<extracted-subdomain>"
  target:
    ip: "<service>.<namespace>.svc.cluster.local"
    port: <port>
    method: http
status:
  # Set by pangolin-operator, read-only for PIC
  url: "https://<subdomain>.<domain>"
  resourceId: "<pangolin-id>"
  phase: "Ready"
  conditions: [...]
```

### 4. Config (Internal)

Runtime configuration loaded from environment.

```go
type Config struct {
    DefaultTunnelName string            // PIC_DEFAULT_TUNNEL_NAME
    TunnelMapping     map[string]string // PIC_TUNNEL_CLASS_MAPPING
    BackendScheme     string            // PIC_BACKEND_SCHEME
    ResyncPeriod      time.Duration     // PIC_RESYNC_PERIOD
    LogLevel          string            // PIC_LOG_LEVEL
    WatchNamespaces   []string          // PIC_WATCH_NAMESPACES
}
```

## Relationships

```
┌─────────────┐         ┌─────────────────┐         ┌────────────────┐
│   Ingress   │ 1:1     │ PangolinResource│   N:1   │ PangolinTunnel │
│             │────────▶│                 │────────▶│                │
│  (watched)  │ creates │   (managed)     │  refs   │  (read-only)   │
└─────────────┘         └─────────────────┘         └────────────────┘
       │                        │
       │ references             │ routes to
       ▼                        ▼
┌─────────────┐         ┌────────────────┐
│   Service   │         │ Backend Target │
│  (read)     │         │ (FQDN + port)  │
└─────────────┘         └────────────────┘
```

## State Transitions

### Ingress Lifecycle → PangolinResource

| Ingress State | PangolinResource Action |
|---------------|------------------------|
| Created with `ingressClassName: pangolin` | Create PangolinResource |
| Updated (host, backend changed) | Update PangolinResource |
| Deleted | Delete PangolinResource (via ownerRef GC) |
| `ingressClassName` changed away from pangolin | Delete PangolinResource |
| Annotation `enabled: "false"` added | Delete PangolinResource |
| Annotation `enabled: "false"` removed | Create PangolinResource |

## Validation Rules

### Ingress Validation

| Rule | Validation | On Failure |
|------|------------|------------|
| Has `ingressClassName` | Must match `pangolin` or `pangolin-*` | Ignore Ingress |
| Has at least one rule | `spec.rules` not empty | Warning event, skip |
| Host is valid | Not wildcard, not IP, parseable | Warning event, skip rule |
| Path is root | `path == "/"` | Warning event, skip path |
| Backend exists | Service can be resolved | Requeue with backoff |

### PangolinResource Validation

| Rule | Validation | On Failure |
|------|------------|------------|
| Tunnel exists | PangolinTunnel with name found | Warning event, requeue |
| Name unique | No collision in namespace | Use deterministic hash |
| Labels set | All `pic.ingress.k8s.io/*` labels | Always set by PIC |
