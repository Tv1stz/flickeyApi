# Flickey API: FastAPI to Go Migration Map

Source of truth: inspected from live FastAPI codebase d:\flickeyApi\app\
Target stack: Go 1.26+, Gin, GORM, pgx/PostgreSQL, Redis (go-redis/v9), slog, go-playground/validator/v10, cleanenv, JWT HS256 (golang-jwt/jwt/v5), AWS SDK v2 (S3-compatible)

---

## 1. Endpoint Inventory

### Auth Endpoints — prefix /api/v1/auth

| Method | Path | Auth | Request | Response | Codes | Notes |
|--------|------|------|---------|----------|-------|-------|
| POST | /request-otp | none | OTPRequestIn {phone:E164, return_url?} | OTPRequestOut {challenge_id, expires_in} | 200,400,429 | Atomic Lua rate limits: phone 5/hr, IP 20/hr, resend cooldown 60s |
| POST | /verify-otp | none | OTPVerifyIn {challenge_id, code:6digits} | OTPVerifyOut {access_token, token_type, is_new_user, return_url?} | 200,400,401 | Sets refresh_token HttpOnly cookie Path=/api/v1/auth SameSite=Lax |
| POST | /complete-profile | Bearer | ProfileCompleteRequest {first_name, last_name, email} | UserRead | 200,400,401,409 | Only if user.status==pending_profile |
| POST | /refresh | Cookie | - | RefreshOut {access_token, token_type} | 200,401 | Refresh token rotation with Lua reuse detection. Sets new cookie. |
| POST | /logout | Cookie | - | - | 204,401 | Revokes single refresh token. Clears cookie. |
| POST | /logout-all | Bearer | - | - | 204,401 | Revokes all user refresh tokens. Clears cookie. |
| GET | /me | Bearer | - | UserRead | 200,401 | Auto-upgrades guest to host if listings exist. |
| POST | /dev-login | none | DevLoginIn {phone, first_name, last_name, role} | {access_token, token_type, user} | 200,404 | Dev env only (ENVIRONMENT==dev) |

### Listings/Drafts Endpoints — prefix /api/v1/listings

| Method | Path | Auth | Request | Response | Codes | Notes |
|--------|------|------|---------|----------|-------|-------|
| POST | /drafts | active user | DraftCreateRequest {type: apartment/house/manor} | DraftCreateResponse {draft_id, current_step:2} | 201,400,401 | Step 1 |
| GET | /drafts/{draft_id} | active user | - | DraftDetailResponse | 200,401,404 | Anti-IDOR: scoped by host_id |
| PATCH | /drafts/{draft_id}/step-2 | active user | DraftStep2Request {name,square,floor,total_floors,max_guests,rooms_count,beds_count,bathrooms_count} | DraftStepResponse {draft_id,current_step:3,status} | 200,400,401,409 | Sanitizes name. Invalidates downstream. |
| PATCH | /drafts/{draft_id}/step-3 | active user | DraftStep3Request {media_ids: UUID[5-15]} | DraftStepResponse {current_step:4} | 200,400,401 | Validates media ownership + uploaded status |
| PATCH | /drafts/{draft_id}/step-4 | active user | DraftStep4Request {amenities: string[]} | DraftStepResponse {current_step:5} | 200,400,401 | Validates registry + housing type compat |
| PATCH | /drafts/{draft_id}/step-5 | active user | DraftStep5Request {price_per_night,currency:BYN,min_nights,checkin_from,checkout_until,rules{...}} | DraftStepResponse {current_step:6} | 200,400,401 | Currency must be BYN only |
| PATCH | /drafts/{draft_id}/step-6 | active user | DraftStep6Request {description: string[30-5000]} | DraftStepResponse {current_step:6} | 200,400,401 | HTML sanitize. Step stays at 6. |
| POST | /drafts/{draft_id}/submit | active user | - | ListingSubmitResponse {listing_id,status} | 201,400,401,409 | Requires Idempotency-Key header (UUIDv4). Redis lock. Atomic DB txn. |
| GET | /my | active user | - | ListingHostReadSchema[] | 200,401 | All statuses, Anti-IDOR host-scoped |
| GET | (empty) | none | - | ListingPublicSchema[] | 200 | Only status=published |
| GET | /{listing_id} | none | - | ListingPublicSchema | 200,404 | Only status=published |

### Media Endpoints — prefix /api/v1/media

| Method | Path | Auth | Request | Response | Codes | Notes |
|--------|------|------|---------|----------|-------|-------|
| POST | /presign | active user | MediaPresignRequest {content_type, file_size_bytes} | MediaPresignResponse {media_id,upload_url,file_key,expires_in:300} | 200,400,401 | Creates pending record. Generates S3 presigned PUT URL. |
| POST | /{media_id}/complete | active user | - | MediaCompleteResponse {media_id,status,file_key} | 200,400,401 | HEAD + range GET for magic bytes verification |
| PUT | /dev-upload/{file_key} | none | raw bytes | - | 200 | Dev only, local file fallback |

### Amenities Endpoints — prefix /api/v1/amenities

| Method | Path | Auth | Query | Response | Codes | Notes |
|--------|------|------|-------|----------|-------|-------|
| GET | (empty) | none | housing_type?: apartment/house/manor | AmenitiesResponse {housing_type?,categories:[{id,name,amenities:[{id,name}]}]} | 200 | Static in-memory registry, Russian names |

---

## 2. FastAPI Module to Go Package Mapping

```
FastAPI Module                     Go Package
app/core/config.py              -> config/
app/core/exceptions.py          -> inline error types per package
app/core/security.py            -> auth/ (hash, verify, mask_phone, generate_token)
app/core/amenities.py           -> amenities/
app/core/logging.py             -> main.go (slog setup)
app/core/error_handlers.py      -> web middleware in main.go

app/database/session.py         -> db/ (GORM + pgx pool)
app/database/redis.py           -> db/ (go-redis/v9 client)

app/auth/otp.py                 -> auth/ (OTP gen, hash, Redis store, Lua verify)
app/auth/jwt.py                 -> auth/ (JWT create/decode, refresh token lifecycle)
app/auth/rate_limiter.py        -> auth/ (Lua rate limit scripts)
app/auth/constants.py           -> auth/ (Redis key prefixes as constants)
app/auth/service.py             -> auth/ (AuthService struct)
app/auth/schemas.py             -> auth/ (request/response structs)
app/auth/dependencies.py        -> auth/ (Gin middleware: RequireAuth, RequireActiveUser)
app/auth/router.py              -> auth/ (Gin handler funcs)

app/users/models.py             -> db/ (User GORM model)
app/users/enums.py              -> db/ (UserRole, UserStatus string types)
app/users/schemas.py            -> auth/ (UserRead, ProfileCompleteRequest)
app/users/service.py            -> auth/ (UserService: get_or_create_by_phone, complete_profile)
app/users/repository.py         -> db/ (UserStore queries)

app/models/listing.py           -> db/ (Listing GORM model)
app/models/listing_draft.py     -> db/ (ListingDraft GORM model)
app/models/listing_amenity.py   -> db/ (ListingAmenity GORM model)
app/models/media.py             -> db/ (Media GORM model)

app/repositories/listing.py     -> db/ (ListingStore queries)
app/repositories/listing_draft.py -> db/ (DraftStore queries)
app/repositories/media.py       -> db/ (MediaStore queries)
app/repositories/verification.py -> db/ (VerificationStore)

app/schemas/listing.py          -> listings/ (request/response structs)
app/schemas/media.py            -> media/ (request/response structs)
app/schemas/amenities.py        -> amenities/ (response structs)

app/services/listing.py         -> listings/ (ListingService: draft state machine + submit)
app/services/media.py           -> media/ (MediaService: presign, complete, magic bytes)

app/infrastructure/storage/base.py -> storage/ (StorageProvider interface)
app/infrastructure/storage/s3.py   -> storage/ (S3Storage impl via AWS SDK v2)
app/infrastructure/storage/local.py -> storage/ (LocalStorage for dev)

app/sms/base.py                 -> sms/ (SMSSender interface)
app/sms/console.py              -> sms/ (ConsoleSMSSender)

app/main.py                     -> main.go (Gin app factory, server startup)
```

---

## 3. Target Go Project Structure

```
go-backend/
  main.go               # entry point: config, DB, Redis, Gin router, graceful shutdown
  go.mod
  go.sum
  Makefile
  Dockerfile
  .env.example
  
  config/
    config.go           # Settings struct loaded from env via cleanenv
  
  db/
    db.go               # GORM DB init, pgx connection pool
    redis.go            # go-redis/v9 client init
    models.go           # User, Listing, ListingDraft, ListingAmenity, Media GORM structs
    queries.go          # All DB query funcs (UserStore, ListingStore, DraftStore, MediaStore)
  
  auth/
    handlers.go         # Gin handlers for /auth/* routes
    service.go          # AuthService (request_otp, verify_otp, refresh, logout)
    jwt.go              # JWT create/decode + refresh token lifecycle (Lua rotation)
    otp.go              # OTP gen, salted hash, Redis store, atomic Lua verify
    ratelimit.go        # Atomic Lua rate limit scripts
    middleware.go       # RequireAuth, RequireActiveUser Gin middleware
    types.go            # Request/Response structs + UserRead, ProfileCompleteRequest
    constants.go        # Redis key prefixes
  
  listings/
    handlers.go         # Gin handlers for /listings/* routes
    service.go          # ListingService (6-step draft machine + submit + get_my + public)
    types.go            # Request/Response structs
  
  media/
    handlers.go         # Gin handlers for /media/* routes
    service.go          # MediaService (presign, complete, magic bytes)
    types.go            # Request/Response structs
  
  amenities/
    registry.go         # AmenityDefinition map, HousingType enum, grouped amenities func
    handlers.go         # GET /amenities handler
  
  storage/
    storage.go          # StorageProvider interface
    s3.go               # S3Storage (AWS SDK v2, Selectel-compatible)
    local.go            # LocalStorage for dev
  
  sms/
    sms.go              # SMSSender interface
    console.go          # ConsoleSMSSender (prints to stdout)
```

---

## 4. Database Schema (exact from Alembic migration)

### users
- id: UUID PK DEFAULT gen_random_uuid()
- phone: VARCHAR(20) NOT NULL UNIQUE INDEX
- first_name: VARCHAR(100) NULL
- last_name: VARCHAR(100) NULL
- email: VARCHAR(255) NULL PARTIAL UNIQUE INDEX (email IS NOT NULL)
- role: ENUM(guest,host,admin) NOT NULL DEFAULT guest
- status: ENUM(pending_profile,active,suspended,banned) NOT NULL DEFAULT pending_profile
- created_at: TIMESTAMPTZ NOT NULL DEFAULT now()
- updated_at: TIMESTAMPTZ NOT NULL DEFAULT now()

### listing_drafts
- id: UUID PK
- host_id: UUID NOT NULL FK->users(id) CASCADE INDEX
- current_step: INTEGER NOT NULL CHECK(1<=step<=6)
- status: VARCHAR(50) NOT NULL DEFAULT draft INDEX
- type: VARCHAR(50) NULL
- name: VARCHAR(100) NULL
- square: NUMERIC(10,2) NULL
- floor: INTEGER NULL
- total_floors: INTEGER NULL
- max_guests: INTEGER NULL
- rooms_count: INTEGER NULL
- beds_count: INTEGER NULL
- bathrooms_count: INTEGER NULL
- media_ids: JSON NULL (array of UUID strings)
- amenities: JSON NULL (array of amenity ID strings)
- price_per_night: NUMERIC(10,2) NULL
- currency: VARCHAR(10) NULL
- min_nights: INTEGER NULL
- checkin_from: TIME NULL
- checkout_until: TIME NULL
- allow_children/pets/smoking/parties/deposit_required/with_invoicing: BOOLEAN NULL
- description: TEXT NULL
- created_at/updated_at: TIMESTAMPTZ NOT NULL DEFAULT now()

### listings
- id: UUID PK
- host_id: UUID NOT NULL FK->users(id) CASCADE INDEX
- status: VARCHAR(50) NOT NULL (pending_review/awaiting_company_verification/published) INDEX
- type: VARCHAR(50) NOT NULL
- name: VARCHAR(100) NOT NULL
- square: NUMERIC(10,2) NOT NULL
- floor, total_floors, max_guests, rooms_count, beds_count, bathrooms_count: INTEGER NOT NULL
- price_per_night: NUMERIC(10,2) NOT NULL
- currency: VARCHAR(10) NOT NULL
- min_nights: INTEGER NOT NULL
- checkin_from, checkout_until: TIME NOT NULL
- allow_children/pets/smoking/parties/deposit_required/with_invoicing: BOOLEAN NOT NULL
- description: TEXT NOT NULL
- created_at/updated_at: TIMESTAMPTZ NOT NULL DEFAULT now()

### listing_amenities
- id: UUID PK
- listing_id: UUID NOT NULL FK->listings(id) CASCADE INDEX
- amenity_id: VARCHAR(100) NOT NULL INDEX
- created_at: TIMESTAMPTZ NOT NULL
- UNIQUE INDEX(listing_id, amenity_id)

### media
- id: UUID PK
- host_id: UUID NOT NULL FK->users(id) CASCADE INDEX
- listing_id: UUID NULL FK->listings(id) SET NULL INDEX
- file_key: VARCHAR(255) NOT NULL UNIQUE INDEX
- status: VARCHAR(50) NOT NULL (pending/uploaded/attached) INDEX
- content_type: VARCHAR(100) NOT NULL
- file_size_bytes: BIGINT NOT NULL
- created_at/updated_at: TIMESTAMPTZ NOT NULL DEFAULT now()
- COMPOSITE INDEX(host_id, status)

---

## 5. Redis Key Schema

| Key Pattern | Type | TTL | Purpose |
|-------------|------|-----|---------|
| otp:challenge:{challenge_id} | HASH {hash,phone,attempts,return_url?} | OTP_TTL_SECONDS | OTP challenge |
| otp:phone_challenge:{phone} | STRING challenge_id | OTP_TTL_SECONDS | Phone->challenge reverse map |
| otp:phone_rate:{sha256(phone)[:16]} | STRING count | 3600s | Per-phone rate limit |
| otp:ip_rate:{ip} | STRING count | 3600s | Per-IP rate limit |
| otp:resend_cooldown:{sha256(phone)[:16]} | STRING "1" | OTP_RESEND_COOLDOWN_SECONDS | Resend cooldown NX |
| refresh:{sha256(raw_token)} | HASH {user_id,session_id,family_id,rotated,created_at} | JWT_REFRESH_TTL_SECONDS | Refresh token |
| refresh:user:{user_id} | SET of token hashes | JWT_REFRESH_TTL_SECONDS | All tokens per user |
| idempotency:listing-submit:{host_id}:{key} | STRING PROCESSING/COMPLETED:{json} | 60s/86400s | Submit idempotency |

---

## 6. Critical Business Logic

### OTP
1. Phone -> phonenumbers E.164 normalization (nyaruka/phonenumbers in Go)
2. OTP = crypto/rand zero-padded 6 digits
3. Stored as salted SHA-256: "salt:sha256(salt+':'+otp)"
4. Atomic Lua: check exists -> check attempts -> increment -> return hash+phone
5. Verify with crypto/subtle.ConstantTimeCompare
6. Invalidate previous challenge via phone->challenge mapping

### JWT
- HS256 only (algorithm restricted in decode to prevent confusion)
- Claims: sub (user UUID string), iat, exp, jti (uuid4), type:"access"
- JWT_SECRET must be >= 32 bytes

### Refresh Token Rotation
- Raw token = crypto/rand 32 bytes urlsafe (256-bit)
- Stored key = sha256(raw_token) hex, unsalted (O(1) lookup)
- Lua script atomically: check rotated -> mark rotated:"1" -> return data
- Reuse detected (rotated=="1") -> revoke entire token family
- Family = scan SMEMBERS(refresh:user:{user_id}), match family_id

### Rate Limiting
- Atomic Lua: GET key -> if count >= limit return TTL else INCR + EXPIRE on first
- Resend cooldown: SET NX EX -> if nil = already set = in cooldown

### Draft State Machine
- Steps must be done in order 1->6
- Step N redoable only if current_step >= N
- Modifying step N < current: null all downstream fields
- Ready to submit: current_step==6, status=="draft"

### Atomic Submit Transaction
1. Redis SETNX idempotency lock (60s TTL)
2. SELECT draft FOR UPDATE -> check owned, not submitted, step==6
3. Full field re-validation
4. Check business verification -> listing_status
5. INSERT listing
6. INSERT listing_amenities (batch)
7. UPDATE media (bulk) -> verify rowcount == len(media_ids)
8. UPDATE draft SET status=submitted
9. SELECT user FOR UPDATE -> if role==guest: UPDATE role=host
10. COMMIT
11. Update Redis key to COMPLETED:{json} (86400s TTL)
On any failure: rollback TX, delete Redis key if PROCESSING

### Media
- Magic bytes JPEG: FF D8 FF
- Magic bytes PNG: 89 50 4E 47 0D 0A 1A 0A
- Magic bytes WebP: bytes[0:4]=="RIFF" AND bytes[8:12]=="WEBP"
- File key pattern: raw/{host_id}/{media_id}{.ext}
- Max size: 15728640 bytes (15MB)

### Cookie
- Name: refresh_token, HttpOnly: true
- Secure: COOKIE_SECURE env (false in dev)
- SameSite: Lax, Path: /api/v1/auth
- MaxAge: JWT_REFRESH_TTL_SECONDS, Domain: COOKIE_DOMAIN env

### Error Response JSON
{ "code": "MACHINE_CODE", "message": "Human message", "details": {} }
HTTP status codes:
- 400 ValidationError, bad request
- 401 AuthenticationError (WWW-Authenticate: Bearer header)
- 403 AuthorizationError
- 404 NotFoundError
- 409 ConflictError
- 429 RateLimitError (Retry-After header)
- 500 InternalError

---

## 7. Go Dependencies

| Module | Purpose |
|--------|---------|
| github.com/gin-gonic/gin | HTTP router |
| gorm.io/gorm + gorm.io/driver/postgres | ORM |
| github.com/redis/go-redis/v9 | Redis client |
| github.com/golang-jwt/jwt/v5 | JWT HS256 |
| github.com/ilyakaznacheev/cleanenv | Config loading |
| github.com/nyaruka/phonenumbers | E.164 phone validation |
| github.com/aws/aws-sdk-go-v2 | S3-compatible storage |
| github.com/microcosm-cc/bluemonday | HTML sanitization |
| github.com/go-playground/validator/v10 | Struct validation |
| github.com/google/uuid | UUID gen |
| log/slog (stdlib) | Structured logging |

---

## 8. Environment Variables

ENVIRONMENT=dev
DATABASE_URL=postgresql://rental:rental_dev@localhost:5432/flickey
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=<required, min 32 bytes>
JWT_ALGORITHM=HS256
JWT_ACCESS_TTL_SECONDS=900
JWT_REFRESH_TTL_SECONDS=2592000
OTP_TTL_SECONDS=300
OTP_LENGTH=6
OTP_MAX_VERIFY_ATTEMPTS=5
OTP_MAX_REQUESTS_PER_PHONE_PER_HOUR=5
OTP_MAX_REQUESTS_PER_IP_PER_HOUR=20
OTP_RESEND_COOLDOWN_SECONDS=60
ALLOWED_REDIRECT_HOSTS=
COOKIE_DOMAIN=
COOKIE_SECURE=true
TRUST_PROXY_HEADERS=false
S3_ENDPOINT_URL=s3.gis-1.storage.selcloud.ru
S3_ACCESS_KEY=
S3_SECRET_KEY=
S3_BUCKET=flickey
S3_REGION=gis-1
S3_PUBLIC_BASE_URL=
S3_PRESIGNED_URL_TTL_SECONDS=300
MEDIA_MAX_FILE_SIZE_BYTES=15728640
MEDIA_MIN_COUNT=5
MEDIA_MAX_COUNT=15

---

## 9. CORS (from main.py)

Allowed origins:
- http://127.0.0.1:5500
- http://localhost:5500
- http://127.0.0.1:8000
- http://localhost:8000
- http://localhost:3000
- http://127.0.0.1:3000

Allow credentials: true
Allow all methods and headers
