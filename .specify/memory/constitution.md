<!--
  Sync Impact Report:
  - Version change: 0.0.0 → 1.0.0 (initial ratification)
  - Modified principles: N/A (initial version)
  - Added sections: Core Principles (5), Technical Standards, Development Workflow, Governance
  - Removed sections: N/A
  - Templates updated:
    - plan-template.md: ✅ Constitution Check section aligns with principles
    - spec-template.md: ✅ User stories and acceptance criteria align
    - tasks-template.md: ✅ Phase structure supports TDD and MVP-first approach
  - Follow-up TODOs: None
-->

# Pangolin Ingress Controller Constitution

## Core Principles

### I. Controller-Runtime First

All Kubernetes integration MUST use controller-runtime patterns and idioms.

- Controllers MUST implement the `Reconciler` interface with idempotent `Reconcile` methods
- CRD type definitions MUST use proper scheme registration via `AddToScheme`
- Watch configurations MUST use `For()`, `Owns()`, and `Watches()` builder patterns
- Leader election MUST be enabled for production deployments
- Metrics and health endpoints MUST be exposed via the manager

**Rationale**: controller-runtime provides battle-tested patterns for Kubernetes controllers, ensuring correctness, performance, and maintainability.

### II. CRD Interoperability (NON-NEGOTIABLE)

PIC MUST NOT duplicate functionality provided by pangolin-operator.

- PIC MUST only read `PangolinTunnel` resources (never create, update, or delete)
- PIC MUST manage the full lifecycle of `PangolinResource` CRs it creates
- PIC MUST NOT call the Pangolin HTTP API directly; all API interaction is delegated to pangolin-operator
- Ownership MUST be established via `ownerReferences` to enable garbage collection
- Labels with `pic.ingress.k8s.io/*` prefix MUST identify PIC-managed resources

**Rationale**: Clear separation of concerns prevents conflicts and reduces complexity. pangolin-operator is the single source of truth for Pangolin API state.

### III. Test-First Development

TDD is mandatory: tests MUST be written and fail before implementation.

- Unit tests MUST cover all reconciliation logic branches
- Integration tests MUST verify Ingress→PangolinResource transformation
- Contract tests MUST validate CRD schema compatibility with pangolin-operator
- Table-driven tests MUST be used for reconciliation scenarios
- Mocks MUST use controller-runtime's `fake.Client` for Kubernetes API interactions

**Rationale**: Kubernetes controllers have complex state machine behavior; comprehensive tests prevent regressions and document expected behavior.

### IV. Observability

All operations MUST be observable through logs, metrics, and Kubernetes events.

- Logging MUST use `logr` with contextual fields: `namespace`, `ingress`, `tunnel`, `pangolinResource`
- Prometheus metrics MUST track: `reconcile_total`, `reconcile_duration`, `resources_managed`
- Kubernetes events MUST be emitted for: resource creation, tunnel resolution failures, spec changes
- Error messages MUST include actionable context for debugging
- Debug-level logs MUST show spec diffs during updates

**Rationale**: Operators need visibility into controller behavior; structured observability reduces mean time to resolution.

### V. Minimal RBAC & Security

Controller permissions MUST follow the principle of least privilege.

- Read-only permissions on: `Ingress`, `Service`, `PangolinTunnel`
- Write permissions ONLY on: `PangolinResource`
- No cluster-wide permissions unless explicitly required
- Secrets MUST NOT be accessed unless strictly necessary
- Namespace-scoped deployments MUST be supported via `PIC_WATCH_NAMESPACES`

**Rationale**: Minimal permissions reduce attack surface and blast radius in multi-tenant clusters.

## Technical Standards

**Language**: Go 1.21+
**Framework**: controller-runtime v0.17+
**Kubernetes**: networking.k8s.io/v1 Ingress API
**CRDs**: tunnel.pangolin.io/v1alpha1 (PangolinTunnel, PangolinResource)
**Testing**: go test with envtest for integration tests
**Linting**: golangci-lint with default configuration

**Project Structure**:
```text
cmd/manager/         # Entry point
internal/controller/ # Reconciliation logic
internal/config/     # Environment configuration
internal/pangolincrd/# CRD type definitions
```

**Naming Conventions**:
- PangolinResource names: `pic-<namespace>-<ingress>-<host-hash>`
- Labels: `pic.ingress.k8s.io/{uid,name,namespace}`

## Development Workflow

1. **Specification**: Define feature via spec.md with acceptance scenarios
2. **Design**: Document approach in plan.md with constitution compliance check
3. **Tests**: Write failing tests covering acceptance scenarios
4. **Implementation**: Implement until tests pass
5. **Observability**: Add logging, metrics, and events
6. **Documentation**: Update README and deployment manifests
7. **Review**: Verify RBAC minimalism and CRD interoperability

**MVP-First**: Features MUST be scoped to deliver working functionality incrementally. Each user story MUST be independently testable and deployable.

## Governance

- This constitution supersedes all other development practices for PIC
- Amendments require: documentation of rationale, version bump, template sync
- All PRs MUST include constitution compliance statement
- Complexity additions MUST be justified in plan.md Complexity Tracking section
- Runtime development guidance: refer to `speckit/constitution.md` for detailed technical specification

**Version**: 1.0.0 | **Ratified**: 2025-11-30 | **Last Amended**: 2025-11-30
