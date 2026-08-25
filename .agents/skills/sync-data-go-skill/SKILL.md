### SKILL: API & Frontend Data Contract Consistency Guard

**Objective:**
Maintain 100% data contract synchronization between the Go backend (Gin/GORM) and the SvelteKit frontend (TypeScript/Svelte 5). Prevent runtime `undefined` errors, casing discrepancies, and contract mismatches.

---

#### 1. Single Source of Truth & Naming Standard
* **Universal JSON standard:** All HTTP API responses and request payloads MUST strictly use `snake_case` (e.g., `created_at`, `first_name`, `phone_number`).
* **Go Structs:** Every struct exported in responses MUST explicitly declare `json:"snake_case_name"` tags. Never rely on default Go PascalCase serialization (`c.JSON(200, model)` without JSON tags is forbidden).
* **TypeScript Types:** All frontend interfaces in `$lib/types/*.ts` MUST mirror the backend JSON keys in exact `snake_case`.

---

#### 2. Pre-Execution Verification Workflow
Before implementing, updating, or touching any frontend API consumer or backend handler:

1. **Verify Backend Response Contract:**
   - Inspect the Go handler and the returned struct/DTO.
   - Confirm all fields have explicit `json:"..."` tags in `snake_case`.
   - Ensure nullable/optional fields in Go (pointers or nullable DB fields) are represented as `T | null` or `T | undefined` in TypeScript.

2. **Verify TypeScript Definitions:**
   - Check `$lib/types/` corresponding interface.
   - Ensure property names match the Go JSON tags 1:1.

3. **Verify Frontend Component Access:**
   - Ensure components access properties exclusively in `snake_case` (`user.first_name`, `user.created_at`).
   - For all optional/nullable fields, enforce safe access and fallbacks (e.g., `u.phone?.toLowerCase() ?? ''`, `u.created_at ? formatDate(u.created_at) : '—'`).

---

#### 3. Prohibited Patterns (Anti-Patterns to Reject)
* ❌ Returning GORM DB models directly from Gin handlers without explicit `json` tags.
* ❌ Mixing `camelCase` and `snake_case` across different API endpoints (e.g., listings using `snake_case` while users use `camelCase` or `PascalCase`).
* ❌ Reading camelCase properties on the frontend when backend outputs snake_case.
* ❌ Duplicating data fetching in both `onMount` and `$effect` in Svelte 5 runes components.

---

#### 4. Automated Check Checklist (Execute before finalizing any task):
- [ ] Are all Go response fields tagged with explicit `json:"<name>"` in `snake_case`?
- [ ] Does the TypeScript interface exactly match the Go JSON output?
- [ ] Are date fields parsed safely in frontend sorting/formatting functions?
- [ ] Does the frontend handle empty/null values gracefully without throwing `TypeError`?