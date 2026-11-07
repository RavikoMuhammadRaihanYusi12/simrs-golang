"use client";

import React from "react";
import { requireAuth, authFetch } from "@/lib/auth";

type Medicine = { id: string; name: string; unit: string };
type Transaction = {
    id: string;
    medicine_id: string;
    quantity: number;
    type: string;
    note: string;
    created_at: string;
};

const fmtDateTime = (iso: string) => {
    if (!iso) return "—";
    const d = new Date(iso);
    if (isNaN(d.getTime())) return "—";
    return d.toLocaleString("id-ID", {
        day: "2-digit",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    });
};

export default function RiwayatObatPage() {
    const [allowed, setAllowed] = React.useState(false);
    const [transactions, setTransactions] = React.useState<Transaction[]>([]);
    const [medicines, setMedicines] = React.useState<Medicine[]>([]);
    const [typeFilter, setTypeFilter] = React.useState("");
    const [q, setQ] = React.useState("");

    React.useEffect(() => {
        if (!requireAuth()) return;
        setAllowed(true);
        authFetch("http://localhost:8080/api/v1/pharmacy/transactions")
            .then((r) => (r.ok ? r.json() : []))
            .then(setTransactions)
            .catch(() => setTransactions([]));
        authFetch("http://localhost:8080/api/v1/pharmacy/medicines")
            .then((r) => (r.ok ? r.json() : []))
            .then(setMedicines)
            .catch(() => setMedicines([]));
    }, []);

    if (!allowed) return null;

    const medName = (id: string) => medicines.find((m) => m.id === id)?.name || id;
    const medUnit = (id: string) => medicines.find((m) => m.id === id)?.unit || "";

    const rows = transactions
        .filter((t) => (typeFilter ? t.type === typeFilter : true))
        .filter((t) => {
            const s = q.toLowerCase();
            return !s || medName(t.medicine_id).toLowerCase().includes(s) || t.note.toLowerCase().includes(s);
        });

    return (
        <main className="mx-auto max-w-5xl px-4 py-8 md:px-6 animate-fade-up">
            <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
                <div>
                    <h2 className="font-display text-2xl font-bold">Riwayat Transaksi Obat</h2>
                    <p className="mt-1 text-sm text-secondary">
                        {rows.length} transaksi stok tercatat
                    </p>
                </div>
                <div className="flex flex-wrap gap-3">
                    <input
                        type="search"
                        className="input max-w-xs"
                        placeholder="Cari obat / catatan…"
                        value={q}
                        onChange={(e) => setQ(e.target.value)}
                        aria-label="Cari transaksi"
                    />
                    <select
                        className="input"
                        value={typeFilter}
                        onChange={(e) => setTypeFilter(e.target.value)}
                        aria-label="Filter jenis"
                    >
                        <option value="">Semua jenis</option>
                        <option value="in">Masuk</option>
                        <option value="out">Keluar</option>
                    </select>
                </div>
            </div>

            <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-card">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-border bg-slate-50 text-left text-xs uppercase tracking-wide text-secondary">
                                <th scope="col" className="px-4 py-3 font-semibold">Waktu</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Obat</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Jenis</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Jumlah</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Catatan</th>
                            </tr>
                        </thead>
                        <tbody>
                            {rows.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-4 py-10 text-center text-secondary">
                                        Belum ada transaksi obat.
                                    </td>
                                </tr>
                            ) : (
                                rows.map((t) => {
                                    const out = t.quantity < 0;
                                    return (
                                        <tr key={t.id} className="border-b border-border last:border-0 hover:bg-slate-50">
                                            <td className="px-4 py-3 tnum text-secondary">{fmtDateTime(t.created_at)}</td>
                                            <td className="px-4 py-3 font-medium">{medName(t.medicine_id)}</td>
                                            <td className="px-4 py-3">
                                                <span
                                                    className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                                                        out ? "bg-error-soft text-error" : "bg-success-soft text-success"
                                                    }`}
                                                >
                                                    {out ? "Keluar" : "Masuk"}
                                                </span>
                                            </td>
                                            <td className="px-4 py-3 tnum">
                                                <span className={out ? "text-error" : "text-success"}>
                                                    {out ? "" : "+"}
                                                    {t.quantity} {medUnit(t.medicine_id)}
                                                </span>
                                            </td>
                                            <td className="px-4 py-3 text-secondary">{t.note || "—"}</td>
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
