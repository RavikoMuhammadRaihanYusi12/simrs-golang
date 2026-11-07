"use client";

import React from "react";
import { requireAuth, getRole } from "@/lib/auth";
import { IconUser } from "@/components/icons";

const ROLE_LABEL: Record<string, string> = {
    admin: "Administrator",
    doctor: "Dokter",
    nurse: "Loket / Perawat",
    cashier: "Kasir",
    pharmacy: "Apoteker",
};

export default function ProfilPage() {
    const [role, setRole] = React.useState<string | null>(null);
    const [allowed, setAllowed] = React.useState(false);

    React.useEffect(() => {
        if (requireAuth()) {
            setAllowed(true);
            setRole(getRole());
        }
    }, []);

    if (!allowed) return null;

    return (
        <main className="mx-auto max-w-3xl px-4 py-8 md:px-6 animate-fade-up">
            <h2 className="mb-6 font-display text-2xl font-bold">Profil Pengguna</h2>
            <div className="rounded-2xl border border-border bg-surface p-6 shadow-card">
                <div className="flex items-center gap-4">
                    <span className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-soft text-primary">
                        <IconUser className="h-8 w-8" />
                    </span>
                    <div>
                        <p className="font-display text-lg font-semibold">
                            {ROLE_LABEL[role || ""] || role || "Pengguna"}
                        </p>
                        <span className="mt-1 inline-block rounded-full bg-accent-soft px-2.5 py-0.5 text-xs font-medium text-accent">
                            {role || "-"}
                        </span>
                    </div>
                </div>
                <dl className="mt-6 grid grid-cols-2 gap-4 border-t border-border pt-4 text-sm">
                    <div>
                        <dt className="text-secondary">Nama</dt>
                        <dd className="font-medium">—</dd>
                    </div>
                    <div>
                        <dt className="text-secondary">Unit</dt>
                        <dd className="font-medium">—</dd>
                    </div>
                </dl>
            </div>
        </main>
    );
}
