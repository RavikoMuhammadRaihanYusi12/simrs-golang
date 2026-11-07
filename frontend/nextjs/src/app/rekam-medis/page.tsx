"use client";

import React from "react";
import { requireAuth } from "@/lib/auth";

export default function RekamMedisPage() {
    const [allowed, setAllowed] = React.useState(false);

    React.useEffect(() => {
        if (requireAuth()) setAllowed(true);
    }, []);

    if (!allowed) return null;

    return (
        <main className="mx-auto max-w-5xl px-4 py-8 md:px-6 animate-fade-up">
            <header className="mb-6">
                <h2 className="font-display text-2xl font-bold">Rekam Medis Elektronik</h2>
                <p className="mt-1 text-sm text-secondary">
                    Pencatatan SOAP (Subjektif, Objektif, Asesmen, Plan) per kunjungan.
                </p>
            </header>

            <div className="rounded-2xl border border-dashed border-border bg-surface p-10 text-center shadow-card">
                <p className="font-medium">Rekam medis terhubung ke kunjungan pasien.</p>
                <p className="mt-1 text-sm text-secondary">
                    Pilih kunjungan dari halaman Antrean untuk mulai mengisi SOAP.
                </p>
            </div>
        </main>
    );
}
