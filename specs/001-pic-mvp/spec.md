# Feature Specification: Pangolin Ingress Controller MVP

**Feature Branch**: `001-pic-mvp`  
**Created**: 2025-11-30  
**Status**: Draft  
**Input**: User description: "Implement Pangolin Ingress Controller MVP"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Expose Service via Ingress (Priority: P1)

As a Kubernetes operator, I want to expose an internal service to the internet via Pangolin by simply creating a standard Kubernetes Ingress resource, so that I don't need to manually configure Pangolin or learn its API.

**Why this priority**: This is the core value proposition—enabling Kubernetes-native experience for Pangolin. Without this, there is no product.

**Independent Test**: Deploy an Ingress with `ingressClassName: pangolin` and verify the service becomes accessible via Pangolin within 60 seconds.

**Acceptance Scenarios**:

1. **Given** pangolin-operator is installed with a working tunnel, **When** I create an Ingress with `ingressClassName: pangolin` and host `app.example.com`, **Then** a corresponding Pangolin resource is created and my service is exposed at that hostname.

2. **Given** an Ingress exists with `ingressClassName: pangolin`, **When** I check the cluster resources, **Then** I see a PangolinResource owned by that Ingress with the correct domain/subdomain configuration.

3. **Given** no tunnel is configured, **When** I create an Ingress with `ingressClassName: pangolin`, **Then** I receive a clear warning event explaining the tunnel is missing.

---

### User Story 2 - Update Exposed Service (Priority: P2)

As a Kubernetes operator, I want changes to my Ingress (host, backend service, port) to automatically update the Pangolin exposure, so that I can manage my external access through standard Kubernetes workflows.

**Why this priority**: Operators expect Ingress updates to propagate automatically. This completes the CRUD lifecycle.

**Independent Test**: Modify an existing Ingress host and verify the Pangolin resource updates within 30 seconds.

**Acceptance Scenarios**:

1. **Given** an Ingress `myapp` is exposed via Pangolin at `app.example.com`, **When** I update the host to `newapp.example.com`, **Then** the Pangolin resource is updated to reflect the new hostname.

2. **Given** an Ingress points to service `frontend:80`, **When** I change the backend to `api:8080`, **Then** the Pangolin resource target is updated accordingly.

---

### User Story 3 - Remove Exposed Service (Priority: P2)

As a Kubernetes operator, I want deleting an Ingress to automatically remove the Pangolin exposure, so that I don't have orphaned external endpoints.

**Why this priority**: Cleanup is essential for security and resource hygiene. Same priority as updates since both complete lifecycle.

**Independent Test**: Delete an Ingress and verify the associated Pangolin resource is removed within 30 seconds.

**Acceptance Scenarios**:

1. **Given** an Ingress is exposed via Pangolin, **When** I delete the Ingress, **Then** the corresponding Pangolin resource is automatically deleted.

2. **Given** an Ingress is exposed via Pangolin, **When** I change its `ingressClassName` to something else, **Then** the Pangolin resource is removed.

---

### User Story 4 - Override Domain Configuration (Priority: P3)

As a Kubernetes operator, I want to override the domain or subdomain derived from the Ingress host using annotations, so that I have flexibility for edge cases where automatic splitting doesn't work.

**Why this priority**: Annotations provide escape hatches for advanced users but aren't required for basic functionality.

**Independent Test**: Create an Ingress with domain override annotations and verify the custom values are used.

**Acceptance Scenarios**:

1. **Given** an Ingress with host `internal.corp.example.com`, **When** I add annotation `pangolin.ingress.k8s.io/subdomain: myapp`, **Then** the Pangolin resource uses subdomain `myapp` instead of the automatically derived value.

---

### User Story 5 - Disable Exposure Temporarily (Priority: P3)

As a Kubernetes operator, I want to temporarily disable Pangolin exposure for an Ingress without deleting it, so that I can perform maintenance without recreating resources.

**Why this priority**: Operational convenience for maintenance scenarios.

**Independent Test**: Set the enabled annotation to false and verify the Pangolin resource is removed; set it back to true and verify recreation.

**Acceptance Scenarios**:

1. **Given** an Ingress is exposed via Pangolin, **When** I add annotation `pangolin.ingress.k8s.io/enabled: "false"`, **Then** the Pangolin resource is deleted but the Ingress remains.

2. **Given** an Ingress has `enabled: "false"` annotation, **When** I remove the annotation or set it to `"true"`, **Then** the Pangolin resource is recreated.

---

### Edge Cases

- **Missing backend service**: System emits a warning event and retries until the service exists
- **Invalid host format (wildcards, IPs)**: System emits an error event and skips the Ingress rule
- **Multiple hosts in one Ingress**: System processes only the first host (MVP limitation) with a warning
- **Non-root paths**: System skips rules with non-`/` paths and emits a warning (MVP limitation)
- **Tunnel disappears after resource creation**: System detects on next reconciliation and emits warning

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST watch Kubernetes Ingress resources and create corresponding Pangolin resources for managed Ingresses
- **FR-002**: System MUST identify managed Ingresses by `ingressClassName` matching `pangolin` or `pangolin-*` pattern
- **FR-003**: System MUST automatically delete Pangolin resources when their source Ingress is deleted
- **FR-004**: System MUST validate that referenced tunnels exist before creating Pangolin resources
- **FR-005**: System MUST emit Kubernetes events on the Ingress for significant state changes (created, updated, failed)
- **FR-006**: System MUST support annotation-based overrides for domain, subdomain, tunnel name, and enabled state
- **FR-007**: System MUST use a default tunnel when `ingressClassName` is exactly `pangolin`
- **FR-008**: System MUST derive tunnel name from `ingressClassName` suffix when pattern is `pangolin-<name>`
- **FR-009**: System MUST split hostnames into domain and subdomain components correctly (including public suffix handling)
- **FR-010**: System MUST log all reconciliation actions with contextual information (namespace, ingress, host)

### Key Entities

- **Ingress**: Standard Kubernetes networking resource that operators create to define external access; contains host, paths, and backend service references
- **PangolinResource**: Custom resource managed by pangolin-operator that represents an exposed endpoint in Pangolin; created/owned by PIC
- **PangolinTunnel**: Custom resource representing a Pangolin tunnel/site; read-only for PIC, must exist before Ingresses can be exposed

## Assumptions

- pangolin-operator is installed and functioning in the cluster
- At least one PangolinTunnel resource exists and is in a ready state
- The PangolinResource CRD schema matches the expected structure (tunnelRef, httpConfig, target fields)
- Operators have RBAC permissions to create Ingress resources
- DNS for exposed hostnames is configured to point to Pangolin

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can expose a service via Ingress in under 2 minutes (time from Ingress creation to service accessibility)
- **SC-002**: 100% of Ingress deletions result in corresponding Pangolin resource cleanup within 60 seconds
- **SC-003**: System correctly handles 100 managed Ingresses without performance degradation
- **SC-004**: Configuration errors (missing tunnel, invalid host) result in clear, actionable event messages within 5 seconds
- **SC-005**: System recovers automatically from transient failures (API errors, network issues) without operator intervention
- **SC-006**: Operators can deploy and configure the controller in under 10 minutes using provided manifests
