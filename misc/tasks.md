# Implementation Plan: Bahasa Daerah Learning Platform

## Overview

Implementasi platform pembelajaran bahasa daerah Indonesia berbasis komunitas dan AI. Stack: Go backend (single binary embed React web dashboard), PostgreSQL, Flutter mobile (Android), React web dashboard, S3-compatible storage (Cloudflare R2 / Supabase Storage). Deployment di Railway.

Pendekatan: bangun dari fondasi (DB schema + auth) → core backend features → AI pipeline → mobile app → web dashboard → integrasi akhir.

## Tasks

- [x] 1. Setup project structure dan database schema
  - Inisialisasi Go module (`go mod init`) dengan struktur direktori: `cmd/server`, `internal/auth`, `internal/phrase`, `internal/validation`, `internal/flashcard`, `internal/search`, `internal/ai`, `internal/admin`, `internal/storage`, `pkg/middleware`, `pkg/db`
  - Buat file migrasi SQL untuk semua tabel: `users`, `languages`, `cultural_contexts`, `phrases`, `words`, `phrase_words`, `votes`, `flags`, `audio_votes`, `script_votes`, `phrase_practice_results`
  - Terapkan semua kolom, constraint CHECK, UNIQUE, foreign key, dan index sesuai design (termasuk kolom denormalized vote counts di tabel `phrases`)
  - Setup koneksi PostgreSQL dengan `pgx` atau `database/sql`, connection pool, dan helper transaksi
  - Buat file `docker-compose.yml` untuk PostgreSQL lokal dan seed data awal (minimal 1 language aktif)
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7_

- [x] 2. Implementasi autentikasi dan manajemen pengguna
  - [x] 2.1 Implementasi endpoint `POST /api/v1/auth/register`
    - Validasi input (name, email, password, role hanya `learner`/`contributor`)
    - Hash password dengan bcrypt cost factor 12
    - Simpan user ke DB, kembalikan JWT 24 jam
    - Kembalikan HTTP 409 jika email duplikat
    - _Requirements: 1.1, 1.2, 1.3, 1.6_

  - [ ]* 2.2 Write property test untuk registrasi
    - **Property 1: Registrasi menghasilkan akun dengan password ter-hash**
    - **Validates: Requirements 1.2, 1.6**
    - **Property 2: Email duplikat selalu ditolak**
    - **Validates: Requirements 1.3**

  - [x] 2.3 Implementasi endpoint `POST /api/v1/auth/login`
    - Validasi kredensial, kembalikan JWT 24 jam + role
    - Kembalikan HTTP 401 tanpa detail jika kredensial salah
    - _Requirements: 1.4, 1.5_

  - [ ]* 2.4 Write property test untuk login
    - **Property 3: Login valid menghasilkan JWT dengan expiry 24 jam**
    - **Validates: Requirements 1.4**
    - **Property 4: Kredensial invalid selalu menghasilkan HTTP 401**
    - **Validates: Requirements 1.5**

  - [x] 2.5 Implementasi JWT middleware dan role-check middleware
    - Validasi JWT pada setiap protected endpoint, inject `user_id` dan `role` ke context
    - RoleCheck middleware untuk learner/contributor/admin
    - _Requirements: 1.4, 16.5_

  - [x] 2.6 Implementasi endpoint `POST /api/v1/auth/upgrade-role`
    - Learner self-upgrade ke contributor tanpa persetujuan
    - Kembalikan JWT baru dengan role `contributor`
    - Kembalikan HTTP 409 jika sudah contributor/admin
    - _Requirements: 23.3, 23.4, 23.6_

  - [ ]* 2.7 Write property test untuk role upgrade
    - **Property 20: Self-upgrade role mengubah role dan menghasilkan JWT baru**
    - **Validates: Requirements 23.3, 23.4, 23.6**

- [x] 3. Checkpoint — Pastikan semua test auth pass
  - Pastikan semua test pass, tanyakan kepada user jika ada pertanyaan.

- [x] 4. Implementasi manajemen bahasa daerah dan pengiriman Phrase
  - [x] 4.1 Implementasi endpoint languages (public + admin)
    - `GET /api/v1/languages` — list semua bahasa
    - `POST /api/v1/admin/languages` — buat bahasa baru (Admin)
    - `PATCH /api/v1/admin/languages/:code` — toggle active/inactive (Admin)
    - _Requirements: 9.1, 9.2, 9.3_

  - [x] 4.2 Implementasi endpoint `POST /api/v1/phrases`
    - Validasi input: `text_latin` wajib (≤500 karakter), `translation` wajib, `language_code` harus aktif
    - Validasi field opsional: jika `text_native_script` ada maka `script_type` wajib; validasi enum `script_type`
    - Simpan Phrase dengan status `pending`, kembalikan UUID + status
    - Kembalikan HTTP 400/422 sesuai jenis pelanggaran validasi
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 9.4, 21.1, 21.2, 21.3, 21.4, 21.5_

  - [ ]* 4.3 Write property test untuk pengiriman Phrase
    - **Property 5: Phrase valid selalu disimpan dengan status pending**
    - **Validates: Requirements 2.1, 2.5**
    - **Property 6: Validasi input Phrase menolak semua input tidak valid**
    - **Validates: Requirements 2.2, 2.3, 2.4, 21.5**

  - [x] 4.4 Implementasi endpoint `GET /api/v1/phrases`, `GET /api/v1/phrases/:id`, `GET /api/v1/phrases/my`
    - List pending phrases untuk voting (Contributor)
    - Detail phrase dengan vote counts
    - Phrase milik contributor yang login beserta status terkini
    - _Requirements: 11.4, 11.5, 12.1_

- [x] 5. Implementasi Validation Engine — Voting dan Flagging
  - [x] 5.1 Implementasi endpoint `POST /api/v1/phrases/:id/votes`
    - Catat vote (upvote/downvote) dalam transaksi atomik bersama update count
    - Cegah duplicate vote (HTTP 409) dan self-vote (HTTP 403)
    - Panggil threshold check setelah update count
    - _Requirements: 5.1, 5.2, 5.7_

  - [ ]* 5.2 Write property test untuk vote recording
    - **Property 9: Vote tercatat dan count bertambah atomically**
    - **Validates: Requirements 5.1, 19.1, 21.7**
    - **Property 10: Duplicate vote selalu ditolak dengan HTTP 409**
    - **Validates: Requirements 5.2, 19.2, 21.8**
    - **Property 11: Self-vote selalu ditolak dengan HTTP 403**
    - **Validates: Requirements 5.7, 19.3, 21.9**

  - [x] 5.3 Implementasi Validation Engine — threshold transitions
    - Fungsi `CheckThresholds(phraseID)`: upvote ≥ 3 → `approved`; downvote ≥ 5 → `rejected`
    - Dipanggil setelah setiap vote/flag operation
    - _Requirements: 5.3, 5.4_

  - [ ]* 5.4 Write property test untuk threshold transitions
    - **Property 12: Threshold vote mengubah status secara deterministik**
    - **Validates: Requirements 5.3, 5.4, 5.6, 19.4, 19.5, 21.10, 21.11**

  - [x] 5.5 Implementasi endpoint `POST /api/v1/phrases/:id/flags`
    - Catat flag dengan alasan (`inaccurate_translation`, `inappropriate_content`, `duplicate`)
    - Satu flag per user per phrase
    - Jika flag count ≥ 3 → ubah status ke `flagged`
    - _Requirements: 5.5, 5.6_

  - [x] 5.6 Implementasi endpoint `POST /api/v1/phrases/:id/audio-votes`
    - Vote audio terpisah dari vote teks
    - Cegah duplicate dan self-vote
    - Threshold: upvote ≥ 3 → `audio_approved`; downvote ≥ 5 → `audio_rejected`
    - _Requirements: 19.1, 19.2, 19.3, 19.4, 19.5_

  - [x] 5.7 Implementasi endpoint `POST /api/v1/phrases/:id/script-votes`
    - Vote aksara terpisah dari vote teks dan audio
    - Cegah duplicate dan self-vote
    - Threshold: upvote ≥ 3 → `script_status: approved`; downvote ≥ 5 → `script_status: rejected`
    - _Requirements: 21.7, 21.8, 21.9, 21.10, 21.11_

- [x] 6. Checkpoint — Pastikan semua test validation engine pass
  - Pastikan semua test pass, tanyakan kepada user jika ada pertanyaan.

- [ ] 7. Implementasi AI Pipeline (async worker)
  - [ ] 7.1 Implementasi channel-based worker pool untuk AI Pipeline
    - Struct `AIPipelineJob` dan `AIResponse` sesuai design
    - Worker goroutine mengambil job dari channel, memanggil AI API eksternal
    - Enqueue job setelah Phrase berhasil disimpan ke DB
    - _Requirements: 3.1, 3.6_

  - [ ] 7.2 Implementasi AI API call dengan retry eksponensial
    - Timeout 10 detik per request
    - Retry maksimal 3 kali dengan jeda eksponensial (1s, 2s, 4s)
    - Jika semua retry gagal → set status `ai_failed` + log error (phrase_id, error_msg)
    - _Requirements: 3.4, 3.5_

  - [ ]* 7.3 Write property test untuk AI Pipeline retry
    - **Property 8: AI Pipeline retry eksponensial dan transisi ke ai_failed**
    - **Validates: Requirements 3.4, 3.5**

  - [ ] 7.4 Implementasi parsing respons AI dan penyimpanan data linguistik
    - Parse JSON `{ "words": [...], "tone": string }` dari AI API
    - Simpan words ke tabel `words`, buat relasi di `phrase_words`, simpan `tone` ke `phrases`
    - Semua dalam satu transaksi DB
    - _Requirements: 3.2, 3.3_

  - [ ]* 7.5 Write property test untuk AI data persistence
    - **Property 7: AI Pipeline menyimpan semua data linguistik dari respons AI**
    - **Validates: Requirements 3.2**

- [ ] 8. Implementasi Audio Upload dan Storage
  - [ ] 8.1 Implementasi `AudioStorageService` interface
    - Implementasi `Upload`, `GenerateSignedURL`, `Delete` untuk S3-compatible storage (Cloudflare R2 / Supabase Storage)
    - Object key berbasis UUID, signed URL expiry minimal 1 jam
    - _Requirements: 18.1, 18.2, 18.5_

  - [ ] 8.2 Integrasi audio upload ke endpoint `POST /api/v1/phrases`
    - Terima file audio opsional (WAV/MP3) via multipart form
    - Validasi format, ukuran (≤5MB), dan durasi (≤30 detik) sebelum upload ke S3
    - Simpan `audio_url` dan `audio_duration_seconds` ke tabel `phrases`
    - Kembalikan HTTP 422 jika validasi gagal; tidak simpan ke S3 jika validasi gagal
    - _Requirements: 17.1, 17.2, 17.3, 17.4, 17.5, 17.6_

  - [ ]* 8.3 Write property test untuk validasi audio
    - **Property 19: Validasi audio menolak semua file tidak valid**
    - **Validates: Requirements 17.1, 17.4, 17.5**

  - [ ] 8.4 Implementasi endpoint untuk aksara PNG (Handwriting Input)
    - Terima file PNG opsional via multipart form bersama data Phrase
    - Upload ke S3, simpan URL ke kolom `native_script_image_url` di tabel `phrases`
    - _Requirements: 22.4, 22.5_

  - [ ] 8.5 Implementasi penghapusan audio dan PNG saat Phrase dihapus
    - Panggil `AudioStorageService.Delete` saat Admin menghapus Phrase
    - _Requirements: 18.4, 24.6_

- [ ] 9. Implementasi Learning Content — Flashcard dan Conversation Scenario
  - [ ] 9.1 Implementasi endpoint `GET /api/v1/flashcards`
    - Hanya kembalikan Phrase berstatus `approved`
    - Filter berdasarkan `language_code` dan `tone`
    - Setiap Flashcard berisi: `text_latin`, `translation`, daftar words (root + POS), `tone`, `cultural_context`
    - Sertakan `text_native_script` jika `script_status = approved`; sertakan `audio_url` jika `audio_status = audio_approved`
    - Urutan acak per sesi, cursor-based pagination (maks 20 item)
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 18.3, 19.6, 21.6_

  - [ ]* 9.2 Write property test untuk Flashcard
    - **Property 13: Hanya konten approved yang disajikan kepada Learner**
    - **Validates: Requirements 6.1, 7.1**
    - **Property 14: Flashcard mengandung semua field yang diperlukan**
    - **Validates: Requirements 6.2, 18.3, 19.6, 21.6**
    - **Property 15: Filter Flashcard menghasilkan hasil yang sesuai filter**
    - **Validates: Requirements 6.3**
    - **Property 16: Pagination tidak pernah melebihi limit**
    - **Validates: Requirements 6.5, 8.4**

  - [ ] 9.3 Implementasi endpoint `GET /api/v1/conversation-scenarios`
    - Hanya gunakan Phrase berstatus `approved`
    - Kembalikan 3–8 Phrase dengan Cultural_Context yang sama atau kompatibel
    - Sertakan metadata: `language_code`, `usage_situation`, jumlah total Phrase
    - Kembalikan HTTP 404 jika approved Phrase < 3 untuk bahasa tersebut
    - _Requirements: 7.1, 7.2, 7.3, 7.4_

  - [ ]* 9.4 Write property test untuk Conversation Scenario
    - **Property 17: Conversation Scenario mengandung 3-8 Phrase dengan Cultural Context kompatibel**
    - **Validates: Requirements 7.2, 7.3**

  - [ ] 9.5 Implementasi endpoint `POST /api/v1/phrase-practice`
    - Simpan hasil practice (tahu/belum_tahu) per Learner per Phrase ke tabel `phrase_practice_results`
    - _Requirements: 14.5, 14.6_

- [ ] 10. Implementasi Search
  - [ ] 10.1 Implementasi endpoint `GET /api/v1/search`
    - Cari Phrase berstatus `approved` berdasarkan teks pada `text_latin` atau `translation`
    - Dukung pencarian berdasarkan `root_form_latin` di tabel `words`
    - Offset-based pagination (maks 50 item)
    - Gunakan index `idx_words_root_form` dan `idx_phrases_language_status` untuk performa
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [ ]* 10.2 Write property test untuk Search
    - **Property 18: Search mengembalikan hanya approved Phrase yang mengandung query**
    - **Validates: Requirements 8.1, 8.3**

- [ ] 11. Implementasi Admin endpoints dan Web Dashboard embed
  - [ ] 11.1 Implementasi Admin endpoints di Go backend
    - `GET /api/v1/admin/phrases/flagged` — list flagged phrases dengan detail
    - `PATCH /api/v1/admin/phrases/:id/status` — approve/reject flagged phrase, catat `moderated_by` + `moderated_at`
    - `GET /api/v1/admin/users` — list users dengan search by name/email
    - `PATCH /api/v1/admin/users/:id/ban` — ban user + reject semua pending phrases miliknya
    - `PATCH /api/v1/admin/users/:id/role` — assign admin role, catat admin yang melakukan
    - `DELETE /api/v1/admin/phrases/:id` — hard delete phrase + cascade (votes, flags, audio di S3)
    - Semua endpoint dilindungi role `admin` (HTTP 403 jika bukan admin)
    - _Requirements: 24.1, 24.2, 24.3, 24.4, 24.5, 24.6, 24.7, 24.8, 24.9_

  - [ ] 11.2 Setup React Web Dashboard project
    - Inisialisasi React project di direktori `web/` dengan Vite atau Create React App
    - Setup routing: halaman Login, Moderasi (flagged phrases), Manajemen Pengguna, Manajemen Bahasa
    - Setup HTTP client dengan JWT header injection
    - _Requirements: 24.1, 24.4, 24.7_

  - [ ] 11.3 Implementasi halaman Moderasi di React
    - Tampilkan daftar Phrase berstatus `flagged` dengan detail, jumlah flag, alasan, info Contributor
    - Tombol "Approve" dan "Reject" yang memanggil `PATCH /api/v1/admin/phrases/:id/status`
    - _Requirements: 24.1, 24.2, 24.3_

  - [ ] 11.4 Implementasi halaman Manajemen Pengguna dan Bahasa di React
    - Halaman manajemen pengguna: list, search, tombol Ban User dan Assign Admin
    - Halaman manajemen bahasa: list semua bahasa, toggle active/inactive, tambah bahasa baru
    - _Requirements: 24.4, 24.5, 24.7, 24.8, 9.1, 9.2, 9.3_

  - [ ] 11.5 Embed React build ke Go binary
    - Build React (`npm run build`) menghasilkan static files di `web/dist/`
    - Gunakan `//go:embed web/dist` di Go untuk embed static files
    - Serve static files dari Go binary untuk semua route non-API
    - _Requirements: (arsitektur: single binary)_

- [ ] 12. Checkpoint — Pastikan semua test backend pass dan web dashboard berfungsi
  - Pastikan semua test pass, tanyakan kepada user jika ada pertanyaan.

- [ ] 13. Implementasi Flutter Mobile App — Foundation
  - [ ] 13.1 Setup Flutter project dan dependency
    - Inisialisasi Flutter project, setup `http` package, `flutter_secure_storage`, `shared_preferences`, `just_audio`, `record`
    - Setup folder struktur: `lib/screens`, `lib/widgets`, `lib/services`, `lib/models`
    - Konfigurasi minimum Android API level 21
    - _Requirements: 16.1, 16.2_

  - [ ] 13.2 Implementasi AuthService dan secure token storage
    - `AuthService`: login, register, logout, upgrade role
    - Simpan JWT ke Android Keystore via `flutter_secure_storage`
    - Inject `Authorization: Bearer <token>` pada setiap request
    - Redirect ke login screen jika token expired atau HTTP 401
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 16.3_

  - [ ] 13.3 Implementasi layar Login dan Register
    - Form login dan register dengan validasi input
    - Tampilkan pesan error sesuai respons API (HTTP 401, 409)
    - Fitur "ingat saya" (token persisten selama valid)
    - _Requirements: 10.1, 10.6_

- [ ] 14. Implementasi Flutter Mobile App — Contributor Features
  - [ ] 14.1 Implementasi layar Submit Phrase
    - Form: `text_latin`, `translation`, `language_code` (dropdown dari API), field opsional `text_native_script` + `script_type`
    - Kirim ke `POST /api/v1/phrases`, tampilkan konfirmasi + ID Phrase
    - Tampilkan error jika offline
    - _Requirements: 11.1, 11.2, 11.3, 21.1_

  - [ ] 14.2 Implementasi audio recording di layar Submit Phrase
    - Tombol rekam mengaktifkan mikrofon, tampilkan indikator level suara real-time dan durasi
    - Batasi durasi maks 30 detik, hentikan otomatis
    - Upload audio bersamaan dengan teks Phrase dalam satu multipart request
    - Minta izin mikrofon sebelum pertama kali; tampilkan pesan jika ditolak
    - _Requirements: 20.4, 20.5, 20.6, 17.1_

  - [ ] 14.3 Implementasi Canvas untuk Handwriting Input aksara
    - Canvas minimal 300×300 piksel, input sentuh/stylus, stroke real-time (3px hitam di atas putih)
    - Tombol Undo (hapus stroke terakhir) dan Clear (hapus semua)
    - Export Canvas sebagai PNG, upload bersama data Phrase
    - _Requirements: 22.1, 22.2, 22.3, 22.4_

  - [ ] 14.4 Implementasi layar daftar Phrase milik Contributor dan voting
    - Tampilkan daftar Phrase milik contributor (status, upvote, downvote, flag count)
    - Tampilkan daftar Phrase pending milik pengguna lain untuk voting
    - Tombol upvote/downvote; update count langsung tanpa reload
    - Sembunyikan tombol vote pada Phrase milik contributor sendiri
    - Tampilkan pesan jika HTTP 409 (sudah vote) atau HTTP 403 (self-vote)
    - _Requirements: 11.4, 11.5, 12.1, 12.2, 12.3, 12.4, 12.5_

  - [ ] 14.5 Implementasi tombol "Become a Contributor" di profil Learner
    - Tampilkan CTA di halaman profil/pengaturan untuk user dengan role `learner`
    - Kirim request upgrade, perbarui sesi dengan JWT baru
    - Tampilkan konfirmasi + penjelasan fitur baru
    - _Requirements: 23.1, 23.2, 23.4, 23.5_

- [ ] 15. Implementasi Flutter Mobile App — Learner Features
  - [ ] 15.1 Implementasi layar Flashcard
    - Pilih bahasa daerah dan filter tone sebelum mulai sesi
    - Tampilkan teks Phrase di sisi depan, terjemahan + metadata (words, tone, cultural_context) di sisi belakang
    - Gesture swipe kiri/kanan untuk navigasi antar Flashcard
    - Tampilkan ringkasan sesi setelah semua Flashcard dilihat
    - Tampilkan tombol putar audio jika `audio_status = audio_approved`
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 20.1, 20.2_

  - [ ] 15.2 Implementasi layar Phrase Practice
    - Tampilkan teks Phrase, tombol "Lihat Jawaban" untuk reveal terjemahan + metadata
    - Tombol self-assessment 👍 (tahu) dan 👎 (belum tahu)
    - Kirim hasil ke `POST /api/v1/phrase-practice`
    - Tampilkan tombol putar audio jika `audio_status = audio_approved`
    - Tampilkan ringkasan sesi setelah semua soal selesai
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.7, 14.8_

  - [ ] 15.3 Implementasi Offline Cache
    - Simpan Flashcard ke `shared_preferences` / local storage setelah berhasil fetch
    - Maks 200 Flashcard per bahasa daerah
    - Sajikan dari cache jika offline; tampilkan pesan jika cache kosong dan offline
    - Cache audio dengan batas 100MB per bahasa
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 20.3_

- [ ] 16. Implementasi Flutter Mobile App — Admin Emergency Actions
  - Implementasi akses terbatas Admin di Mobile_App: ban user dan hapus konten instan
  - Tampilkan menu admin hanya jika role = `admin`
  - Panggil endpoint `PATCH /api/v1/admin/users/:id/ban` dan `DELETE /api/v1/admin/phrases/:id`
  - _Requirements: 24.10_

- [ ] 17. Implementasi error handling dan middleware final
  - [ ] 17.1 Implementasi CORS, RateLimit, dan Logger middleware di Go
    - CORS untuk request dari web dashboard dan mobile
    - Per-IP rate limiting
    - Request/response logger
    - _Requirements: (arsitektur middleware stack)_

  - [ ] 17.2 Implementasi error response format konsisten di semua endpoint
    - Format: `{ "error": { "code": string, "message": string, "details": {} } }`
    - HTTP 5xx → log detail, kembalikan pesan generik ke client
    - _Requirements: 16.6_

  - [ ] 17.3 Implementasi URL versioning `/api/v1/` di semua endpoint
    - Pastikan semua route menggunakan prefix `/api/v1/`
    - _Requirements: 16.5_

- [ ] 18. Checkpoint Final — Pastikan semua test pass dan integrasi end-to-end berfungsi
  - Jalankan semua unit test, property test, dan integration test
  - Verifikasi Go binary dapat di-build dengan React dashboard ter-embed
  - Pastikan semua test pass, tanyakan kepada user jika ada pertanyaan.

## Notes

- Task bertanda `*` bersifat opsional dan dapat dilewati untuk MVP yang lebih cepat
- Setiap task mereferensikan requirements spesifik untuk traceability
- Property tests menggunakan library `gopter` (Go) sesuai design document
- Checkpoint memastikan validasi inkremental di setiap fase
- Web Dashboard di-embed ke Go binary menggunakan `//go:embed` — build React dulu sebelum build Go
- Deployment: Railway untuk Go binary + PostgreSQL; Cloudflare R2 atau Supabase Storage untuk file audio dan PNG aksara
