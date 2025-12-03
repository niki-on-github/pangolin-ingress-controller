# Feature Specification: SSO & Multi-Path Support

**Feature Branch**: `002-sso-multipath-support`  
**Created**: 2025-12-03  
**Status**: Draft  
**Input**: User description: "Implement SSO/blockAccess support by calling updateResource after creation, and add multi-path support by creating one target per Ingress path"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Disable SSO Authentication on Ingress (Priority: P1)

As a platform operator, I want to deploy an Ingress resource with SSO disabled so that my service is publicly accessible without authentication.

**Why this priority**: By default, Pangolin enables SSO on new resources, blocking access. Operators need to explicitly disable SSO for public services. This is the most critical user need since services are currently inaccessible.

**Independent Test**: Deploy an Ingress with `pangolin.ingress.k8s.io/sso: "false"` annotation and verify the service is publicly accessible without authentication prompt.

**Acceptance Scenarios**:

1. **Given** an Ingress with annotation `pangolin.ingress.k8s.io/sso: "false"`, **When** the PangolinResource is created and reconciled, **Then** the Pangolin resource has SSO disabled and the service is publicly accessible.
2. **Given** an Ingress without SSO annotation, **When** the PangolinResource is created, **Then** SSO defaults to disabled (open access by default).
3. **Given** an existing PangolinResource with SSO enabled, **When** the Ingress annotation is changed to `sso: "false"`, **Then** the resource is updated and SSO is disabled.

---

### User Story 2 - Enable SSO Authentication on Ingress (Priority: P2)

As a platform operator, I want to enable SSO authentication on specific Ingress resources to restrict access to authenticated users only.

**Why this priority**: Some services require authentication. This builds on the SSO disable capability by allowing explicit enablement.

**Independent Test**: Deploy an Ingress with `pangolin.ingress.k8s.io/sso: "true"` and `pangolin.ingress.k8s.io/block-access: "true"` annotations and verify authentication is required to access the service.

**Acceptance Scenarios**:

1. **Given** an Ingress with annotation `pangolin.ingress.k8s.io/sso: "true"`, **When** the PangolinResource is created, **Then** the Pangolin resource has SSO enabled.
2. **Given** an Ingress with annotations `sso: "true"` and `block-access: "true"`, **When** an unauthenticated user accesses the service, **Then** they are redirected to authentication.
3. **Given** an Ingress with `sso: "true"` and `block-access: "false"`, **When** an unauthenticated user accesses the service, **Then** they can access it but are not identified.

---

### User Story 3 - Multi-Path Ingress Support (Priority: P3)

As a platform operator, I want to define multiple paths in a single Ingress resource so that different backends can serve different URL paths under the same hostname.

**Why this priority**: This enables more complex routing scenarios but is less critical than basic SSO support. Many deployments use single-path Ingress resources.

**Independent Test**: Deploy an Ingress with multiple paths (e.g., `/api` and `/web`) pointing to different services, and verify each path routes to its correct backend.

**Acceptance Scenarios**:

1. **Given** an Ingress with paths `/api` and `/web`, **When** the PangolinResource is created, **Then** two targets are created in Pangolin, one for each path.
2. **Given** a request to `https://example.com/api/users`, **When** the request is processed, **Then** it is routed to the backend configured for `/api`.
3. **Given** a request to `https://example.com/web/dashboard`, **When** the request is processed, **Then** it is routed to the backend configured for `/web`.
4. **Given** an Ingress path is removed, **When** the PangolinResource is reconciled, **Then** the corresponding target is deleted from Pangolin.

---

### Edge Cases

- What happens when an Ingress has overlapping paths (e.g., `/api` and `/api/v2`)? → System should use path priority based on specificity (longer paths have higher priority).
- What happens when SSO annotation value is invalid (not "true" or "false")? → System should treat invalid values as "false" (default to open access).
- What happens when a target creation fails for one path but succeeds for others? → System should report partial failure in status and retry failed targets.
- What happens when the Pangolin API updateResource call fails? → System should retry with exponential backoff and report status condition.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Pangolin Operator MUST call the Pangolin API `updateResource` endpoint after resource creation to set SSO and blockAccess settings.
- **FR-002**: PIC MUST read `pangolin.ingress.k8s.io/sso` annotation and propagate the value to PangolinResource spec.
- **FR-003**: PIC MUST read `pangolin.ingress.k8s.io/block-access` annotation and propagate the value to PangolinResource spec.
- **FR-004**: SSO MUST default to `false` (disabled) when annotation is absent or invalid.
- **FR-005**: blockAccess MUST default to `false` (access allowed) when annotation is absent or invalid.
- **FR-006**: PIC MUST create one Pangolin target per Ingress path entry.
- **FR-007**: Each target MUST include the path, pathMatchType, and priority from the Ingress path configuration.
- **FR-008**: Pangolin Operator MUST support the `path`, `pathMatchType`, and `priority` fields when creating targets via the API.
- **FR-009**: System MUST reconcile target additions and deletions when Ingress paths change.
- **FR-010**: PangolinResource status MUST reflect the SSO/blockAccess configuration state.

### Key Entities

- **PangolinResource**: Kubernetes CRD representing a Pangolin resource; extended with `httpConfig.sso` and `httpConfig.blockAccess` boolean fields.
- **Target**: Backend service endpoint within a PangolinResource; extended with `path`, `pathMatchType`, `rewritePath`, and `priority` fields.
- **Ingress**: Kubernetes native resource; annotations control SSO settings, paths define routing rules.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Services deployed with `sso: "false"` annotation are publicly accessible without authentication within 30 seconds of Ingress creation.
- **SC-002**: Services deployed with `sso: "true"` and `block-access: "true"` require authentication before access is granted.
- **SC-003**: Multi-path Ingress resources correctly route traffic to the appropriate backend for each defined path.
- **SC-004**: 100% of SSO/blockAccess annotation changes are reflected in Pangolin within one reconciliation cycle.
- **SC-005**: Target count in Pangolin matches the number of paths defined in the Ingress resource.

## Assumptions

- The Pangolin API `updateResource` endpoint is available and accepts `sso` and `blockAccess` fields (confirmed in source code analysis).
- The Pangolin API `createTarget` endpoint accepts `path`, `pathMatchType`, `rewritePath`, `rewritePathType`, and `priority` fields (confirmed in source code analysis).
- The Pangolin Operator has network access to the Pangolin API with valid credentials.
- Path matching in Pangolin follows standard prefix matching semantics by default.
