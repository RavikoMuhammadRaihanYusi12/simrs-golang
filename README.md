# SIMRS — Sistem Informasi Manajemen Rumah Sakit

Go Clean Architecture backend + Next.js (React) frontend for a hospital management
information system (SIMRS), covering patient registration, queue management,
medical records, pharmacy/dispensing, billing, and BPJS integration.

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.27, Clean Architecture (Handler → Usecase → Repository) |
| Database | SQLite (in-memory, via modernc.org/sqlite) — PostgreSQL planned |
| Auth | JWT (golang-jwt v5), HS256, role-based access control |
| Frontend | Next.js 15 (App Router), React 18, TypeScript, Tailwind CSS 3 |

## Repository layout

```
backend/
  cmd/server/          # entrypoint, runs :8080
  internal/
    delivery/          # HTTP handlers, routes, guards
    usecase/           # business logic
    repository/        # SQLite persistence
    domain/            # entities
  pkg/jwt/             # JWT service + role allowlists

frontend/
  nextjs/              # Next.js app
    src/app/           # pages: / (pendaftaran), /login, /dashboard, /pasien, /antrean,
                       #    /dokter, /farmasi, /riwayat, /riwayat-obat, /rekam-medis, /profil
    src/components/    # AppShell, Sidebar, icons
    src/lib/           # auth client (token in localStorage, authFetch wrapper)

Planning documentation/
  simrs_product_requirements_document.md
  simrs_system_design_document.md
STATUS.md
```

## Run

### Backend

```bash
cd backend
go run ./cmd/server     # serves on :8080
```

### Frontend

```bash
cd frontend/nextjs
npm install
npm run dev             # serves on :3000
```

## Demo accounts

| Username | Password   | Role        |
|---|---|---|
| admin    | admin123   | admin       |
| dokter   | dokter123  | doctor      |
| loket    | loket123   | nurse       |
| kasir    | kasir123   | cashier     |
| apoteker | apoteker123| pharmacy    |

## API overview

- `POST /api/v1/auth/login` — login, returns `{ token }`
- `GET  /api/v1/patients` — list patients (nurse/admin, pharmacy read-only)
- `POST /api/v1/patients` — register patient (nurse/admin)
- `GET  /api/v1/encounters` — list encounters (optional `?polyclinic=` filter)
- `POST /api/v1/encounters` — create encounter (nurse/admin, doctor)
- `GET  /api/v1/encounters/:id` — get encounter
- `PATCH /api/v1/encounters/:id` — status transition waiting → in_examination → completed / cancelled
- `GET  /api/v1/doctors`, `POST /api/v1/doctors`, `PUT/DELETE /api/v1/doctors/:id` — doctor CRUD (admin)
- `GET  /api/v1/polyclinics` — list polyclinics
- `GET  /api/v1/pharmacy/medicines` — list medicines (pharmacy/admin)
- `POST /api/v1/pharmacy/medicines` — add medicine (pharmacy/admin)
- `PUT/DELETE /api/v1/pharmacy/medicines/:id` — update/remove medicine (pharmacy/admin)
- `GET  /api/v1/pharmacy/transactions` — stock transactions (pharmacy/admin)
- `POST /api/v1/pharmacy/transactions` — record stock take-out (pharmacy/admin)
- `POST /api/v1/pharmacy/prescriptions` — doctor creates prescription (doctor/pharmacy)
- `GET  /api/v1/pharmacy/prescriptions` — list prescriptions (no `encounter_id` = dispensing queue)
- `PATCH /api/v1/pharmacy/prescriptions/:id` — status transition (pharmacy/admin)

## Features

- **Auth / RBAC**: JWT with per-endpoint role guards; demo users for each role.
- **Registration & queue**: patient registration → encounter (poli + doctor) → queue number.
- **Antrean**: per-polyclinic queue list; nurse/admin/doctor actions (Panggil/Selesai/Batal).
- **Doctor management**: admin-only CRUD for doctors and their polyclinics.
- **Pharmacy**: medicine stock CRUD, stock take-out transactions, prescription workflow
  (`pending` → `processing` → `dispensed`) with automatic stock deduction on dispense.
- **Billing / BPJS / SATUSEHAT**: Phase 3 modules (billing, payment, BPJS SEP/V-claim scaffold,
  SATUSEHAT payload) implemented on top of encounters.
- **Dashboard & history**: operational dashboard, patient list (with latest poli + doctor),
  visit history (`/riwayat`), medicine stock history (`/riwayat-obat`).

## Status

See [STATUS.md](STATUS.md) for current phase status, implemented modules, and
remaining work (PostgreSQL migration, bcrypt, audit trail, Redis real-time queue, BPJS V-Claim SOAP keys).
