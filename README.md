# Jobbin — Backend API

REST API untuk aplikasi Jobbin, dibangun dengan [Goravel](https://goravel.dev) (Go framework).

## Tech Stack

- **Language:** Go 1.24
- **Framework:** Goravel
- **Database:** PostgreSQL (Neon)
- **Auth:** JWT
- **Email:** Resend
- **Deploy:** Azure Container Apps + Docker

## Production

| Service | URL |
|---------|-----|
| API | https://api.jobbin.site/api/v1 |
| Frontend | https://www.jobbin.site |

## Setup Local

### Prerequisites

- Go 1.24+
- PostgreSQL (lokal atau Neon)
- [Goravel CLI](https://goravel.dev/getting-started/installation.html)

### 1. Clone repo

```bash
git clone https://github.com/ghanisiapfullstack/Jobbin.git
cd Jobbin
```

### 2. Copy env

```bash
cp .env.example .env
```

### 3. Isi `.env`

```env
APP_NAME=Jobbin
APP_ENV=local
APP_DEBUG=true
APP_HOST=127.0.0.1
APP_PORT=3000
APP_KEY=your-32-char-secret-key-here!!!

JWT_SECRET=your-32-char-jwt-secret-here!!!

DB_CONNECTION=postgres
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=jobbin
DB_USERNAME=jobbin
DB_PASSWORD=jobbin_secret
DB_SSLMODE=disable
DB_SCHEMA=public

RESEND_API_KEY=re_xxxxx
MAIL_FROM_ADDRESS=noreply@yourdomain.com
MAIL_FROM_NAME=Jobbin

FRONTEND_URL=http://localhost:5173
```

### 4. Install dependencies

```bash
go mod download
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

Base URL: `/api/v1`

### Auth

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| POST | `/auth/register` | ❌ | Registrasi user baru |
| POST | `/auth/login` | ❌ | Login, dapat JWT token |
| POST | `/auth/logout` | ✅ | Logout |
| GET | `/auth/me` | ✅ | Data user yang login |
| POST | `/auth/verify-email` | ❌ | Verifikasi email dengan token |
| POST | `/auth/resend-verification` | ❌ | Kirim ulang email verifikasi |

### Profile

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| GET | `/profile` | ✅ | Get profil user |
| PUT | `/profile` | ✅ | Update nama |
| PUT | `/profile/password` | ✅ | Ganti password |

### Applications

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| GET | `/applications` | ✅ | List lamaran (filter: status, archived) |
| POST | `/applications` | ✅ | Tambah lamaran baru |
| GET | `/applications/:id` | ✅ | Detail lamaran |
| PUT | `/applications/:id` | ✅ | Update lamaran |
| DELETE | `/applications/:id` | ✅ | Hapus lamaran |
| PATCH | `/applications/:id/position` | ✅ | Update posisi di board |
| PATCH | `/applications/:id/archive` | ✅ | Toggle archive/restore |

### Reminders

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| GET | `/reminders` | ✅ | List reminder hari ini + besok |
| POST | `/reminders/test` | ✅ | Trigger test reminder email |

---

## Database

### ERD

```
users
├── id                    (PK)
├── name                  (varchar 100)
├── email                 (varchar 255, unique)
├── password              (varchar 255, bcrypt)
├── email_verified_at     (timestamp, nullable)
├── email_verify_token    (varchar, nullable)
├── created_at
└── updated_at

applications
├── id                    (PK)
├── user_id               (FK → users.id)
├── job_title             (varchar 255)
├── company               (varchar 255)
├── url                   (varchar, nullable)
├── status                (enum: wishlist|applied|interview|offer|rejected)
├── notes                 (text, nullable)
├── applied_date          (date, nullable)
├── reminder_date         (date, nullable)
├── reminder_sent_day_before (boolean, default false)
├── reminder_sent_day_of  (boolean, default false)
├── position              (float, default 0)
├── is_archived           (boolean, default false)
├── created_at
└── updated_at
```

### Migration commands

```bash
# Jalankan semua migration
go run . artisan migrate

# Rollback migration terakhir
go run . artisan migrate:rollback

# Reset semua migration
go run . artisan migrate:reset
```

---

## Deploy ke Azure

### Prerequisites

- Docker Desktop
- Azure CLI (`az login`)
- Akses ke `jobbinregistry.azurecr.io`

### Steps

```bash
# 1. Login ke ACR
az acr login --name jobbinregistry

# 2. Build image
docker build -t jobbinregistry.azurecr.io/jobbin-backend:latest .

# 3. Push ke ACR
docker push jobbinregistry.azurecr.io/jobbin-backend:latest

# 4. Deploy ke Container Apps
az containerapp update \
  --name jobbin-backend \
  --resource-group jobbin-rg \
  --image jobbinregistry.azurecr.io/jobbin-backend:latest

# 5. Jalankan migration (kalau ada schema baru)
az containerapp exec \
  --name jobbin-backend \
  --resource-group jobbin-rg \
  --command "./main artisan migrate"
```

### Environment Variables Production

| Key | Keterangan |
|-----|------------|
| `APP_KEY` | Random 32 char string |
| `JWT_SECRET` | Random 32 char string |
| `DB_HOST` | Neon PostgreSQL host |
| `DB_DATABASE` | Nama database |
| `DB_USERNAME` | Username database |
| `DB_PASSWORD` | Password database |
| `DB_SSLMODE` | `require` untuk Neon |
| `RESEND_API_KEY` | API key dari resend.com |
| `MAIL_FROM_ADDRESS` | Email pengirim (domain verified) |
| `FRONTEND_URL` | URL frontend production |

---

## Project Structure

```
jobbin-backend/
├── app/
│   ├── http/
│   │   ├── controllers/     # Request handlers
│   │   └── middleware/      # JWT auth, rate limiter
│   ├── models/              # Database models
│   ├── services/            # Business logic (email, reminder)
│   └── facades/             # Goravel facade wrappers
├── config/                  # App, database, CORS, mail config
├── database/
│   └── migrations/          # Schema migrations
├── routes/                  # API route definitions
├── resources/               # Email templates
├── Dockerfile
├── render.yaml
└── .github/workflows/       # GitHub Actions CI/CD
```

---

## Security

- Password di-hash dengan bcrypt
- JWT dengan expiry 24 jam
- Rate limiting pada endpoint login (10 req/menit)
- HTTPS enforced di production
- SQL injection dicegah via ORM parameterized query
- Ownership check pada semua resource endpoint

---

## Known Issues & Backlog

Lihat [TASKS.md](../Jobbin/docs/TASKS.md) untuk backlog lengkap.

| ID | Issue | Priority |
|----|-------|----------|
| BL-11 | CORS wildcard — perlu restrict ke specific origins | Medium |
| BL-12 | XSS server-side sanitasi | Low |
| BL-13 | Cold start Azure Container Apps (min-replicas: 0) | Low |
