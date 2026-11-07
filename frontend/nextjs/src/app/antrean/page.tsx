"use client";

import React from "react";
import { requireAuth, authFetch } from "@/lib/auth";

type Polyclinic = { id: string; name: string };
type Patient = { id: string; no_rm: string; full_name: string };
type Doctor = { id: string; name: string; poli: string };
type Encounter = {
    id: string;
    patient_id: string;
    doctor_id: string;
    polyclinic: string;
    queue_number: number;
    status: string;
    visit_date: string;
};

const fmtDate = (iso: string) => {
    if (!iso) return "—";
    const d = new Date(iso);
    if (isNaN(d.getTime())) return "—";
    return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
};

const STATUS_STYLE: Record<string, string> = {
    waiting: "bg-warning-soft text-warning",
    in_examination: "bg-primary-soft text-primary",
    completed: "bg-success-soft text-success",
    cancelled: "bg-error-soft text-error",
};

const STATUS_LABEL: Record<string, string> = {
    waiting: "Menunggu",
    in_examination: "Diperiksa",
    completed: "Selesai",
    cancelled: "Batal",
};

export default function AntreanPage() {
    const [polys, setPolys] = React.useState<Polyclinic[]>([]);
    const [poli, setPoli] = React.useState("");
    const [encounters, setEncounters] = React.useState<Encounter[]>([]);
    const [patients, setPatients] = React.useState<Patient[]>([]);
    const [doctors, setDoctors] = React.useState<Doctor[]>([]);
    const [allowed, setAllowed] = React.useState(false);

    React.useEffect(() => {
        if (!requireAuth()) return;
        setAllowed(true);
        authFetch("http://localhost:8080/api/v1/polyclinics")
            .then((r) => (r.ok ? r.json() : []))
            .then((d: Polyclinic[]) => {
                setPolys(d);
                const def = d.find((p) => p.id === "poli-umum") || d[0];
                if (def) setPoli(def.id);
            })
            .catch(() => setPolys([]));
        authFetch("http://localhost:8080/api/v1/patients")
            .then((r) => (r.ok ? r.json() : []))
            .then(setPatients)
            .catch(() => setPatients([]));
        authFetch("http://localhost:8080/api/v1/doctors")
            .then((r) => (r.ok ? r.json() : []))
            .then(setDoctors)
            .catch(() => setDoctors([]));
    }, []);

    const loadEncounters = async () => {
        if (!poli) return;
        const res = await authFetch(`http://localhost:8080/api/v1/encounters?polyclinic=${encodeURIComponent(poli)}`);
        if (res.ok) setEncounters(await res.json());
    };

    React.useEffect(() => {
        loadEncounters();
    }, [poli]);

    const updateStatus = async (id: string, status: string) => {
        const res = await authFetch(`http://localhost:8080/api/v1/encounters/${id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ status }),
        });
        if (res.ok) loadEncounters();
    };

    if (!allowed) return null;

    const patientLabel = (id: string) => {
        const p = patients.find((x) => x.id === id);
        return p ? `${p.full_name} (${p.no_rm})` : id;
    };
    const doctorLabel = (id: string) => doctors.find((d) => d.id === id)?.name || id || "—";

    return (
        <main className="mx-auto max-w-5xl px-4 py-8 md:px-6 animate-fade-up">
            <header className="mb-6">
                <h2 className="font-display text-2xl font-bold">Antrean Poli</h2>
                <p className="mt-1 text-sm text-secondary">
                    Daftar antrean pasien per poliklinik.
                </p>
            </header>

            <div className="mb-4 flex flex-wrap gap-2" role="tablist" aria-label="Pilih poliklinik">
                {polys.map((p) => (
                    <button
                        key={p.id}
                        role="tab"
                        aria-selected={poli === p.id}
                        onClick={() => setPoli(p.id)}
                        className={`rounded-xl px-4 py-2 text-sm font-medium transition-colors ${
                            poli === p.id
                                ? "bg-primary text-white"
                                : "border border-border bg-surface text-secondary hover:bg-slate-100"
                        }`}
                    >
                        {p.name}
                    </button>
                ))}
            </div>

            <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-card">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-border bg-slate-50 text-left text-xs uppercase tracking-wide text-secondary">
                                <th scope="col" className="px-4 py-3 font-semibold">No.</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Pasien</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Dokter</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Tanggal</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Status</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Aksi</th>
                            </tr>
                        </thead>
                        <tbody>
                            {encounters.length === 0 ? (
                                <tr>
                                    <td colSpan={6} className="px-4 py-10 text-center text-secondary">
                                        Belum ada antrean di poliklinik ini.
                                    </td>
                                </tr>
                            ) : (
                                encounters.map((e) => (
                                    <tr key={e.id} className="border-b border-border last:border-0 hover:bg-slate-50">
                                        <td className="px-4 py-3 font-display text-lg font-bold tnum">
                                            {e.queue_number}
                                        </td>
                                        <td className="px-4 py-3">{patientLabel(e.patient_id)}</td>
                                        <td className="px-4 py-3">{doctorLabel(e.doctor_id)}</td>
                                        <td className="px-4 py-3 tnum text-secondary">{fmtDate(e.visit_date)}</td>
                                        <td className="px-4 py-3">
                                            <span
                                                className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                                                    STATUS_STYLE[e.status] || "bg-slate-100 text-secondary"
                                                }`}
                                            >
                                                {STATUS_LABEL[e.status] || e.status}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3">
                                            <div className="flex gap-2">
                                                {e.status === "waiting" && (
                                                    <ActionButton onClick={() => updateStatus(e.id, "in_examination")}>
                                                        Panggil
                                                    </ActionButton>
                                                )}
                                                {e.status === "in_examination" && (
                                                    <ActionButton onClick={() => updateStatus(e.id, "completed")}>
                                                        Selesai
                                                    </ActionButton>
                                                )}
                                                {e.status !== "completed" && e.status !== "cancelled" && (
                                                    <ActionButton
                                                        variant="danger"
                                                        onClick={() => updateStatus(e.id, "cancelled")}
                                                    >
                                                        Batal
                                                    </ActionButton>
                                                )}
                                            </div>
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

function ActionButton({
    children,
    onClick,
    variant = "default",
}: {
    children: React.ReactNode;
    onClick: () => void;
    variant?: "default" | "danger";
}) {
    return (
        <button
            onClick={onClick}
            className={`rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
                variant === "danger"
                    ? "bg-error-soft text-error hover:bg-error hover:text-white"
                    : "bg-primary-soft text-primary hover:bg-primary hover:text-white"
            }`}
        >
            {children}
        </button>
    );
}
