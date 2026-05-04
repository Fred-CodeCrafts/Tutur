# Bahasa Daerah Learning Platform

Platform pembelajaran bahasa daerah dan dialek lokal Indonesia berbasis komunitas (UGC) dan AI.

## Structure

```
bahasa-daerah-platform/
├── be/                           ← Go backend (REST API + AI pipeline)
│   ├── cmd/server/
│   ├── internal/
│   │   ├── domain/               ← shared models & constants
│   │   ├── auth/
│   │   ├── phrase/
│   │   ├── validation/
│   │   ├── flashcard/
│   │   ├── search/
│   │   ├── ai/
│   │   ├── admin/
│   │   └── storage/
│   └── pkg/
│       ├── db/                   ← pool, migrations (embedded SQL)
│       ├── middleware/
│       ├── response/
│       └── validator/
├── fe/
│   ├── web/src/                  ← React admin dashboard
│   │   ├── context/, store/, utils/, services/, hooks/, pages/, types/
│   └── mobile/lib/               ← Flutter app
│       ├── context/, store/, utils/, services/, models/, screens/, widgets/
└── misc/
    └── docker-compose.yml        ← PostgreSQL + MinIO
```

## Quick Start

```bash
# Start local services (PostgreSQL + MinIO)
cd misc && docker-compose up -d

# Run backend
cd be && go run ./cmd/server

# Run web dashboard
cd fe/web && npm install && npm run dev

# Run mobile app
cd fe/mobile && flutter run
```
