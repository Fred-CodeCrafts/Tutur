# Bahasa Daerah Learning Platform

Platform pembelajaran bahasa daerah dan dialek lokal Indonesia berbasis komunitas (UGC) dan AI.

## Structure

```
bahasa-daerah-platform/
├── be/          # Go backend (REST API + AI pipeline)
├── fe/
│   ├── web/     # React web dashboard (Admin)
│   └── mobile/  # Flutter mobile app (Android/iOS)
└── misc/        # Docker, docs, scripts
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
