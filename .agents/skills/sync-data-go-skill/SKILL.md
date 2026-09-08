---
name: sync-data-go
description: >
  Ensures 100% data contract synchronization and schema consistency between the Go backend (Gin/GORM)
  and SvelteKit frontend (TypeScript/Svelte 5 Runes). Covers snake_case contract enforcement,
  frontend-to-backend DTO adapters, multi-step draft wizard pipelines, presigned media uploads,
  and zero-backend-modification client synchronization.
---

# API & Frontend Data Contract Consistency Guard

## Objective
Maintain 100% data contract synchronization between the Go backend (Gin/GORM) and the SvelteKit frontend (TypeScript/Svelte 5). Prevent runtime `undefined` errors, casing discrepancies, and contract mismatches without requiring breaking changes on the backend.

---

## 1. Single Source of Truth & Naming Standard
* **Universal JSON standard:** All HTTP API responses and request payloads MUST strictly use `snake_case` (e.g., `created_at`, `first_name`, `phone_number`, `price_per_night`).
* **Go Structs:** Every struct exported in responses MUST explicitly declare `json:"snake_case_name"` tags. Never rely on default Go PascalCase serialization (`c.JSON(200, model)` without JSON tags is forbidden).
* **TypeScript Types:** All frontend interfaces in `$lib/types/*.ts` MUST mirror the backend JSON keys in exact `snake_case`.

---

## 2. Frontend-Backend Schema Adapters Pattern
When frontend UI components or third-party design systems require `camelCase` models, DO NOT modify the Go backend. Instead, create and use pure converter functions:
* **`apiListingToCardListing(api: ListingPublic): Listing`** — converts backend DTO to UI model.
* **`formToCreateRequest(form: FormValues): CreateListingRequest`** — maps form state to backend DTO.
* **`backendUserToStoreUser(user: BackendUser): StoreUser`** — synchronizes backend user sessions with reactive stores.

---

## 3. Multi-Step Draft & Upload Pipeline Standard
When implementing multi-step forms / wizards (e.g. Host listing creation):
1. **Draft Initialization:** `POST /api/v1/listings/drafts` with `{ type }` -> returns `draft_id`.
2. **Stepwise Updates:** Use `PATCH /api/v1/listings/drafts/{id}/step-{N}` with validated typed payloads.
3. **Media Upload Flow:**
   - Request presign: `POST /api/v1/media/presign` -> `{ media_id, upload_url, file_key }`.
   - Direct PUT to `upload_url`.
   - Finalize inspection: `POST /api/v1/media/{id}/complete`.
   - Attach IDs to draft: `PATCH /api/v1/listings/drafts/{id}/step-3` with `{ media_ids }`.
4. **Idempotent Publish:** Submit via `POST /api/v1/listings/drafts/{id}/submit` with `Idempotency-Key: <UUIDv4>`.

---

## 4. Prohibited Patterns (Anti-Patterns to Reject)
* ❌ Returning GORM DB models directly from Gin handlers without explicit `json` tags.
* ❌ Mixing `camelCase` and `snake_case` across different API endpoints.
* ❌ Reading camelCase properties on the frontend when backend outputs snake_case without an adapter.
* ❌ Duplicating data fetching in both `onMount` and `$effect` in Svelte 5 runes components.

---

## 5. Pre-Execution & Verification Checklist
- [ ] Are all Go response fields tagged with explicit `json:"<name>"` in `snake_case`?
- [ ] Does the TypeScript interface in `$lib/types/` exactly match the Go JSON output?
- [ ] Are converters used when binding API data to design-specific UI components?
- [ ] Are optional/nullable fields accessed with safe fallback operators (`??`, `?.`)?
- [ ] Does `npx svelte-check` pass with 0 errors and `npm run build` succeed?