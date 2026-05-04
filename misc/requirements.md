# Requirements Document

## Introduction

Platform pembelajaran bahasa daerah dan dialek lokal Indonesia berbasis komunitas (UGC) dan AI. Penutur asli berkontribusi kalimat sehari-hari beserta terjemahan bahasa Indonesianya. Backend Go memproses input tersebut melalui AI API untuk menghasilkan struktur linguistik terstandar (SVO, akar kata, part of speech, tone). Data terstruktur disimpan di PostgreSQL dengan skema relasional. Komunitas memvalidasi konten melalui sistem vote/flagging sebelum konten digunakan sebagai materi belajar. Pelajar mengakses flashcard dan skenario percakapan yang dihasilkan hanya dari konten terverifikasi.

## Glossary

- **Platform**: Sistem web/API bahasa daerah learning platform ini secara keseluruhan.
- **Mobile_App**: Aplikasi mobile berbasis Flutter yang berjalan di Android dan berinteraksi dengan Backend_API.
- **Web_Dashboard**: Antarmuka web yang digunakan oleh Admin untuk moderasi konten, manajemen bahasa daerah, dan manajemen pengguna.
- **Backend_API**: Komponen backend Go yang menyediakan REST API untuk Mobile_App dan Web_Dashboard.
- **Learner**: Pengguna terdaftar yang menggunakan Mobile_App untuk mempelajari bahasa daerah. Learner dapat mengakses flashcard, skenario percakapan, dan latihan menulis aksara, serta dapat melaporkan (flag) konten.
- **Contributor**: Pengguna terdaftar yang memiliki semua akses Learner ditambah kemampuan untuk mengirimkan Phrase, merekam audio, menggambar aksara di Canvas, dan memberikan vote (upvote/downvote) pada konten pengguna lain. Contributor menggunakan Mobile_App sebagai platform utama.
- **Admin**: Pengguna dengan hak akses penuh untuk moderasi konten, manajemen bahasa daerah, dan manajemen pengguna. Admin menggunakan Web_Dashboard sebagai platform utama, dengan akses terbatas ke Mobile_App untuk tindakan darurat seperti ban pengguna dan penghapusan konten instan.
- **Phrase**: Satu kalimat atau frasa bahasa daerah beserta terjemahan bahasa Indonesianya yang dikirimkan oleh Contributor.
- **Word**: Satuan kata yang diekstrak dari Phrase, beserta metadata linguistiknya (akar kata, part of speech).
- **Cultural_Context**: Informasi konteks budaya, register (formal/kasar/netral), dan situasi penggunaan yang terkait dengan sebuah Phrase.
- **AI_Pipeline**: Komponen backend Go yang mengirim teks mentah ke AI API dan menerima respons JSON terstruktur.
- **Validation_Engine**: Komponen backend yang mengelola status pending/approved/rejected sebuah Phrase berdasarkan vote komunitas.
- **Flashcard**: Unit materi belajar yang menampilkan satu Phrase terverifikasi beserta terjemahan dan metadata linguistiknya.
- **Conversation_Scenario**: Rangkaian Phrase terverifikasi yang disusun menjadi simulasi percakapan untuk Learner.
- **Vote**: Aksi upvote atau downvote yang diberikan Contributor terhadap sebuah Phrase berstatus pending.
- **Flag**: Laporan dari Learner atau Contributor bahwa sebuah Phrase mengandung konten tidak akurat atau tidak pantas.
- **Offline_Cache**: Penyimpanan lokal di Mobile_App yang menyimpan data Flashcard dan Conversation_Scenario untuk akses offline.
- **Audio_Recording**: File audio berformat WAV atau MP3 yang merekam pengucapan sebuah Phrase oleh Contributor.
- **Audio_Storage**: Komponen backend yang menyimpan dan menyajikan Audio_Recording menggunakan object storage S3-compatible.
- **Audio_Vote**: Aksi upvote atau downvote yang diberikan Contributor terhadap keakuratan Audio_Recording sebuah Phrase, terpisah dari Vote teks.
- **Native_Script**: Sistem aksara tradisional yang digunakan oleh suatu bahasa daerah, seperti aksara Jawa (Hanacaraka), aksara Sunda, aksara Bali, aksara Lontara (Bugis/Makassar), atau aksara Batak.
- **Script_Type**: Enum yang mengidentifikasi jenis aksara yang digunakan dalam representasi teks, dengan nilai yang valid: `latin`, `javanese`, `sundanese`, `balinese`, `lontara`, `batak`, dan `other`.
- **Handwriting_Input**: Fitur yang memungkinkan Contributor menggambar aksara di Canvas dan menyimpan hasilnya sebagai gambar PNG.
- **Canvas**: Area gambar interaktif di Mobile_App tempat pengguna dapat menggambar stroke aksara menggunakan input sentuh atau stylus.
- **OCR_Service**: Komponen backend yang menerima gambar aksara dan mengonversinya menjadi teks Unicode menggunakan layanan handwriting recognition eksternal.
- **Phrase_Practice**: Sesi latihan di mana Learner menebak arti Phrase bahasa daerah dan menilai sendiri pemahamannya.

---

## Requirements

### Requirement 1: Registrasi dan Autentikasi Pengguna

**User Story:** Sebagai pengguna baru, saya ingin mendaftar dan masuk ke Platform, agar saya dapat berkontribusi atau belajar bahasa daerah.

#### Acceptance Criteria

1. THE Platform SHALL menyediakan endpoint registrasi yang menerima nama, email, kata sandi, dan peran pengguna (`learner` atau `contributor`); peran `admin` hanya dapat ditetapkan oleh Admin yang sudah ada melalui Web_Dashboard.
2. WHEN pengguna mengirimkan data registrasi yang valid, THE Platform SHALL membuat akun baru dengan peran yang dipilih dan mengembalikan token autentikasi JWT.
3. IF email yang didaftarkan sudah terdaftar sebelumnya, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 409 dan pesan yang menjelaskan konflik.
4. WHEN pengguna mengirimkan kredensial login yang valid, THE Platform SHALL mengembalikan token JWT dengan masa berlaku 24 jam beserta peran pengguna yang aktif.
5. IF pengguna mengirimkan kredensial login yang tidak valid, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 401 tanpa mengungkap detail mana yang salah.
6. THE Platform SHALL menyimpan kata sandi pengguna menggunakan algoritma hashing bcrypt dengan cost factor minimal 12.

---

### Requirement 2: Pengiriman Frasa oleh Contributor

**User Story:** Sebagai Contributor, saya ingin mengirimkan kalimat bahasa daerah beserta terjemahannya, agar kalimat tersebut dapat diproses dan dipelajari orang lain.

#### Acceptance Criteria

1. WHEN Contributor yang terautentikasi mengirimkan sebuah Phrase beserta terjemahan bahasa Indonesia dan kode bahasa daerah yang valid, THE Platform SHALL menyimpan Phrase tersebut dengan status `pending`.
2. THE Platform SHALL mendukung kode bahasa daerah mengacu pada daftar bahasa yang aktif di tabel `languages`; kode bahasa tidak di-hardcode dan dapat bertambah kapan saja melalui Admin Web Dashboard.
3. IF Contributor mengirimkan Phrase tanpa terjemahan bahasa Indonesia, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 400.
4. IF Contributor mengirimkan Phrase dengan panjang teks melebihi 500 karakter, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 422.
5. WHEN sebuah Phrase berhasil disimpan, THE Platform SHALL mengembalikan ID unik Phrase tersebut beserta status `pending` kepada Contributor.

---

### Requirement 3: Pemrosesan AI Pipeline

**User Story:** Sebagai sistem, saya ingin memproses setiap Phrase yang dikirimkan melalui AI API, agar struktur linguistiknya dapat diekstrak secara otomatis.

#### Acceptance Criteria

1. WHEN sebuah Phrase baru disimpan dengan status `pending`, THE AI_Pipeline SHALL mengirimkan teks Phrase dan terjemahannya ke AI API dalam waktu 30 detik.
2. WHEN AI API mengembalikan respons sukses, THE AI_Pipeline SHALL mengurai respons JSON dan menyimpan data berikut ke database: daftar Word dengan akar kata dan part of speech, serta nilai tone (`formal`, `netral`, atau `kasar`).
3. THE AI_Pipeline SHALL mengirimkan prompt ke AI API dalam format yang mengharuskan AI API mengembalikan JSON dengan skema: `{ "words": [{"surface": string, "root": string, "pos": string}], "tone": string }`.
4. IF AI API mengembalikan respons error atau timeout setelah 10 detik, THEN THE AI_Pipeline SHALL mencoba ulang permintaan sebanyak maksimal 3 kali dengan jeda eksponensial.
5. IF semua percobaan ulang gagal, THEN THE AI_Pipeline SHALL menandai Phrase dengan status `ai_failed` dan mencatat log error yang mencakup ID Phrase dan pesan error dari AI API.
6. THE AI_Pipeline SHALL memproses antrian Phrase secara asinkron sehingga pengiriman Phrase oleh Contributor tidak terblokir oleh proses AI.

---

### Requirement 4: Penyimpanan Data Relasional

**User Story:** Sebagai sistem, saya ingin menyimpan data linguistik dalam skema relasional yang terstruktur, agar data dapat dikueri secara efisien untuk kebutuhan belajar.

#### Acceptance Criteria

1. THE Platform SHALL menyimpan data dalam tabel `words` yang berelasi many-to-many dengan tabel `phrases` melalui tabel junction `phrase_words`.
2. THE Platform SHALL menyimpan data dalam tabel `phrases` yang berelasi many-to-one dengan tabel `cultural_contexts`.
3. THE Platform SHALL menyimpan setiap Word dengan kolom: `id`, `surface_form_latin`, `surface_form_native_script`, `root_form_latin`, `root_form_native_script`, `script_type`, `part_of_speech`, `language_code`, `created_at`; kolom `surface_form_native_script`, `root_form_native_script`, dan `script_type` bersifat opsional (nullable).
4. THE Platform SHALL menyimpan setiap Phrase dengan kolom: `id`, `text_latin`, `text_native_script`, `script_type`, `translation`, `language_code`, `tone`, `status`, `script_status`, `contributor_id`, `cultural_context_id`, `audio_url`, `audio_duration_seconds`, `created_at`, `updated_at`; kolom `text_native_script` dan `script_type` bersifat opsional (nullable), dan `script_status` memiliki nilai default `none` dengan nilai valid: `none`, `pending`, `approved`, `rejected`.
5. THE Platform SHALL menyimpan setiap Cultural_Context dengan kolom: `id`, `language_code`, `region`, `usage_situation`, `created_at`.
6. THE Platform SHALL menerapkan foreign key constraint pada semua relasi antar tabel untuk menjaga integritas referensial.
7. THE Platform SHALL menerapkan constraint pada kolom `script_type` di tabel `phrases` dan `words` sehingga hanya menerima nilai dari enum Script_Type yang valid: `latin`, `javanese`, `sundanese`, `balinese`, `lontara`, `batak`, `other`.

---

### Requirement 5: Validasi Komunitas (Voting dan Flagging)

**User Story:** Sebagai Contributor, saya ingin memberikan vote pada Phrase yang dikirimkan pengguna lain, agar hanya konten yang akurat yang digunakan sebagai materi belajar.

#### Acceptance Criteria

1. WHEN Contributor yang terautentikasi mengirimkan upvote atau downvote pada sebuah Phrase berstatus `pending`, THE Validation_Engine SHALL mencatat vote tersebut dan memperbarui jumlah upvote atau downvote Phrase.
2. THE Validation_Engine SHALL mencegah satu Contributor memberikan lebih dari satu Vote pada Phrase yang sama; IF Contributor mencoba vote kedua kali pada Phrase yang sama, THEN THE Validation_Engine SHALL mengembalikan respons error dengan kode HTTP 409.
3. WHEN jumlah upvote sebuah Phrase mencapai minimal 3, THE Validation_Engine SHALL mengubah status Phrase menjadi `approved` dan menjadikannya visible kepada semua pengguna.
4. WHEN jumlah downvote sebuah Phrase mencapai 5 atau lebih, THE Validation_Engine SHALL mengubah status Phrase menjadi `rejected`.
5. WHEN Learner atau Contributor yang terautentikasi mengirimkan Flag pada sebuah Phrase, THE Validation_Engine SHALL mencatat Flag tersebut beserta alasan yang dipilih dari daftar: `inaccurate_translation`, `inappropriate_content`, `duplicate`.
6. WHEN jumlah Flag pada sebuah Phrase mencapai 3 atau lebih, THE Validation_Engine SHALL mengubah status Phrase menjadi `flagged` dan menyembunyikan Phrase dari tampilan publik hingga ditinjau oleh Admin.
7. THE Validation_Engine SHALL mencegah Contributor yang mengirimkan sebuah Phrase untuk memberikan Vote pada Phrase miliknya sendiri; IF hal tersebut terjadi, THEN THE Validation_Engine SHALL mengembalikan respons error dengan kode HTTP 403.

---

### Requirement 6: Penyajian Materi Belajar — Flashcard

**User Story:** Sebagai Learner, saya ingin mengakses flashcard dari Phrase yang telah diverifikasi, agar saya dapat belajar kosakata bahasa daerah secara terstruktur.

#### Acceptance Criteria

1. THE Platform SHALL hanya menyertakan Phrase berstatus `approved` dalam kumpulan Flashcard yang disajikan kepada Learner.
2. WHEN Learner meminta sesi Flashcard untuk kode bahasa daerah tertentu, THE Platform SHALL mengembalikan kumpulan Flashcard yang masing-masing berisi: teks Phrase, terjemahan bahasa Indonesia, daftar Word beserta root dan part of speech, nilai tone, dan informasi Cultural_Context.
3. THE Platform SHALL mendukung parameter kueri untuk memfilter Flashcard berdasarkan tone (`formal`, `netral`, `kasar`) dan kode bahasa daerah.
4. THE Platform SHALL mengembalikan Flashcard dalam urutan acak per sesi untuk setiap permintaan Learner.
5. THE Platform SHALL membatasi jumlah Flashcard per respons API maksimal 20 item, dengan dukungan pagination menggunakan cursor-based pagination.

---

### Requirement 7: Penyajian Materi Belajar — Conversation Scenario

**User Story:** Sebagai Learner, saya ingin mengakses skenario percakapan yang dibangun dari Phrase terverifikasi, agar saya dapat memahami penggunaan bahasa daerah dalam konteks nyata.

#### Acceptance Criteria

1. THE Platform SHALL hanya menggunakan Phrase berstatus `approved` sebagai komponen Conversation_Scenario.
2. WHEN Learner meminta Conversation_Scenario untuk kode bahasa daerah tertentu, THE Platform SHALL mengembalikan rangkaian 3 hingga 8 Phrase yang memiliki Cultural_Context yang sama atau kompatibel.
3. THE Platform SHALL menyertakan metadata berikut pada setiap Conversation_Scenario: kode bahasa daerah, nama situasi penggunaan dari Cultural_Context, dan jumlah total Phrase dalam skenario.
4. IF jumlah Phrase berstatus `approved` untuk kode bahasa daerah tertentu kurang dari 3, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 404 dan pesan yang menjelaskan bahwa konten belum tersedia untuk bahasa tersebut.

---

### Requirement 8: Pencarian Kata dan Frasa

**User Story:** Sebagai Learner atau Contributor, saya ingin mencari kata atau frasa dalam bahasa daerah tertentu, agar saya dapat menemukan konten yang relevan dengan cepat.

#### Acceptance Criteria

1. WHEN Learner atau Contributor yang terautentikasi mengirimkan kueri pencarian dengan teks dan kode bahasa daerah, THE Platform SHALL mengembalikan daftar Phrase berstatus `approved` yang mengandung teks kueri pada kolom `text` atau `translation`.
2. THE Platform SHALL mengembalikan hasil pencarian dalam waktu maksimal 500ms untuk kueri dengan panjang teks hingga 100 karakter pada database dengan hingga 100.000 Phrase.
3. THE Platform SHALL mendukung pencarian berdasarkan `root_form` Word, sehingga pencarian kata dasar mengembalikan semua Phrase yang mengandung Word dengan root tersebut.
4. THE Platform SHALL membatasi hasil pencarian maksimal 50 item per respons, dengan dukungan pagination menggunakan offset-based pagination.

---

### Requirement 9: Manajemen Bahasa Daerah

**User Story:** Sebagai Admin, saya ingin mengelola daftar bahasa daerah yang didukung melalui Web_Dashboard, agar Contributor hanya dapat mengirimkan konten untuk bahasa yang terdaftar.

#### Acceptance Criteria

1. THE Platform SHALL menyimpan daftar bahasa daerah yang didukung dalam tabel `languages` dengan kolom: `code`, `name`, `region`, `is_active`.
2. WHEN Admin mengaktifkan sebuah bahasa daerah melalui Web_Dashboard, THE Platform SHALL mengizinkan Contributor untuk mengirimkan Phrase dengan kode bahasa tersebut.
3. WHEN Admin menonaktifkan sebuah bahasa daerah melalui Web_Dashboard, THE Platform SHALL menghentikan penerimaan Phrase baru untuk bahasa tersebut, namun tetap menyajikan Phrase yang sudah `approved` kepada Learner.
4. IF Contributor mengirimkan Phrase dengan kode bahasa yang tidak terdaftar atau tidak aktif, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 400.

---

### Requirement 10: Mobile App — Autentikasi dan Sesi Pengguna

**User Story:** Sebagai pengguna Mobile_App, saya ingin masuk menggunakan akun yang sama dengan web, agar saya dapat mengakses Platform dari perangkat Android tanpa membuat akun baru.

#### Acceptance Criteria

1. THE Mobile_App SHALL menyediakan layar login yang mengirimkan kredensial ke endpoint autentikasi Backend_API yang sudah ada.
2. WHEN Backend_API mengembalikan token JWT yang valid, THE Mobile_App SHALL menyimpan token tersebut di secure storage perangkat (Android Keystore).
3. WHILE token JWT tersimpan dan belum kedaluwarsa, THE Mobile_App SHALL menyertakan token tersebut pada setiap request ke Backend_API melalui header `Authorization: Bearer <token>`.
4. WHEN token JWT kedaluwarsa, THE Mobile_App SHALL mengarahkan pengguna ke layar login dan menghapus token yang tersimpan dari secure storage.
5. IF Backend_API mengembalikan respons HTTP 401, THEN THE Mobile_App SHALL menghapus token yang tersimpan dan menampilkan pesan bahwa sesi telah berakhir.
6. THE Mobile_App SHALL mendukung fitur "ingat saya" sehingga pengguna tidak perlu login ulang selama token masih valid.

---

### Requirement 11: Mobile App — Kontribusi Frasa oleh Contributor

**User Story:** Sebagai Contributor, saya ingin mengirimkan Phrase baru melalui Mobile_App, agar saya dapat berkontribusi kapan saja dari perangkat Android.

#### Acceptance Criteria

1. WHEN Contributor yang terautentikasi mengisi form pengiriman Phrase di Mobile_App dan menekan tombol kirim, THE Mobile_App SHALL mengirimkan data Phrase ke endpoint pengiriman Backend_API yang sudah ada.
2. WHEN Backend_API mengembalikan respons sukses dengan ID Phrase, THE Mobile_App SHALL menampilkan konfirmasi kepada Contributor beserta ID Phrase yang baru dibuat.
3. IF perangkat tidak memiliki koneksi internet saat Contributor menekan tombol kirim, THEN THE Mobile_App SHALL menampilkan pesan error yang menginformasikan bahwa pengiriman Phrase memerlukan koneksi internet.
4. THE Mobile_App SHALL menampilkan daftar Phrase yang pernah dikirimkan oleh Contributor yang sedang login, diambil dari Backend_API, beserta status terkini setiap Phrase (`pending`, `approved`, `rejected`, `ai_failed`).
5. WHEN Contributor membuka detail Phrase miliknya di Mobile_App, THE Mobile_App SHALL menampilkan jumlah upvote, downvote, dan flag yang diterima Phrase tersebut.

---

### Requirement 12: Mobile App — Voting oleh Contributor

**User Story:** Sebagai Contributor, saya ingin memberikan vote pada Phrase milik pengguna lain melalui Mobile_App, agar saya dapat berpartisipasi dalam validasi komunitas dari perangkat Android.

#### Acceptance Criteria

1. WHEN Contributor yang terautentikasi membuka daftar Phrase berstatus `pending` di Mobile_App, THE Mobile_App SHALL mengambil dan menampilkan daftar tersebut dari Backend_API.
2. WHEN Contributor menekan tombol upvote atau downvote pada sebuah Phrase di Mobile_App, THE Mobile_App SHALL mengirimkan Vote ke endpoint voting Backend_API yang sudah ada.
3. WHEN Backend_API mengembalikan respons sukses setelah Vote dikirimkan, THE Mobile_App SHALL memperbarui tampilan jumlah vote pada Phrase tersebut secara langsung tanpa memuat ulang seluruh daftar.
4. IF Backend_API mengembalikan HTTP 409 karena Contributor sudah pernah vote pada Phrase yang sama, THEN THE Mobile_App SHALL menampilkan pesan bahwa Contributor sudah memberikan vote pada Phrase tersebut.
5. IF Backend_API mengembalikan HTTP 403 karena Contributor mencoba vote pada Phrase miliknya sendiri, THEN THE Mobile_App SHALL menyembunyikan tombol vote pada Phrase milik Contributor yang sedang login.

---

### Requirement 13: Mobile App — Belajar dengan Flashcard untuk Learner

**User Story:** Sebagai Learner, saya ingin mengakses sesi Flashcard melalui Mobile_App, agar saya dapat belajar bahasa daerah kapan saja dari perangkat Android.

#### Acceptance Criteria

1. WHEN Learner yang terautentikasi memilih bahasa daerah dan memulai sesi Flashcard di Mobile_App, THE Mobile_App SHALL mengambil kumpulan Flashcard dari endpoint Backend_API yang sudah ada.
2. THE Mobile_App SHALL menampilkan setiap Flashcard dengan tata letak yang menampilkan teks Phrase di sisi depan, dan terjemahan beserta metadata linguistik (daftar Word, tone, Cultural_Context) di sisi belakang.
3. WHEN Learner menggeser Flashcard ke kanan atau ke kiri, THE Mobile_App SHALL menampilkan Flashcard berikutnya dalam sesi.
4. WHEN Learner telah melihat semua Flashcard dalam sesi, THE Mobile_App SHALL menampilkan ringkasan sesi yang mencakup jumlah Flashcard yang telah dilihat.
5. THE Mobile_App SHALL mendukung filter Flashcard berdasarkan tone (`formal`, `netral`, `kasar`) sebelum sesi dimulai, dengan mengirimkan parameter filter ke Backend_API.

---

### Requirement 14: Mobile App — Phrase Practice untuk Learner

**User Story:** Sebagai Learner, saya ingin berlatih menebak arti kata/frasa bahasa daerah, agar saya dapat menguji pemahaman saya terhadap kosakata yang sudah dipelajari.

#### Acceptance Criteria

1. WHEN Learner memulai sesi Phrase Practice untuk kode bahasa daerah tertentu, THE Mobile_App SHALL mengambil kumpulan Phrase berstatus `approved` dari Backend_API dan menyajikannya sebagai soal latihan.
2. THE Mobile_App SHALL menampilkan teks Phrase bahasa daerah kepada Learner dan meminta Learner menebak terjemahan bahasa Indonesianya sebelum menampilkan jawaban.
3. WHEN Learner menekan tombol "Lihat Jawaban", THE Mobile_App SHALL menampilkan terjemahan bahasa Indonesia beserta metadata linguistik (daftar Word, tone, Cultural_Context) dari Phrase tersebut.
4. WHEN jawaban ditampilkan, THE Mobile_App SHALL menampilkan dua tombol self-assessment: 👍 (tahu) dan 👎 (belum tahu).
5. WHEN Learner menekan tombol self-assessment, THE Mobile_App SHALL mengirimkan hasil (tahu atau belum_tahu) beserta ID Phrase ke Backend_API untuk disimpan sebagai data tracking progress.
6. THE Platform SHALL menyimpan hasil Phrase Practice per Learner per Phrase di tabel `phrase_practice_results` dengan kolom: `id`, `learner_id`, `phrase_id`, `result` (tahu/belum_tahu), `created_at`.
7. THE Mobile_App SHALL menampilkan ringkasan sesi Phrase Practice setelah semua soal selesai, mencakup jumlah soal yang dijawab "tahu" dan "belum tahu".
8. IF Phrase yang ditampilkan memiliki Audio_Recording berstatus `audio_approved`, THE Mobile_App SHALL menampilkan tombol putar audio sehingga Learner dapat mendengar pengucapan setelah melihat jawaban.

---

### Requirement 15: Mobile App — Offline Capability

**User Story:** Sebagai Learner, saya ingin tetap dapat mengakses materi belajar meskipun tidak ada koneksi internet, agar proses belajar tidak terganggu oleh kondisi jaringan.

#### Acceptance Criteria

1. WHEN Learner berhasil mengambil kumpulan Flashcard dari Backend_API, THE Mobile_App SHALL menyimpan data Flashcard tersebut ke Offline_Cache di penyimpanan lokal perangkat.
2. WHILE perangkat tidak memiliki koneksi internet, THE Mobile_App SHALL menyajikan Flashcard dari Offline_Cache kepada Learner.
3. THE Offline_Cache SHALL menyimpan maksimal 200 Flashcard terakhir yang diakses per bahasa daerah untuk membatasi penggunaan penyimpanan lokal.
4. IF Offline_Cache kosong dan perangkat tidak memiliki koneksi internet, THEN THE Mobile_App SHALL menampilkan pesan yang menginformasikan bahwa koneksi internet diperlukan untuk memuat materi pertama kali.

---

### Requirement 16: Mobile App — Cross-Platform dan Kompatibilitas

**User Story:** Sebagai tim pengembang, saya ingin Mobile_App dibangun dengan Flutter, agar codebase yang sama dapat digunakan untuk mendukung Android sekarang dan iOS di masa depan.

#### Acceptance Criteria

1. THE Mobile_App SHALL dibangun menggunakan Flutter SDK sehingga satu codebase dapat dikompilasi menjadi aplikasi native untuk Android dan iOS.
2. THE Mobile_App SHALL mendukung Android API level 21 (Android 5.0 Lollipop) ke atas sebagai target minimum deployment.
3. THE Mobile_App SHALL berkomunikasi dengan Backend_API menggunakan protokol HTTPS dengan validasi sertifikat TLS untuk semua request.
4. THE Mobile_App SHALL menangani perbedaan ukuran layar Android dengan menggunakan layout responsif sehingga tampilan tetap konsisten pada layar berukuran 4 hingga 7 inci.
5. THE Backend_API SHALL menyediakan versi API yang stabil (menggunakan URL versioning, contoh: `/api/v1/`) sehingga Mobile_App dapat beroperasi tanpa perubahan saat Backend_API diperbarui secara backward-compatible.
6. IF Backend_API mengembalikan kode error HTTP 5xx, THEN THE Mobile_App SHALL menampilkan pesan error generik kepada pengguna dan mencatat detail error ke log lokal tanpa mengekspos detail teknis.

---

### Requirement 17: Rekam Audio Pengucapan saat Submit Phrase

**User Story:** Sebagai Contributor, saya ingin merekam audio pengucapan saat mengirimkan Phrase, agar pelajar dapat mendengar intonasi dan nada yang tidak bisa direpresentasikan oleh teks saja.

#### Acceptance Criteria

1. WHEN Contributor yang terautentikasi mengirimkan Phrase, THE Platform SHALL menerima satu file Audio_Recording opsional berformat WAV atau MP3 dengan ukuran maksimal 5MB dan durasi maksimal 30 detik.
2. WHEN Audio_Recording diterima bersama Phrase, THE Audio_Storage SHALL menyimpan file tersebut di object storage S3-compatible dan mengembalikan URL permanen yang terkait dengan Phrase tersebut.
3. THE Platform SHALL menyimpan referensi Audio_Recording pada tabel `phrases` dengan kolom tambahan `audio_url` dan `audio_duration_seconds`.
4. IF file yang diunggah bukan berformat WAV atau MP3, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 422 dan pesan yang menjelaskan format yang didukung.
5. IF ukuran file Audio_Recording melebihi 5MB atau durasi melebihi 30 detik, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 422 dan pesan yang menjelaskan batas yang diizinkan.
6. THE Platform SHALL mengizinkan Contributor mengirimkan Phrase tanpa Audio_Recording; keberadaan audio bersifat opsional.

---

### Requirement 18: Penyimpanan dan Penyajian Audio

**User Story:** Sebagai sistem, saya ingin menyimpan Audio_Recording secara aman dan menyajikannya dengan efisien, agar Learner dapat mengakses audio tanpa latensi tinggi.

#### Acceptance Criteria

1. THE Audio_Storage SHALL menyimpan setiap Audio_Recording dengan nama file unik berbasis UUID untuk menghindari konflik.
2. THE Backend_API SHALL menyajikan Audio_Recording melalui URL yang dapat diakses langsung oleh Mobile_App menggunakan protokol HTTPS.
3. WHEN Learner meminta Flashcard yang memiliki Audio_Recording, THE Platform SHALL menyertakan `audio_url` dalam respons Flashcard.
4. THE Audio_Storage SHALL mendukung penghapusan Audio_Recording ketika Phrase terkait dihapus oleh Admin, untuk menjaga konsistensi data.
5. THE Backend_API SHALL mengembalikan URL Audio_Recording dengan masa berlaku (signed URL) minimal 1 jam untuk mencegah akses tidak sah ke file audio.

---

### Requirement 19: Audio Validation — Voting Keakuratan Audio

**User Story:** Sebagai Contributor, saya ingin memberikan vote pada keakuratan Audio_Recording pengucapan Phrase milik Contributor lain, agar hanya audio yang benar-benar akurat yang digunakan sebagai referensi belajar.

#### Acceptance Criteria

1. WHEN Contributor yang terautentikasi mengirimkan Audio_Vote pada Audio_Recording sebuah Phrase, THE Validation_Engine SHALL mencatat Audio_Vote tersebut secara terpisah dari Vote teks Phrase.
2. THE Validation_Engine SHALL mencegah satu Contributor memberikan lebih dari satu Audio_Vote pada Audio_Recording Phrase yang sama; IF Contributor mencoba Audio_Vote kedua kali, THEN THE Validation_Engine SHALL mengembalikan respons error dengan kode HTTP 409.
3. THE Validation_Engine SHALL mencegah Contributor yang mengunggah Audio_Recording memberikan Audio_Vote pada audio miliknya sendiri; IF hal tersebut terjadi, THEN THE Validation_Engine SHALL mengembalikan respons error dengan kode HTTP 403.
4. WHEN jumlah upvote Audio_Vote sebuah Audio_Recording mencapai minimal 3, THE Validation_Engine SHALL menandai Audio_Recording tersebut dengan status `audio_approved`.
5. WHEN jumlah downvote Audio_Vote sebuah Audio_Recording mencapai 5 atau lebih, THE Validation_Engine SHALL menandai Audio_Recording tersebut dengan status `audio_rejected` dan menghentikan penyajian audio tersebut kepada Learner.
6. THE Platform SHALL hanya menyertakan Audio_Recording berstatus `audio_approved` dalam respons Flashcard yang dikirimkan kepada Learner.

---

### Requirement 20: Mobile App — Antarmuka Audio untuk Learner dan Contributor

**User Story:** Sebagai pengguna Mobile_App, saya ingin antarmuka yang intuitif untuk memutar, merekam, dan berinteraksi dengan audio, agar pengalaman belajar dan berkontribusi berbasis suara terasa alami.

#### Acceptance Criteria

1. WHEN Learner membuka Flashcard yang memiliki Audio_Recording berstatus `audio_approved`, THE Mobile_App SHALL menampilkan tombol putar audio yang memutar Audio_Recording menggunakan audio player bawaan perangkat.
2. WHEN Learner menekan tombol putar audio, THE Mobile_App SHALL mengunduh dan memutar Audio_Recording dari `audio_url` yang disediakan Backend_API; IF perangkat tidak memiliki koneksi internet, THEN THE Mobile_App SHALL menampilkan pesan bahwa audio memerlukan koneksi internet.
3. THE Mobile_App SHALL menyimpan Audio_Recording ke Offline_Cache bersama data Flashcard terkait, dengan batas penyimpanan audio maksimal 100MB per bahasa daerah.
4. WHEN Contributor membuka form pengiriman Phrase di Mobile_App, THE Mobile_App SHALL menyediakan tombol rekam yang mengaktifkan mikrofon perangkat, menampilkan indikator level suara secara real-time, dan menampilkan indikator durasi rekaman; THE Mobile_App SHALL membatasi durasi perekaman maksimal 30 detik dan menghentikan perekaman secara otomatis ketika batas tersebut tercapai.
5. WHEN Contributor selesai merekam dan menekan tombol kirim, THE Mobile_App SHALL mengunggah Audio_Recording ke Backend_API bersamaan dengan data teks Phrase dalam satu request multipart.
6. THE Mobile_App SHALL meminta izin akses mikrofon kepada pengguna sebelum pertama kali mengaktifkan fitur rekam; IF pengguna menolak izin, THEN THE Mobile_App SHALL menampilkan pesan yang menjelaskan bahwa izin mikrofon diperlukan untuk fitur rekam audio.

---

### Requirement 21: Multi-Script Support — Pengiriman Phrase dengan Aksara Asli

**User Story:** Sebagai Contributor penutur asli, saya ingin mengirimkan Phrase beserta representasi aksara aslinya (misalnya aksara Jawa atau Bali), agar pelajar dapat mempelajari tulisan tradisional bahasa daerah tersebut.

#### Acceptance Criteria

1. WHEN Contributor yang terautentikasi mengirimkan Phrase, THE Platform SHALL menerima field opsional `text_native_script` dan `script_type` di samping field `text_latin` yang wajib ada; Contributor memasukkan `text_native_script` sebagai teks biasa menggunakan keyboard perangkat.
2. IF Contributor mengirimkan `text_native_script` tanpa menyertakan `script_type`, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 400 dan pesan yang menjelaskan bahwa `script_type` wajib disertakan ketika `text_native_script` diberikan.
3. IF Contributor mengirimkan `script_type` dengan nilai di luar enum Script_Type yang valid, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 422 dan daftar nilai Script_Type yang diizinkan.
4. WHEN Phrase dengan `text_native_script` berhasil disimpan, THE Platform SHALL menyimpan `script_status` dengan nilai `pending` dan mengembalikan field `script_status` dalam respons kepada Contributor.
5. THE Platform SHALL memastikan kolom `text_latin` selalu terisi pada setiap Phrase; IF `text_latin` kosong atau tidak dikirimkan, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 400.
6. WHEN Learner meminta Flashcard untuk sebuah Phrase yang memiliki `text_native_script` dengan `script_status` bernilai `approved`, THE Platform SHALL menyertakan `text_native_script` dan `script_type` dalam respons Flashcard.
7. WHEN Contributor yang terautentikasi mengirimkan upvote atau downvote pada `text_native_script` sebuah Phrase berstatus `script_status: pending`, THE Validation_Engine SHALL mencatat vote tersebut secara terpisah dari Vote teks Latin dan Audio_Vote.
8. THE Validation_Engine SHALL mencegah satu Contributor memberikan lebih dari satu vote aksara pada Phrase yang sama; IF Contributor mencoba vote aksara kedua kali pada Phrase yang sama, THEN THE Validation_Engine SHALL mengembalikan respons error dengan kode HTTP 409.
9. THE Validation_Engine SHALL mencegah Contributor yang mengirimkan `text_native_script` sebuah Phrase memberikan vote aksara pada miliknya sendiri; IF hal tersebut terjadi, THEN THE Validation_Engine SHALL mengembalikan respons error dengan kode HTTP 403.
10. WHEN jumlah upvote vote aksara sebuah Phrase mencapai minimal 3, THE Validation_Engine SHALL mengubah `script_status` Phrase tersebut menjadi `approved`.
11. WHEN jumlah downvote vote aksara sebuah Phrase mencapai 5 atau lebih, THE Validation_Engine SHALL mengubah `script_status` Phrase tersebut menjadi `rejected` dan menghentikan penyajian `text_native_script` tersebut kepada Learner.

---

### Requirement 22: Handwriting Input — Canvas untuk Menulis Aksara

**User Story:** Sebagai Contributor, saya ingin menggambar aksara langsung di canvas menggunakan stylus atau jari, agar saya dapat mengirimkan representasi aksara asli tanpa memerlukan keyboard aksara khusus.

#### Acceptance Criteria

1. THE Mobile_App SHALL menyediakan Canvas dengan ukuran minimal 300×300 piksel yang dapat menerima input sentuh untuk menggambar aksara menggunakan stylus atau jari.
2. THE Mobile_App SHALL menampilkan stroke yang sedang digambar secara real-time di Canvas dengan ketebalan garis 3 piksel dan warna hitam di atas latar putih.
3. THE Mobile_App SHALL menyediakan tombol Undo yang menghapus stroke terakhir dan tombol Clear yang menghapus semua stroke di Canvas.
4. WHEN Contributor selesai menggambar dan menekan tombol "Simpan Aksara", THE Mobile_App SHALL mengekspor tampilan Canvas sebagai file gambar PNG dan mengirimkannya ke Backend_API bersamaan dengan data Phrase.
5. THE Backend_API SHALL menyimpan file gambar PNG aksara di object storage S3-compatible dan menyimpan URL-nya pada kolom `native_script_image_url` di tabel `phrases`.
6. WHEN Learner membuka Flashcard yang memiliki `native_script_image_url`, THE Mobile_App SHALL menampilkan gambar aksara tersebut di dalam Flashcard.

---

### Requirement 23: Self-Upgrade Learner ke Contributor

**User Story:** Sebagai Learner, saya ingin meningkatkan peran saya menjadi Contributor melalui tombol CTA di Mobile_App, agar saya dapat mulai berkontribusi konten tanpa proses persetujuan yang rumit.

#### Acceptance Criteria

1. THE Mobile_App SHALL menampilkan tombol CTA "Become a Contributor" yang terlihat jelas pada profil atau halaman pengaturan Learner yang sedang login.
2. WHEN Learner menekan tombol "Become a Contributor" dan mengkonfirmasi tindakan tersebut, THE Mobile_App SHALL mengirimkan permintaan upgrade peran ke Backend_API.
3. WHEN Backend_API menerima permintaan upgrade peran yang valid dari Learner yang terautentikasi, THE Platform SHALL mengubah peran pengguna tersebut dari `learner` menjadi `contributor` secara langsung tanpa memerlukan persetujuan pihak lain.
4. WHEN upgrade peran berhasil, THE Backend_API SHALL mengembalikan token JWT baru yang mencerminkan peran `contributor` dan THE Mobile_App SHALL memperbarui sesi pengguna dengan token baru tersebut.
5. WHEN upgrade peran berhasil, THE Mobile_App SHALL menampilkan konfirmasi kepada pengguna bahwa peran telah berubah menjadi Contributor beserta penjelasan singkat fitur baru yang dapat diakses.
6. IF pengguna yang mengirimkan permintaan upgrade sudah memiliki peran `contributor` atau `admin`, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 409.

---

### Requirement 24: Admin Web Dashboard — Moderasi dan Manajemen

**User Story:** Sebagai Admin, saya ingin mengelola konten dan pengguna melalui Web_Dashboard, agar Platform tetap aman, akurat, dan berkualitas tinggi.

#### Acceptance Criteria

1. THE Web_Dashboard SHALL menyediakan halaman moderasi yang menampilkan daftar Phrase berstatus `flagged` beserta detail Phrase, jumlah flag, alasan flag, dan informasi Contributor pengirim.
2. WHEN Admin membuka detail Phrase berstatus `flagged` di Web_Dashboard, THE Web_Dashboard SHALL menampilkan tombol "Approve" untuk mengubah status menjadi `approved` dan tombol "Reject" untuk mengubah status menjadi `rejected`.
3. WHEN Admin menekan tombol "Approve" atau "Reject" pada Phrase berstatus `flagged`, THE Platform SHALL mengubah status Phrase sesuai tindakan Admin dan mencatat ID Admin serta timestamp tindakan tersebut.
4. THE Web_Dashboard SHALL menyediakan halaman manajemen pengguna yang menampilkan daftar pengguna beserta peran, status akun, dan tanggal registrasi, dengan dukungan pencarian berdasarkan nama atau email.
5. WHEN Admin menekan tombol "Ban User" pada halaman manajemen pengguna di Web_Dashboard, THE Platform SHALL menonaktifkan akun pengguna tersebut sehingga pengguna tidak dapat login dan semua Phrase berstatus `pending` milik pengguna tersebut diubah menjadi `rejected`.
6. WHEN Admin menekan tombol "Hapus Konten" pada sebuah Phrase di Web_Dashboard, THE Platform SHALL menghapus Phrase tersebut beserta semua data terkait (Vote, Flag, Audio_Recording) secara permanen.
7. THE Web_Dashboard SHALL menyediakan halaman manajemen bahasa daerah yang menampilkan daftar semua bahasa beserta status aktif/nonaktif, dengan tombol toggle untuk mengaktifkan atau menonaktifkan setiap bahasa.
8. WHEN Admin menetapkan peran `admin` kepada pengguna lain melalui Web_Dashboard, THE Platform SHALL mengubah peran pengguna target menjadi `admin` dan mencatat ID Admin yang melakukan tindakan beserta timestamp.
9. THE Web_Dashboard SHALL membatasi akses seluruh fitur moderasi dan manajemen hanya kepada pengguna dengan peran `admin`; IF pengguna dengan peran lain mencoba mengakses endpoint Admin, THEN THE Platform SHALL mengembalikan respons error dengan kode HTTP 403.
10. WHERE Admin menggunakan Mobile_App untuk tindakan darurat, THE Mobile_App SHALL menyediakan akses terbatas kepada Admin untuk melakukan ban pengguna dan penghapusan konten instan tanpa memerlukan Web_Dashboard.
