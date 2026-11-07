"use client";

import React from "react";
import { requireAuth, authFetch } from "@/lib/auth";
import { IconClose } from "@/components/icons";

type Medicine = {
    id: string;
    name: string;
    description: string;
    stock: number;
    unit: string;
    price: number;
    expired_date: string;
};
type Prescription = {
    id: string;
    encounter_id: string;
    medicine_id: string;
    title: string;
    description: string;
    quantity: number;
    status: string;
    created_at: string;
};

const fmtDate = (iso: string) => {
    if (!iso) return "—";
    const d = new Date(iso);
    if (isNaN(d.getTime())) return "—";
    return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
};

const RX_LABEL: Record<string, string> = {
    pending: "Menunggu",
    processing: "Sedang Diolah",
    dispensed: "Selesai",
};
const RX_STYLE: Record<string, string> = {
    pending: "bg-warning-soft text-warning",
    processing: "bg-primary-soft text-primary",
    dispensed: "bg-success-soft text-success",
};

export default function FarmasiPage() {
    const [allowed, setAllowed] = React.useState(false);
    const [medicines, setMedicines] = React.useState<Medicine[]>([]);
    const [prescriptions, setPrescriptions] = React.useState<Prescription[]>([]);
    const [msg, setMsg] = React.useState<{ kind: "ok" | "err"; text: string } | null>(null);

    // medicine form (create + edit)
    const [editing, setEditing] = React.useState<Medicine | null>(null);
    const [showForm, setShowForm] = React.useState(false);
    const [name, setName] = React.useState("");
    const [description, setDescription] = React.useState("");
    const [stock, setStock] = React.useState("0");
    const [unit, setUnit] = React.useState("tablet");
    const [price, setPrice] = React.useState("0");
    const [expired, setExpired] = React.useState("");

    // take-medicine modal
    const [takeOf, setTakeOf] = React.useState<Medicine | null>(null);
    const [takeQty, setTakeQty] = React.useState("1");
    const [takeNote, setTakeNote] = React.useState("");

    const load = async () => {
        const [m, p] = await Promise.all([
            authFetch("http://localhost:8080/api/v1/pharmacy/medicines"),
            authFetch("http://localhost:8080/api/v1/pharmacy/prescriptions"),
        ]);
        if (m.ok) setMedicines(await m.json());
        if (p.ok) setPrescriptions(await p.json());
    };

    React.useEffect(() => {
        if (requireAuth()) {
            setAllowed(true);
            load();
        }
    }, []);

    const medName = (id: string) => medicines.find((m) => m.id === id)?.name || id;

    const openCreate = () => {
        setEditing(null);
        setName("");
        setDescription("");
        setStock("0");
        setUnit("tablet");
        setPrice("0");
        setExpired("");
        setShowForm(true);
    };

    const openEdit = (m: Medicine) => {
        setEditing(m);
        setName(m.name);
        setDescription(m.description);
        setStock(String(m.stock));
        setUnit(m.unit);
        setPrice(String(m.price));
        setExpired(m.expired_date);
        setShowForm(true);
    };

    const saveMedicine = async (e: React.FormEvent) => {
        e.preventDefault();
        setMsg(null);
        const payload = {
            name,
            description,
            stock: Number(stock),
            unit,
            price: Number(price),
            expired_date: expired,
        };
        const res = editing
            ? await authFetch(`http://localhost:8080/api/v1/pharmacy/medicines/${editing.id}`, {
                  method: "PUT",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify(payload),
              })
            : await authFetch("http://localhost:8080/api/v1/pharmacy/medicines", {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify(payload),
              });
        if (!res.ok) {
            setMsg({ kind: "err", text: `Gagal menyimpan obat (${res.status}).` });
            return;
        }
        setShowForm(false);
        setMsg({ kind: "ok", text: editing ? "Obat diperbarui." : "Obat ditambahkan." });
        load();
    };

    const removeMedicine = async (m: Medicine) => {
        if (!confirm(`Hapus obat "${m.name}"?`)) return;
        const res = await authFetch(`http://localhost:8080/api/v1/pharmacy/medicines/${m.id}`, {
            method: "DELETE",
        });
        if (res.ok) {
            setMsg({ kind: "ok", text: "Obat dihapus." });
            load();
        } else {
            setMsg({ kind: "err", text: `Gagal menghapus (${res.status}).` });
        }
    };

    const takeMedicine = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!takeOf) return;
        const res = await authFetch("http://localhost:8080/api/v1/pharmacy/transactions", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                medicine_id: takeOf.id,
                quantity: Number(takeQty),
                note: takeNote,
            }),
        });
        if (!res.ok) {
            setMsg({ kind: "err", text: `Gagal mengambil obat (${res.status}).` });
            return;
        }
        setTakeOf(null);
        setTakeQty("1");
        setTakeNote("");
        setMsg({ kind: "ok", text: "Pengambilan obat dicatat." });
        load();
    };

    const setRxStatus = async (id: string, status: string) => {
        const res = await authFetch(`http://localhost:8080/api/v1/pharmacy/prescriptions/${id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ status }),
        });
        if (res.ok) {
            setMsg({ kind: "ok", text: status === "dispensed" ? "Resep selesai." : "Resep mulai diolah." });
            load();
        } else {
            setMsg({ kind: "err", text: `Gagal memperbarui resep (${res.status}).` });
        }
    };

    if (!allowed) return null;

    const active = prescriptions.filter((p) => p.status !== "dispensed");

    return (
        <main className="mx-auto max-w-6xl px-4 py-8 md:px-6 animate-fade-up">
            <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
                <div>
                    <h2 className="font-display text-2xl font-bold">Farmasi</h2>
                    <p className="mt-1 text-sm text-secondary">
                        Kelola stok obat, olah resep, dan catat pengambilan obat.
                    </p>
                </div>
                <button
                    onClick={openCreate}
                    className="rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-hover"
                >
                    + Tambah Obat
                </button>
            </div>

            {msg && (
                <p
                    role="status"
                    aria-live="polite"
                    className={`mb-4 rounded-xl px-4 py-3 text-sm ${
                        msg.kind === "ok" ? "bg-success-soft text-success" : "bg-error-soft text-error"
                    }`}
                >
                    {msg.text}
                </p>
            )}

            <section className="mb-8">
                <h3 className="mb-3 font-display text-lg font-semibold">
                    Resep Aktif <span className="text-secondary">({active.length})</span>
                </h3>
                <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-card">
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="border-b border-border bg-slate-50 text-left text-xs uppercase tracking-wide text-secondary">
                                    <th scope="col" className="px-4 py-3 font-semibold">Nama Resep</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Obat</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Deskripsi</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Jumlah</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Tanggal</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Status</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Aksi</th>
                                </tr>
                            </thead>
                            <tbody>
                                {active.length === 0 ? (
                                    <tr>
                                        <td colSpan={7} className="px-4 py-10 text-center text-secondary">
                                            Tidak ada resep aktif.
                                        </td>
                                    </tr>
                                ) : (
                                    active.map((p) => (
                                        <tr key={p.id} className="border-b border-border last:border-0 hover:bg-slate-50">
                                            <td className="px-4 py-3 font-medium">{p.title || p.id}</td>
                                            <td className="px-4 py-3">{medName(p.medicine_id)}</td>
                                            <td className="max-w-[14rem] truncate px-4 py-3 text-secondary">
                                                {p.description || "—"}
                                            </td>
                                            <td className="px-4 py-3 tnum">{p.quantity}</td>
                                            <td className="px-4 py-3 tnum text-secondary">{fmtDate(p.created_at)}</td>
                                            <td className="px-4 py-3">
                                                <span
                                                    className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                                                        RX_STYLE[p.status] || "bg-slate-100 text-secondary"
                                                    }`}
                                                >
                                                    {RX_LABEL[p.status] || p.status}
                                                </span>
                                            </td>
                                            <td className="px-4 py-3">
                                                <div className="flex gap-2">
                                                    {p.status === "pending" && (
                                                        <button
                                                            onClick={() => setRxStatus(p.id, "processing")}
                                                            className="rounded-lg bg-primary-soft px-3 py-1.5 text-xs font-medium text-primary transition-colors hover:bg-primary hover:text-white"
                                                        >
                                                            Olah
                                                        </button>
                                                    )}
                                                    {p.status === "processing" && (
                                                        <button
                                                            onClick={() => setRxStatus(p.id, "dispensed")}
                                                            className="rounded-lg bg-success-soft px-3 py-1.5 text-xs font-medium text-success transition-colors hover:bg-success hover:text-white"
                                                        >
                                                            Selesai
                                                        </button>
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
            </section>

            <section>
                <h3 className="mb-3 font-display text-lg font-semibold">
                    Stok Obat <span className="text-secondary">({medicines.length})</span>
                </h3>
                <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-card">
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="border-b border-border bg-slate-50 text-left text-xs uppercase tracking-wide text-secondary">
                                    <th scope="col" className="px-4 py-3 font-semibold">Nama</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Deskripsi</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Stok</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Satuan</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Harga</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Kadaluarsa</th>
                                    <th scope="col" className="px-4 py-3 font-semibold">Aksi</th>
                                </tr>
                            </thead>
                            <tbody>
                                {medicines.length === 0 ? (
                                    <tr>
                                        <td colSpan={7} className="px-4 py-10 text-center text-secondary">
                                            Belum ada obat.
                                        </td>
                                    </tr>
                                ) : (
                                    medicines.map((m) => (
                                        <tr key={m.id} className="border-b border-border last:border-0 hover:bg-slate-50">
                                            <td className="px-4 py-3 font-medium">{m.name}</td>
                                            <td className="max-w-[14rem] truncate px-4 py-3 text-secondary">
                                                {m.description || "—"}
                                            </td>
                                            <td className="px-4 py-3 tnum">
                                                <span
                                                    className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                                                        m.stock <= 5 ? "bg-error-soft text-error" : "bg-success-soft text-success"
                                                    }`}
                                                >
                                                    {m.stock}
                                                </span>
                                            </td>
                                            <td className="px-4 py-3 text-secondary">{m.unit}</td>
                                            <td className="px-4 py-3 tnum">Rp {m.price.toLocaleString("id-ID")}</td>
                                            <td className="px-4 py-3 tnum text-secondary">{fmtDate(m.expired_date)}</td>
                                            <td className="px-4 py-3">
                                                <div className="flex flex-wrap gap-2">
                                                    <button
                                                        onClick={() => {
                                                            setTakeOf(m);
                                                            setTakeQty("1");
                                                            setTakeNote("");
                                                        }}
                                                        className="rounded-lg bg-primary-soft px-3 py-1.5 text-xs font-medium text-primary transition-colors hover:bg-primary hover:text-white"
                                                    >
                                                        Ambil
                                                    </button>
                                                    <button
                                                        onClick={() => openEdit(m)}
                                                        className="rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-medium text-secondary transition-colors hover:bg-slate-200"
                                                    >
                                                        Edit
                                                    </button>
                                                    <button
                                                        onClick={() => removeMedicine(m)}
                                                        className="rounded-lg bg-error-soft px-3 py-1.5 text-xs font-medium text-error transition-colors hover:bg-error hover:text-white"
                                                    >
                                                        Hapus
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </section>

            {showForm && (
                <Modal title={editing ? "Edit Obat" : "Tambah Obat"} onClose={() => setShowForm(false)}>
                    <form onSubmit={saveMedicine} className="space-y-4">
                        <Field label="Nama Obat" required>
                            <input className="input" value={name} onChange={(e) => setName(e.target.value)} required />
                        </Field>
                        <Field label="Deskripsi">
                            <input
                                className="input"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="mis. pereda nyeri, demam"
                            />
                        </Field>
                        <div className="grid grid-cols-2 gap-4">
                            <Field label="Stok">
                                <input type="number" min="0" className="input" value={stock} onChange={(e) => setStock(e.target.value)} />
                            </Field>
                            <Field label="Satuan">
                                <input className="input" value={unit} onChange={(e) => setUnit(e.target.value)} />
                            </Field>
                            <Field label="Harga">
                                <input type="number" min="0" className="input" value={price} onChange={(e) => setPrice(e.target.value)} />
                            </Field>
                            <Field label="Kadaluarsa">
                                <input type="date" className="input" value={expired} onChange={(e) => setExpired(e.target.value)} />
                            </Field>
                        </div>
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

            {takeOf && (
                <Modal title={`Ambil Obat — ${takeOf.name}`} onClose={() => setTakeOf(null)}>
                    <form onSubmit={takeMedicine} className="space-y-4">
                        <p className="text-sm text-secondary">
                            Stok tersedia:{" "}
                            <span className="font-semibold text-foreground">
                                {takeOf.stock} {takeOf.unit}
                            </span>
                        </p>
                        <Field label="Jumlah Diambil" required>
                            <input
                                type="number"
                                min="1"
                                max={takeOf.stock}
                                className="input"
                                value={takeQty}
                                onChange={(e) => setTakeQty(e.target.value)}
                                required
                            />
                        </Field>
                        <Field label="Catatan">
                            <input
                                className="input"
                                value={takeNote}
                                onChange={(e) => setTakeNote(e.target.value)}
                                placeholder="mis. untuk pasien Budi / resep RX-1"
                            />
                        </Field>
                        <div className="flex justify-end gap-3 pt-2">
                            <button
                                type="button"
                                onClick={() => setTakeOf(null)}
                                className="rounded-xl border border-border px-5 py-2.5 text-sm font-medium text-secondary transition-colors hover:bg-slate-100"
                            >
                                Batal
                            </button>
                            <button
                                type="submit"
                                className="rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-hover"
                            >
                                Catat Pengambilan
                            </button>
                        </div>
                    </form>
                </Modal>
            )}
        </main>
    );
}

function Field({ label, required, children }: { label: string; required?: boolean; children: React.ReactNode }) {
    return (
        <label className="flex flex-col gap-1.5 text-sm">
            <span className="font-medium">
                {label} {required && <span className="text-error">*</span>}
            </span>
            {children}
        </label>
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
