# Product Requirements Document (PRD): Sistem Informasi Manajemen Rumah Sakit (SIMRS)

## 1. Overview & Project Goals
### 1.1. Executive Summary
Sistem Informasi Manajemen Rumah Sakit (SIMRS) adalah aplikasi berbasis web enterprise yang dirancang untuk mengintegrasikan dan mengotomatisasi seluruh alur kerja operasional, administratif, dan klinis di rumah sakit. Sistem ini bertujuan untuk meningkatkan efisiensi pelayanan pasien, akurasi pencatatan rekam medis, transparansi keuangan, serta mempermudah pelaporan wajib ke ekosistem kesehatan nasional (SATUSEHAT & BPJS Kesehatan).

### 1.2. Target Users
*   **Pasien / Keluarga:** Pendaftaran online, melihat jadwal dokter, melihat riwayat kunjungan.
*   **Petugas Pendaftaran (Front Office / Admisi):** Registrasi pasien baru/lama, manajemen antrean, penjadwalan poliklinik.
*   **Dokter:** Input rekam medis elektronik (RME), melihat hasil lab/radiologi, e-Prescription (resep elektronik), surat rujukan.
*   **Perawat:** Triage, pencatatan tanda-tanda vital (vital signs), asuhan keperawatan, administrasi tindakan.
*   **Kasir / Billing:** Pembayaran rawat jalan/inap, integrasi asuransi (BPJS/Asuransi Swasta), pencetakan kuitansi.
*   **Apotek / Farmasi:** Validasi resep dokter, manajemen stok obat, dispensing obat, laporan narkotika/psikotropika.
*   **Manajemen / Direksi:** Dashboard eksekutif, laporan keuangan, analisis BOR (Bed Occupancy Rate), statistik kunjungan rumah sakit.

---

## 2. Technical Stack Specifications
*   **Frontend (FE):** Next.js (React Framework) with TypeScript & Tailwind CSS for responsive, performant, and type-safe UI.
*   **Backend (BE):** Go (Golang) leveraging clean architecture for high-concurrency, low-latency microservices/monolith API services.
*   **Database (DB):** PostgreSQL (Primary relational database for ACID compliance regarding financial records and medical history) + Redis (In-memory cache for session management, queueing systems, and real-time counter).

---

## 3. Core Modules & Functional Requirements

### 3.1. Modul Pendaftaran & Antrean (Front Office)
*   **Registrasi Pasien:** Pencatatan data demografi pasien baru dengan validasi Nomor Induk Kependudukan (NIK) terintegrasi Dukcapil (opsional/mock).
*   **Manajemen Kunjungan:** Pendaftaran poli rawat jalan, IGD, atau rawat inap.
*   **Sistem Antrean Real-Time:** Menampilkan nomor antrean poli secara *live* menggunakan Redis Pub/Sub / WebSockets.

### 3.2. Modul Rekam Medis Elektronik (RME) / Electronic Medical Record (EMR)
*   ** SOAP Integration:** Form pencatatan medis berbasis Subjektif, Objektif, Asesmen, Plan.
*   **Riwayat Medis Pasien:** Timeline komprehensif kunjungan pasien sebelumnya.
*   **Integrasi SATUSEHAT:** Sinkronisasi otomatis data Encounter, Condition, Observation, dan Procedure ke platform Kemenkes RI via REST API.

### 3.3. Modul Farmasi & Stok Obat
*   **Manajemen Inventori Obat:** Pencatatan stok berdasarkan batch dan tanggal kadaluarsa (FIFO/FEFO).
*   **e-Prescription Dispensing:** Validasi resep digital dari dokter oleh apoteker, pengurangan stok otomatis, dan kalkulasi harga obat.

### 3.4. Modul Kasir, Billing & Kasir Pintar
*   **Billing Terpusat:** Penggabungan tagihan dari poli, tindakan perawat, laboratorium, dan farmasi.
*   **Integrasi BPJS (V-Claim):** Pengecekan eligibility peserta BPJS dan pembuatan SEP (Surat Eligibilitas Peserta).
*   **Laporan Keuangan:** Rekonsiliasi harian pendapatan kasir, piutang, dan klaim asuransi.

---

## 4. Non-Functional Requirements (NFR)
*   **Performance:** API response time $< 200\text{ms}$ untuk operasi umum; beban puncak hingga 500 concurrent users tanpa degradasi signifikan (didukung oleh Go concurrency).
*   **Security & Compliance:** 
    *   Enkripsi data sensitif (password dengan bcrypt, data rekam medis terenkripsi di database).
    *   Role-Based Access Control (RBAC) yang ketat untuk setiap divisi.
    *   *Audit Trail* untuk mencatat setiap aktivitas akses/perubahan data rekam medis.
*   **Availability:** Sistem dirancang dengan target uptime 99.9%.

---

## 5. Milestone & Release Phases
*   **Phase 1 (MVP):** Autentikasi, RBAC, Pendaftaran Pasien Dasar, Poli Rawat Jalan, dan Antrean Redis.
*   **Phase 2:** Rekam Medis Elektronik (RME) dan Integrasi Farmasi/Stok Obat.
*   **Phase 3:** Billing, Kasir, dan Integrasi BPJS / SATUSEHAT.
*   **Phase 4:** Dashboard Manajemen, Rawat Inap (Bed Management), dan Optimasi Performa Go.