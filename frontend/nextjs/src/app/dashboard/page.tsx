"use client";

import React from "react";
import Link from "next/link";
import { requireAuth, authFetch, getRole } from "@/lib/auth";

const ROLE_LABEL: Record<string, string> = {
    admin: "Administrator",
    doctor: "Dokter",
    nurse: "Loket / Perawat",
    cashier: "Kasir",
    pharmacy: "Apoteker",
};

type Stat = { label: string; value: string; accent: string };

export default function DashboardPage() {
    const [role, setRole] = React.useState<string | null>(null);
    const [allowed, setAllowed] = React.useState(false);
    const [patientCount, setPatientCount] = React.useState<number | null>(null);

    React.useEffect(() => {
        if (!requireAuth()) return;
        setAllowed(true);
        setRole(getRole());
        authFetch("http://localhost:8080/api/v1/patients")
            .then((r) => (r.ok ? r.json() : []))
            .then((d) => setPatientCount(Array.isArray(d) ? d.length : 0))
            .catch(() => setPatientCount(0));
    }, []);

    if (!allowed) return null;

    const stats: Stat[] = [
        {
            label: "Total Pasien",
            value: patientCount === null ? "…" : String(patientCount),
            accent: "bg-primary-soft text-primary",
        },
        { label: "Antrean Aktif", value: "—", accent: "bg-accent-soft text-accent" },
        { label: "Kunjungan Hari Ini", value: "—", accent: "bg-warning-soft text-warning" },
        { label: "Pendapatan Hari Ini", value: "—", accent: "bg-success-soft text-success" },
    ];

    return (
        <main className="mx-auto max-w-6xl px-4 py-8 md:px-6 animate-fade-up">
            <header className="mb-6">
                <h2 className="font-display text-2xl font-bold">Dashboard</h2>
                <p className="mt-1 text-sm text-secondary">
                    Selamat datang,{" "}
                    <span className="font-medium text-foreground">
                        {ROLE_LABEL[role || ""] || role || "Pengguna"}
                    </span>
                    .
                </p>
            </header>

            <section className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                {stats.map((s) => (
                    <div
                        key={s.label}
                        className="rounded-2xl border border-border bg-surface p-5 shadow-card transition-shadow hover:shadow-card-hover"
                    >
                        <span
                            className={`inline-flex h-10 w-10 items-center justify-center rounded-xl text-lg font-bold ${s.accent}`}
                            aria-hidden="true"
                        >
                            {s.value === "…" ? "…" : "#"}
                        </span>
                        <p className="mt-3 text-sm text-secondary">{s.label}</p>
                        <p className="font-display text-2xl font-bold tnum">{s.value}</p>
                    </div>
                ))}
            </section>

            <section className="mt-6 grid gap-4 lg:grid-cols-2">
                <div className="rounded-2xl border border-border bg-surface p-6 shadow-card">
                    <h3 className="font-display font-semibold">Antrean Poli</h3>
                    <p className="mt-1 text-sm text-secondary">
                        Status antrean real-time per poliklinik.
                    </p>
                    <div className="mt-4 rounded-xl border border-dashed border-border bg-slate-50 p-6 text-center text-sm text-secondary">
                        Belum ada antrean aktif.
                    </div>
                </div>

                <div className="rounded-2xl border border-border bg-surface p-6 shadow-card">
                    <h3 className="font-display font-semibold">Aksi Cepat</h3>
                    <p className="mt-1 text-sm text-secondary">Tugas umum sesuai peran Anda.</p>
                    <div className="mt-4 flex flex-wrap gap-3">
                        <QuickLink href="/" label="Pendaftaran Pasien" />
                        <QuickLink href="/pasien" label="Cari Pasien" />
                    </div>
                </div>
            </section>
        </main>
    );
}

function QuickLink({ href, label }: { href: string; label: string }) {
    return (
        <Link
            href={href}
            className="rounded-xl border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-primary hover:bg-primary-soft hover:text-primary"
        >
            {label}
        </Link>
    );
}
