# Routes & Request Flow — Bahasa Daerah Learning Platform

---

## Global Middleware Stack

Every request passes through this chain before reaching any handler, applied in `main.go`:

```
Incoming Request
       │
       ▼
  CORS Handler          — sets Access-Control-* headers, handles OPTIONS preflight
       │
       ▼
  RequestID             — attaches X-Request-ID to each request
       │
       ▼
  RealIP                — reads X-Forwarded-For / X-Real-IP into r.RemoteAddr
       │
       ▼
  Logger                — logs method, path, status, latency
       │
       ▼
  Recoverer             — catches panics, returns 500 instead of crashing
       │
       ▼
  Timeout (30s)         — cancels context after 30 seconds
       │
       ▼
  Route-specific middleware (Authenticate, RequireRole)
       │
       ▼
  Handler function
```

---

## Route Table

### Legend

| Symbol | Meaning |
|---|---|
| 🔓 | No auth required |
| 🔑 | JWT required (any valid role) |
| 👤 | learner, contributor, or admin |
| ✏️ | contributor or admin only |
| 🛡️ | admin only |
| ✅ | Implemented |
| 🔲 | Planned (not yet implemented) |

---

### System

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| `GET` | `/health` | 🔓 | inline | ✅ |

---

### Auth — `/api/v1/auth`

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| `POST` | `/api/v1/auth/register` | 🔓 | `auth.Handler.Register` | ✅ |
| `POST` | `/api/v1/auth/login` | 🔓 | `auth.Handler.Login` | ✅ |
| `POST` | `/api/v1/auth/upgrade-role` | 🔑 👤 | `auth.Handler.UpgradeRole` | ✅ |

---

### Languages — `/api/v1/languages` and `/api/v1/admin/languages`

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| `GET` | `/api/v1/languages` | 🔓 | `language.Handler.ListLanguages` | ✅ |
| `POST` | `/api/v1/admin/languages` | 🛡️ | `language.Handler.CreateLanguage` | ✅ |
| `PATCH` | `/api/v1/admin/languages/:code` | 🛡️ | `language.Handler.ToggleActive` | ✅ |

---

### Phrases — `/api/v1/phrases`

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| `POST` | `/api/v1/phrases` | ✏️ | `phrase.Handler.SubmitPhrase` | ✅ |
| `GET` | `/api/v1/phrases` | ✏️ | `phrase.Handler.ListPendingPhrases` | ✅ |
| `GET` | `/api/v1/phrases/my` | ✏️ | `phrase.Handler.ListMyPhrases` | ✅ |
| `GET` | `/api/v1/phrases/:id` | 🔑 👤 | `phrase.Handler.GetPhraseByID` | ✅ |

> **Routing note:** `/my` is registered before `/:id` in chi to prevent the static segment from being swallowed by the wildcard.

---

### Voting & Flagging — `/api/v1/phrases/:id/...`

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| `POST` | `/api/v1/phrases/:id/votes` | ✏️ | `validation.Handler.VotePhrase` | ✅ |
| `POST` | `/api/v1/phrases/:id/flags` | 🔑 👤 | `validation.Handler.FlagPhrase` | ✅ |
| `POST` | `/api/v1/phrases/:id/audio-votes` | ✏️ | `validation.Handler.VoteAudio` | ✅ |
| `POST` | `/api/v1/phrases/:id/script-votes` | ✏️ | `validation.Handler.VoteScript` | ✅ |

> **Mount note:** Both `phrase.Handler.Routes` and `validation.Handler.Routes` are mounted on the same `/api/v1/phrases` prefix in `main.go`. Each applies its own `Authenticate` + `RequireRole` middleware independently.

---

### Learning Content (Planned)

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| `GET` | `/api/v1/flashcards` | 🔑 👤 | `flashcard.Handler.List` | 🔲 |
| `GET` | `/api/v1/conversation-scenarios` | 🔑 👤 | `flashcard.Handler.Scenarios` | 🔲 |
| `POST` | `/api/v1/phrase-practice` | 🔑 👤 | `flashcard.Handler.RecordPractice` | 🔲 |

---

### Search (Planned)

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| `GET` | `/api/v1/search` | 🔑 👤 | `search.Handler.Search` | 🔲 |

---

### Admin (Planned)

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| `GET` | `/api/v1/admin/phrases/flagged` | 🛡️ | `admin.Handler.ListFlagged` | 🔲 |
| `PATCH` | `/api/v1/admin/phrases/:id/status` | 🛡️ | `admin.Handler.UpdateStatus` | 🔲 |
| `DELETE` | `/api/v1/admin/phrases/:id` | 🛡️ | `admin.Handler.DeletePhrase` | 🔲 |
| `GET` | `/api/v1/admin/users` | 🛡️ | `admin.Handler.ListUsers` | 🔲 |
| `PATCH` | `/api/v1/admin/users/:id/ban` | 🛡️ | `admin.Handler.BanUser` | 🔲 |
| `PATCH` | `/api/v1/admin/users/:id/role` | 🛡️ | `admin.Handler.AssignRole` | 🔲 |

---

## Request Flows

### Flow 1 — Register

```
POST /api/v1/auth/register
        │
        ▼ [global middleware: CORS, Logger, Recoverer, Timeout]
        │
        ▼ auth.Handler.Register
        │   decode JSON body → RegisterRequest
        │   validator.New()
        │     .Check name not empty
        │     .Check email format (regex)
        │     .Check password >= 8 chars
        │     .Check role is learner|contributor
        │   if !v.Valid() → 400 VALIDATION_ERROR
        │
        ▼ auth.Service.Register(ctx, req)
        │   bcrypt.GenerateFromPassword(password, cost=12)
        │   domain.User{ID: uuid.New(), ...}
        │
        ▼ auth.Repository.CreateUser(ctx, user)
        │   INSERT INTO users ... RETURNING created_at, updated_at
        │   if unique_violation (23505) → ErrDuplicateEmail
        │
        ├── ErrDuplicateEmail → 409 DUPLICATE_EMAIL
        │
        ▼ generateJWT(user)
        │   jwt.MapClaims{user_id, role, exp: now+24h}
        │   sign with HS256 + JWT_SECRET
        │
        ▼ 201 { token, user }
```

---

### Flow 2 — Login

```
POST /api/v1/auth/login
        │
        ▼ [global middleware]
        │
        ▼ auth.Handler.Login
        │   decode JSON → LoginRequest
        │   validate email + password not empty
        │   if !v.Valid() → 400 VALIDATION_ERROR
        │
        ▼ auth.Service.Login(ctx, req)
        │
        ▼ auth.Repository.GetUserByEmail(ctx, email)
        │   SELECT ... FROM users WHERE email = $1
        │   if no rows → ErrNotFound → ErrInvalidCredentials
        │
        ▼ bcrypt.CompareHashAndPassword(hash, password)
        │   if mismatch → ErrInvalidCredentials
        │
        ├── ErrInvalidCredentials → 401 UNAUTHORIZED (no detail)
        │
        ▼ generateJWT(user) → 200 { token, user }
```

---

### Flow 3 — Upgrade Role (Learner → Contributor)

```
POST /api/v1/auth/upgrade-role
  Authorization: Bearer <token>
        │
        ▼ [global middleware]
        │
        ▼ middleware.Authenticate(jwtSecret)
        │   parse Bearer token
        │   verify HMAC-SHA256 signature
        │   extract user_id, role from claims
        │   inject into context
        │   if invalid → 401
        │
        ▼ middleware.RequireRole(learner, contributor, admin)
        │   read role from context
        │   if not in allowed set → 403
        │
        ▼ auth.Handler.UpgradeRole
        │   UserIDFromContext(ctx)
        │
        ▼ auth.Service.UpgradeRole(ctx, userID)
        │
        ▼ auth.Repository.GetUserByID(ctx, userID)
        │   if user.Role == contributor|admin → ErrRoleAlreadyUpgraded
        │
        ▼ auth.Repository.UpdateUserRole(ctx, userID, contributor)
        │   UPDATE users SET role = 'contributor' WHERE id = $1
        │
        ├── ErrRoleAlreadyUpgraded → 409 ROLE_ALREADY_UPGRADED
        │
        ▼ generateJWT(user{role: contributor}) → 200 { token, user }
```

---

### Flow 4 — Submit Phrase

```
POST /api/v1/phrases
  Authorization: Bearer <token>
        │
        ▼ [global middleware]
        │
        ▼ middleware.Authenticate(jwtSecret)
        │   validate JWT → inject user_id, role into context
        │   if invalid → 401
        │
        ▼ middleware.RequireRole(contributor, admin)
        │   if role == learner → 403
        │
        ▼ phrase.Handler.SubmitPhrase
        │   UserIDFromContext → contributorID
        │   decode JSON → SubmitPhraseRequest
        │   validator.New()
        │     .Check text_latin not empty          → 400 if missing
        │     .Check text_latin <= 500 chars        → 422 if too long
        │     .Check translation not empty          → 400 if missing
        │     .Check language_code not empty        → 400 if missing
        │     if text_native_script provided:
        │       .Check script_type present          → 422 if missing
        │       .Check script_type in valid enum    → 422 if invalid
        │
        ▼ phrase.Service.SubmitPhrase(ctx, contributorID, req)
        │
        ▼ phrase.Repository.IsLanguageActive(ctx, language_code)
        │   SELECT is_active FROM languages WHERE code = $1
        │   if not active or not found → ErrInactiveLanguage → 400
        │
        │   determine script_status:
        │     text_native_script present → ScriptStatusPending
        │     otherwise                  → ScriptStatusNone
        │
        ▼ phrase.Repository.CreatePhrase(ctx, phrase)
        │   INSERT INTO phrases (...) VALUES (...)
        │   status = 'pending', audio_status = 'none'
        │   all vote counts = 0
        │
        ▼ 201 { id: uuid, status: "pending" }
        │
        (async, future) → AI Pipeline enqueued
```

---

### Flow 5 — Vote on a Phrase

```
POST /api/v1/phrases/:id/votes
  Authorization: Bearer <token>
  { "vote_type": "upvote" }
        │
        ▼ [global middleware]
        │
        ▼ middleware.Authenticate(jwtSecret)
        │   validate JWT → inject user_id, role
        │   if invalid → 401
        │
        ▼ middleware.RequireRole(contributor, admin)
        │   if role == learner → 403
        │
        ▼ validation.Handler.VotePhrase
        │   parsePhraseID → uuid.Parse(chi.URLParam "id")
        │   if invalid UUID → 400 INVALID_ID
        │   UserIDFromContext → contributorID
        │   decode JSON → VoteRequest
        │
        ▼ validation.Service.VotePhrase(ctx, phraseID, contributorID, req)
        │
        │   validateVoteType(req.VoteType)
        │   if not upvote|downvote → 400 INVALID_VOTE_TYPE
        │
        ▼ validation.Repository.GetPhraseContributorID(ctx, phraseID)
        │   SELECT contributor_id FROM phrases WHERE id = $1
        │   if no rows → ErrPhraseNotFound → 404
        │   if ownerID == contributorID → ErrSelfVote → 403
        │
        ▼ validation.Repository.InsertVoteAndUpdateCount(ctx, ...)
        │   BEGIN TRANSACTION
        │     INSERT INTO votes (phrase_id, contributor_id, vote_type)
        │     if unique_violation → ErrDuplicateVote → 409
        │     UPDATE phrases SET upvote_count = upvote_count + 1
        │       RETURNING upvote_count, downvote_count
        │   COMMIT
        │
        ▼ validation.Engine.CheckPhraseThresholds(ctx, phraseID, up, down)
        │   if upvotes >= 3  → UPDATE phrases SET status = 'approved'
        │   if downvotes >= 5 → UPDATE phrases SET status = 'rejected'
        │   otherwise → no-op
        │
        ▼ 201 { "message": "Vote recorded." }
```

---

### Flow 6 — Flag a Phrase

```
POST /api/v1/phrases/:id/flags
  Authorization: Bearer <token>
  { "reason": "inaccurate_translation" }
        │
        ▼ [global middleware]
        │
        ▼ middleware.Authenticate(jwtSecret)
        │   validate JWT → inject user_id, role
        │
        ▼ middleware.RequireRole(learner, contributor, admin)
        │   all authenticated roles allowed
        │
        ▼ validation.Handler.FlagPhrase
        │   parsePhraseID → uuid
        │   UserIDFromContext → userID
        │   decode JSON → FlagRequest
        │
        ▼ validation.Service.FlagPhrase(ctx, phraseID, userID, req)
        │
        │   validateFlagReason(req.Reason)
        │   if not valid enum → 400 INVALID_FLAG_REASON
        │
        ▼ validation.Repository.GetPhraseContributorID(ctx, phraseID)
        │   verify phrase exists (no self-flag check for flags)
        │   if not found → ErrPhraseNotFound → 404
        │
        ▼ validation.Repository.InsertFlagAndUpdateCount(ctx, ...)
        │   BEGIN TRANSACTION
        │     INSERT INTO flags (phrase_id, user_id, reason)
        │     if unique_violation → ErrDuplicateVote → 409 DUPLICATE_FLAG
        │     UPDATE phrases SET flag_count = flag_count + 1
        │       RETURNING flag_count
        │   COMMIT
        │
        ▼ validation.Engine.CheckFlagThreshold(ctx, phraseID, flagCount)
        │   if flag_count >= 3 → UPDATE phrases SET status = 'flagged'
        │
        ▼ 201 { "message": "Flag recorded." }
```

---

### Flow 7 — Audio Vote

```
POST /api/v1/phrases/:id/audio-votes
  Authorization: Bearer <token>
  { "vote_type": "upvote" }
        │
        ▼ [same auth + role middleware as text vote]
        │
        ▼ validation.Handler.VoteAudio
        │   parsePhraseID, UserIDFromContext, decode VoteRequest
        │
        ▼ validation.Service.VoteAudio(ctx, phraseID, contributorID, req)
        │   validateVoteType
        │   GetPhraseContributorID → self-vote check → 403
        │
        ▼ validation.Repository.InsertAudioVoteAndUpdateCount(ctx, ...)
        │   BEGIN TRANSACTION
        │     INSERT INTO audio_votes (...)
        │     if unique_violation → ErrDuplicateVote → 409
        │     UPDATE phrases SET audio_upvote_count = audio_upvote_count + 1
        │       RETURNING audio_upvote_count, audio_downvote_count
        │   COMMIT
        │
        ▼ validation.Engine.CheckAudioThresholds(ctx, phraseID, up, down)
        │   if audio_upvotes >= 3  → UPDATE phrases SET audio_status = 'audio_approved'
        │   if audio_downvotes >= 5 → UPDATE phrases SET audio_status = 'audio_rejected'
        │
        ▼ 201 { "message": "Audio vote recorded." }
```

Script vote (`/script-votes`) follows the identical flow, operating on `script_upvote_count` / `script_downvote_count` and updating `script_status`.

---

### Flow 8 — List Pending Phrases (for voting)

```
GET /api/v1/phrases
  Authorization: Bearer <token>
        │
        ▼ [global middleware]
        │
        ▼ middleware.Authenticate → inject identity
        ▼ middleware.RequireRole(contributor, admin)
        │
        ▼ phrase.Handler.ListPendingPhrases
        │
        ▼ phrase.Service.ListPendingPhrases(ctx)
        │
        ▼ phrase.Repository.ListPendingPhrases(ctx)
        │   SELECT ... FROM phrases WHERE status = 'pending'
        │   ORDER BY created_at ASC
        │
        ▼ 200 { "phrases": [ ...Phrase ] }
```

---

### Flow 9 — Get Phrase Detail

```
GET /api/v1/phrases/:id
  Authorization: Bearer <token>
        │
        ▼ [global middleware]
        │
        ▼ middleware.Authenticate → inject identity
        ▼ middleware.RequireRole(learner, contributor, admin)
        │
        ▼ phrase.Handler.GetPhraseByID
        │   uuid.Parse(chi.URLParam "id")
        │   if invalid → 400 INVALID_ID
        │
        ▼ phrase.Service.GetPhraseByID(ctx, id)
        │
        ▼ phrase.Repository.GetPhraseByID(ctx, id)
        │   SELECT all columns FROM phrases WHERE id = $1
        │   if no rows → ErrNotFound → 404
        │
        ▼ 200 Phrase (full object with all vote counts)
```

---

### Flow 10 — Admin: Toggle Language Active

```
PATCH /api/v1/admin/languages/:code
  Authorization: Bearer <admin-token>
  { "is_active": false }
        │
        ▼ [global middleware]
        │
        ▼ middleware.Authenticate → inject identity
        ▼ middleware.RequireRole(admin)
        │   if role != admin → 403
        │
        ▼ language.Handler.ToggleActive
        │   chi.URLParam "code"
        │   decode JSON → ToggleActiveRequest
        │
        ▼ language.Service.ToggleActive(ctx, code, isActive)
        │
        ▼ language.Repository.SetLanguageActive(ctx, code, false)
        │   UPDATE languages SET is_active = $1 WHERE code = $2
        │   if rows_affected == 0 → ErrNotFound → 404
        │
        ▼ language.Repository.GetLanguageByCode(ctx, code)
        │   SELECT ... FROM languages WHERE code = $1
        │
        ▼ 200 Language (updated)
```

---

## Middleware Application Map

This shows exactly which middleware wraps each route group, as defined in the handler `Routes()` functions.

```
/health                                    — [global only]

/api/v1/auth/register                      — [global only]
/api/v1/auth/login                         — [global only]
/api/v1/auth/upgrade-role                  — [global] → Authenticate → RequireRole(learner|contributor|admin)

/api/v1/languages                          — [global only]
/api/v1/admin/languages                    — [global] → Authenticate → RequireRole(admin)
/api/v1/admin/languages/:code              — [global] → Authenticate → RequireRole(admin)

/api/v1/phrases          GET               — [global] → Authenticate → RequireRole(contributor|admin)
/api/v1/phrases          POST              — [global] → Authenticate → RequireRole(contributor|admin)
/api/v1/phrases/my       GET               — [global] → Authenticate → RequireRole(contributor|admin)
/api/v1/phrases/:id      GET               — [global] → Authenticate → RequireRole(learner|contributor|admin)

/api/v1/phrases/:id/votes         POST     — [global] → Authenticate → RequireRole(contributor|admin)
/api/v1/phrases/:id/audio-votes   POST     — [global] → Authenticate → RequireRole(contributor|admin)
/api/v1/phrases/:id/script-votes  POST     — [global] → Authenticate → RequireRole(contributor|admin)
/api/v1/phrases/:id/flags         POST     — [global] → Authenticate → RequireRole(learner|contributor|admin)
```

---

## Phrase Status State Machine

Phrase status transitions driven by votes, flags, and admin actions:

```
                    ┌─────────┐
                    │ pending │  ◄── initial state on submit
                    └────┬────┘
                         │
          ┌──────────────┼──────────────┐
          │              │              │
          ▼              ▼              ▼
    upvotes >= 3   downvotes >= 5   flags >= 3
          │              │              │
          ▼              ▼              ▼
      approved        rejected        flagged
                                         │
                                    Admin review
                                    /         \
                               approved     rejected
```

Audio and script statuses follow the same pattern independently:

```
audio_status:   none → pending → audio_approved | audio_rejected
script_status:  none → pending → approved       | rejected
```

---

## Dependency Wiring (main.go)

How all components are assembled at startup:

```
pgxpool.Pool (shared)
    │
    ├── auth.NewRepository(pool)       → auth.Repository
    │       └── auth.NewService(repo)  → auth.Service
    │               └── auth.NewHandler(svc) → mounted at /api/v1/auth
    │
    ├── language.NewRepository(pool)   → language.Repository
    │       └── language.NewService(repo) → language.Service
    │               └── language.NewHandler(svc)
    │                       ├── PublicRoutes() → /api/v1/languages
    │                       └── AdminRoutes()  → /api/v1/admin/languages
    │
    ├── phrase.NewRepository(pool)     → phrase.Repository
    │       └── phrase.NewService(repo) → phrase.Service
    │               └── phrase.NewHandler(svc) ──┐
    │                                             │ both mounted at
    ├── validation.NewRepository(pool) → validation.Repository    │ /api/v1/phrases
    │       └── validation.NewService(repo) → validation.Service  │
    │               └── validation.NewHandler(svc) ───────────────┘
    │
    └── (future) ai, flashcard, search, admin repositories/services/handlers
```
