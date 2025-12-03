# Research: SSO & Multi-Path Support

**Feature**: 002-sso-multipath-support  
**Date**: 2025-12-03

## Research Tasks

### 1. Pangolin API SSO Fields

**Question**: How does Pangolin API handle SSO configuration?

**Finding**: 
- `createResource` API does NOT accept `sso` or `blockAccess` fields
- `updateResource` API DOES accept these fields (lines 42-43 in updateResource.ts)
- Schema: `sso: z.boolean().optional()`, `blockAccess: z.boolean().optional()`

**Decision**: Call `updateResource` after `createResource` to set SSO fields.

**Rationale**: Cannot modify Pangolin upstream API; two-phase approach is the only option.

**Alternatives Considered**:
- Modify Pangolin createResource API: Rejected (upstream project, requires PR)
- Skip SSO support: Rejected (user requirement)

---

### 2. Pangolin Target Path Configuration

**Question**: How does Pangolin handle path-based routing?

**Finding**:
- Path fields are at TARGET level, not resource level
- Schema in createTarget.ts (lines 56-66):
  ```typescript
  path: z.string().optional().nullable(),
  pathMatchType: z.enum(["exact", "prefix", "regex"]).optional().nullable(),
  rewritePath: z.string().optional().nullable(),
  rewritePathType: z.enum(["exact", "prefix", "regex", "stripPrefix"]).optional().nullable(),
  priority: z.number().int().min(1).max(1000).optional().nullable()
  ```
- Database schema confirms: `path`, `pathMatchType`, `rewritePath`, `rewritePathType`, `priority` columns on targets table

**Decision**: Create one target per Ingress path with appropriate path configuration.

**Rationale**: Aligns with Pangolin's data model; each target can have unique path routing.

**Alternatives Considered**:
- Single target with multiple paths: Not supported by Pangolin
- Resource-level path config: Not available in Pangolin API

---

### 3. Ingress PathType Mapping

**Question**: How to map Kubernetes Ingress `pathType` to Pangolin `pathMatchType`?

**Finding**:
- Kubernetes pathTypes: `Exact`, `Prefix`, `ImplementationSpecific`
- Pangolin pathMatchTypes: `exact`, `prefix`, `regex`

**Decision**: Direct mapping:
| Ingress PathType | Pangolin pathMatchType |
|------------------|----------------------|
| Exact | exact |
| Prefix | prefix |
| ImplementationSpecific | prefix (default) |

**Rationale**: Most intuitive mapping; ImplementationSpecific defaults to prefix as safest option.

---

### 4. Default SSO Behavior

**Question**: What should be the default when annotations are absent?

**Finding**:
- Current behavior: Pangolin defaults SSO to `true` (enabled)
- User expectation: Public services should be accessible by default

**Decision**: Default to `sso: false, blockAccess: false` (open access).

**Rationale**: Aligns with Kubernetes Ingress semantics where resources are publicly accessible unless explicitly restricted. Explicit enablement is safer than implicit blocking.

---

### 5. Operator UpdateResource Implementation

**Question**: Where to add updateResource call in operator flow?

**Finding**:
- Current flow in `reconcilePangolinResource`:
  1. Check if resource exists (by ResourceID)
  2. If not, create via API
  3. Create targets
  4. Update status
- updateResource needs to happen after create, before target creation

**Decision**: Add updateResource call immediately after successful createResource.

**Rationale**: SSO settings should be applied before targets are created to ensure consistent state.

---

## Summary of Decisions

| Topic | Decision | Confidence |
|-------|----------|------------|
| SSO API flow | Create + Update (two-phase) | High |
| Path routing | One target per Ingress path | High |
| PathType mapping | Direct (Exact→exact, Prefix→prefix) | High |
| Default SSO | Disabled (open access) | High |
| Update timing | After create, before targets | High |

## Open Questions

None - all research questions resolved.
