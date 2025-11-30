# Pangolin Ingress Controller (PIC) — Technical Specification (MVP, Interop with pangolin-operator)

**Status:** Draft v2  
**Owner:** (to be decided)  
**Language:** Go (Golang)  
**Scope:** Kubernetes controller that exposes selected Kubernetes Ingresses via Pangolin by **creating/updating `PangolinResource` CRDs**, delegating all Pangolin API interaction to the existing `pangolin-operator` project.

---

## 1. High-Level Overview

### 1.1 Problem Statement

Pangolin is a tunneled reverse proxy that can expose internal services securely via “sites” (Newt agents + tunnels) and a central control plane. The existing `pangolin-operator` already provides:

- Custom Resource Definitions (CRDs) to model Pangolin entities (`PangolinTunnel`, `PangolinResource`, etc.).
- A controller that watches these CRDs and drives the Pangolin HTTP API accordingly.

What is missing:

- A **Kubernetes-native ingress experience** where operators can simply define an `Ingress` to expose a service, and Pangolin is automatically configured.

We want:

- A **Kubernetes Ingress Controller** that:
  - Watches `Ingress` resources.
  - Translates selected `Ingress` definitions into `PangolinResource` CRDs.
  - Lets `pangolin-operator` handle all low-level interactions with the Pangolin API.

### 1.2 MVP Scope

The **Pangolin Ingress Controller (PIC)**:

- Runs in a Kubernetes cluster as a standard controller.
- Watches `Ingress` resources (`networking.k8s.io/v1`).
- Uses `ingressClassName` and/or annotations to decide which Ingresses it manages.
- Computes the desired Pangolin exposure (host, domain, subdomain, backend).
- Creates/updates/deletes **`PangolinResource` CRDs** accordingly.

**Out of scope for MVP:**

- Managing Pangolin tunnels (Newt, WireGuard, Sites).
- Direct calls to the Pangolin HTTP API (this is exclusively handled by `pangolin-operator`).
- Managing TLS certificates.
- Creating new CRDs (we rely on those provided by `pangolin-operator`).
- Multi-cluster and multi-organization scenarios.

### 1.3 Key Assumptions

- `pangolin-operator` is installed in the cluster and its CRDs are available:
  - `PangolinTunnel`
  - `PangolinResource`
  - (Optionally `PangolinBinding`, etc.)
- There is at least one working Pangolin **tunnel** (`PangolinTunnel` CR) that represents a Newt-based site/tunnel already connected to the Pangolin instance.
- The `PangolinTunnel` CR and `PangolinResource` CR behave according to the upstream operator’s contract (status is updated, resources are created in Pangolin, etc.).
- PIC runs with enough RBAC permissions to:
  - Read `Ingress` resources.
  - Read `Services`.
  - Create/update/delete `PangolinResource` resources.
  - Read `PangolinTunnel` resources (to validate tunnel references).

---

## 2. Kubernetes Interface Design

### 2.1 Conceptual Model

PIC behaves like a **specialized Ingress Controller**, but instead of directly terminating network traffic, it:

- Watches `Ingress` objects.
- For those that are designated for Pangolin:
  - Derives a **desired set of `PangolinResource` CRs** representing the external exposure:
    - Domain / subdomain.
    - Tunnel.
    - Backend target (service + port).
- Ensures that `PangolinResource` CRs exist and match the desired state.
- Deletion of Ingress -> deletion of corresponding `PangolinResource` CRs (via ownerReferences or explicit cleanup).

### 2.2 Ingress Class & Management Decision

PIC decides whether to manage an Ingress based on `spec.ingressClassName` and optional annotations.

#### 2.2.1 Ingress Class Names

Two main operating modes:

1. **Single tunnel mode (MVP baseline)**

   ```yaml
   spec:
     ingressClassName: pangolin
   ```

   - All Ingresses with `ingressClassName: pangolin` are managed by PIC.
   - PIC uses a **default tunnel** name, e.g. `"default"`, resolved to a `PangolinTunnel` via configuration.
   - PIC sets `spec.tunnelRef.name` on `PangolinResource` to that default tunnel name.

2. **Multi-tunnel mode**

   ```yaml
   spec:
     ingressClassName: pangolin-edge-eu
   ```

   - PIC treats `pangolin-<suffix>` as valid classes:
     - Prefix: `pangolin-`
     - Suffix: `<tunnelName>` (e.g. `edge-eu`, `edge-us`).
   - PIC resolves `<tunnelName>` to a `PangolinTunnel` CR name using a configurable mapping.

Any Ingress with an `ingressClassName` not matching `pangolin` or `pangolin-*` is **ignored**.

#### 2.2.2 Annotations

PIC supports these annotations (all optional):

- `pangolin.ingress.k8s.io/enabled: "true|false"`  
  - Forces enable/disable for this Ingress.
  - If `"false"`, PIC deletes any previously-created `PangolinResource` CRs for this Ingress.

- `pangolin.ingress.k8s.io/tunnel-name: "<name>"`  
  - Overrides tunnel name derived from `ingressClassName`.
  - PIC must verify that a `PangolinTunnel` with this name exists (or treat as invalid and emit warning).

- `pangolin.ingress.k8s.io/domain-name: "<example.com>"`  
  - Overrides domain name derived from the Ingress host.

- `pangolin.ingress.k8s.io/subdomain: "<app>"`  
  - Overrides subdomain derived from the Ingress host.

### 2.3 Ingress Rules Mapping

Example Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  namespace: prod
spec:
  ingressClassName: pangolin
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 8080
```

PIC will:

1. Resolve **tunnel name** (`default` in single-tunnel mode).
2. Split host into `subdomain = "app"`, `domainName = "example.com"` unless overridden by annotations.
3. Construct the backend endpoint:
   - Target host: `"my-app.prod.svc.cluster.local"`
   - Target port: `8080`
   - Method: typically `http` (or `https` if configured).

PIC then defines a **desired `PangolinResource`**:

```yaml
apiVersion: tunnel.pangolin.io/v1alpha1
kind: PangolinResource
metadata:
  name: pic-prod-my-app-app-example-com
  namespace: prod
  labels:
    pic.ingress.k8s.io/uid: "<ingress-uid>"
    pic.ingress.k8s.io/name: "my-app"
    pic.ingress.k8s.io/namespace: "prod"
spec:
  enabled: true
  protocol: http
  tunnelRef:
    name: default
  httpConfig:
    domainName: "example.com"
    subdomain: "app"
  target:
    ip: "my-app.prod.svc.cluster.local"
    port: 8080
    method: http
```

> **MVP simplification:**  
> - One `PangolinResource` per `(Ingress, host)` combination.  
> - Paths:
>   - For MVP, PIC supports `path: "/"` only.  
>   - Non-root paths may be rejected with a clear event + status, or mapped in a simple “prefix” fashion in a later iteration once the Pangolin-side routing model is confirmed.

---

## 3. Interoperability Design with `pangolin-operator`

### 3.1 CRDs Used

PIC assumes the following CRDs are installed (provided by `pangolin-operator`):

- `apiVersion: tunnel.pangolin.io/v1alpha1, kind: PangolinTunnel`
- `apiVersion: tunnel.pangolin.io/v1alpha1, kind: PangolinResource`

We treat them as the authoritative Kubernetes representations of Pangolin tunnels and resources.

### 3.2 `PangolinTunnel` Usage

PIC uses `PangolinTunnel` as **reference anchors** for Ingress:

- PIC never creates or deletes `PangolinTunnel` CRs.
- It only **reads** them to:
  - Validate that a tunnel name used by Ingress exists.
  - Potentially read additional hints (e.g. siteId), if needed.

Tunnel resolution:

1. Single-tunnel mode:
   - Environment variable `PANGOLIN_DEFAULT_TUNNEL_NAME` defines a default `PangolinTunnel.metadata.name`.
   - `Ingress.spec.ingressClassName: pangolin` is mapped to this tunnel.

2. Multi-tunnel mode:
   - PIC derives a `tunnelAlias` from `ingressClassName: pangolin-<alias>`.
   - It uses a mapping (see configuration) to resolve `tunnelAlias` → actual `PangolinTunnel.metadata.name`.
   - Alternatively, `tunnelAlias` may coincide with `PangolinTunnel` name directly.

3. Annotation override:
   - `pangolin.ingress.k8s.io/tunnel-name` explicitly sets `spec.tunnelRef.name`.

If PIC cannot resolve a tunnel:

- It emits a Kubernetes event on the Ingress.
- It logs an error.
- It requeues the reconcile with backoff, but does not create any `PangolinResource`.

### 3.3 `PangolinResource` Lifecycle

PIC is responsible for **full lifecycle management** of `PangolinResource` objects that it creates:

- **Create** on Ingress create (if managed).
- **Update** on Ingress update (if managed).
- **Delete** when:
  - Ingress is deleted, or
  - Ingress changes to a non-managed class, or
  - `pangolin.ingress.k8s.io/enabled: "false"` is set.

`pangolin-operator` watches these `PangolinResource` CRDs and:

- Creates/updates/deletes the corresponding resources in the Pangolin API.
- Updates `status` fields (e.g. `status.url`, `status.resourceId`, `status.conditions`).

PIC does not use the Pangolin API directly; it only uses the CRDs and their status as needed.

---

## 4. Reconciliation Logic

### 4.1 Trigger Conditions

PIC runs reconciliation when:

- An `Ingress` is created, updated, or deleted.
- A periodic resync occurs (configurable, e.g. every 5 minutes).
- Optionally, when a `PangolinResource` with `pic.ingress.k8s.io/*` labels changes (for sanity check / drift detection).

### 4.2 Reconcile Algorithm (Ingress)

Given a reconciliation request for a key `(namespace, name)`:

1. **Fetch Ingress**
   - If not found (deleted):
     - Perform cleanup: delete any `PangolinResource` CRs that refer to this Ingress and are owned by PIC (via ownerReferences or labels).
     - Return.

2. **Determine if managed**
   - Inspect `spec.ingressClassName`.
   - If not `pangolin` or `pangolin-*`, and no overriding annotations → **not managed**.
   - If annotation `pangolin.ingress.k8s.io/enabled: "false"` → **not managed**.
     - If `PangolinResource` exists: delete it.
     - Return.

3. **Resolve tunnel**
   - Derive tunnel name via:
     - `pangolin.ingress.k8s.io/tunnel-name` annotation, OR
     - `ingressClassName == "pangolin"` → default tunnel name, OR
     - `ingressClassName == "pangolin-<alias>"` → alias mapping.
   - Validate with `GET PangolinTunnel` (Kubernetes API).
   - If tunnel not found:
     - Emit warning event on Ingress.
     - Requeue with backoff.
     - Return.

4. **Compute desired `PangolinResource` set**
   - For MVP:
     - Require exactly one host and one path `/` in `spec.rules`.
       - If multiple hosts or non-root paths, emit a warning and **skip** (or handle only the first host/path).
   - Split host into domain/subdomain, factoring in override annotations.
   - Compute backend endpoint from the referenced `Service`:
     - `serviceName.namespace.svc.cluster.local`
     - Port number from backend.
     - Method = `"http"` by default.
   - Compute deterministic name for `PangolinResource`:
     - e.g. `pic-<namespace>-<ingressName>-<hostHash>`.

5. **Fetch current `PangolinResource` (if any)**
   - Query for `PangolinResource` with:
     - `metadata.namespace = Ingress.namespace`
     - `metadata.labels["pic.ingress.k8s.io/uid"] == <Ingress UID>`
   - Or directly by deterministic name.

6. **Create / Update / Delete**

   - If no `PangolinResource` exists:
     - Create a new one:
       - Set `metadata.ownerReferences` pointing to the Ingress (to enable garbage collection).
       - Set `labels` as above.
       - Set desired spec (`enabled`, `protocol`, `tunnelRef`, `httpConfig`, `target`).
   - If one exists:
     - Compare current spec with desired spec.
     - If differences, update the spec accordingly.
   - If more than one exists (data corruption / drift):
     - Keep the one matching deterministic name; delete others (with caution & logging).

7. **Observe status (optional)**
   - PIC may read `status` of `PangolinResource` to:
     - Emit events on Ingress when the resource becomes ready or fails.
   - MVP requirement: optional — it’s useful but not mandatory for basic functionality.

### 4.3 Deletion Path

When an Ingress is deleted:

- If ownerReferences were set correctly on `PangolinResource`, Kubernetes garbage collector will delete the `PangolinResource` CR automatically.
- If ownerReferences are not used (alternative model), PIC must explicitly:
  - List `PangolinResource` CRs with `pic.ingress.k8s.io/uid == <old UID>`.
  - Delete them.

Design choice:

- **MVP recommendation**: use `ownerReferences` for automatic cascading deletion, plus labels for matching when needed.

---

## 5. Configuration & Deployment

### 5.1 Environment Variables

PIC configuration via environment variables (or ConfigMap):

- `PIC_DEFAULT_TUNNEL_NAME`  
  - Name of the default `PangolinTunnel` for `ingressClassName: pangolin`.

- `PIC_TUNNEL_CLASS_MAPPING`  
  - Mapping of ingress class suffix → tunnel name.  
  - Example (multi-line, key=value per line):
    ```text
    edge-eu=edge-eu-tunnel
    edge-us=edge-us-tunnel
    staging=staging-tunnel
    ```

- `PIC_BACKEND_SCHEME`  
  - Default `"http"` or `"https"` for backend services.

- `PIC_RESYNC_PERIOD`  
  - Controller resync period (e.g. `"5m"`).

- `PIC_LOG_LEVEL`  
  - `"debug"`, `"info"`, `"warn"`, `"error"`.

- `PIC_WATCH_NAMESPACES` (optional)  
  - Comma-separated list of namespaces to watch.
  - If empty, PIC watches all namespaces.

### 5.2 RBAC Requirements

PIC’s ServiceAccount needs:

- `get`, `list`, `watch` on:
  - `Ingress` (`networking.k8s.io/v1`).
  - `Service` (`v1`).
  - `PangolinTunnel` (`tunnel.pangolin.io/v1alpha1`).
  - `PangolinResource` (`tunnel.pangolin.io/v1alpha1`).

- `create`, `update`, `patch`, `delete` on:
  - `PangolinResource` only.

No write permissions on `Ingress`, `Service`, or `PangolinTunnel` are needed.

### 5.3 Deployment

- PIC is packaged as a container image with a statically linked Go binary.
- Deploy as a `Deployment`:
  - `replicas: 1` for MVP; enable leader election to support scaling later.
- Add:
  - Liveness probe: HTTP `/healthz` endpoint.
  - Readiness probe: may reuse `/healthz` or `/readyz`.
- Expose metrics endpoint (e.g. `:8080/metrics`) for Prometheus scraping.

---

## 6. Go Implementation Design

### 6.1 Project Layout

Recommended layout:

```text
pangolin-ingress-controller/
├── cmd/
│   └── manager/main.go
├── internal/
│   ├── controller/
│   │   └── ingress_controller.go
│   ├── pangolincrd/
│   │   ├── types.go          # Go structs mirroring PangolinTunnel/PangolinResource
│   │   └── scheme.go         # register CRD scheme
│   ├── config/
│   │   └── config.go         # env parsing & config struct
│   └── util/
│       ├── naming.go         # deterministic names
│       └── hostsplit.go      # host → (subdomain, domain)
├── config/
│   ├── rbac/
│   ├── manager/
│   └── samples/
├── go.mod
└── go.sum
```

We use `controller-runtime` as the basis for the manager + controller.

### 6.2 Configuration Module

`internal/config/config.go`:

- Loads environment variables.
- Provides a struct:

```go
type TunnelMapping map[string]string // ingressClass suffix -> tunnel name

type Config struct {
    DefaultTunnelName string
    TunnelMapping     TunnelMapping
    BackendScheme     string
    ResyncPeriod      time.Duration
    LogLevel          string
    WatchNamespaces   []string
}
```

### 6.3 CRD Types Module

`internal/pangolincrd/types.go`:

- Defines Go structs for `PangolinTunnel` and `PangolinResource`:
  - Either:
    - Import types from `bovf/pangolin-operator` if it publishes them as a Go module.
  - Or:
    - Reimplement minimal struct definitions matching the CRDs’ API fields.

Example (simplified):

```go
type PangolinResourceSpec struct {
    Enabled    bool   `json:"enabled"`
    Protocol   string `json:"protocol,omitempty"`
    TunnelRef  TunnelRef `json:"tunnelRef,omitempty"`
    HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
    Target     *Target     `json:"target,omitempty"`
}

type TunnelRef struct {
    Name string `json:"name,omitempty"`
}

type HTTPConfig struct {
    DomainName string `json:"domainName,omitempty"`
    Subdomain  string `json:"subdomain,omitempty"`
}

type Target struct {
    IP     string `json:"ip,omitempty"`
    Port   int32  `json:"port,omitempty"`
    Method string `json:"method,omitempty"`
}
```

And the full CR types:

```go
type PangolinResource struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   PangolinResourceSpec   `json:"spec,omitempty"`
    Status PangolinResourceStatus `json:"status,omitempty"`
}

type PangolinResourceList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []PangolinResource `json:"items"`
}
```

### 6.4 Main Entry

`cmd/manager/main.go`:

- Parse configuration.
- Create a `ctrl.Manager`.
- Register the CRD schemes.
- Add the Ingress reconciler.

Pseudo-code:

```go
func main() {
    cfg := config.MustLoad()

    scheme := runtime.NewScheme()
    _ = clientgoscheme.AddToScheme(scheme)
    _ = pangolincrd.AddToScheme(scheme) // register Pangolin CRDs

    mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
        Scheme:             scheme,
        MetricsBindAddress: ":8080",
        LeaderElection:     true,
        LeaderElectionID:   "pangolin-ingress-controller",
    })
    if err != nil {
        // log & exit
    }

    reconciler := controller.NewIngressReconciler(
        mgr.GetClient(),
        cfg,
        mgr.GetScheme(),
    )

    if err := reconciler.SetupWithManager(mgr); err != nil {
        // log & exit
    }

    if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
        // log & exit
    }
}
```

### 6.5 Ingress Reconciler

`internal/controller/ingress_controller.go`:

Key fields:

```go
type IngressReconciler struct {
    client.Client
    Scheme *runtime.Scheme
    Config *config.Config
    Log    logr.Logger
}
```

`Reconcile` logic:

1. Fetch Ingress.
2. If not found:
   - Delete or rely on GC for `PangolinResource`.
3. Determine if managed (ingressClassName, annotations).
4. Resolve tunnel (via Config + `PangolinTunnel` lookup).
5. Compute desired `PangolinResource` spec.
6. Get or create `PangolinResource`:
   - Use deterministic name.
   - Set ownerReference to Ingress.
7. Compare and update spec if needed.
8. Handle errors with proper `ctrl.Result` and backoff.

`SetupWithManager`:

```go
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&networkingv1.Ingress{}).
        Owns(&pangolincrd.PangolinResource{}).
        Complete(r)
}
```

This ensures:

- Reconcile is triggered on Ingress and on `PangolinResource` changes that we own.

---

## 7. Observability, Logging & Metrics

### 7.1 Logging

- Use `logr` from `controller-runtime`.
- Include contextual fields:
  - `namespace`, `ingress`, `host`, `tunnel`, `pangolinResource`.
- At:

  - `Info` for normal operations (create/update/delete).
  - `Error` for failures.
  - `Debug` for detailed diff and spec comparisons.

### 7.2 Metrics

Expose Prometheus metrics:

- `pic_reconcile_total{result="success|error"}`  
- `pic_pangolinresource_managed` — number of `PangolinResource` CRs owned by PIC.
- Optional histograms for reconcile duration.

Metrics registration via `controller-runtime` metrics registry.

---

## 8. Security Considerations

- PIC never handles Pangolin API tokens directly (that’s `pangolin-operator`’s job).
- RBAC is minimal and strictly read-only on Ingress/Service/Tunnel, write-only on `PangolinResource`.
- `ownerReferences` ensure automatic cleanup if Ingress is removed.
- In multi-tenant clusters:
  - `PIC_WATCH_NAMESPACES` can restrict PIC to specific namespaces.
  - Different PIC instances could be deployed per-namespace or per-tenant with separate config.

---

## 9. Future Enhancements

- **Path routing support**
  - Map Kubernetes `path`/`pathType` to Pangolin’s path-based routing (via either `PangolinResource` fields or new CRDs).
- **HTTPS / TLS integration**
  - Configure TLS certificates via Cert-Manager and propagate to Pangolin (possibly via additional `PangolinResource` fields).
- **CRD `PangolinIngress`**
  - Define a more expressive API (beyond standard Ingress) but keep compatibility with PIC.
- **Status reflection**
  - Update Ingress conditions based on `PangolinResource.status` (e.g. “PangolinReady”, “PangolinError”).
- **Direct Pangolin API mode**
  - Optional alternative mode where PIC talks directly to the Pangolin API, bypassing `pangolin-operator` for users who don’t want extra CRDs.

---

## 10. MVP Acceptance Criteria

1. **Functional**

   - With `pangolin-operator` installed and at least one `PangolinTunnel` configured:
     - Creating an Ingress with `ingressClassName: pangolin` and a single host/root path creates a matching `PangolinResource` in the same namespace.
     - Updating the Ingress host or backend updates the corresponding `PangolinResource`.
     - Deleting the Ingress removes the `PangolinResource` (via ownerReference GC or explicit delete).

2. **Interoperability**

   - `pangolin-operator` sees the `PangolinResource` created by PIC and successfully creates/updates/deletes the corresponding resources in Pangolin.
   - PIC does not interfere with `PangolinResource` objects that it does **not** own (no labels / no relevant ownerReferences).

3. **Robustness**

   - PIC handles temporary API errors with retries and does not crash.
   - Misconfigured ingressClass or tunnel mapping results in clear Kubernetes events and logs.

4. **Deployability**

   - PIC can be deployed with a provided set of manifests (RBAC, Deployment).
   - Works with existing `pangolin-operator` installation without modification.

---

_End of MVP Technical Specification for Pangolin Ingress Controller — Interop with pangolin-operator._
