# Jobbin — Backend

Job application tracker. Built with Goravel (Go) + PostgreSQL.

## Tech Stack

- **Language:** Go 1.26
- **Framework:** [Goravel](https://goravel.dev) v1.17
- **Database:** PostgreSQL 16
- **Auth:** JWT
- **Email:** Resend

## Requirements

- Go 1.22+
- Docker + Docker Compose
- PostgreSQL 16 (via Docker)

## Quick Start

### 1. Clone repo

```bash
git clone https://github.com/ghanisiapfullstack/Jobbin.git
cd Jobbin
```

### 2. Setup environment

```bash
cp .env.example .env
```

Isi nilai berikut di `.env`:

| Variable | Keterangan |
|----------|------------|
| `APP_KEY` | Auto-generated, jalankan `go run . artisan key:generate` |
| `JWT_SECRET` | Auto-generated, jalankan `go run . artisan jwt:secret` |
| `DB_USERNAME` | Username PostgreSQL |
| `DB_PASSWORD` | Password PostgreSQL |
| `RESEND_API_KEY` | API key dari [resend.com](https://resend.com) |
| `FRONTEND_URL` | URL frontend, default `http://localhost:5173` |

### 3. Jalankan PostgreSQL via Docker

```bash
docker compose up -d postgres
```

> **Note:** PostgreSQL default di port `5433` karena port `5432` sudah dipakai PostgreSQL lokal.
> Kalau port `5432` kosong, ubah `docker-compose.yml` dan `.env` ke port `5432`.

### 4. Generate keys

```bash
go run . artisan key:generate
go run . artisan jwt:secret
```

### 5. Jalankan migration

```bash
go run . artisan migrate
```

### 6. Jalankan server

```bash
go run .
```

Server berjalan di `http://localhost:3000`

---

## API Endpoints

### Auth
| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| POST | `/api/v1/auth/register` | - | Registrasi user baru |
| POST | `/api/v1/auth/verify-email` | - | Verifikasi email |
| POST | `/api/v1/auth/resend-verification` | - | Kirim ulang email verifikasi |
| POST | `/api/v1/auth/login` | - | Login (rate limit: 5 req/menit) |
| GET | `/api/v1/auth/me` | ✅ | Data user yang login |
| POST | `/api/v1/auth/logout` | ✅ | Logout |

### Applications
| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/api/v1/applications` | ✅ | List lamaran (`?status=&archived=`) |
| POST | `/api/v1/applications` | ✅ | Tambah lamaran |
| GET | `/api/v1/applications/:id` | ✅ | Detail lamaran |
| PUT | `/api/v1/applications/:id` | ✅ | Update lamaran |
| PATCH | `/api/v1/applications/:id/position` | ✅ | Update posisi (drag & drop) |
| PATCH | `/api/v1/applications/:id/archive` | ✅ | Archive/restore lamaran |
| DELETE | `/api/v1/applications/:id` | ✅ | Hapus lamaran |

### Reminders
| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/api/v1/reminders` | ✅ | Reminder hari ini + besok |
| POST | `/api/v1/reminders/test` | ✅ | Test kirim email reminder |

### Profile
| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/api/v1/profile` | ✅ | Data profil |
| PUT | `/api/v1/profile` | ✅ | Update nama |
| PUT | `/api/v1/profile/password` | ✅ | Ganti password |

---

## Branch Strategy

```
main        → Production-ready code
develop     → Integration branch
feature/*   → Feature branches (contoh: feature/kanban-board)
fix/*       → Bug fix branches
```

## Project Structure

```
jobbin-backend/
├── app/
│   ├── http/
│   │   ├── controllers/    # Request handlers
│   │   └── middleware/     # JWT, rate limiter
│   ├── models/             # Database models
│   └── services/           # Business logic (email, reminder)
├── bootstrap/              # App + migration + schedule setup
├── config/                 # App configuration
├── database/
│   └── migrations/         # Database migrations
└── routes/                 # Route definitions
```

---

## Environment Variables

Lihat `.env.example` untuk lengkapnya.
