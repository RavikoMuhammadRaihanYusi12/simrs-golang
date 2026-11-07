import React from "react";
import { Inter, Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";
import AppShell from "@/components/AppShell";

const inter = Inter({
    subsets: ["latin"],
    variable: "--font-inter",
    display: "swap",
});

const jakarta = Plus_Jakarta_Sans({
    subsets: ["latin"],
    variable: "--font-jakarta",
    display: "swap",
});

export const metadata = {
    title: "SIMRS — Sistem Informasi Manajemen Rumah Sakit",
    description: "Sistem Informasi Manajemen Rumah Sakit berbasis web",
};

export default function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <html lang="id" className={`${inter.variable} ${jakarta.variable}`} suppressHydrationWarning>
            <body className="min-h-screen bg-background text-foreground font-sans antialiased">
                <AppShell>{children}</AppShell>
            </body>
        </html>
    );
}
