# API Contract — Bahasa Daerah Learning Platform

**Base URL:** `https://<host>/api/v1`  
**Protocol:** HTTPS only  
**Content-Type:** `application/json` (all requests and responses)  
**Auth:** `Authorization: Bearer <JWT>` on all protected endpoints

---

## Common Conventions

### Authentication

JWT is issued on register/login/upgrade-role. It expires after **24 hours**.

JWT payload:
```json
{ "user_id": "<uuid>", "role": "learner|contributor|admin", "exp": <unix_ts> }
```

Role hierarchy for access control:

| Role | Can access |
|---|---|
| `learner` | Read-only learning content, flags |
| `contributor` | All learner access + submit phrases, vote |
| `admin` | All contributor access + admin endpoints |

### Error Response Envelope

All errors use this shape:

```json
{
  "error": {
    "code": "SNAKE_CASE_CODE",
    "message": "Human-readable description."
  }
}
```

### HTTP Status Codes

| Code | Meaning |
|---|---|
| `200` | OK |
| `201` | Created |
| `400` | Bad request — missing required field or malformed JSON |
| `401` | Unauthenticated — missing, expired, or invalid JWT |
| `403` | Forbidden — insufficient role or self-vote attempt |
| `404` | Resource not found |
| `409` | Conflict — duplicate email, duplicate vote |
| `422` | Unprocessable entity — constraint violation (length, format, enum) |
| `500` | Internal server error |

### Shared Object Schemas

#### `User`
```json
{
  "id": "uuid",
  "name": "string",
  "email": "string",
  "role": "learner | contributor | admin",
  "is_active": true,
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```
> `password_hash` is never returned.

#### `Language`
```json
{
  "code": "string (≤20 chars, e.g. 'jv', 'su', 'ban')",
  "name": "string",
  "region": "string | omitted",
  "is_active": true,
  "created_at": "RFC3339"
}
```

#### `Phrase`
```json
{
  "id": "uuid",
  "text_latin": "string",
  "text_native_script": "string | omitted",
  "script_type": "latin | javanese | sundanese | balinese | lontara | batak | other | omitted",
  "translation": "string",
  "language_code": "string",
  "tone": "formal | netral | kasar | omitted",
  "status": "pending | approved | rejected | flagged | ai_failed",
  "script_status": "none | pending | approved | rejected",
  "audio_status": "none | pending | audio_approved | audio_rejected",
  "contributor_id": "uuid",
  "cultural_context_id": "uuid | omitted",
  "audio_url": "string | omitted",
  "audio_duration_seconds": "float | omitted",
  "native_script_image_url": "string | omitted",
  "moderated_by": "uuid | omitted",
  "moderated_at": "RFC3339 | omitted",
  "upvote_count": 0,
  "downvote_count": 0,
  "flag_count": 0,
  "audio_upvote_count": 0,
  "audio_downvote_count": 0,
  "script_upvote_count": 0,
  "script_downvote_count": 0,
  "created_at": "RFC3339",
  "updated_at": "RFC3339",
  "words": [ "<Word>" ],
  "cultural_context": "<CulturalContext> | omitted"
}
```

#### `Word`
```json
{
  "id": "uuid",
  "surface_form_latin": "string",
  "surface_form_native_script": "string | omitted",
  "root_form_latin": "string",
  "root_form_native_script": "string | omitted",
  "script_type": "latin | javanese | ... | omitted",
  "part_of_speech": "string | omitted",
  "language_code": "string",
  "position": 0,
  "created_at": "RFC3339"
}
```

#### `CulturalContext`
```json
{
  "id": "uuid",
  "language_code": "string",
  "region": "string | omitted",
  "usage_situation": "string | omitted",
  "created_at": "RFC3339"
}
```

---

## Endpoints

---

### Auth

#### `POST /api/v1/auth/register`

Register a new user. Role must be `learner` or `contributor` — `admin` can only be assigned by an existing admin.

**Auth:** None

**Request:**
```json
{
  "name": "string (required)",
  "email": "string (required, valid email)",
  "password": "string (required, min 8 chars)",
  "role": "learner | contributor (required)"
}
```

**Response `201`:**
```json
{
  "token": "JWT string",
  "user": { "<User>" }
}
```

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `VALIDATION_ERROR` | Missing or invalid field |
| `400` | `INVALID_JSON` | Malformed request body |
| `409` | `DUPLICATE_EMAIL` | Email already registered |

---

#### `POST /api/v1/auth/login`

Authenticate and receive a 24-hour JWT.

**Auth:** None

**Request:**
```json
{
  "email": "string (required)",
  "password": "string (required)"
}
```

**Response `200`:**
```json
{
  "token": "JWT string",
  "user": { "<User>" }
}
```

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `VALIDATION_ERROR` | Missing field |
| `400` | `INVALID_JSON` | Malformed request body |
| `401` | `UNAUTHORIZED` | Invalid credentials (no detail exposed) |

---

#### `POST /api/v1/auth/upgrade-role`

Self-upgrade from `learner` to `contributor`. Returns a new JWT with the updated role.

**Auth:** Required (any role)

**Request:** _(empty body)_

**Response `200`:**
```json
{
  "token": "JWT string",
  "user": { "<User>" }
}
```

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `401` | `UNAUTHORIZED` | Missing or invalid JWT |
| `409` | `ROLE_ALREADY_UPGRADED` | User is already `contributor` or `admin` |

---

### Languages

#### `GET /api/v1/languages`

List all languages (active and inactive).

**Auth:** None

**Response `200`:**
```json
{
  "languages": [ "<Language>" ]
}
```

---

#### `POST /api/v1/admin/languages`

Create a new language entry.

**Auth:** Required — `admin`

**Request:**
```json
{
  "code": "string (required, ≤20 chars)",
  "name": "string (required)",
  "region": "string (optional)"
}
```

**Response `201`:** `<Language>`

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `VALIDATION_ERROR` | Missing `code` or `name` |
| `400` | `INVALID_JSON` | Malformed request body |
| `401` | `UNAUTHORIZED` | Missing or invalid JWT |
| `403` | `FORBIDDEN` | Not an admin |
| `409` | `DUPLICATE_CODE` | Language code already exists |

---

#### `PATCH /api/v1/admin/languages/:code`

Toggle a language active or inactive.

**Auth:** Required — `admin`

**Path param:** `code` — language code (e.g. `jv`)

**Request:**
```json
{
  "is_active": true
}
```

**Response `200`:** `<Language>`

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `INVALID_JSON` | Malformed request body |
| `401` | `UNAUTHORIZED` | Missing or invalid JWT |
| `403` | `FORBIDDEN` | Not an admin |
| `404` | `NOT_FOUND` | Language code not found |

---

### Phrases

#### `POST /api/v1/phrases`

Submit a new phrase. Saved with `status = pending`.

**Auth:** Required — `contributor`, `admin`

**Request:**
```json
{
  "text_latin": "string (required, ≤500 chars)",
  "translation": "string (required)",
  "language_code": "string (required, must be active)",
  "text_native_script": "string (optional)",
  "script_type": "latin | javanese | sundanese | balinese | lontara | batak | other (required if text_native_script provided)",
  "cultural_context_id": "uuid (optional)"
}
```

**Response `201`:**
```json
{
  "id": "uuid",
  "status": "pending"
}
```

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `VALIDATION_ERROR` | Missing `text_latin`, `translation`, or `language_code` |
| `400` | `INACTIVE_LANGUAGE` | `language_code` not active or not found |
| `400` | `INVALID_JSON` | Malformed request body |
| `401` | `UNAUTHORIZED` | Missing or invalid JWT |
| `403` | `FORBIDDEN` | Not a contributor or admin |
| `422` | `VALIDATION_ERROR` | `text_latin` > 500 chars, or invalid `script_type` enum |

---

#### `GET /api/v1/phrases`

List all `pending` phrases available for voting.

**Auth:** Required — `contributor`, `admin`

**Response `200`:**
```json
{
  "phrases": [ "<Phrase>" ]
}
```

Ordered by `created_at ASC`.

---

#### `GET /api/v1/phrases/my`

List all phrases submitted by the authenticated contributor, newest first.

**Auth:** Required — `contributor`, `admin`

**Response `200`:**
```json
{
  "phrases": [ "<Phrase>" ]
}
```

---

#### `GET /api/v1/phrases/:id`

Get full detail of a single phrase including vote counts.

**Auth:** Required — any role

**Path param:** `id` — phrase UUID

**Response `200`:** `<Phrase>`

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `INVALID_ID` | `id` is not a valid UUID |
| `401` | `UNAUTHORIZED` | Missing or invalid JWT |
| `404` | `NOT_FOUND` | Phrase not found |

---

### Voting & Flagging

All vote/flag endpoints share the same path prefix `/api/v1/phrases/:id/`.

#### `POST /api/v1/phrases/:id/votes`

Cast a text vote (upvote or downvote) on a phrase.

**Auth:** Required — `contributor`, `admin`

**Path param:** `id` — phrase UUID

**Request:**
```json
{
  "vote_type": "upvote | downvote"
}
```

**Response `201`:**
```json
{ "message": "Vote recorded." }
```

**Side effects:**
- Atomically increments `upvote_count` or `downvote_count` on the phrase.
- Triggers threshold check: `upvote_count ≥ 3` → `status = approved`; `downvote_count ≥ 5` → `status = rejected`.

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `INVALID_JSON` | Malformed request body |
| `400` | `INVALID_VOTE_TYPE` | `vote_type` not `upvote` or `downvote` |
| `401` | `UNAUTHORIZED` | Missing or invalid JWT |
| `403` | `FORBIDDEN` | Voting on own phrase |
| `404` | `NOT_FOUND` | Phrase not found |
| `409` | `DUPLICATE_VOTE` | Already voted on this phrase |

---

#### `POST /api/v1/phrases/:id/flags`

Flag a phrase for review.

**Auth:** Required — any role (`learner`, `contributor`, `admin`)

**Path param:** `id` — phrase UUID

**Request:**
```json
{
  "reason": "inaccurate_translation | inappropriate_content | duplicate"
}
```

**Response `201`:**
```json
{ "message": "Flag recorded." }
```

**Side effects:**
- Atomically increments `flag_count` on the phrase.
- Triggers threshold check: `flag_count ≥ 3` → `status = flagged`.

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `INVALID_JSON` | Malformed request body |
| `400` | `INVALID_FLAG_REASON` | `reason` not a valid enum value |
| `401` | `UNAUTHORIZED` | Missing or invalid JWT |
| `404` | `NOT_FOUND` | Phrase not found |
| `409` | `DUPLICATE_FLAG` | Already flagged this phrase |

---

#### `POST /api/v1/phrases/:id/audio-votes`

Vote on the accuracy of a phrase's audio recording.

**Auth:** Required — `contributor`, `admin`

**Path param:** `id` — phrase UUID

**Request:**
```json
{
  "vote_type": "upvote | downvote"
}
```

**Response `201`:**
```json
{ "message": "Audio vote recorded." }
```

**Side effects:**
- Atomically increments `audio_upvote_count` or `audio_downvote_count`.
- Triggers threshold check: `audio_upvote_count ≥ 3` → `audio_status = audio_approved`; `audio_downvote_count ≥ 5` → `audio_status = audio_rejected`.

**Errors:** Same as `/votes` — `INVALID_VOTE_TYPE`, `DUPLICATE_VOTE`, `FORBIDDEN` (self-vote), `NOT_FOUND`.

---

#### `POST /api/v1/phrases/:id/script-votes`

Vote on the accuracy of a phrase's native script representation.

**Auth:** Required — `contributor`, `admin`

**Path param:** `id` — phrase UUID

**Request:**
```json
{
  "vote_type": "upvote | downvote"
}
```

**Response `201`:**
```json
{ "message": "Script vote recorded." }
```

**Side effects:**
- Atomically increments `script_upvote_count` or `script_downvote_count`.
- Triggers threshold check: `script_upvote_count ≥ 3` → `script_status = approved`; `script_downvote_count ≥ 5` → `script_status = rejected`.

**Errors:** Same as `/votes`.

---

### Learning Content

> These endpoints are **planned** — handlers not yet implemented.

#### `GET /api/v1/flashcards`

Fetch a randomised set of approved flashcards.

**Auth:** Required — any role

**Query params:**

| Param | Type | Required | Description |
|---|---|---|---|
| `language_code` | string | Yes | Filter by language |
| `tone` | string | No | Filter by tone: `formal`, `netral`, `kasar` |
| `cursor` | string | No | Cursor for pagination |
| `limit` | int | No | Max items per page (default 20, max 20) |

**Response `200`:**
```json
{
  "flashcards": [
    {
      "id": "uuid",
      "text_latin": "string",
      "text_native_script": "string | omitted (only if script_status = approved)",
      "translation": "string",
      "language_code": "string",
      "tone": "formal | netral | kasar | omitted",
      "audio_url": "string | omitted (only if audio_status = audio_approved)",
      "audio_duration_seconds": "float | omitted",
      "words": [ "<Word>" ],
      "cultural_context": "<CulturalContext> | omitted"
    }
  ],
  "next_cursor": "string | null"
}
```

Only phrases with `status = approved` are returned. `text_native_script` is included only when `script_status = approved`. `audio_url` is included only when `audio_status = audio_approved`.

---

#### `GET /api/v1/conversation-scenarios`

Fetch a conversation scenario (3–8 approved phrases sharing a cultural context).

**Auth:** Required — any role

**Query params:**

| Param | Type | Required | Description |
|---|---|---|---|
| `language_code` | string | Yes | Language to fetch scenario for |

**Response `200`:**
```json
{
  "language_code": "string",
  "usage_situation": "string",
  "phrase_count": 5,
  "phrases": [ "<Phrase>" ]
}
```

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `404` | `NOT_FOUND` | Fewer than 3 approved phrases for this language |

---

#### `POST /api/v1/phrase-practice`

Record a self-assessment result for a practice session.

**Auth:** Required — `learner`, `contributor`, `admin`

**Request:**
```json
{
  "phrase_id": "uuid",
  "result": "tahu | belum_tahu"
}
```

**Response `201`:**
```json
{
  "id": "uuid",
  "phrase_id": "uuid",
  "result": "tahu | belum_tahu",
  "created_at": "RFC3339"
}
```

---

### Search

> **Planned** — handler not yet implemented.

#### `GET /api/v1/search`

Search approved phrases by text or root word form.

**Auth:** Required — any role

**Query params:**

| Param | Type | Required | Description |
|---|---|---|---|
| `q` | string | Yes | Search query (max 100 chars) |
| `language_code` | string | No | Filter by language |
| `search_type` | string | No | `text` (default) or `root` |
| `offset` | int | No | Pagination offset (default 0) |
| `limit` | int | No | Max items (default 50, max 50) |

**Response `200`:**
```json
{
  "phrases": [ "<Phrase>" ],
  "total": 42,
  "offset": 0,
  "limit": 50
}
```

Only `status = approved` phrases are returned. `search_type=root` searches `words.root_form_latin`.

---

### Admin

> **Planned** — handlers not yet implemented.

#### `GET /api/v1/admin/phrases/flagged`

List all phrases with `status = flagged`.

**Auth:** Required — `admin`

**Response `200`:**
```json
{
  "phrases": [ "<Phrase>" ]
}
```

---

#### `PATCH /api/v1/admin/phrases/:id/status`

Approve or reject a flagged phrase.

**Auth:** Required — `admin`

**Path param:** `id` — phrase UUID

**Request:**
```json
{
  "status": "approved | rejected"
}
```

**Response `200`:** `<Phrase>` (with `moderated_by` and `moderated_at` populated)

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `INVALID_STATUS` | `status` not `approved` or `rejected` |
| `404` | `NOT_FOUND` | Phrase not found |

---

#### `DELETE /api/v1/admin/phrases/:id`

Hard-delete a phrase and all related data (votes, flags, audio in S3).

**Auth:** Required — `admin`

**Path param:** `id` — phrase UUID

**Response `204`:** _(no body)_

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `404` | `NOT_FOUND` | Phrase not found |

---

#### `GET /api/v1/admin/users`

List all users with optional search.

**Auth:** Required — `admin`

**Query params:**

| Param | Type | Required | Description |
|---|---|---|---|
| `q` | string | No | Search by name or email |
| `offset` | int | No | Pagination offset |
| `limit` | int | No | Max items (default 50) |

**Response `200`:**
```json
{
  "users": [ "<User>" ],
  "total": 100,
  "offset": 0,
  "limit": 50
}
```

---

#### `PATCH /api/v1/admin/users/:id/ban`

Ban a user and reject all their pending phrases.

**Auth:** Required — `admin`

**Path param:** `id` — user UUID

**Request:** _(empty body)_

**Response `200`:** `<User>` (with `is_active = false`)

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `404` | `NOT_FOUND` | User not found |

---

#### `PATCH /api/v1/admin/users/:id/role`

Assign the `admin` role to a user.

**Auth:** Required — `admin`

**Path param:** `id` — user UUID

**Request:**
```json
{
  "role": "admin"
}
```

**Response `200`:** `<User>`

**Errors:**

| Status | Code | Condition |
|---|---|---|
| `400` | `INVALID_ROLE` | `role` is not a valid assignable role |
| `404` | `NOT_FOUND` | User not found |

---

## Validation Thresholds (Reference)

| Vote type | Approve threshold | Reject threshold |
|---|---|---|
| Text vote (`votes`) | `upvote_count ≥ 3` → `status = approved` | `downvote_count ≥ 5` → `status = rejected` |
| Flag (`flags`) | — | `flag_count ≥ 3` → `status = flagged` |
| Audio vote (`audio-votes`) | `audio_upvote_count ≥ 3` → `audio_status = audio_approved` | `audio_downvote_count ≥ 5` → `audio_status = audio_rejected` |
| Script vote (`script-votes`) | `script_upvote_count ≥ 3` → `script_status = approved` | `script_downvote_count ≥ 5` → `script_status = rejected` |

All count updates and status transitions happen atomically in a single DB transaction.

---

## Enum Reference

### `role`
`learner` · `contributor` · `admin`

### `status` (phrase)
`pending` · `approved` · `rejected` · `flagged` · `ai_failed`

### `script_status`
`none` · `pending` · `approved` · `rejected`

### `audio_status`
`none` · `pending` · `audio_approved` · `audio_rejected`

### `script_type`
`latin` · `javanese` · `sundanese` · `balinese` · `lontara` · `batak` · `other`

### `tone`
`formal` · `netral` · `kasar`

### `vote_type`
`upvote` · `downvote`

### `flag reason`
`inaccurate_translation` · `inappropriate_content` · `duplicate`

### `practice result`
`tahu` · `belum_tahu`

---

## Misc

### Health Check

```
GET /health
```

**Auth:** None  
**Response `200`:** `{"status":"ok"}`

### Audio Upload (Planned — Task 8)

Audio is submitted as a `multipart/form-data` request alongside the phrase text fields. Constraints:
- Format: `WAV` or `MP3`
- Max size: `5 MB`
- Max duration: `30 seconds`
- Violations return `422`

### Handwriting / Native Script Image (Planned — Task 8)

PNG exported from the Canvas is submitted as `multipart/form-data` alongside phrase data. Stored in S3, URL saved to `native_script_image_url`.

### Signed Audio URLs

`audio_url` values are signed S3 URLs with a minimum expiry of **1 hour**. Clients should not cache these URLs beyond their expiry.
