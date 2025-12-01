# Tasks: Pangolin Ingress Controller MVP

**Input**: Design documents from `/specs/001-pic-mvp/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: TDD mandatory per constitution (Principle III)

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1, US2, US3, US4, US5 (maps to spec.md user stories)

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Initialize Go project and basic structure

- [x] T001 Initialize Go module with `go mod init github.com/your-org/pangolin-ingress-controller` in go.mod
- [x] T002 [P] Create directory structure: cmd/manager/, internal/controller/, internal/config/, internal/pangolincrd/, internal/util/
- [x] T003 [P] Create directory structure: config/rbac/, config/manager/, config/samples/, tests/unit/, tests/integration/
- [x] T004 [P] Create Makefile with targets: test, build, docker-build, docker-push, manifests
- [x] T005 [P] Create Dockerfile with multi-stage build in Dockerfile
- [x] T006 Add controller-runtime, client-go, publicsuffix dependencies to go.mod

**Checkpoint**: Project compiles with `go build ./...`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure required by ALL user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests (Write First, Must Fail)

- [x] T007 [P] Write unit tests for Config loading in tests/unit/config_test.go
- [x] T008 [P] Write unit tests for host splitting in tests/unit/hostsplit_test.go
- [x] T009 [P] Write unit tests for deterministic naming in tests/unit/naming_test.go

### Implementation

- [x] T010 [P] Implement Config struct and environment loading in internal/config/config.go
- [x] T011 [P] Implement SplitHost function with publicsuffix in internal/util/hostsplit.go
- [x] T012 [P] Implement GenerateName function in internal/util/naming.go
- [x] T013 Define PangolinResource and PangolinTunnel Go types in internal/pangolincrd/types.go
- [x] T014 Implement scheme registration (AddToScheme) in internal/pangolincrd/scheme.go
- [x] T015 Create RBAC ClusterRole with minimal permissions in config/rbac/role.yaml
- [x] T016 [P] Create ClusterRoleBinding in config/rbac/role_binding.yaml
- [x] T017 Create sample Ingress manifest in config/samples/ingress.yaml

**Checkpoint**: `make test` passes for unit tests, CRD types compile

---

## Phase 3: User Story 1 - Expose Service via Ingress (Priority: P1) 🎯 MVP

**Goal**: Create PangolinResource when Ingress with `ingressClassName: pangolin` is created

**Independent Test**: Deploy Ingress → verify PangolinResource created with correct spec

### Tests (Write First, Must Fail)

- [x] T018 [P] [US1] Write integration test: Ingress created → PangolinResource created in tests/integration/reconciler_test.go
- [x] T019 [P] [US1] Write integration test: Ingress with missing tunnel → Warning event emitted in tests/integration/reconciler_test.go
- [x] T020 [P] [US1] Write integration test: ownerReference set correctly for GC in tests/integration/reconciler_test.go

### Implementation

- [x] T021 [US1] Create IngressReconciler struct with Client, Scheme, Config, Logger in internal/controller/ingress_controller.go
- [x] T022 [US1] Implement Reconcile method skeleton (fetch Ingress, check deletion) in internal/controller/ingress_controller.go
- [x] T023 [US1] Implement isManaged() to check ingressClassName matches pangolin or pangolin-* in internal/controller/ingress_controller.go
- [x] T024 [US1] Implement resolveTunnel() to get tunnel name from class/annotation/default in internal/controller/ingress_controller.go
- [x] T025 [US1] Implement validateTunnel() to verify PangolinTunnel exists in internal/controller/ingress_controller.go
- [x] T026 [US1] Implement buildDesiredPangolinResource() with spec, labels, ownerRef in internal/controller/ingress_controller.go
- [x] T027 [US1] Implement createPangolinResource() with event emission in internal/controller/ingress_controller.go
- [x] T028 [US1] Implement SetupWithManager() with For(Ingress).Owns(PangolinResource) in internal/controller/ingress_controller.go
- [x] T029 [US1] Create main.go with manager setup, scheme registration, reconciler init in cmd/manager/main.go
- [ ] T030 [US1] Add Prometheus metrics: reconcile_total, reconcile_duration in internal/controller/ingress_controller.go
- [x] T031 [US1] Add structured logging with logr (namespace, ingress, tunnel fields) in internal/controller/ingress_controller.go

**Checkpoint**: US1 complete - `make test` passes, Ingress creates PangolinResource

---

## Phase 4: User Story 2 - Update Exposed Service (Priority: P2)

**Goal**: Update PangolinResource when Ingress host or backend changes

**Independent Test**: Modify Ingress host → verify PangolinResource spec updated

### Tests (Write First, Must Fail)

- [x] T032 [P] [US2] Write integration test: Ingress host updated → PangolinResource httpConfig updated in tests/integration/lifecycle_test.go
- [x] T033 [P] [US2] Write integration test: Ingress backend updated → PangolinResource target updated in tests/integration/lifecycle_test.go

### Implementation

- [x] T034 [US2] Implement getPangolinResource() to fetch existing resource by labels in internal/controller/ingress_controller.go
- [x] T035 [US2] Implement specChanged() to compare current vs desired spec in internal/controller/ingress_controller.go
- [x] T036 [US2] Implement updatePangolinResource() with event emission in internal/controller/ingress_controller.go
- [x] T037 [US2] Update Reconcile() to handle update path (fetch, compare, update) in internal/controller/ingress_controller.go

**Checkpoint**: US2 complete - Ingress updates propagate to PangolinResource

---

## Phase 5: User Story 3 - Remove Exposed Service (Priority: P2)

**Goal**: Delete PangolinResource when Ingress is deleted or unmanaged

**Independent Test**: Delete Ingress → verify PangolinResource garbage collected

### Tests (Write First, Must Fail)

- [x] T038 [P] [US3] Write integration test: Ingress deleted → PangolinResource deleted via GC in tests/integration/lifecycle_test.go
- [x] T039 [P] [US3] Write integration test: ingressClassName changed → PangolinResource deleted in tests/integration/lifecycle_test.go

### Implementation

- [x] T040 [US3] Implement deletePangolinResource() for explicit deletion case in internal/controller/ingress_controller.go
- [x] T041 [US3] Update Reconcile() to handle not-managed transition (delete existing) in internal/controller/ingress_controller.go
- [x] T042 [US3] Verify ownerReference GC works (no additional code, test only) in tests/integration/lifecycle_test.go

**Checkpoint**: US3 complete - Full CRUD lifecycle works

---

## Phase 6: User Story 4 - Override Domain Configuration (Priority: P3)

**Goal**: Support annotation-based overrides for domain/subdomain

**Independent Test**: Ingress with override annotation → PangolinResource uses custom values

### Tests (Write First, Must Fail)

- [x] T043 [P] [US4] Write unit test: parseAnnotations extracts domain/subdomain overrides in tests/unit/config_test.go
- [x] T044 [P] [US4] Write integration test: annotation override applied to PangolinResource in tests/integration/reconciler_test.go

### Implementation

- [x] T045 [US4] Implement parseAnnotations() for domain-name, subdomain, tunnel-name in internal/controller/ingress_controller.go
- [x] T046 [US4] Update buildDesiredPangolinResource() to use annotation overrides in internal/controller/ingress_controller.go

**Checkpoint**: US4 complete - Annotations override auto-derived values

---

## Phase 7: User Story 5 - Disable Exposure Temporarily (Priority: P3)

**Goal**: Support enabled annotation to temporarily disable exposure

**Independent Test**: Set enabled=false → PangolinResource deleted; enabled=true → recreated

### Tests (Write First, Must Fail)

- [x] T047 [P] [US5] Write integration test: enabled=false → PangolinResource deleted in tests/integration/lifecycle_test.go
- [x] T048 [P] [US5] Write integration test: enabled removed → PangolinResource recreated in tests/integration/lifecycle_test.go

### Implementation

- [x] T049 [US5] Update isManaged() to check enabled annotation in internal/controller/ingress_controller.go
- [x] T050 [US5] Update Reconcile() to delete when enabled=false in internal/controller/ingress_controller.go

**Checkpoint**: US5 complete - Temporary disable/enable works

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Production readiness

- [x] T051 [P] Create Deployment manifest with health probes, resources, env vars in config/manager/deployment.yaml
- [x] T052 [P] Create ServiceAccount manifest in config/rbac/service_account.yaml
- [x] T053 [P] Create IngressClass manifest (pangolin) in config/samples/ingressclass.yaml
- [x] T054 Add /healthz and /readyz endpoints via manager options in cmd/manager/main.go
- [x] T055 Add leader election configuration in cmd/manager/main.go
- [x] T056 [P] Create README.md with installation and usage instructions in README.md
- [x] T057 [P] Create install.yaml combining all manifests via kustomize in deploy/install.yaml
- [ ] T058 Run golangci-lint and fix any issues
- [ ] T059 Verify quickstart.md scenarios work end-to-end

**Checkpoint**: Production-ready, deployable controller

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1: Setup ─────────────────┐
                                │
Phase 2: Foundational ──────────┤ (BLOCKS all user stories)
                                │
         ┌──────────────────────┴──────────────────────┐
         │                                             │
         ▼                                             ▼
Phase 3: US1 (P1) ──┬── Phase 4: US2 (P2) ──┬── Phase 6: US4 (P3)
                    │                       │
                    └── Phase 5: US3 (P2) ──┴── Phase 7: US5 (P3)
                                                       │
                                                       ▼
                                            Phase 8: Polish
```

### User Story Dependencies

- **US1 (P1)**: Foundation only - standalone MVP
- **US2 (P2)**: Depends on US1 (needs create to test update)
- **US3 (P2)**: Depends on US1 (needs create to test delete)
- **US4 (P3)**: Depends on US1 (annotation override for create)
- **US5 (P3)**: Depends on US1, US3 (enabled toggle uses delete logic)

### Parallel Opportunities

**Phase 2 (Foundation)**:
```bash
# Run in parallel:
T007, T008, T009  # All unit tests
T010, T011, T012  # All utility implementations
T015, T016, T017  # All manifests
```

**Phase 3 (US1)**:
```bash
# Run in parallel:
T018, T019, T020  # All integration tests for US1
```

**Phase 8 (Polish)**:
```bash
# Run in parallel:
T051, T052, T053, T056, T057  # All manifests and docs
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: US1 (Expose Service)
4. **STOP and VALIDATE**: `make test` passes, manual verification
5. Deploy/demo - controller creates PangolinResource from Ingress

### Incremental Delivery

| Increment | Stories | Value Delivered |
|-----------|---------|-----------------|
| MVP | US1 | Basic exposure works |
| +CRUD | US1, US2, US3 | Full lifecycle |
| +Annotations | US1-US4 | Annotation overrides |
| +Disable | US1-US5 | Temporary disable |
| Production | All + Polish | Deployable |

---

## Summary

| Metric | Value |
|--------|-------|
| **Total Tasks** | 59 |
| **Setup Tasks** | 6 |
| **Foundational Tasks** | 11 |
| **US1 Tasks** | 14 |
| **US2 Tasks** | 6 |
| **US3 Tasks** | 5 |
| **US4 Tasks** | 4 |
| **US5 Tasks** | 4 |
| **Polish Tasks** | 9 |
| **Parallelizable** | 28 (47%) |

### Files Created/Modified

| File | Tasks |
|------|-------|
| `internal/controller/ingress_controller.go` | 18 |
| `tests/integration/*` | 11 |
| `tests/unit/*` | 4 |
| `internal/util/*` | 2 |
| `config/*` | 6 |
| `cmd/manager/main.go` | 3 |
