# Implementation Plan: SSO & Multi-Path Support

**Branch**: `002-sso-multipath-support` | **Date**: 2025-12-03 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-sso-multipath-support/spec.md`

## Summary

Enable SSO/blockAccess control and multi-path routing for Pangolin Ingress Controller. PIC will propagate SSO annotations to PangolinResource CRD, and pangolin-operator will call Pangolin API `updateResource` after creation. Multi-path support requires creating one target per Ingress path with appropriate path matching configuration.

**Key Insight**: Pangolin API `createResource` does not accept SSO fields; they must be set via `updateResource` post-creation. Path routing is configured at the target level, not resource level.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: controller-runtime v0.17+, k8s.io/api networking/v1  
**Storage**: N/A (Kubernetes CRDs only)  
**Testing**: go test with envtest for integration tests  
**Target Platform**: Kubernetes 1.27+  
**Project Type**: Multi-project (PIC + pangolin-operator)  
**Performance Goals**: Reconciliation < 5s for SSO updates  
**Constraints**: No direct Pangolin API calls from PIC (constitution)  
**Scale/Scope**: Support Ingress with up to 10 paths

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| **I. Controller-Runtime First** | ✅ PASS | Uses existing reconciler pattern, no new controller needed |
| **II. CRD Interoperability** | ✅ PASS | PIC sets CRD fields only; operator handles API calls |
| **III. Test-First Development** | ⏳ PENDING | Tests to be written in Phase 2 |
| **IV. Observability** | ⏳ PENDING | Events for SSO changes to be added |
| **V. Minimal RBAC** | ✅ PASS | No new permissions required |

**Critical Compliance Note (Principle II)**:
- PIC MUST NOT call Pangolin API directly
- PIC sets `httpConfig.sso` and `httpConfig.blockAccess` in PangolinResource CRD
- pangolin-operator reads these fields and calls `updateResource` API
- This maintains clean separation of concerns

## Project Structure

### Documentation (this feature)

```text
specs/002-sso-multipath-support/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

**PIC (pangolin-ingress-controller)**:
```text
internal/
├── controller/
│   └── ingress_controller.go    # Add SSO annotation parsing, multi-path target generation
├── pangolincrd/
│   └── types.go                 # HTTPConfig.SSO, HTTPConfig.BlockAccess already added
│                                # Add Target.Path, Target.PathMatchType, Target.Priority
└── config/
    └── config.go                # No changes needed
```

**Operator (pangolin-operator)**:
```text
internal/controller/
└── pangolinresource_controller.go  # Add updateResource call after creation
                                    # Add multi-target support with path fields
pkg/pangolin/
├── client.go                       # Add UpdateResource method
└── types.go                        # Add SSO/BlockAccess to ResourceCreateSpec (done)
                                    # Add path fields to TargetCreateSpec
api/v1alpha1/
└── pangolinresource_types.go       # HTTPConfig.SSO, BlockAccess already added
                                    # Add Target.Path, PathMatchType, Priority
```

**Structure Decision**: Multi-project structure following existing patterns. Changes span both PIC and pangolin-operator repositories.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Two-phase API call (create + update) | Pangolin API limitation | Cannot modify Pangolin API createResource schema |
| Multi-target per resource | Ingress path semantics | Single target would break multi-path routing |
