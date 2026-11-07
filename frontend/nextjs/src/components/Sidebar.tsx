"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
    IconDashboard,
    IconPatients,
    IconQueue,
    IconRecords,
    IconDoctor,
    IconPharmacy,
    IconHistory,
    IconClose,
} from "@/components/icons";
import { getRole } from "@/lib/auth";

// roles: undefined = everyone. Otherwise the nav item shows only for those roles.
const NAV: { href: string; label: string; Icon: (p: { className?: string }) => React.ReactElement; roles?: string[] }[] = [
    { href: "/dashboard", label: "Dashboard", Icon: IconDashboard },
    { href: "/", label: "Pendaftaran", Icon: IconPatients, roles: ["nurse", "admin"] },
    { href: "/pasien", label: "Pasien", Icon: IconPatients },
    { href: "/antrean", label: "Antrean", Icon: IconQueue },
    { href: "/dokter", label: "Dokter", Icon: IconDoctor },
    { href: "/rekam-medis", label: "Rekam Medis", Icon: IconRecords, roles: ["doctor", "admin"] },
    { href: "/farmasi", label: "Farmasi", Icon: IconPharmacy, roles: ["pharmacy", "admin"] },
    { href: "/riwayat-obat", label: "Riwayat Obat", Icon: IconHistory, roles: ["pharmacy", "admin"] },
    { href: "/riwayat", label: "Riwayat", Icon: IconHistory },
];

export default function Sidebar({
    open,
    onClose,
}: {
    open: boolean;
    onClose: () => void;
}) {
    const pathname = usePathname();
    const [role, setRole] = React.useState<string | null>(null);

    React.useEffect(() => {
        setRole(getRole());
    }, []);

    const items = NAV.filter((item) => !item.roles || (role !== null && item.roles.includes(role)));

    return (
        <>
            {open && (
                <div
                    className="fixed inset-0 z-40 bg-slate-900/40 backdrop-blur-sm md:hidden"
                    onClick={onClose}
                    aria-hidden="true"
                />
            )}
            <aside
                className={`fixed inset-y-0 left-0 z-50 w-64 transform border-r border-border bg-surface transition-transform duration-200 ease-out md:static md:translate-x-0 ${
                    open ? "translate-x-0" : "-translate-x-full"
                }`}
            >
                <div className="flex h-16 items-center justify-between border-b border-border px-5">
                    <Link href="/" className="flex items-center gap-2.5">
                        <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-white">
                            <IconLogoMini />
                        </span>
                        <span className="font-display text-lg font-bold tracking-tight">
                            SIMRS
                        </span>
                    </Link>
                    <button
                        onClick={onClose}
                        className="rounded-lg p-2 text-secondary hover:bg-slate-100 md:hidden"
                        aria-label="Tutup menu"
                    >
                        <IconClose className="h-5 w-5" />
                    </button>
                </div>

                <nav className="flex flex-col gap-1 p-3" aria-label="Navigasi utama">
                    {items.map(({ href, label, Icon }) => {
                        const active =
                            pathname === href || pathname.startsWith(href + "/");
                        return (
                            <Link
                                key={href}
                                href={href}
                                onClick={onClose}
                                aria-current={active ? "page" : undefined}
                                className={`flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                                    active
                                        ? "bg-primary-soft text-primary"
                                        : "text-secondary hover:bg-slate-100 hover:text-foreground"
                                }`}
                            >
                                <Icon className="h-5 w-5 shrink-0" />
                                {label}
                            </Link>
                        );
                    })}
                </nav>
            </aside>
        </>
    );
}

function IconLogoMini() {
    return (
        <svg
            className="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth={2.2}
            strokeLinecap="round"
            aria-hidden="true"
        >
            <path d="M12 4v16M4 12h16" />
        </svg>
    );
}
