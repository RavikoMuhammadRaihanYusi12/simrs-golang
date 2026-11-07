"use client";

import React from "react";
import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import Sidebar from "@/components/Sidebar";
import { IconMenu, IconUser, IconLogout } from "@/components/icons";
import { getToken, getRole, logout } from "@/lib/auth";

const ROLE_LABEL: Record<string, string> = {
    admin: "Administrator",
    doctor: "Dokter",
    nurse: "Loket / Perawat",
    cashier: "Kasir",
    pharmacy: "Apoteker",
    patient: "Pasien",
};

export default function AppShell({ children }: { children: React.ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();
    const [sidebarOpen, setSidebarOpen] = React.useState(false);
    const [role, setRole] = React.useState<string | null>(null);
    const [loggedIn, setLoggedIn] = React.useState(false);

    React.useEffect(() => {
        setLoggedIn(!!getToken());
        setRole(getRole());
    }, [pathname]);

    const handleLogout = () => {
        logout();
        router.push("/login");
    };

    // Login page renders standalone (no chrome).
    if (pathname === "/login") {
        return <>{children}</>;
    }

    return (
        <div className="flex min-h-screen">
            {loggedIn && (
                <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />
            )}

            <div className="flex min-w-0 flex-1 flex-col">
                <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-surface/80 px-4 backdrop-blur md:px-6">
                    <div className="flex items-center gap-3">
                        {loggedIn && (
                            <button
                                onClick={() => setSidebarOpen(true)}
                                className="rounded-lg p-2 text-secondary hover:bg-slate-100 md:hidden"
                                aria-label="Buka menu"
                            >
                                <IconMenu className="h-5 w-5" />
                            </button>
                        )}
                        <h1 className="font-display text-base font-semibold">
                            Sistem Informasi Manajemen Rumah Sakit
                        </h1>
                    </div>

                    <div className="flex items-center gap-2">
                        {loggedIn ? (
                            <>
                                <Link
                                    href="/profil"
                                    className="flex items-center gap-2 rounded-xl px-2.5 py-1.5 text-sm transition-colors hover:bg-slate-100"
                                >
                                    <span className="flex h-8 w-8 items-center justify-center rounded-full bg-primary-soft text-primary">
                                        <IconUser className="h-4 w-4" />
                                    </span>
                                    <span className="hidden font-medium sm:inline">
                                        {ROLE_LABEL[role || ""] || role || "Pengguna"}
                                    </span>
                                </Link>
                                <button
                                    onClick={handleLogout}
                                    className="flex items-center gap-2 rounded-xl px-2.5 py-1.5 text-sm font-medium text-secondary transition-colors hover:bg-error-soft hover:text-error"
                                >
                                    <IconLogout className="h-4 w-4" />
                                    <span className="hidden sm:inline">Keluar</span>
                                </button>
                            </>
                        ) : (
                            <Link
                                href="/login"
                                className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
                            >
                                Masuk
                            </Link>
                        )}
                    </div>
                </header>

                <div className="flex-1">{children}</div>
            </div>
        </div>
    );
}
