"use client";

import React from "react";
import { requireAuth, authFetch, getRole } from "@/lib/auth";
import { IconClose } from "@/components/icons";

type Doctor = { id: string; name: string; poli: string };
type Polyclinic = { id: string; name: string };

export default function DokterPage() {
    const [doctors, setDoctors] = React.useState<Doctor[]>([]);
    const [polys, setPolys] = React.useState<Polyclinic[]>([]);
    const [allowed, setAllowed] = React.useState(false);
    const [isAdmin, setIsAdmin] = React.useState(false);
    const [editing, setEditing] = React.useState<Doctor | null>(null);
    const [showForm, setShowForm] = React.useState(false);
    const [name, setName] = React.useState("");
    const [poli, setPoli] = React.useState("");
    const [error, setError] = React.useState("");
    const [loading, setLoading] = React.useState(true);

    const load = async () => {
        setLoading(true);
        try {
            const res = await authFetch("http://localhost:8080/api/v1/doctors");
            if (res.ok) setDoctors(await res.json());
        } finally {
            setLoading(false);
        }
    };

    React.useEffect(() => {
        if (!requireAuth()) return;
        setAllowed(true);
        setIsAdmin(getRole() === "admin");
        load();
        authFetch("http://localhost:8080/api/v1/polyclinics")
            .then((r) => (r.ok ? r.json() : []))
            .then(setPolys)
            .catch(() => setPolys([]));
    }, []);

    const openCreate = () => {
        setEditing(null);
        setName("");
        setPoli("");
        setError("");
        setShowForm(true);
    };

    const openEdit = (d: Doctor) => {
        setEditing(d);
        setName(d.name);
        setPoli(d.poli);
        setError("");
        setShowForm(true);
    };

    const save = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        if (!name || !poli) {
            setError("Nama dan poliklinik wajib diisi.");
            return;
        }
        const res = editing
            ? await authFetch(`http://localhost:8080/api/v1/doctors/${editing.id}`, {
                  method: "PUT",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({ name, poli }),
              })
            : await authFetch("http://localhost:8080/api/v1/doctors", {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({ name, poli }),
              });
        if (!res.ok) {
            setError(`Gagal menyimpan (${res.status}).`);
            return;
        }
        setShowForm(false);
        load();
    };

    const remove = async (d: Doctor) => {
        if (!confirm(`Hapus dokter "${d.name}"?`)) return;
        const res = await authFetch(`http://localhost:8080/api/v1/doctors/${d.id}`, {
            method: "DELETE",
        });
        if (res.ok) load();
    };

    if (!allowed) return null;

    const poliName = (id: string) => polys.find((p) => p.id === id)?.name || id;

    return (
        <main className="mx-auto max-w-5xl px-4 py-8 md:px-6 animate-fade-up">
            <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
                <div>
                    <h2 className="font-display text-2xl font-bold">Manajemen Dokter</h2>
                    <p className="mt-1 text-sm text-secondary">
                        {doctors.length} dokter terdaftar
                    </p>
                </div>
                {isAdmin && (
                    <button
                        onClick={openCreate}
                        className="rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-hover"
                    >
                        + Tambah Dokter
                    </button>
                )}
            </div>

            {!isAdmin && (
                <p className="mb-4 rounded-xl border border-border bg-slate-50 px-4 py-3 text-sm text-secondary">
                    Hanya administrator yang dapat mengubah data dokter.
                </p>
            )}

            <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-card">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-border bg-slate-50 text-left text-xs uppercase tracking-wide text-secondary">
                                <th scope="col" className="px-4 py-3 font-semibold">Nama</th>
                                <th scope="col" className="px-4 py-3 font-semibold">Poliklinik</th>
                                {isAdmin && <th scope="col" className="px-4 py-3 font-semibold">Aksi</th>}
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr>
                                    <td colSpan={3} className="px-4 py-10 text-center text-secondary">
                                        Memuat data…
                                    </td>
                                </tr>
                            ) : doctors.length === 0 ? (
                                <tr>
                                    <td colSpan={3} className="px-4 py-10 text-center text-secondary">
                                        Belum ada dokter.
                                    </td>
                                </tr>
                            ) : (
                                doctors.map((d) => (
                                    <tr
                                        key={d.id}
                                        className="border-b border-border last:border-0 hover:bg-slate-50"
                                    >
                                        <td className="px-4 py-3 font-medium">{d.name}</td>
                                        <td className="px-4 py-3 text-secondary">{poliName(d.poli)}</td>
                                        {isAdmin && (
                                            <td className="px-4 py-3">
                                                <div className="flex gap-2">
                                                    <button
                                                        onClick={() => openEdit(d)}
                                                        className="rounded-lg bg-primary-soft px-3 py-1.5 text-xs font-medium text-primary transition-colors hover:bg-primary hover:text-white"
                                                    >
                                                        Edit
                                                    </button>
                                                    <button
                                                        onClick={() => remove(d)}
                                                        className="rounded-lg bg-error-soft px-3 py-1.5 text-xs font-medium text-error transition-colors hover:bg-error hover:text-white"
                                                    >
                                                        Hapus
                                                    </button>
                                                </div>
                                            </td>
                                        )}
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            {showForm && (
                <Modal
                    title={editing ? "Edit Dokter" : "Tambah Dokter"}
                    onClose={() => setShowForm(false)}
                >
                    <form onSubmit={save} className="space-y-4">
                        <div className="space-y-1.5">
                            <label htmlFor="doc-name" className="text-sm font-medium">
                                Nama Dokter <span className="text-error">*</span>
                            </label>
                            <input
                                id="doc-name"
                                className="input"
                                placeholder="mis. dr. Andi Pratama"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                required
                            />
                        </div>
                        <div className="space-y-1.5">
                            <label htmlFor="doc-poli" className="text-sm font-medium">
                                Poliklinik <span className="text-error">*</span>
                            </label>
                            <select
                                id="doc-poli"
                                className="input"
                                value={poli}
                                onChange={(e) => setPoli(e.target.value)}
                                required
                            >
                                <option value="">— Pilih poliklinik —</option>
                                {polys.map((p) => (
                                    <option key={p.id} value={p.id}>
                                        {p.name}
                                    </option>
                                ))}
                            </select>
                        </div>
                        {error && (
                            <p role="alert" aria-live="polite" className="text-sm text-error">
                                {error}
                            </p>
                        )}
                        <div className="flex justify-end gap-3 pt-2">
                            <button
                                type="button"
                                onClick={() => setShowForm(false)}
                                className="rounded-xl border border-border px-5 py-2.5 text-sm font-medium text-secondary transition-colors hover:bg-slate-100"
                            >
                                Batal
                            </button>
                            <button
                                type="submit"
                                className="rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-hover"
                            >
                                Simpan
                            </button>
                        </div>
                    </form>
                </Modal>
            )}
        </main>
    );
}

function Modal({
    title,
    onClose,
    children,
}: {
    title: string;
    onClose: () => void;
    children: React.ReactNode;
}) {
    React.useEffect(() => {
        const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
        window.addEventListener("keydown", onKey);
        return () => window.removeEventListener("keydown", onKey);
    }, [onClose]);

    return (
        <div
            className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-sm"
            onClick={onClose}
            role="dialog"
            aria-modal="true"
            aria-label={title}
        >
            <div
                className="w-full max-w-md animate-fade-up rounded-2xl border border-border bg-surface p-6 shadow-card-hover"
                onClick={(e) => e.stopPropagation()}
            >
                <div className="mb-4 flex items-center justify-between">
                    <h3 className="font-display text-lg font-semibold">{title}</h3>
                    <button
                        onClick={onClose}
                        className="rounded-lg p-1.5 text-secondary transition-colors hover:bg-slate-100"
                        aria-label="Tutup"
                    >
                        <IconClose className="h-5 w-5" />
                    </button>
                </div>
                {children}
            </div>
        </div>
    );
}
