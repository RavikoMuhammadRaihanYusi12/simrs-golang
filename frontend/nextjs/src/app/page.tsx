"use client";

import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { requireAuth, authFetch } from "@/lib/auth";

type Polyclinic = { id: string; name: string };
type Doctor = { id: string; name: string; poli: string };

export default function PendaftaranPage() {
    const router = useRouter();
    const [name, setName] = React.useState("");
    const [nik, setNik] = React.useState("");
    const [gender, setGender] = React.useState("L");
    const [birthDate, setBirthDate] = React.useState("");
    const [phone, setPhone] = React.useState("");
    const [complaint, setComplaint] = React.useState("");
    const [description, setDescription] = React.useState("");
    const [poli, setPoli] = React.useState("");
    const [doctorId, setDoctorId] = React.useState("");

    const [polys, setPolys] = React.useState<Polyclinic[]>([]);
    const [doctors, setDoctors] = React.useState<Doctor[]>([]);
    const [msg, setMsg] = React.useState<{ kind: "ok" | "err"; text: string } | null>(null);
    const [loading, setLoading] = React.useState(false);

    React.useEffect(() => {
        if (!requireAuth()) return;
        authFetch("http://localhost:8080/api/v1/polyclinics")
            .then((r) => (r.ok ? r.json() : []))
            .then(setPolys)
            .catch(() => setPolys([]));
    }, []);

    React.useEffect(() => {
        if (!poli) return setDoctors([]);
        authFetch(`http://localhost:8080/api/v1/doctors?poli=${encodeURIComponent(poli)}`)
            .then((r) => (r.ok ? r.json() : []))
            .then(setDoctors)
            .catch(() => setDoctors([]));
    }, [poli]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setMsg(null);
        if (!nik || !name) {
            setMsg({ kind: "err", text: "Nama dan NIK wajib diisi." });
            return;
        }
        setLoading(true);
        try {
            const patient = {
                id: "patient-" + Date.now(),
                nik,
                no_rm: "RM" + Date.now(),
                full_name: name,
                birth_date: birthDate || "1990-01-01",
                gender,
                phone,
                address: "",
                complaint,
                description,
            };
            const res = await authFetch("http://localhost:8080/api/v1/patients", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(patient),
            });
            if (!res.ok) {
                setMsg({ kind: "err", text: `Gagal menyimpan pendaftaran (${res.status}).` });
                return;
            }
            // Buat kunjungan bila poli dipilih.
            if (poli) {
                const enc = await authFetch("http://localhost:8080/api/v1/encounters", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        patient_id: patient.id,
                        polyclinic: poli,
                        doctor_id: doctorId,
                    }),
                });
                if (!enc.ok) {
                    setMsg({
                        kind: "err",
                        text: `Pasien tersimpan, tapi pembuatan kunjungan gagal (${enc.status}).`,
                    });
                    return;
                }
            }
            setMsg({ kind: "ok", text: "Pendaftaran berhasil disimpan." });
            setName("");
            setNik("");
            setGender("L");
            setBirthDate("");
            setPhone("");
            setComplaint("");
            setDescription("");
            setPoli("");
            setDoctorId("");
            router.push("/pasien");
        } finally {
            setLoading(false);
        }
    };

    return (
        <main className="mx-auto max-w-5xl px-4 py-8 md:px-6 animate-fade-up">
            <header className="mb-6">
                <h2 className="font-display text-2xl font-bold">Pendaftaran Pasien</h2>
                <p className="mt-1 text-sm text-secondary">
                    Registrasi pasien baru beserta keluhan dan tujuan poliklinik.
                </p>
            </header>

            <form
                onSubmit={handleSubmit}
                className="grid gap-6 rounded-2xl border border-border bg-surface p-6 shadow-card md:grid-cols-2"
            >
                {/* Kolom identitas */}
                <fieldset className="space-y-4">
                    <legend className="mb-2 text-sm font-semibold uppercase tracking-wide text-secondary">
                        Identitas Pasien
                    </legend>
                    <Field label="Nama Lengkap" htmlFor="name" required>
                        <input
                            id="name"
                            className="input"
                            placeholder="mis. Budi Setiawan"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            required
                        />
                    </Field>
                    <Field label="NIK" htmlFor="nik" required>
                        <input
                            id="nik"
                            inputMode="numeric"
                            pattern="[0-9]*"
                            className="input tnum"
                            placeholder="16 digit NIK"
                            value={nik}
                            onChange={(e) => setNik(e.target.value)}
                            required
                        />
                    </Field>
                    <div className="grid grid-cols-2 gap-4">
                        <Field label="Tanggal Lahir" htmlFor="birth">
                            <input
                                id="birth"
                                type="date"
                                className="input"
                                value={birthDate}
                                onChange={(e) => setBirthDate(e.target.value)}
                            />
                        </Field>
                        <Field label="Jenis Kelamin" htmlFor="gender">
                            <select
                                id="gender"
                                className="input"
                                value={gender}
                                onChange={(e) => setGender(e.target.value)}
                            >
                                <option value="L">Laki-laki</option>
                                <option value="P">Perempuan</option>
                            </select>
                        </Field>
                    </div>
                    <Field label="No. Telepon" htmlFor="phone">
                        <input
                            id="phone"
                            type="tel"
                            inputMode="tel"
                            className="input"
                            placeholder="08xx xxxx xxxx"
                            value={phone}
                            onChange={(e) => setPhone(e.target.value)}
                        />
                    </Field>
                </fieldset>

                {/* Kolom keluhan + tujuan */}
                <fieldset className="space-y-4">
                    <legend className="mb-2 text-sm font-semibold uppercase tracking-wide text-secondary">
                        Keluhan &amp; Tujuan
                    </legend>
                    <Field label="Keluhan Utama" htmlFor="complaint">
                        <input
                            id="complaint"
                            className="input"
                            placeholder="mis. Demam 3 hari"
                            value={complaint}
                            onChange={(e) => setComplaint(e.target.value)}
                        />
                    </Field>
                    <Field label="Deskripsi Keluhan" htmlFor="description">
                        <textarea
                            id="description"
                            rows={3}
                            className="input resize-none"
                            placeholder="Rincian keluhan, riwayat, gejala penyerta…"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                        />
                    </Field>
                    <Field label="Poliklinik" htmlFor="poli">
                        <select
                            id="poli"
                            className="input"
                            value={poli}
                            onChange={(e) => {
                                setPoli(e.target.value);
                                setDoctorId("");
                            }}
                        >
                            <option value="">— Pilih poliklinik —</option>
                            {polys.map((p) => (
                                <option key={p.id} value={p.id}>
                                    {p.name}
                                </option>
                            ))}
                        </select>
                    </Field>
                    <Field label="Dokter" htmlFor="doctor">
                        <select
                            id="doctor"
                            className="input"
                            value={doctorId}
                            onChange={(e) => setDoctorId(e.target.value)}
                            disabled={!poli}
                        >
                            <option value="">
                                {poli ? "— Pilih dokter —" : "Pilih poliklinik dulu"}
                            </option>
                            {doctors.map((d) => (
                                <option key={d.id} value={d.id}>
                                    {d.name}
                                </option>
                            ))}
                        </select>
                    </Field>
                </fieldset>

                <div className="md:col-span-2">
                    {msg && (
                        <p
                            role="alert"
                            aria-live="polite"
                            className={`mb-3 text-sm ${
                                msg.kind === "ok" ? "text-success" : "text-error"
                            }`}
                        >
                            {msg.text}
                        </p>
                    )}
                    <div className="flex flex-wrap gap-3">
                        <button
                            type="submit"
                            disabled={loading}
                            className="rounded-xl bg-primary px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-hover disabled:opacity-60"
                        >
                            {loading ? "Menyimpan…" : "Simpan Pendaftaran"}
                        </button>
                        <Link
                            href="/pasien"
                            className="rounded-xl border border-border px-6 py-2.5 text-sm font-medium text-secondary transition-colors hover:bg-slate-100"
                        >
                            Lihat Daftar Pasien
                        </Link>
                    </div>
                </div>
            </form>
        </main>
    );
}

function Field({
    label,
    htmlFor,
    required,
    children,
}: {
    label: string;
    htmlFor: string;
    required?: boolean;
    children: React.ReactNode;
}) {
    return (
        <div className="space-y-1.5">
            <label htmlFor={htmlFor} className="text-sm font-medium">
                {label}
                {required && <span className="text-error"> *</span>}
            </label>
            {children}
        </div>
    );
}
