# Design Document — Bahasa Daerah Learning Platform

## Overview

Platform pembelajaran bahasa daerah dan dialek lokal Indonesia berbasis komunitas (UGC) dan AI. Penutur asli berkontribusi kalimat sehari-hari beserta terjemahan bahasa Indonesianya. Backend Go memproses input melalui AI API eksternal untuk menghasilkan struktur linguistik terstandar. Komunitas memvalidasi konten melalui sistem vote/flagging sebelum konten digunakan sebagai materi belajar.

### Tujuan Sistem

- Menyediakan platform bagi penutur asli untuk mendokumentasikan bahasa daerah Indonesia
- Mengekstrak metadata linguistik (POS, root word, tone) secara otomatis via AI
- Memvalidasi konten secara komunitas sebelum digunakan sebagai materi belajar
- Menyajikan flashcard, conversation scenario, dan phrase practice kepada Learner
- Mendukung aksara tradisional (Jawa, Sunda, Bali, Lontara, Batak) dan audio pengucapan

### Batasan Proyek

- 9 hari development, 5 orang tim
- Android primary, iOS future
- Offline capability terbatas (teks saja, 200 flashcard per bahasa)
- Audio offline terbatas (100MB per bahasa)

---

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Clients                                   │
│  ┌──────────────────┐          ┌──────────────────────────────┐  │
│  │  Flutter Mobile  │          │   React Web Dashboard        │  │
│  │  (Android/iOS)   │          │   (Admin only)               │  │
│  └────────┬─────────┘          └──────────────┬───────────────┘  │
└───────────┼──────────────────────────────────┼──────────────────┘
            │ HTTPS / REST                      │ HTTPS / REST
            ▼                                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Backend API (Go)                              │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │  Auth        │  │  Phrase      │  │  Validation Engine   │   │
│  │  Handler     │  │  Handler     │  │  (Vote/Flag/Status)  │   │
│  └──────────────┘  └──────────────┘  └──────────────────────┘   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │  Flashcard   │  │  Search      │  │  Admin Handler       │   │
│  │  Handler     │  │  Handler     │  │                      │   │
│  └──────────────┘  └──────────────┘  └──────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │              AI Pipeline (Async Worker)                  │    │
│  │  Queue → Process → Extract → Store → Retry               │    │
│  └──────────────────────────────────────────────────────────┘    │
└──────────────┬──────────────────────────────┬───────────────────┘
               │                              │
               ▼                              ▼
┌──────────────────────┐        ┌─────────────────────────────┐
│   PostgreSQL         │        │   S3-Compatible Storage      │
│   (Primary DB)       │        │   (Audio WAV/MP3, PNG)       │
└──────────────────────┘        └─────────────────────────────┘
                                               │
                                               ▼
                                ┌─────────────────────────────┐
                                │   External AI API            │
                                │   (LLM for POS/tone/root)   │
                                └─────────────────────────────┘
```

### Komponen Utama

| Komponen | Teknologi | Tanggung Jawab |
|---|---|---|
| Backend API | Go (net/http atau Gin) | REST API, business logic, auth |
| AI Pipeline | Go goroutine + channel | Async processing, retry logic |
| Database | PostgreSQL | Persistent storage, relational data |
| Object Storage | S3-compatible (MinIO/AWS S3) | Audio files, aksara PNG images |
| Mobile App | Flutter (Dart) | Android/iOS client |
| Web Dashboard | React | Admin moderation UI |

### API Versioning

Semua endpoint menggunakan URL versioning: `/api/v1/`. Ini memastikan backward compatibility saat backend diperbarui.

### Authentication Flow

```
Client → POST /api/v1/auth/login
       ← JWT (24h expiry) + role

Client → GET /api/v1/phrases (Authorization: Bearer <token>)
       ← Data (role-based access)
```

JWT payload berisi: `user_id`, `role`, `exp`. Middleware memvalidasi token pada setiap protected endpoint.

---

## Components and Interfaces

### REST API Endpoints

#### Auth
```
POST   /api/v1/auth/register          # Registrasi (learner/contributor)
POST   /api/v1/auth/login             # Login, return JWT
POST   /api/v1/auth/upgrade-role      # Learner → Contributor (self-service)
```

#### Phrases
```
POST   /api/v1/phrases                # Submit phrase (Contributor)
GET    /api/v1/phrases                # List pending phrases (Contributor, for voting)
GET    /api/v1/phrases/:id            # Detail phrase
GET    /api/v1/phrases/my             # Phrase milik contributor yang login
```

#### Voting & Flagging
```
POST   /api/v1/phrases/:id/votes      # Upvote/downvote phrase (Contributor)
POST   /api/v1/phrases/:id/flags      # Flag phrase (Learner/Contributor)
POST   /api/v1/phrases/:id/audio-votes  # Vote audio (Contributor)
POST   /api/v1/phrases/:id/script-votes # Vote aksara (Contributor)
```

#### Learning Content
```
GET    /api/v1/flashcards             # Flashcard list (filter: lang, tone, cursor)
GET    /api/v1/conversation-scenarios # Conversation scenarios (filter: lang)
POST   /api/v1/phrase-practice        # Submit practice result
```

#### Search
```
GET    /api/v1/search                 # Search phrases (text + root_form)
```

#### Languages (Admin)
```
GET    /api/v1/languages              # List all languages
POST   /api/v1/admin/languages        # Create language (Admin)
PATCH  /api/v1/admin/languages/:code  # Toggle active/inactive (Admin)
```

#### Admin
```
GET    /api/v1/admin/phrases/flagged  # List flagged phrases
PATCH  /api/v1/admin/phrases/:id/status  # Approve/reject flagged phrase
GET    /api/v1/admin/users            # List users (search by name/email)
PATCH  /api/v1/admin/users/:id/ban    # Ban user
PATCH  /api/v1/admin/users/:id/role   # Assign admin role
DELETE /api/v1/admin/phrases/:id      # Hard delete phrase + related data
```

### Middleware Stack

```
Request → CORS → RateLimit → Logger → Auth (JWT validate) → RoleCheck → Handler
```

- **Auth Middleware**: Validasi JWT, inject `user_id` dan `role` ke context
- **RoleCheck Middleware**: Verifikasi role minimum (learner/contributor/admin)
- **RateLimit**: Per-IP rate limiting untuk mencegah abuse

### AI Pipeline Interface

```go
type AIPipelineJob struct {
    PhraseID    uuid.UUID
    TextLatin   string
    Translation string
    Attempt     int
}

type AIResponse struct {
    Words []AIWord `json:"words"`
    Tone  string   `json:"tone"`
}

type AIWord struct {
    Surface string `json:"surface"`
    Root    string `json:"root"`
    POS     string `json:"pos"`
}
```

Pipeline menggunakan channel-based worker pool. Phrase baru di-enqueue setelah berhasil disimpan ke DB. Worker mengambil job dari channel, memanggil AI API, menyimpan hasil, dan retry dengan exponential backoff jika gagal.

### Audio Storage Interface

```go
type AudioStorageService interface {
    Upload(ctx context.Context, fileBytes []byte, mimeType string) (url string, err error)
    GenerateSignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
    Delete(ctx context.Context, objectKey string) error
}
```

Signed URL dengan expiry minimal 1 jam. Object key berbasis UUID untuk menghindari konflik.

### Validation Engine

Validation Engine adalah komponen stateless yang dipanggil setelah setiap vote/flag operation:

```
VoteUpserted → CheckThresholds(phraseID) → UpdateStatus(phraseID, newStatus)
```

Threshold:
- Phrase: upvote ≥ 3 → `approved`; downvote ≥ 5 → `rejected`
- Flag: count ≥ 3 → `flagged`
- Audio: upvote ≥ 3 → `audio_approved`; downvote ≥ 5 → `audio_rejected`
- Script: upvote ≥ 3 → `script_status: approved`; downvote ≥ 5 → `script_status: rejected`

---

## Data Models

### Entity Relationship Diagram

Diagram lengkap tersedia di [`erd.md`](./erd.md) dalam format Mermaid.

Ringkasan relasi antar entitas:

```
languages ──< cultural_contexts
languages ──< phrases
languages ──< words

cultural_contexts ──< phrases

users ──< phrases (contributor_id)
users ──< phrases (moderated_by)
users ──< votes (contributor_id)
users ──< flags (user_id)
users ──< audio_votes (contributor_id)
users ──< script_votes (contributor_id)
users ──< phrase_practice_results (learner_id)

phrases ──< phrase_words >── words
phrases ──< votes
phrases ──< flags
phrases ──< audio_votes
phrases ──< script_votes
phrases ──< phrase_practice_results
```

### Table Definitions

#### `users`
```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,  -- bcrypt cost 12
    role            VARCHAR(50) NOT NULL CHECK (role IN ('learner', 'contributor', 'admin')),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### `languages`
```sql
CREATE TABLE languages (
    code        VARCHAR(20) PRIMARY KEY,  -- e.g. 'jv', 'su', 'ban'
    name        VARCHAR(255) NOT NULL,
    region      VARCHAR(255),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### `cultural_contexts`
```sql
CREATE TABLE cultural_contexts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language_code    VARCHAR(20) NOT NULL REFERENCES languages(code),
    region           VARCHAR(255),
    usage_situation  VARCHAR(500),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### `phrases`
```sql
CREATE TABLE phrases (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text_latin              VARCHAR(500) NOT NULL,
    text_native_script      TEXT,                    -- nullable
    script_type             VARCHAR(50) CHECK (script_type IN (
                                'latin','javanese','sundanese',
                                'balinese','lontara','batak','other'
                            )),
    translation             VARCHAR(500) NOT NULL,
    language_code           VARCHAR(20) NOT NULL REFERENCES languages(code),
    tone                    VARCHAR(50) CHECK (tone IN ('formal','netral','kasar')),
    status                  VARCHAR(50) NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending','approved','rejected','flagged','ai_failed')),
    script_status           VARCHAR(50) NOT NULL DEFAULT 'none'
                                CHECK (script_status IN ('none','pending','approved','rejected')),
    contributor_id          UUID NOT NULL REFERENCES users(id),
    cultural_context_id     UUID REFERENCES cultural_contexts(id),
    audio_url               TEXT,                    -- nullable
    audio_duration_seconds  FLOAT,                   -- nullable
    audio_status            VARCHAR(50) NOT NULL DEFAULT 'none'
                                CHECK (audio_status IN ('none','pending','audio_approved','audio_rejected')),
    native_script_image_url TEXT,                    -- nullable, S3 URL for handwriting PNG
    moderated_by            UUID REFERENCES users(id),  -- Admin who last moderated
    moderated_at            TIMESTAMPTZ,
    upvote_count            INTEGER NOT NULL DEFAULT 0,
    downvote_count          INTEGER NOT NULL DEFAULT 0,
    flag_count              INTEGER NOT NULL DEFAULT 0,
    audio_upvote_count      INTEGER NOT NULL DEFAULT 0,
    audio_downvote_count    INTEGER NOT NULL DEFAULT 0,
    script_upvote_count     INTEGER NOT NULL DEFAULT 0,
    script_downvote_count   INTEGER NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_phrases_language_status ON phrases(language_code, status);
CREATE INDEX idx_phrases_contributor ON phrases(contributor_id);
CREATE INDEX idx_phrases_status ON phrases(status);
```

#### `words`
```sql
CREATE TABLE words (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surface_form_latin      VARCHAR(255) NOT NULL,
    surface_form_native_script VARCHAR(255),         -- nullable
    root_form_latin         VARCHAR(255) NOT NULL,
    root_form_native_script VARCHAR(255),            -- nullable
    script_type             VARCHAR(50) CHECK (script_type IN (
                                'latin','javanese','sundanese',
                                'balinese','lontara','batak','other'
                            )),
    part_of_speech          VARCHAR(50),
    language_code           VARCHAR(20) NOT NULL REFERENCES languages(code),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_words_root_form ON words(root_form_latin, language_code);
CREATE INDEX idx_words_language ON words(language_code);
```

#### `phrase_words` (junction table)
```sql
CREATE TABLE phrase_words (
    phrase_id   UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    word_id     UUID NOT NULL REFERENCES words(id),
    position    INTEGER NOT NULL,  -- word order in phrase
    PRIMARY KEY (phrase_id, word_id)
);
```

#### `votes`
```sql
CREATE TABLE votes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase_id       UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    contributor_id  UUID NOT NULL REFERENCES users(id),
    vote_type       VARCHAR(10) NOT NULL CHECK (vote_type IN ('upvote','downvote')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (phrase_id, contributor_id)  -- prevent duplicate votes
);
```

#### `flags`
```sql
CREATE TABLE flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase_id   UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    reason      VARCHAR(50) NOT NULL CHECK (reason IN (
                    'inaccurate_translation','inappropriate_content','duplicate'
                )),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (phrase_id, user_id)  -- one flag per user per phrase
);
```

#### `audio_votes`
```sql
CREATE TABLE audio_votes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase_id       UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    contributor_id  UUID NOT NULL REFERENCES users(id),
    vote_type       VARCHAR(10) NOT NULL CHECK (vote_type IN ('upvote','downvote')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (phrase_id, contributor_id)
);
```

#### `script_votes`
```sql
CREATE TABLE script_votes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase_id       UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    contributor_id  UUID NOT NULL REFERENCES users(id),
    vote_type       VARCHAR(10) NOT NULL CHECK (vote_type IN ('upvote','downvote')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (phrase_id, contributor_id)
);
```

#### `phrase_practice_results`
```sql
CREATE TABLE phrase_practice_results (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id  UUID NOT NULL REFERENCES users(id),
    phrase_id   UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    result      VARCHAR(20) NOT NULL CHECK (result IN ('tahu','belum_tahu')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_practice_learner ON phrase_practice_results(learner_id, phrase_id);
```

### Vote Count Denormalization

Untuk performa query, tabel `phrases` menyimpan denormalized vote counts langsung sebagai kolom (`upvote_count`, `downvote_count`, `flag_count`, `audio_upvote_count`, `audio_downvote_count`, `script_upvote_count`, `script_downvote_count`). Count diperbarui atomically dalam satu transaksi bersama insert vote. Threshold check dilakukan setelah update count.

---

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property Reflection

Sebelum menulis properties final, saya mereview semua kandidat property untuk menghilangkan redundansi:

**Redundansi yang ditemukan:**
1. Properties voting untuk text (5.1-5.7), audio (19.1-19.6), dan script (21.7-21.11) mengikuti pola yang sama persis. Ketiga set ini dapat digeneralisasi menjadi satu property "vote recording" dan satu property "threshold transition" yang berlaku untuk semua jenis vote.
2. Properties threshold approval (5.3 upvote≥3 → approved) dan (5.4 downvote≥5 → rejected) dapat digabung menjadi satu property "threshold state transition" yang berlaku untuk semua vote types.
3. Properties filtering untuk flashcard (6.1 hanya approved) dan conversation scenario (7.1 hanya approved) mengikuti pola yang sama — keduanya adalah "only approved content is served" property.
4. Properties pagination (6.5 flashcard ≤20, 8.4 search ≤50) adalah instance dari property yang sama: "response count tidak melebihi limit yang dikonfigurasi".
5. Properties completeness untuk flashcard (6.2) dan scenario metadata (7.3) adalah instance dari "response contains all required fields" pattern.

**Properties yang dikonsolidasi:**
- Vote recording + duplicate prevention + self-vote prevention → 3 properties yang berlaku untuk semua vote types
- Threshold transitions → 1 property yang berlaku untuk semua vote types dan thresholds
- Content filtering → 1 property "only approved content served"
- Validation rejection → 1 property untuk input validation

---

### Property 1: Registrasi menghasilkan akun dengan password ter-hash

*For any* kombinasi (name, email, password, role) yang valid, setelah registrasi berhasil, password yang tersimpan di database SHALL berupa bcrypt hash dengan cost factor ≥ 12, bukan plaintext.

**Validates: Requirements 1.2, 1.6**

---

### Property 2: Email duplikat selalu ditolak

*For any* email yang sudah terdaftar di sistem, percobaan registrasi ulang dengan email yang sama SHALL selalu mengembalikan HTTP 409, terlepas dari nilai name, password, atau role yang dikirimkan.

**Validates: Requirements 1.3**

---

### Property 3: Login valid menghasilkan JWT dengan expiry 24 jam

*For any* user yang terdaftar, login dengan kredensial yang benar SHALL menghasilkan JWT yang ketika di-decode memiliki `exp - iat = 86400 detik` (24 jam), dan payload berisi `user_id` dan `role` yang benar.

**Validates: Requirements 1.4**

---

### Property 4: Kredensial invalid selalu menghasilkan HTTP 401

*For any* kombinasi (email, password) di mana email tidak terdaftar atau password tidak cocok, endpoint login SHALL selalu mengembalikan HTTP 401 tanpa mengungkap detail mana yang salah.

**Validates: Requirements 1.5**

---

### Property 5: Phrase valid selalu disimpan dengan status pending

*For any* Phrase dengan `text_latin` non-empty (≤500 karakter), `translation` non-empty, dan `language_code` yang aktif di tabel `languages`, submission oleh Contributor yang terautentikasi SHALL menyimpan Phrase dengan `status = pending` dan mengembalikan UUID yang valid.

**Validates: Requirements 2.1, 2.5**

---

### Property 6: Validasi input Phrase menolak semua input tidak valid

*For any* Phrase submission yang melanggar salah satu constraint berikut: (a) `text_latin` kosong atau tidak ada, (b) `translation` tidak ada, (c) `text_latin` panjangnya > 500 karakter, (d) `language_code` tidak aktif atau tidak ada — sistem SHALL mengembalikan HTTP error (400 atau 422) dan tidak menyimpan Phrase ke database.

**Validates: Requirements 2.2, 2.3, 2.4, 21.5**

---

### Property 7: AI Pipeline menyimpan semua data linguistik dari respons AI

*For any* AI API response yang valid dengan skema `{ "words": [...], "tone": string }`, AI Pipeline SHALL menyimpan semua words (dengan surface, root, POS) ke tabel `words`, membuat relasi di `phrase_words`, dan menyimpan `tone` ke tabel `phrases` — tidak ada data yang hilang atau terpotong.

**Validates: Requirements 3.2**

---

### Property 8: AI Pipeline retry eksponensial dan transisi ke ai_failed

*For any* Phrase yang AI API-nya selalu gagal (error atau timeout), AI Pipeline SHALL melakukan tepat 3 retry dengan jeda yang mengikuti pola eksponensial (jeda ke-n ≥ 2^(n-1) detik), dan setelah semua retry habis, status Phrase SHALL berubah menjadi `ai_failed`.

**Validates: Requirements 3.4, 3.5**

---

### Property 9: Vote tercatat dan count bertambah atomically

*For any* Contributor yang memberikan vote (upvote atau downvote) pada Phrase yang belum pernah ia vote, vote SHALL tersimpan di tabel `votes` dan count yang sesuai (`upvote_count` atau `downvote_count`) di tabel `phrases` SHALL bertambah tepat 1 dalam satu transaksi atomik.

**Validates: Requirements 5.1, 19.1, 21.7**

---

### Property 10: Duplicate vote selalu ditolak dengan HTTP 409

*For any* Contributor yang sudah pernah memberikan vote pada sebuah Phrase (text vote, audio vote, atau script vote), percobaan vote kedua pada Phrase yang sama SHALL selalu mengembalikan HTTP 409 dan count di database SHALL tidak berubah.

**Validates: Requirements 5.2, 19.2, 21.8**

---

### Property 11: Self-vote selalu ditolak dengan HTTP 403

*For any* Contributor yang mencoba memberikan vote (text, audio, atau script) pada Phrase yang ia sendiri submit, sistem SHALL selalu mengembalikan HTTP 403 dan vote tidak tersimpan.

**Validates: Requirements 5.7, 19.3, 21.9**

---

### Property 12: Threshold vote mengubah status secara deterministik

*For any* Phrase, ketika jumlah upvote mencapai tepat 3, status SHALL berubah menjadi `approved`; ketika jumlah downvote mencapai tepat 5, status SHALL berubah menjadi `rejected`; ketika jumlah flag mencapai tepat 3, status SHALL berubah menjadi `flagged`. Transisi ini berlaku untuk text status, audio_status, dan script_status dengan threshold yang sama.

**Validates: Requirements 5.3, 5.4, 5.6, 19.4, 19.5, 21.10, 21.11**

---

### Property 13: Hanya konten approved yang disajikan kepada Learner

*For any* request Flashcard atau Conversation Scenario dari Learner, semua Phrase yang dikembalikan SHALL memiliki `status = approved`. Tidak ada Phrase dengan status `pending`, `rejected`, `flagged`, atau `ai_failed` yang boleh muncul dalam respons.

**Validates: Requirements 6.1, 7.1**

---

### Property 14: Flashcard mengandung semua field yang diperlukan

*For any* Flashcard yang dikembalikan oleh API, respons SHALL mengandung: `text_latin`, `translation`, daftar `words` (dengan `root_form_latin` dan `part_of_speech`), `tone`, dan `cultural_context`. Jika Phrase memiliki `text_native_script` dengan `script_status = approved`, field tersebut SHALL disertakan. Jika Phrase memiliki `audio_url` dengan `audio_status = audio_approved`, field tersebut SHALL disertakan.

**Validates: Requirements 6.2, 18.3, 19.6, 21.6**

---

### Property 15: Filter Flashcard menghasilkan hasil yang sesuai filter

*For any* kombinasi filter (language_code, tone) yang valid, semua Flashcard yang dikembalikan SHALL memiliki `language_code` dan `tone` yang sesuai dengan filter yang diberikan. Tidak ada Flashcard dari bahasa atau tone lain yang boleh muncul.

**Validates: Requirements 6.3**

---

### Property 16: Pagination tidak pernah melebihi limit

*For any* request Flashcard (limit 20) atau Search (limit 50), jumlah item dalam respons SHALL selalu ≤ limit yang ditentukan, dan jika total data melebihi limit, respons SHALL menyertakan cursor/offset untuk halaman berikutnya.

**Validates: Requirements 6.5, 8.4**

---

### Property 17: Conversation Scenario mengandung 3-8 Phrase dengan Cultural Context kompatibel

*For any* bahasa daerah yang memiliki ≥ 3 approved Phrase dengan Cultural Context yang sama, Conversation Scenario yang dikembalikan SHALL mengandung antara 3 dan 8 Phrase, semua dari Cultural Context yang sama atau kompatibel.

**Validates: Requirements 7.2, 7.3**

---

### Property 18: Search mengembalikan hanya approved Phrase yang mengandung query

*For any* search query (teks atau root_form) dan language_code, semua Phrase yang dikembalikan SHALL: (a) berstatus `approved`, (b) mengandung query text pada kolom `text_latin` atau `translation` (untuk text search), atau memiliki Word dengan `root_form_latin` yang cocok (untuk root search).

**Validates: Requirements 8.1, 8.3**

---

### Property 19: Validasi audio menolak semua file tidak valid

*For any* file audio yang diunggah bersama Phrase submission: jika format bukan WAV atau MP3, sistem SHALL mengembalikan HTTP 422; jika ukuran > 5MB atau durasi > 30 detik, sistem SHALL mengembalikan HTTP 422; dan tidak ada file yang tersimpan ke S3 dalam kasus penolakan.

**Validates: Requirements 17.1, 17.4, 17.5**

---

### Property 20: Self-upgrade role mengubah role dan menghasilkan JWT baru

*For any* user dengan role `learner` yang melakukan self-upgrade, setelah request berhasil: (a) role di database SHALL berubah menjadi `contributor`, (b) JWT baru yang dikembalikan SHALL mengandung role `contributor`, (c) user dengan role `contributor` atau `admin` yang mencoba upgrade SHALL mendapatkan HTTP 409.

**Validates: Requirements 23.3, 23.4, 23.6**

---

## Error Handling

### Strategi Error Handling

#### HTTP Error Codes

| Kode | Kondisi |
|---|---|
| 400 | Input tidak valid (field wajib kosong, format salah) |
| 401 | Token JWT tidak ada, expired, atau invalid |
| 403 | Akses ditolak (role tidak cukup, self-vote) |
| 404 | Resource tidak ditemukan |
| 409 | Konflik (email duplikat, duplicate vote) |
| 422 | Unprocessable entity (constraint violation: panjang teks, format file) |
| 500 | Internal server error |

#### Error Response Format

Semua error response menggunakan format konsisten:

```json
{
  "error": {
    "code": "DUPLICATE_EMAIL",
    "message": "Email sudah terdaftar. Gunakan email lain atau login.",
    "details": {}
  }
}
```

#### AI Pipeline Error Handling

```
Phrase submitted
    │
    ▼
Enqueue to AI worker channel
    │
    ▼
AI API call (timeout: 10s)
    │
    ├── Success → Parse JSON → Save words/tone → Done
    │
    └── Error/Timeout
            │
            ├── Attempt 1 → Wait 1s → Retry
            ├── Attempt 2 → Wait 2s → Retry
            ├── Attempt 3 → Wait 4s → Retry
            │
            └── All failed → Set status = ai_failed → Log error (phrase_id, error_msg)
```

Exponential backoff: jeda ke-n = 2^(n-1) detik (1s, 2s, 4s).

#### Database Error Handling

- Semua operasi DB menggunakan transaksi untuk operasi multi-step (vote + count update)
- Connection pool dengan retry untuk transient connection errors
- Deadlock detection dengan retry otomatis (maksimal 3x)

#### File Upload Error Handling

- Validasi format, ukuran, dan durasi sebelum upload ke S3
- Jika S3 upload gagal setelah file divalidasi, kembalikan HTTP 500 dan log error
- Tidak menyimpan referensi URL ke DB jika S3 upload gagal

#### Mobile App Error Handling

- HTTP 5xx → tampilkan pesan generik, log detail ke local log
- HTTP 401 → hapus token, redirect ke login screen
- Network timeout → tampilkan pesan "Koneksi bermasalah, coba lagi"
- Offline → sajikan dari Offline_Cache jika tersedia

---

## Testing Strategy

### Pendekatan Dual Testing

Platform ini menggunakan dua pendekatan testing yang saling melengkapi:

1. **Unit Tests + Property-Based Tests**: Untuk business logic murni (validation, vote counting, status transitions, search filtering)
2. **Integration Tests**: Untuk interaksi dengan external services (AI API, S3, PostgreSQL)

### Property-Based Testing

Library yang digunakan: **[`gopter`](https://github.com/leanovate/gopter)** untuk Go.

Setiap property test dikonfigurasi dengan minimum **100 iterasi** untuk memastikan coverage yang memadai.

Tag format untuk setiap test:
```go
// Feature: bahasa-daerah-learning-platform, Property N: <property_text>
```

#### Property Tests yang Diimplementasikan

| Property | Test | Generators |
|---|---|---|
| P1: Password hashing | Verifikasi bcrypt hash dari random passwords | `gen.AlphaString()` untuk password |
| P2: Email duplikat → 409 | Register dua kali dengan email sama | `gen.AlphaString()` untuk email |
| P3: JWT expiry 24 jam | Login dengan random valid user | Random user generator |
| P4: Invalid credentials → 401 | Login dengan random invalid creds | Random email/password generator |
| P5: Phrase valid → pending | Submit random valid phrase | Random phrase generator |
| P6: Invalid phrase → error | Submit phrase dengan berbagai invalid inputs | Random invalid phrase generator |
| P7: AI data persistence | Mock AI response → verifikasi DB | Random AI response generator |
| P8: AI retry + ai_failed | Mock AI always fail → verifikasi retry count | N/A (deterministic mock) |
| P9: Vote count atomic | Random contributor + phrase → vote | Random contributor/phrase generator |
| P10: Duplicate vote → 409 | Vote dua kali pada phrase yang sama | Random contributor/phrase generator |
| P11: Self-vote → 403 | Contributor vote pada phrase sendiri | Random contributor generator |
| P12: Threshold transitions | Tambahkan votes hingga threshold | Random vote sequence generator |
| P13: Only approved in flashcard | Mix phrases → request flashcard | Random phrase status generator |
| P14: Flashcard completeness | Random approved phrase → flashcard | Random phrase generator |
| P15: Filter correctness | Random filter → verify results | Random (language, tone) generator |
| P16: Pagination limit | > limit phrases → verify count ≤ limit | Random large dataset generator |
| P17: Scenario 3-8 phrases | Random approved phrases → scenario | Random phrase count generator |
| P18: Search correctness | Random query → verify results | Random query + phrase generator |
| P19: Audio validation | Random audio files → verify accept/reject | Random file size/format generator |
| P20: Role upgrade | Random learner → upgrade → verify | Random user generator |

#### Contoh Property Test (Go + gopter)

```go
// Feature: bahasa-daerah-learning-platform, Property 12: Threshold vote mengubah status secara deterministik
func TestVoteThresholdTransition(t *testing.T) {
    properties := gopter.NewProperties(gopter.DefaultTestParameters())
    
    properties.Property("upvote threshold 3 changes status to approved", prop.ForAll(
        func(phraseID uuid.UUID, contributors []uuid.UUID) bool {
            // Setup: create phrase with pending status
            // Add 3 upvotes from different contributors
            // Verify status = approved
            return verifyStatusAfterVotes(phraseID, contributors[:3], "upvote") == "approved"
        },
        gen.UUID(),
        gen.SliceOfN(3, gen.UUID()),
    ))
    
    properties.TestingRun(t)
}
```

### Unit Tests

Unit tests fokus pada:
- Specific examples untuk edge cases
- Integration points antar komponen
- Error conditions yang tidak tercakup property tests

```
pkg/
├── auth/
│   ├── auth_test.go          # Login, register, JWT validation
│   └── bcrypt_test.go        # Password hashing
├── phrase/
│   ├── phrase_test.go        # Submit, validation
│   └── validation_test.go    # Vote/flag logic
├── flashcard/
│   └── flashcard_test.go     # Filtering, pagination
├── search/
│   └── search_test.go        # Text search, root search
└── ai/
    └── pipeline_test.go      # Retry logic, response parsing
```

### Integration Tests

Integration tests menggunakan test database (PostgreSQL) dan mock S3:

```
tests/integration/
├── auth_integration_test.go
├── phrase_integration_test.go
├── voting_integration_test.go
├── flashcard_integration_test.go
├── search_integration_test.go
└── ai_pipeline_integration_test.go
```

### Mobile App Testing

- **Widget Tests**: Untuk setiap screen (login, flashcard, phrase submission)
- **Integration Tests**: Menggunakan mock Backend_API untuk end-to-end flows
- **Offline Tests**: Verifikasi Offline_Cache behavior saat network tidak tersedia

### Web Dashboard Testing

- **Component Tests**: React Testing Library untuk setiap komponen moderasi
- **E2E Tests**: Playwright untuk critical admin flows (approve/reject, ban user)

### Test Coverage Target

| Layer | Target Coverage |
|---|---|
| Backend business logic | ≥ 80% |
| API handlers | ≥ 70% |
| Mobile critical paths | ≥ 60% |
| Web Dashboard | ≥ 60% |
