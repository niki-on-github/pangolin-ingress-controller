# Tasks: SSO & Multi-Path Support

**Input**: Design documents from `/specs/002-sso-multipath-support/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Not explicitly requested - implementation tasks only.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

**Multi-project structure**:
- **PIC**: `/Users/stefb/src/pangolin-ingress-controller/`
- **Operator**: `/Users/stefb/src/pangolin-operator/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify existing code state and prepare for changes

- [x] T001 Verify SSO/BlockAccess fields exist in PIC types at `internal/pangolincrd/types.go`
- [x] T002 Verify SSO/BlockAccess fields exist in operator types at `api/v1alpha1/pangolinresource_types.go`
- [x] T003 [P] Verify SSO annotations defined in PIC controller at `internal/controller/ingress_controller.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Add `UpdateResource` method to Pangolin client at `pkg/pangolin/client.go` (pangolin-operator)
- [x] T005 Add `ResourceUpdateSpec` type with SSO/BlockAccess fields at `pkg/pangolin/types.go` (pangolin-operator)
- [x] T006 [P] Add path fields (Path, PathMatchType, Priority) to Target struct at `internal/pangolincrd/types.go` (PIC)
- [x] T007 [P] Add path fields to TargetConfig struct at `api/v1alpha1/pangolinresource_types.go` (pangolin-operator)
- [x] T008 [P] Add path fields to TargetCreateSpec at `pkg/pangolin/types.go` (pangolin-operator)
- [x] T009 Regenerate CRDs for operator: run `make generate && make manifests` in pangolin-operator
- [x] T010 Apply updated CRDs to cluster: `kubectl apply -f config/crd/bases/`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Disable SSO Authentication (Priority: P1) 🎯 MVP

**Goal**: Services with `sso: "false"` annotation are publicly accessible without authentication

**Independent Test**: Deploy Ingress with `pangolin.ingress.k8s.io/sso: "false"`, verify service is publicly accessible

### Implementation for User Story 1

- [x] T011 [US1] Add `updateResourceSSO` method to operator reconciler at `internal/controller/pangolinresource_controller.go`
- [x] T012 [US1] Call `updateResourceSSO` after successful resource creation in `reconcilePangolinResource` at `internal/controller/pangolinresource_controller.go`
- [x] T013 [US1] Add SSO status fields to PangolinResourceStatus at `api/v1alpha1/pangolinresource_types.go`
- [x] T014 [US1] Update status with SSO configuration after updateResource call at `internal/controller/pangolinresource_controller.go`
- [x] T015 [US1] Add Kubernetes event for SSO configuration changes at `internal/controller/pangolinresource_controller.go`

**Checkpoint**: User Story 1 complete - SSO disable works independently

---

## Phase 4: User Story 2 - Enable SSO Authentication (Priority: P2)

**Goal**: Services with `sso: "true"` and `block-access: "true"` require authentication

**Independent Test**: Deploy Ingress with SSO enabled, verify authentication is required

### Implementation for User Story 2

- [x] T016 [US2] Handle SSO enable case in `updateResourceSSO` at `internal/controller/pangolinresource_controller.go`
- [x] T017 [US2] Add blockAccess handling to updateResource call at `internal/controller/pangolinresource_controller.go`
- [x] T018 [US2] Add validation for SSO/blockAccess combination (blockAccess requires SSO) at `internal/controller/pangolinresource_controller.go`
- [x] T019 [US2] Add status condition for authentication state at `internal/controller/pangolinresource_controller.go`

**Checkpoint**: User Stories 1 AND 2 complete - SSO enable/disable works independently

---

## Phase 5: User Story 3 - Multi-Path Ingress Support (Priority: P3)

**Goal**: Multiple paths in Ingress create multiple targets with path-based routing

**Independent Test**: Deploy Ingress with `/api` and `/web` paths, verify each routes to correct backend

### Implementation for User Story 3

- [x] T020 [US3] Change PIC Target field from single to array (Targets []Target) at `internal/pangolincrd/types.go`
- [x] T021 [US3] Update `buildDesiredPangolinResource` to iterate paths and create target per path at `internal/controller/ingress_controller.go`
- [x] T022 [US3] Map Ingress pathType to Pangolin pathMatchType (Exact→exact, Prefix→prefix) at `internal/controller/ingress_controller.go`
- [x] T023 [US3] Calculate priority based on path specificity (longer paths = higher priority) at `internal/controller/ingress_controller.go`
- [x] T024 [US3] Update `specChanged` to compare target arrays at `internal/controller/ingress_controller.go`
- [x] T025 [US3] Change operator TargetConfig from single to array at `api/v1alpha1/pangolinresource_types.go`
- [x] T026 [US3] Update operator to create multiple targets with path fields at `internal/controller/pangolinresource_controller.go`
- [x] T027 [US3] Handle target deletion when paths are removed at `internal/controller/pangolinresource_controller.go`
- [x] T028 [US3] Add status field for target count at `api/v1alpha1/pangolinresource_types.go`

**Checkpoint**: All user stories complete - multi-path routing works independently

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T029 [P] Add debug logging for SSO annotation parsing at `internal/controller/ingress_controller.go` (PIC) - already present
- [x] T030 [P] Add debug logging for updateResource API calls at `internal/controller/pangolinresource_controller.go` (operator) - added in updateResourceSSO
- [ ] T031 [P] Add metrics for SSO configuration changes at `internal/controller/pangolinresource_controller.go` (operator) - deferred
- [x] T032 Build and push new PIC image: `docker buildx build --platform linux/amd64 --no-cache -t registry.wizzz.net/stefb/pangolin-ingress-controller:v0.2.0 --push .`
- [x] T033 Build and push new operator image: `docker buildx build --platform linux/amd64 --no-cache -t registry.wizzz.net/stefb/pangolin-operator:v0.2.0 --push .`
- [x] T034 Deploy updated images to cluster and verify reconciliation
- [ ] T035 Run quickstart.md validation scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - verification only
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can proceed sequentially in priority order (P1 → P2 → P3)
  - US2 builds on US1 (same updateResourceSSO method)
  - US3 is independent of US1/US2
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Builds on US1's updateResourceSSO method - Start after US1
- **User Story 3 (P3)**: Independent of US1/US2 - Can start after Foundational (Phase 2)

### Within Each User Story

- Models/types before services
- Services before controller logic
- Core implementation before status updates
- Story complete before moving to next priority

### Parallel Opportunities

- T006, T007, T008 can run in parallel (different files in different projects)
- T029, T030, T031 can run in parallel (different files)
- US3 can be worked on in parallel with US1/US2 (independent functionality)

---

## Parallel Example: Foundational Phase

```bash
# Launch all type updates together:
Task T006: "Add path fields to Target struct in PIC types.go"
Task T007: "Add path fields to TargetConfig in operator types.go"
Task T008: "Add path fields to TargetCreateSpec in operator types.go"
```

## Parallel Example: User Story 3

```bash
# PIC changes can proceed in parallel with operator changes:
# Developer A (PIC):
Task T020: "Change Target to Targets array"
Task T021: "Update buildDesiredPangolinResource"
Task T022: "Map pathType to pathMatchType"

# Developer B (Operator):
Task T025: "Change TargetConfig to array"
Task T026: "Update operator for multi-target"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verification)
2. Complete Phase 2: Foundational (T004-T010)
3. Complete Phase 3: User Story 1 (T011-T015)
4. **STOP and VALIDATE**: Test SSO disable independently
5. Deploy and verify services are publicly accessible

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test SSO disable → Deploy (MVP!)
3. Add User Story 2 → Test SSO enable → Deploy
4. Add User Story 3 → Test multi-path → Deploy
5. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- PIC changes require image rebuild and deployment
- Operator changes require CRD regeneration and image rebuild
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
