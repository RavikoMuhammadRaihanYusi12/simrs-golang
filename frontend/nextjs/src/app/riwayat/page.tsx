"use client";

import React from "react";
import { requireAuth, authFetch } from "@/lib/auth";

type Encounter = {
    id: string;
    patient_id: string;
    doctor_id: string;
    polyclinic: string;
    queue_number: number;
    status: string;
    visit_date: string;
};
type Patient = { id: string; no_rm: string; full_name: string };
type Polyclinic = { id: string; name: string };
type Doctor = { id: string; name: string };

const STATUS_LABEL: Record<string, string> = {
    waiting: "Menunggu",
    in_examination: "Diperiksa",
    completed: "Selesai",
    cancelled: "Batal",
};
const STATUS_STYLE: Record<string, string> = {
    waiting: "bg-warning-soft text-warning",
    in_examination: "bg-primary-soft text-primary",
    completed: "bg-success-soft text-success",
    cancelled: "bg-error-soft text-error",
};

const fmtDate = (iso: string) => {
    if (!iso) return "—";
    const d = new Date(iso);
    if (isNaN(d.getTime())) return "—";
    return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
};

export default function RiwayatPage() {
    const [allowed, setAllowed] = React.useState(false);
    const [encounters, setEncounters] = React.useState<Encounter[]>([]);
    const [patients, setPatients] = React.useState<Patient[]>([]);
    const [polys, setPolys] = React.useState<Polyclinic[]>([]);
    const [doctors, setDoctors] = React.useState<Doctor[]>([]);
    const [statusFilter, setStatusFilter] = React.useState("");
    const [q, setQ] = React.useState("");

    React.useEffect(() => {
        if (!requireAuth()) return;
        setAllowed(true);
        authFetch("http://localhost:8080/api/v1/encounters")
            .then((r) => (r.ok ? r.json() : []))
            .then(setEncounters)
            .catch(() => setEncounters([]));
        authFetch("http://localhost:8080/api/v1/patients")
            .then((r) => (r.ok ? r.json() : []))
            .then(setPatients)
            .catch(() => setPatients([]));
        authFetch("http://localhost:8080/api/v1/polyclinics")
            .then((r) => (r.ok ? r.json() : []))
            .then(setPolys)
            .catch(() => setPolys([]));
        authFetch("http://localhost:8080/api/v1/doctors")
            .then((r) => (r.ok ? r.json() : []))
            .then(setDoctors)
            .catch(() => setDoctors([]));
    }, []);

    if (!allowed) return null;

    const patientName = (id: string) => patients.find((p) => p.id === id)?.full_name || id;
    const poliName = (id: string) => polys.find((p) => p.id === id)?.name || id;
    const doctorName = (id: string) => doctors.find((d) => d.id === id)?.name || id || "—";

    const rows = encounters
        .filter((e) => (statusFilter ? e.status === statusFilter : true))
        .filter((e) => {
            const s = q.toLowerCase();
            return (
                !s ||
                patientName(e.patient_id).toLowerCase().includes(s) ||
                poliName(e.polyclinic).toLowerCase().includes(s) ||
                e.id.toLowerCase().includes(s)
            );
        })
        .sort((a, b) => (a.visit_date < b.visit_date ? 1 : -1));

    return (
        <main className="mx-auto max-w-6xl px-4 py-8 md:px-6 animate-fade-up">
            <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
                <div>
                    <h2 className="font-display text-2xl font-bold">Riwayat Kunjungan</h2>
                    <p className="mt-1 text-sm text-secondary">
                        {rows.length} kunjungan tercatat
                    </p>
                </div>
                <div className="flex flex-wrap gap-3">
                    <input
                        type="search"
                        className="input max-w-xs"
                        placeholder="Cari pasien / poli / ID…"
                        value={q}
                        onChange={(e) => setQ(e.target.value)}
                        aria-label="Cari riwayat"
                    />
                    <select
                        className="input"
                        value={statusFilter}
                        onChange={(e) => setStatusFilter(e.target.value)}
                        aria-label="Filter status"
                    >
                        <option value="">Semua status</option>
                        <option value="waiting">Menunggu</option>
                        <option value="in_examination">Diperiksa</option>
                        <option value="completed">Selesai</option>
                        <option value="cancelled">Batal</option>
                    </select>
                </div>
            </div>

            <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-card">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-border bg-slate-50 text-left text-xs uppercase tracking-wide text-secondary">
                                <th scope="col" className="px-4 py-3 font-semibold">Tanggal</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Pasien</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Poliklinik</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Dokter</th>
                                <th scope="col" className="px-4 py-3 font-semibold">No. Antrean</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Status</th>
                            </tr>
                        </thead>
                        <tbody>
                            {rows.length === 0 ? (
                                <tr>
                                    <td colSpan={6} className="px-4 py-10 text-center text-secondary">
                                        Belum ada riwayat kunjungan.
                                    </td>
                                </tr>
                            ) : (
                                rows.map((e) => (
                                    <tr key={e.id} className="border-b border-border last:border-0 hover:bg-slate-50">
                                        <td className="px-4 py-3 tnum text-secondary">{fmtDate(e.visit_date)}</td>
                                        <td className="px-4 py-3 font-medium">{patientName(e.patient_id)}</td>
                                        <td className="px-4 py-3">
                                            <span className="rounded-full bg-primary-soft px-2.5 py-0.5 text-xs font-medium text-primary">
                                                {poliName(e.polyclinic)}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3 text-secondary">{doctorName(e.doctor_id)}</td>
                                        <td className="px-4 py-3 font-display text-lg font-bold tnum">
                                            {e.queue_number}
                                        </td>
                                        <td className="px-4 py-3">
                                            <span
                                                className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                                                    STATUS_STYLE[e.status] || "bg-slate-100 text-secondary"
                                                }`}
                                            >
                                                {STATUS_LABEL[e.status] || e.status}
                                            </span>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </main>
    );
}
