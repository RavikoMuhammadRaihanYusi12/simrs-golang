"use client";

import React from "react";
import Link from "next/link";
import { requireAuth, authFetch } from "@/lib/auth";

type Patient = {
    id: string;
    nik: string;
    no_rm: string;
    full_name: string;
    gender: string;
    phone: string;
    complaint: string;
};

type Encounter = {
    id: string;
    patient_id: string;
    doctor_id: string;
    polyclinic: string;
    queue_number: number;
    status: string;
    visit_date: string;
};

type Doctor = { id: string; name: string; poli: string };
type Polyclinic = { id: string; name: string };

const fmtDate = (iso: string) => {
    if (!iso) return "—";
    const d = new Date(iso);
    if (isNaN(d.getTime())) return "—";
    return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
};

export default function PasienPage() {
    const [patients, setPatients] = React.useState<Patient[]>([]);
    const [encounters, setEncounters] = React.useState<Encounter[]>([]);
    const [doctors, setDoctors] = React.useState<Doctor[]>([]);
    const [polys, setPolys] = React.useState<Polyclinic[]>([]);
    const [allowed, setAllowed] = React.useState(false);
    const [q, setQ] = React.useState("");
    const [loading, setLoading] = React.useState(true);

    const load = async () => {
        setLoading(true);
        try {
            const res = await authFetch("http://localhost:8080/api/v1/patients");
            if (res.ok) setPatients(await res.json());
        } finally {
            setLoading(false);
        }
    };

    // Encounters + master data enrich the patient rows with poli + doctor.
    const loadVisits = async () => {
        const [enc, doc, pol] = await Promise.all([
            authFetch("http://localhost:8080/api/v1/encounters"),
            authFetch("http://localhost:8080/api/v1/doctors"),
            authFetch("http://localhost:8080/api/v1/polyclinics"),
        ]);
        if (enc.ok) setEncounters(await enc.json());
        if (doc.ok) setDoctors(await doc.json());
        if (pol.ok) setPolys(await pol.json());
    };

    React.useEffect(() => {
        if (requireAuth()) {
            setAllowed(true);
            load();
            loadVisits();
        }
    }, []);

    if (!allowed) return null;

    // Most recent visit per patient (encounters come back ordered by queue number).
    const latestVisit = (patientId: string) => {
        let found: Encounter | undefined;
        for (const e of encounters) {
            if (e.patient_id === patientId) found = e;
        }
        return found;
    };
    const poliName = (id: string) => polys.find((p) => p.id === id)?.name || id;
    const doctorName = (id: string) => doctors.find((d) => d.id === id)?.name || id;

    const filtered = patients.filter((p) => {
        const s = q.toLowerCase();
        return (
            p.full_name.toLowerCase().includes(s) ||
            p.nik.includes(s) ||
            p.no_rm.toLowerCase().includes(s)
        );
    });

    return (
        <main className="mx-auto max-w-6xl px-4 py-8 md:px-6 animate-fade-up">
            <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
                <div>
                    <h2 className="font-display text-2xl font-bold">Daftar Pasien</h2>
                    <p className="mt-1 text-sm text-secondary">
                        {patients.length} pasien terdaftar
                    </p>
                </div>
                <Link
                    href="/"
                    className="rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-hover"
                >
                    + Pendaftaran Baru
                </Link>
            </div>

            <div className="mb-4">
                <input
                    type="search"
                    className="input max-w-sm"
                    placeholder="Cari nama, NIK, atau No. RM…"
                    value={q}
                    onChange={(e) => setQ(e.target.value)}
                    aria-label="Cari pasien"
                />
            </div>

            <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-card">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-border bg-slate-50 text-left text-xs uppercase tracking-wide text-secondary">
                                <th scope="col" className="px-4 py-3 font-semibold">No. RM</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Nama</th>
                                <th scope="col" className="px-4 py-3 font-semibold">NIK</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Gender</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Poliklinik</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Dokter</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Tanggal</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Keluhan</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr>
                                    <td colSpan={8} className="px-4 py-10 text-center text-secondary">
                                        Memuat data…
                                    </td>
                                </tr>
                            ) : filtered.length === 0 ? (
                                <tr>
                                    <td colSpan={8} className="px-4 py-10 text-center text-secondary">
                                        {q ? "Tidak ada pasien yang cocok." : "Belum ada pasien terdaftar."}
                                    </td>
                                </tr>
                            ) : (
                                filtered.map((p) => {
                                    const v = latestVisit(p.id);
                                    return (
                                    <tr
                                        key={p.id}
                                        className="border-b border-border last:border-0 transition-colors hover:bg-slate-50"
                                    >
                                        <td className="px-4 py-3 font-mono text-xs tnum">{p.no_rm}</td>
                                        <td className="px-4 py-3 font-medium">{p.full_name}</td>
                                        <td className="px-4 py-3 tnum">{p.nik}</td>
                                        <td className="px-4 py-3">
                                            <span
                                                className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                                                    p.gender === "P"
                                                        ? "bg-accent-soft text-accent"
                                                        : "bg-primary-soft text-primary"
                                                }`}
                                            >
                                                {p.gender === "P" ? "Perempuan" : "Laki-laki"}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3">
                                            {v ? (
                                                <span className="rounded-full bg-primary-soft px-2.5 py-0.5 text-xs font-medium text-primary">
                                                    {poliName(v.polyclinic)}
                                                </span>
                                            ) : (
                                                <span className="text-secondary">—</span>
                                            )}
                                        </td>
                                        <td className="px-4 py-3 text-secondary">
                                            {v?.doctor_id ? doctorName(v.doctor_id) : "—"}
                                        </td>
                                        <td className="px-4 py-3 tnum text-secondary">
                                            {v ? fmtDate(v.visit_date) : "—"}
                                        </td>
                                        <td className="max-w-[16rem] truncate px-4 py-3 text-secondary">
                                            {p.complaint || "—"}
                                        </td>
                                    </tr>
                                    );
                                })
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </main>
    );
}
