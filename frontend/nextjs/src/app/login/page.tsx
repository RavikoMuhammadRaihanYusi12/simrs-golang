"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { saveAuth, homeForRole } from "@/lib/auth";
import { IconLogo } from "@/components/icons";

const DEMO = [
    { role: "Administrator", username: "admin", password: "admin123" },
    { role: "Loket / Perawat", username: "loket", password: "loket123" },
    { role: "Dokter", username: "dokter", password: "dokter123" },
    { role: "Kasir", username: "kasir", password: "kasir123" },
    { role: "Apoteker", username: "apoteker", password: "apoteker123" },
];

export default function LoginPage() {
    const router = useRouter();
    const [username, setUsername] = React.useState("");
    const [password, setPassword] = React.useState("");
    const [error, setError] = React.useState("");
    const [loading, setLoading] = React.useState(false);

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        setLoading(true);
        try {
            const res = await fetch("http://localhost:8080/api/v1/auth/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ username, password }),
            });
            if (!res.ok) {
                setError("Username atau password salah.");
                return;
            }
            const { token } = await res.json();
            const role = saveAuth(token);
            router.push(homeForRole(role));
        } catch {
            setError("Tidak dapat menghubungi server.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="grid min-h-screen lg:grid-cols-2">
            {/* Brand panel */}
            <div className="relative hidden flex-col justify-between overflow-hidden bg-slate-900 p-12 text-white lg:flex">
                <div
                    className="pointer-events-none absolute inset-0 opacity-30"
                    style={{
                        backgroundImage:
                            "radial-gradient(60rem 40rem at 80% -10%, #2563eb55, transparent), radial-gradient(40rem 30rem at 0% 100%, #0d948855, transparent)",
                    }}
                    aria-hidden="true"
                />
                <div className="relative flex items-center gap-3">
                    <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary">
                        <IconLogo className="h-6 w-6" />
                    </span>
                    <span className="font-display text-xl font-bold tracking-tight">SIMRS</span>
                </div>
                <div className="relative space-y-4">
                    <h1 className="font-display text-4xl font-bold leading-tight text-balance">
                        Sistem Informasi Manajemen Rumah Sakit
                    </h1>
                    <p className="max-w-md text-slate-300">
                        Kelola pendaftaran pasien, antrean poli, rekam medis elektronik, farmasi,
                        dan billing dalam satu platform terpadu.
                    </p>
                </div>
                <p className="relative text-sm text-slate-400">
                    &copy; {new Date().getFullYear()} SIMRS. Terintegrasi BPJS &amp; SATUSEHAT.
                </p>
            </div>

            {/* Form panel */}
            <div className="flex items-center justify-center px-6 py-12">
                <div className="w-full max-w-sm animate-fade-up">
                    <div className="mb-8 flex items-center gap-3 lg:hidden">
                        <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-white">
                            <IconLogo className="h-6 w-6" />
                        </span>
                        <span className="font-display text-xl font-bold">SIMRS</span>
                    </div>

                    <h2 className="font-display text-2xl font-bold">Masuk</h2>
                    <p className="mt-1 text-sm text-secondary">
                        Silakan masuk dengan akun Anda.
                    </p>

                    <form onSubmit={handleLogin} className="mt-8 space-y-4">
                        <div className="space-y-1.5">
                            <label htmlFor="username" className="text-sm font-medium">
                                Username
                            </label>
                            <input
                                id="username"
                                name="username"
                                autoComplete="username"
                                spellCheck={false}
                                className="w-full rounded-xl border border-border bg-surface px-3.5 py-2.5 text-sm outline-none transition-colors placeholder:text-slate-400 focus:border-primary"
                                placeholder="mis. dokter"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                required
                            />
                        </div>
                        <div className="space-y-1.5">
                            <label htmlFor="password" className="text-sm font-medium">
                                Password
                            </label>
                            <input
                                id="password"
                                name="password"
                                type="password"
                                autoComplete="current-password"
                                className="w-full rounded-xl border border-border bg-surface px-3.5 py-2.5 text-sm outline-none transition-colors placeholder:text-slate-400 focus:border-primary"
                                placeholder="••••••••"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                required
                            />
                        </div>

                        {error && (
                            <p role="alert" aria-live="polite" className="text-sm text-error">
                                {error}
                            </p>
                        )}

                        <button
                            type="submit"
                            disabled={loading}
                            className="w-full rounded-xl bg-primary py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-hover disabled:opacity-60"
                        >
                            {loading ? "Memproses…" : "Masuk"}
                        </button>
                    </form>

                    <div className="mt-8 rounded-xl border border-border bg-surface p-4">
                        <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-secondary">
                            Akun Demo
                        </p>
                        <ul className="space-y-1 text-sm text-secondary">
                            {DEMO.map((d) => (
                                <li key={d.username} className="flex items-center justify-between gap-2">
                                    <span>{d.role}</span>
                                    <span className="font-mono text-xs">
                                        {d.username} / {d.password}
                                    </span>
                                </li>
                            ))}
                        </ul>
                    </div>
                </div>
            </div>
        </div>
    );
}
