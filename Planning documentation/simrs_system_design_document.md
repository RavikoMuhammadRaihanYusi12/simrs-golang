# System Design Document: Sistem Informasi Manajemen Rumah Sakit (SIMRS)

## 1. High-Level Architecture

Sistem SIMRS ini dirancang menggunakan arsitektur **Modular Monolith** berbasis layanan backend Go yang bersih (*Clean Architecture*) dan frontend modern berbasis Next.js Server-Side Rendering (SSR). Pemisahan yang jelas antara lapisan presentation, business logic, dan data persistence memastikan kemudahan pemeliharaan (*maintainability*) dan skalabilitas di masa depan.

```
+-----------------------------------------------------------------+
|                       CLIENT LAYER                              |
|   Next.js (Web Frontend) + Tailwind CSS + TypeScript            |
|   (Runs on Client Browser / Vercel / Node Server)               |
+--------------------------------+--------------------------------+
                                 | HTTPS / REST / WebSockets
                                 v
+-----------------------------------------------------------------+
|                       API GATEWAY / REVERSE PROXY               |
|   Nginx / Traefik (Load Balancing & SSL Termination)            |
+--------------------------------+--------------------------------+
                                 |
                                 v
+-----------------------------------------------------------------+
|                       BACKEND CORE (Go / Golang)                |
|   - Clean Architecture (Handler -> UseCase -> Repository)       |
|   - JWT Authentication & RBAC Middleware                        |
|   - Background Workers (Async Tasks for BPJS / SATUSEHAT Sync)  |
+-----------------+------------------------------+----------------+
                  |                              |
                  v                              v
+---------------------------------+    +--------------------------+
|      POSTGRE_SQL DB             |    |          REDIS           |
|  - Relational Data / ACID       |    |  - Real-time Queue / PubSub|
|  - Medical Records & Billing    |    |  - Session Cache / Rate  |
+---------------------------------+    +--------------------------+
```

---

## 2. Backend Architecture (Go - Clean Architecture)

Backend dibangun menggunakan Go dengan struktur direktori Clean Architecture untuk memisahkan *concern*:

```text
/backend
├── cmd/
│   └── server/          # Entrypoint aplikasi (main.go)
├── internal/
│   ├── delivery/        # HTTP handlers, Router (Fiber / Gin / Chi)
│   ├── domain/          # Struct entity, interfaces (Business models)
│   ├── usecase/         # Business logic implementation
│   └── repository/      # Database queries (PostgreSQL & Redis integration)
├── pkg/                 # Shared packages (Logger, JWT, Validator)
└── go.mod
```

---

## 3. Database Schema Design (PostgreSQL)

Berikut adalah rancangan entitas utama relasional pada PostgreSQL:

### 3.1. Tabel `users` & `roles` (Autentikasi & RBAC)
*   `id` (UUID, PK)
*   `username` (VARCHAR, Unique)
*   `password_hash` (VARCHAR)
*   `role` (ENUM: `admin`, `doctor`, `nurse`, `cashier`, `pharmacy`, `patient`)
*   `created_at`, `updated_at`

### 3.2. Tabel `patients` (Data Pasien)
*   `id` (UUID, PK)
*   `nik` (VARCHAR(16), Unique) - *Nomor Induk Kependudukan*
*   `no_rm` (VARCHAR(20), Unique) - *Nomor Rekam Medis*
*   `full_name` (VARCHAR)
*   `birth_date` (DATE)
*   `gender` (VARCHAR)
*   `phone` (VARCHAR)
*   `address` (TEXT)

### 3.3. Tabel `encounters` (Kunjungan / Pendaftaran Poli)
*   `id` (UUID, PK)
*   `patient_id` (UUID, FK to `patients`)
*   `doctor_id` (UUID, FK to `users`)
*   `polyclinic` (VARCHAR) - *Contoh: Poli Anak, Poli Umum, IGD*
*   `queue_number` (INT)
*   `status` (ENUM: `waiting`, `in_examination`, `completed`, `cancelled`)
*   `visit_date` (TIMESTAMP)

### 3.4. Tabel `medical_records` (RME / SOAP)
*   `id` (UUID, PK)
*   `encounter_id` (UUID, FK to `encounters`)
*   `subjective` (TEXT)
*   `objective` (TEXT)
*   `assessment` (TEXT)
*   `plan` (TEXT)
*   `created_at` (TIMESTAMP)

### 3.5. Tabel `medicines` & `prescriptions` (Farmasi)
*   `medicines`: `id` (UUID, PK), `name` (VARCHAR), `stock` (INT), `unit` (VARCHAR), `price` (DECIMAL), `expired_date` (DATE)
*   `prescriptions`: `id` (UUID, PK), `encounter_id` (UUID, FK), `medicine_id` (UUID, FK), `quantity` (INT), `status` (ENUM: `pending`, `dispensed`)

#### Alur Bisnis Apoteker / Farmasi (RBAC: role `pharmacy`)
1.  **Akses Data Pasien:** Apoteker dapat membaca (`GET`) daftar pasien dan kunjungan (`/patients`, `/encounters`) — tidak boleh menulis pasien/kunjungan.
2.  **Resep Masuk:** Dokter membuat resep (`POST /pharmacy/prescriptions`) → status awal `pending`, stok obat otomatis berkurang sebesar `quantity` saat resep dibuat.
3.  **Antrean Resep:** Apoteker melihat seluruh resep via `GET /pharmacy/prescriptions` (tanpa `encounter_id` = semua resep pending).
4.  **Validasi & Dispensing:** Apoteker menyerahkan obat → `PATCH /pharmacy/prescriptions/:id` dengan `{"status":"dispensed"}` → status transisi `pending → dispensed`.
5.  **Manajemen Stok:** Apoteker kelola inventori obat via `/pharmacy/medicines` (`GET`/`POST`).
6.  **Evaluasi:** Farmasi tidak mengakses endpoint billing/BPJS (khusus `cashier`).

---

## 4. Redis Usage Strategy

Redis digunakan untuk mengoptimalkan performa pada tiga area kritikal:
1.  **Real-Time Queue (Antrean Poli):** Menggunakan Redis Sorted Sets (`ZSET`) atau Pub/Sub untuk menyiarkan nomor antrean aktif secara langsung ke layar ruang tunggu pasien tanpa membebani database PostgreSQL.
2.  **Session Management:** Menyimpan token sesi aktif atau *Blacklisted JWT* untuk logout yang aman.
3.  **Caching Data Master:** Menyimpan cache data master obat atau daftar poliklinik yang jarang berubah guna mempercepat waktu respons API.

---

## 5. Security & Integration Layer

*   **Authentication & Authorization:** Token berbasis JWT (JSON Web Token) dengan middleware role-based di Go untuk memastikan hanya *doctor* yang dapat mengakses modul rekam medis, dan *cashier* untuk modul billing.
*   **External API Integration:**
    *   **BPJS (V-Claim):** Menggunakan HTTP client Go dengan enkripsi HMAC-SHA256 untuk tanda tangan digital (*signature* header) ke server BPJS.
    *   **SATUSEHAT (Kemenkes):** Menggunakan OAuth2 token exchange untuk mengirim data FHIR (Fast Healthcare Interoperability Resources) secara asinkronus via background worker.