# Implementation Plan: Pangolin Ingress Controller MVP

**Branch**: `001-pic-mvp` | **Date**: 2025-11-30 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/001-pic-mvp/spec.md`

## Summary

Build a Kubernetes Ingress Controller that watches Ingress resources and creates corresponding `PangolinResource` CRDs, enabling operators to expose services via Pangolin using standard Kubernetes workflows. The controller delegates all Pangolin API interaction to the existing `pangolin-operator`.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: controller-runtime v0.17+, client-go, golang.org/x/net/publicsuffix  
**Storage**: N/A (Kubernetes API server is the data store)  
**Testing**: go test with envtest for integration tests, testify/assert  
**Target Platform**: Linux containers (amd64/arm64), Kubernetes 1.28+  
**Project Type**: Single project (Kubernetes controller)  
**Performance Goals**: Reconcile 100 Ingresses in <10 seconds, individual reconcile <500ms  
**Constraints**: <100MB memory, minimal RBAC, leader election for HA  
**Scale/Scope**: 100+ managed Ingresses per cluster

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Status |
|-----------|-------------|--------|
| I. Controller-Runtime First | Use Reconciler interface, scheme registration, For/Owns builders | ✅ Planned |
| II. CRD Interoperability | Read-only PangolinTunnel, manage PangolinResource lifecycle, ownerReferences | ✅ Planned |
| III. Test-First Development | Unit tests for logic, integration tests with envtest, table-driven | ✅ Planned |
| IV. Observability | logr logging, Prometheus metrics, Kubernetes events | ✅ Planned |
| V. Minimal RBAC | Read Ingress/Service/Tunnel, Write only PangolinResource | ✅ Planned |

**Gate Status**: ✅ PASS - All principles addressed in design

## Project Structure

### Documentation (this feature)

```text
specs/001-pic-mvp/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CRD schemas)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── manager/
    └── main.go              # Entry point, manager setup

internal/
├── controller/
│   └── ingress_controller.go    # Reconciliation logic
├── config/
│   └── config.go                # Environment configuration
├── pangolincrd/
│   ├── types.go                 # PangolinResource, PangolinTunnel structs
│   └── scheme.go                # Scheme registration
└── util/
    ├── hostsplit.go             # Host → domain/subdomain splitting
    └── naming.go                # Deterministic resource naming

config/
├── rbac/
│   ├── role.yaml
│   └── role_binding.yaml
├── manager/
│   └── deployment.yaml
└── samples/
    └── ingress.yaml

tests/
├── unit/
│   ├── hostsplit_test.go
│   ├── naming_test.go
│   └── config_test.go
└── integration/
    ├── reconciler_test.go
    └── lifecycle_test.go

Dockerfile
Makefile
go.mod
go.sum
```

**Structure Decision**: Standard Go Kubernetes controller layout following controller-runtime conventions. Single binary deployed as a Kubernetes Deployment.

## Complexity Tracking

> No violations - design follows constitution principles
